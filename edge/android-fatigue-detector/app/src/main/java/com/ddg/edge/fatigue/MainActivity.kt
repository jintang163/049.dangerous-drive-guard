package com.ddg.edge.fatigue

import android.Manifest
import android.content.ComponentName
import android.content.Context
import android.content.Intent
import android.content.ServiceConnection
import android.content.pm.PackageManager
import android.graphics.Bitmap
import android.graphics.SurfaceTexture
import android.os.Build
import android.os.Bundle
import android.os.IBinder
import android.view.TextureView
import android.view.WindowManager
import androidx.activity.result.contract.ActivityResultContracts
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.camera.core.Camera
import androidx.camera.core.CameraSelector
import androidx.camera.core.ImageAnalysis
import androidx.camera.core.Preview
import androidx.camera.lifecycle.ProcessCameraProvider
import androidx.core.content.ContextCompat
import androidx.core.view.ViewCompat
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModelProvider
import androidx.work.Constraints
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.NetworkType
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import com.ddg.edge.fatigue.alert.AlarmReportManager
import com.ddg.edge.fatigue.alert.LocalVideoRecorder
import com.ddg.edge.fatigue.alert.SeatVibrationController
import com.ddg.edge.fatigue.alert.VoiceAlertManager
import com.ddg.edge.fatigue.data.local.AlarmDatabase
import com.ddg.edge.fatigue.data.local.OfflineAlarmRepository
import com.ddg.edge.fatigue.data.remote.AlarmUploadWorker
import com.ddg.edge.fatigue.data.remote.AlarmUploadWorkerDelegate
import com.ddg.edge.fatigue.data.remote.DetectionFrameUploadRequest
import com.ddg.edge.fatigue.data.remote.PlatformApiService
import com.ddg.edge.fatigue.detector.FaceLandmarkDetector
import com.ddg.edge.fatigue.detector.FatigueMetricsCalculator
import com.ddg.edge.fatigue.detector.FatigueScoreAggregator
import com.ddg.edge.fatigue.detector.GazeDirectionDetector
import com.ddg.edge.fatigue.detector.PERCLOSCalculator
import com.ddg.edge.fatigue.detector.YawnDetector
import com.ddg.edge.fatigue.databinding.ActivityMainBinding
import com.ddg.edge.fatigue.model.AlarmLevel
import com.ddg.edge.fatigue.model.DetectionFrameResult
import com.ddg.edge.fatigue.model.FatigueAlarm
import com.ddg.edge.fatigue.tracking.GPSTracker
import com.ddg.edge.fatigue.utils.CameraSizeSelector
import com.ddg.edge.fatigue.utils.FrameMetadata
import com.ddg.edge.fatigue.utils.FrameUtils
import com.google.mediapipe.framework.image.MPImage
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancelChildren
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.util.UUID
import java.util.concurrent.Executors
import java.util.concurrent.TimeUnit

class MainActivity : AppCompatActivity() {

    private lateinit var binding: ActivityMainBinding
    private val viewModel: FatigueViewModel by viewModels()

    private var cameraProvider: ProcessCameraProvider? = null
    private var camera: Camera? = null
    private val cameraExecutor = Executors.newFixedThreadPool(4)

    private val requiredPermissions = buildList {
        add(Manifest.permission.CAMERA)
        add(Manifest.permission.RECORD_AUDIO)
        add(Manifest.permission.ACCESS_FINE_LOCATION)
        add(Manifest.permission.ACCESS_COARSE_LOCATION)
        add(Manifest.permission.FOREGROUND_SERVICE)
        add(Manifest.permission.WAKE_LOCK)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            add(Manifest.permission.POST_NOTIFICATIONS)
        }
    }

    private val permissionLauncher = registerForActivityResult(
        ActivityResultContracts.RequestMultiplePermissions()
    ) { permissions ->
        val allGranted = permissions.all { it.value }
        if (allGranted) {
            initializeAll()
        }
    }

    private var detectionService: FatigueDetectionService? = null
    private var isServiceBound = false

    private val serviceConnection = object : ServiceConnection {
        override fun onServiceConnected(name: ComponentName?, binder: IBinder?) {
            val serviceBinder = binder as FatigueDetectionService.LocalBinder
            detectionService = serviceBinder.getService()
            isServiceBound = true
            observeServiceData()
        }

        override fun onServiceDisconnected(name: ComponentName?) {
            detectionService = null
            isServiceBound = false
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityMainBinding.inflate(layoutInflater)
        setContentView(binding.root)

        setupFullscreen()
        window.addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)

        checkPermissions()
    }

    private fun setupFullscreen() {
        WindowCompat.setDecorFitsSystemWindows(window, false)
        ViewCompat.setOnApplyWindowInsetsListener(binding.root) { _, insets ->
            val systemBars = insets.getInsets(WindowInsetsCompat.Type.systemBars())
            binding.root.setPadding(systemBars.left, systemBars.top, systemBars.right, systemBars.bottom)
            insets
        }
        WindowInsetsControllerCompat(window, binding.root).let { controller ->
            controller.hide(WindowInsetsCompat.Type.systemBars())
            controller.systemBarsBehavior =
                WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
        }
    }

    private fun checkPermissions() {
        val needToRequest = requiredPermissions.filter {
            ContextCompat.checkSelfPermission(this, it) != PackageManager.PERMISSION_GRANTED
        }
        if (needToRequest.isEmpty()) {
            initializeAll()
        } else {
            permissionLauncher.launch(needToRequest.toTypedArray())
        }
    }

    private fun initializeAll() {
        viewModel.initialize(applicationContext)
        setupTextureView()
        observeViewModel()
        bindAndStartService()
        setupCameraX()
        setupPeriodicUploads()
    }

    private fun setupTextureView() {
        binding.textureView.surfaceTextureListener = object : TextureView.SurfaceTextureListener {
            override fun onSurfaceTextureAvailable(
                surface: SurfaceTexture,
                width: Int,
                height: Int
            ) {
            }

            override fun onSurfaceTextureSizeChanged(
                surface: SurfaceTexture,
                width: Int,
                height: Int
            ) {
            }

            override fun onSurfaceTextureDestroyed(surface: SurfaceTexture): Boolean = true

            override fun onSurfaceTextureUpdated(surface: SurfaceTexture) {
                viewModel.onSurfaceUpdated(binding.textureView)
            }
        }
    }

    private fun observeViewModel() {
        viewModel.frameResult.observe(this) { result ->
            updateUI(result)
        }

        viewModel.fatigueScore.observe(this) { score ->
            binding.scoreProgress.progress = score.coerceIn(0, 100)
            binding.scoreText.text = score.toString()
            updateScoreColor(score)
        }

        viewModel.currentAlarm.observe(this) { alarm ->
            updateAlarmIndicator(alarm?.level)
        }
    }

    private fun observeServiceData() {
    }

    private fun updateUI(result: DetectionFrameResult) {
        binding.earText.text = getString(R.string.ear_label).format(result.ear)
            .let { "EAR: %.2f".format(result.ear) }
        binding.marText.text = "MAR: %.2f".format(result.mar)
        binding.perclosText.text = "PERCLOS: %.3f".format(result.perclos)
        binding.poseText.text = "Pitch: %.1f° Yaw: %.1f°".format(result.headPitch, result.headYaw)
    }

    private fun updateScoreColor(score: Int) {
        val colorRes = when {
            score >= 80 -> R.color.fatigue_normal
            score >= 60 -> R.color.fatigue_warning
            score >= 40 -> R.color.fatigue_alert
            else -> R.color.fatigue_danger
        }
        val color = ContextCompat.getColor(this, colorRes)
        binding.scoreProgress.progressTintList =
            android.content.res.ColorStateList.valueOf(color)
        binding.scoreText.setTextColor(color)
    }

    private fun updateAlarmIndicator(level: AlarmLevel?) {
        val drawableRes = when (level) {
            AlarmLevel.LEVEL_1 -> R.drawable.indicator_warning
            AlarmLevel.LEVEL_2, AlarmLevel.LEVEL_3 -> R.drawable.indicator_danger
            else -> R.drawable.indicator_normal
        }
        binding.alarmIndicator.setBackgroundResource(drawableRes)
    }

    private fun bindAndStartService() {
        val serviceIntent = Intent(this, FatigueDetectionService::class.java)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            startForegroundService(serviceIntent)
        } else {
            startService(serviceIntent)
        }
        bindService(serviceIntent, serviceConnection, Context.BIND_AUTO_CREATE)
    }

    private fun setupCameraX() {
        val cameraProviderFuture = ProcessCameraProvider.getInstance(this)
        cameraProviderFuture.addListener({
            cameraProvider = cameraProviderFuture.get()
            bindCameraUseCases()
        }, ContextCompat.getMainExecutor(this))
    }

    private fun bindCameraUseCases() {
        val provider = cameraProvider ?: return

        val cameraSelector = CameraSelector.Builder()
            .requireLensFacing(CameraSelector.LENS_FACING_FRONT)
            .build()

        val resolutionSelector = CameraSizeSelector.createResolutionSelector(640, 480)

        val preview = Preview.Builder()
            .setResolutionSelector(resolutionSelector)
            .setTargetRotation(binding.textureView.display.rotation)
            .build()
            .also {
                it.setSurfaceProvider { surfaceRequest ->
                    val surfaceTexture = binding.textureView.surfaceTexture ?: return@setSurfaceProvider
                    surfaceTexture.setDefaultBufferSize(
                        surfaceRequest.resolution.width,
                        surfaceRequest.resolution.height
                    )
                    val executor = ContextCompat.getMainExecutor(this)
                    surfaceRequest.provideSurface(
                        android.view.Surface(surfaceTexture),
                        executor
                    ) { }
                }
            }

        val imageAnalysis = ImageAnalysis.Builder()
            .setResolutionSelector(resolutionSelector)
            .setBackpressureStrategy(ImageAnalysis.STRATEGY_KEEP_ONLY_LATEST)
            .setTargetRotation(binding.textureView.display.rotation)
            .build()
            .also {
                it.setAnalyzer(cameraExecutor) { imageProxy ->
                    viewModel.analyzeFrame(imageProxy)
                }
            }

        provider.unbindAll()

        camera = provider.bindToLifecycle(
            this,
            cameraSelector,
            preview,
            imageAnalysis
        )
    }

    private fun setupPeriodicUploads() {
        val constraints = Constraints.Builder()
            .setRequiredNetworkType(NetworkType.CONNECTED)
            .build()

        val uploadWorkRequest = PeriodicWorkRequestBuilder<AlarmUploadWorker>(
            15, TimeUnit.MINUTES
        )
            .setConstraints(constraints)
            .build()

        WorkManager.getInstance(this).enqueueUniquePeriodicWork(
            AlarmUploadWorker.WORK_NAME,
            ExistingPeriodicWorkPolicy.KEEP,
            uploadWorkRequest
        )
    }

    override fun onDestroy() {
        super.onDestroy()
        if (isServiceBound) {
            unbindService(serviceConnection)
            isServiceBound = false
        }
        cameraExecutor.shutdown()
        cameraProvider?.unbindAll()
    }
}

class FatigueViewModel(application: android.app.Application) : AndroidViewModel(application) {

    private val _frameResult = MutableLiveData<DetectionFrameResult>()
    val frameResult: LiveData<DetectionFrameResult> = _frameResult

    private val _fatigueScore = MutableLiveData(100)
    val fatigueScore: LiveData<Int> = _fatigueScore

    private val _currentAlarm = MutableLiveData<FatigueAlarm?>(null)
    val currentAlarm: LiveData<FatigueAlarm?> = _currentAlarm

    private lateinit var faceLandmarkDetector: FaceLandmarkDetector
    private lateinit var perclosCalculator: PERCLOSCalculator
    private lateinit var scoreAggregator: FatigueScoreAggregator
    private lateinit var yawnDetector: YawnDetector
    private lateinit var gazeDetector: GazeDirectionDetector
    private lateinit var gpsTracker: GPSTracker
    private lateinit var voiceAlertManager: VoiceAlertManager
    private lateinit var vibrationController: SeatVibrationController
    private lateinit var videoRecorder: LocalVideoRecorder
    private lateinit var alarmReportManager: AlarmReportManager
    private lateinit var offlineAlarmRepository: OfflineAlarmRepository
    private lateinit var apiService: PlatformApiService

    private val detectorScope = CoroutineScope(SupervisorJob() + Dispatchers.Default)
    private var analysisJob: Job? = null
    private var isInitialized = false
    private var frameIdCounter = 0L
    private var lastAlarmTime: Long = 0L
    private var lastVideoTriggerTime: Long = 0L

    private val alarmCooldownMs = 10_000L
    private val videoCooldownMs = 30_000L

    private var lastTextureBitmap: Bitmap? = null

    fun initialize(context: Context) {
        if (isInitialized) return

        detectorScope.launch {
            faceLandmarkDetector = FaceLandmarkDetector(context)
            faceLandmarkDetector.initialize()

            perclosCalculator = PERCLOSCalculator()
            scoreAggregator = FatigueScoreAggregator()
            yawnDetector = YawnDetector()
            gazeDetector = GazeDirectionDetector()

            val database = AlarmDatabase.getDatabase(context)
            offlineAlarmRepository = OfflineAlarmRepository(database.alarmDao(), Dispatchers.IO)

            apiService = AlarmUploadWorkerDelegate.createApiService(
                AlarmUploadWorker.DEFAULT_BASE_URL
            )

            gpsTracker = GPSTracker(context)
            gpsTracker.startTracking()
            gpsTracker.startPeriodicReports(3000)

            voiceAlertManager = VoiceAlertManager(context)
            voiceAlertManager.initialize()

            vibrationController = SeatVibrationController(
                context.getSystemService(Context.USB_SERVICE) as android.hardware.usb.UsbManager
            )
            detectorScope.launch { vibrationController.connect() }

            videoRecorder = LocalVideoRecorder(context)
            detectorScope.launch { videoRecorder.initialize() }

            alarmReportManager = AlarmReportManager(
                context = context,
                apiService = apiService,
                alarmRepository = offlineAlarmRepository
            )
            alarmReportManager.start()

            isInitialized = true
        }
    }

    fun analyzeFrame(imageProxy: androidx.camera.core.ImageProxy) {
        if (!isInitialized) {
            imageProxy.close()
            return
        }

        analysisJob = detectorScope.launch {
            try {
                val timestamp = System.currentTimeMillis()
                val frameId = ++frameIdCounter

                val bitmap = withContext(Dispatchers.Default) {
                    imageProxy.toBitmap().let {
                        FrameUtils.rotateBitmap(it, 0, true)
                    }
                }

                val mpImage = FrameUtils.bitmapToMPImage(
                    FrameUtils.resizeBitmap(bitmap, 640, 480)
                )

                val metadata = FrameMetadata.create(
                    frameId = frameId,
                    width = bitmap.width,
                    height = bitmap.height,
                    rotationDegrees = 0,
                    isMirrored = true
                )

                val landmarks = faceLandmarkDetector.detectAsync(
                    mpImage,
                    metadata.timestampMs = timestamp
                )

                val ear = if (landmarks != null) {
                    FatigueMetricsCalculator.calculateEAR(landmarks)
                } else 0f

                val mar = if (landmarks != null) {
                    FatigueMetricsCalculator.calculateMAR10Point(landmarks)
                } else 0f

                val headPose = if (landmarks != null) {
                    FatigueMetricsCalculator.calculateHeadPose(landmarks)
                } else FatigueMetricsCalculator.HeadPose(0f, 0f, 0f)

                val gazeResult = if (landmarks != null) {
                    gazeDetector.detect(landmarks)
                } else GazeDirectionDetector.GazeResult(
                    com.ddg.edge.fatigue.model.GazeDirection.UNKNOWN,
                    0f, 0f, 0f, false
                )

                val perclos = perclosCalculator.addFrame(timestamp, ear)
                val isYawning = yawnDetector.processFrame(timestamp, mar)
                val isHeadDown = headPose.pitch > 15f
                val isGazeAway = gazeResult.isGazeAway

                val aggregation = scoreAggregator.aggregate(
                    perclos = perclos,
                    mar = mar,
                    headPitch = headPose.pitch,
                    gazeAngleDeviation = gazeResult.angleDeviation,
                    isYawning = isYawning
                )

                val gpsLoc = gpsTracker.getCurrentLocationSnapshot()

                val result = DetectionFrameResult(
                    timestamp = timestamp,
                    frameId = frameId,
                    landmarks = landmarks,
                    ear = ear,
                    mar = mar,
                    perclos = perclos,
                    headPitch = headPose.pitch,
                    headYaw = headPose.yaw,
                    headRoll = headPose.roll,
                    gazeDirection = gazeResult.direction,
                    gazeAngleDeviation = gazeResult.angleDeviation,
                    isYawning = isYawning,
                    isHeadDown = isHeadDown,
                    isGazeAway = isGazeAway,
                    fatigueScore = aggregation.score,
                    alarmLevel = aggregation.level,
                    gpsLatitude = gpsLoc?.latitude,
                    gpsLongitude = gpsLoc?.longitude,
                    gpsSpeed = gpsLoc?.speed
                )

                withContext(Dispatchers.Main) {
                    _frameResult.value = result
                    _fatigueScore.value = aggregation.score
                }

                if (aggregation.level >= AlarmLevel.LEVEL_1) {
                    handleAlarm(result, aggregation)
                } else {
                    _currentAlarm.postValue(null)
                }

                if (frameId % 10 == 0L) {
                    reportFrameIfNeeded(result)
                }

            } catch (e: Exception) {
            } finally {
                imageProxy.close()
            }
        }
    }

    private suspend fun handleAlarm(
        result: DetectionFrameResult,
        aggregation: FatigueScoreAggregator.AggregationResult
    ) {
        val now = System.currentTimeMillis()

        val alarm = FatigueAlarm(
            alarmId = UUID.randomUUID().toString(),
            timestamp = result.timestamp,
            level = aggregation.level,
            score = aggregation.score,
            triggers = aggregation.triggers,
            ear = result.ear,
            mar = result.mar,
            perclos = result.perclos,
            headPitch = result.headPitch,
            headYaw = result.headYaw,
            gazeAngle = result.gazeAngleDeviation,
            isYawning = result.isYawning,
            gpsLatitude = result.gpsLatitude,
            gpsLongitude = result.gpsLongitude,
            gpsSpeed = result.gpsSpeed
        )

        _currentAlarm.postValue(alarm)

        if (now - lastAlarmTime >= alarmCooldownMs) {
            lastAlarmTime = now

            voiceAlertManager.speakForAlarmLevel(
                level = alarm.level,
                isYawning = alarm.isYawning,
                isHeadDown = result.isHeadDown,
                isGazeAway = result.isGazeAway
            )

            detectorScope.launch {
                vibrationController.vibrateForAlarmLevel(alarm.level)
            }

            if (now - lastVideoTriggerTime >= videoCooldownMs) {
                lastVideoTriggerTime = now
                detectorScope.launch {
                    if (!videoRecorder.isRecordingActive()) {
                        videoRecorder.startRecording()
                        delay(100)
                    }
                    videoRecorder.triggerAlarm()
                }
            }

            alarmReportManager.reportAlarm(alarm)
        }
    }

    private fun reportFrameIfNeeded(result: DetectionFrameResult) {
        val prefs = getApplication<android.app.Application>()
            .getSharedPreferences("device_prefs", Context.MODE_PRIVATE)
        val deviceId = prefs.getString("device_id", "unknown") ?: "unknown"

        alarmReportManager.reportFrame(
            AlarmReportManager.FrameReport(
                deviceId = deviceId,
                timestamp = result.timestamp,
                score = result.fatigueScore,
                alarmLevel = result.alarmLevel.value,
                ear = result.ear,
                mar = result.mar,
                perclos = result.perclos,
                gpsLatitude = result.gpsLatitude,
                gpsLongitude = result.gpsLongitude
            )
        )
    }

    fun onSurfaceUpdated(textureView: TextureView) {
        if (!textureView.isAvailable) return
        runCatching {
            val bitmap = textureView.bitmap
            if (bitmap != null) {
                lastTextureBitmap = bitmap
            }
        }
    }

    override fun onCleared() {
        super.onCleared()
        detectorScope.coroutineContext.cancelChildren()
        runCatching { faceLandmarkDetector.close() }
        runCatching { voiceAlertManager.release() }
        runCatching { vibrationController.setEnabled(false) }
        runCatching { gpsTracker.stopPeriodicReports() }
        runCatching { alarmReportManager.release() }
    }
}

private fun androidx.camera.core.ImageProxy.toBitmap(): Bitmap {
    val buffer = planes[0].buffer
    val pixelStride = planes[0].pixelStride
    val rowStride = planes[0].rowStride
    val rowPadding = rowStride - pixelStride * width

    val bitmap = Bitmap.createBitmap(
        width + rowPadding / pixelStride,
        height,
        Bitmap.Config.ARGB_8888
    )
    buffer.rewind()
    bitmap.copyPixelsFromBuffer(buffer)
    return Bitmap.createBitmap(bitmap, 0, 0, width, height)
}
