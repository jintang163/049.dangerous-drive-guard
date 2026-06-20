package com.ddg.edge.fatigue

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.graphics.Bitmap
import android.graphics.PixelFormat
import android.hardware.usb.UsbManager
import android.os.Binder
import android.os.Build
import android.os.Handler
import android.os.IBinder
import android.os.Looper
import android.view.Gravity
import android.view.LayoutInflater
import android.view.WindowManager
import android.widget.ProgressBar
import android.widget.TextView
import androidx.core.app.NotificationCompat
import androidx.camera.core.CameraSelector
import androidx.camera.core.ImageAnalysis
import androidx.camera.core.Preview
import androidx.camera.core.resolutionselector.ResolutionSelector
import androidx.camera.lifecycle.ProcessCameraProvider
import androidx.core.content.ContextCompat
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleOwner
import androidx.lifecycle.LifecycleRegistry
import androidx.work.Constraints
import androidx.work.ExistingWorkPolicy
import androidx.work.NetworkType
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.WorkManager
import com.ddg.edge.fatigue.alert.AlarmReportManager
import com.ddg.edge.fatigue.alert.LocalVideoRecorder
import com.ddg.edge.fatigue.alert.SeatVibrationController
import com.ddg.edge.fatigue.alert.VoiceAlertManager
import com.ddg.edge.fatigue.data.local.AlarmDatabase
import com.ddg.edge.fatigue.data.local.OfflineAlarmRepository
import com.ddg.edge.fatigue.data.remote.AlarmUploadWorker
import com.ddg.edge.fatigue.data.remote.AlarmUploadWorkerDelegate
import com.ddg.edge.fatigue.data.remote.PlatformApiService
import com.ddg.edge.fatigue.detector.FaceLandmarkDetector
import com.ddg.edge.fatigue.detector.FatigueMetricsCalculator
import com.ddg.edge.fatigue.detector.FatigueScoreAggregator
import com.ddg.edge.fatigue.detector.GazeDirectionDetector
import com.ddg.edge.fatigue.detector.PERCLOSCalculator
import com.ddg.edge.fatigue.detector.YawnDetector
import com.ddg.edge.fatigue.model.AlarmLevel
import com.ddg.edge.fatigue.model.DetectionFrameResult
import com.ddg.edge.fatigue.model.FatigueAlarm
import com.ddg.edge.fatigue.model.GazeDirection
import com.ddg.edge.fatigue.tracking.GPSTracker
import com.ddg.edge.fatigue.utils.CameraSizeSelector
import com.ddg.edge.fatigue.utils.FrameMetadata
import com.ddg.edge.fatigue.utils.FrameUtils
import com.google.mediapipe.tasks.components.containers.NormalizedLandmark
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancelChildren
import kotlinx.coroutines.channels.awaitClose
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.callbackFlow
import kotlinx.coroutines.flow.flowOn
import kotlinx.coroutines.flow.launchIn
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.mapLatest
import kotlinx.coroutines.flow.onEach
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.util.UUID
import java.util.concurrent.Executors
import kotlin.math.abs

class FatigueDetectionService : Service(), LifecycleOwner {

    private val binder = LocalBinder()
    private lateinit var lifecycleRegistry: LifecycleRegistry

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

    private val serviceScope = CoroutineScope(SupervisorJob() + Dispatchers.Default)
    private val cameraExecutor = Executors.newFixedThreadPool(4)
    private val uiHandler = Handler(Looper.getMainLooper())

    private var cameraProvider: ProcessCameraProvider? = null
    private var frameIdCounter = 0L
    private var lastAlarmTime: Long = 0L
    private var lastVideoTriggerTime: Long = 0L
    private val alarmCooldownMs = 10_000L
    private val videoCooldownMs = 30_000L

    private val _detectionState = MutableStateFlow(DetectionState())
    val detectionState: StateFlow<DetectionState> = _detectionState.asStateFlow()

    private val _lastResult = MutableStateFlow<DetectionFrameResult?>(null)
    val lastResult: StateFlow<DetectionFrameResult?> = _lastResult.asStateFlow()

    private val _lastAlarm = MutableStateFlow<FatigueAlarm?>(null)
    val lastAlarm: StateFlow<FatigueAlarm?> = _lastAlarm.asStateFlow()

    private var floatWindow: android.view.View? = null
    private var windowManager: WindowManager? = null

    data class DetectionState(
        val isRunning: Boolean = false,
        val score: Int = 100,
        val level: AlarmLevel = AlarmLevel.NONE,
        val framesProcessed: Long = 0L,
        val lastFrameTimestamp: Long = 0L,
        val gpsConnected: Boolean = false,
        val detectorReady: Boolean = false,
        val wsConnected: Boolean = false,
        val pendingAlarms: Int = 0
    )

    inner class LocalBinder : Binder() {
        fun getService(): FatigueDetectionService = this@FatigueDetectionService
    }

    override fun getLifecycle(): Lifecycle = lifecycleRegistry

    override fun onCreate() {
        super.onCreate()
        lifecycleRegistry = LifecycleRegistry(this)
        lifecycleRegistry.currentState = Lifecycle.State.CREATED

        createNotificationChannel()
        startForeground(NOTIFICATION_ID, buildNotification(DetectionState()))

        initializeComponents()
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        lifecycleRegistry.currentState = Lifecycle.State.STARTED
        lifecycleRegistry.currentState = Lifecycle.State.RESUMED

        serviceScope.launch {
            startDetection()
        }

        return START_STICKY
    }

    override fun onBind(intent: Intent?): IBinder {
        return binder
    }

    private fun initializeComponents() {
        serviceScope.launch {
            faceLandmarkDetector = FaceLandmarkDetector(this@FatigueDetectionService)
            faceLandmarkDetector.initialize()

            perclosCalculator = PERCLOSCalculator()
            scoreAggregator = FatigueScoreAggregator()
            yawnDetector = YawnDetector()
            gazeDetector = GazeDirectionDetector()

            val database = AlarmDatabase.getDatabase(this@FatigueDetectionService)
            offlineAlarmRepository = OfflineAlarmRepository(database.alarmDao(), Dispatchers.IO)

            apiService = AlarmUploadWorkerDelegate.createApiService(
                AlarmUploadWorker.DEFAULT_BASE_URL
            )

            gpsTracker = GPSTracker(this@FatigueDetectionService)
            gpsTracker.startTracking()
            gpsTracker.startPeriodicReports(3000)

            voiceAlertManager = VoiceAlertManager(this@FatigueDetectionService)
            voiceAlertManager.initialize()

            vibrationController = SeatVibrationController(
                getSystemService(Context.USB_SERVICE) as UsbManager
            )
            launch { vibrationController.connect() }

            videoRecorder = LocalVideoRecorder(this@FatigueDetectionService)
            launch { videoRecorder.initialize() }

            alarmReportManager = AlarmReportManager(
                context = this@FatigueDetectionService,
                apiService = apiService,
                alarmRepository = offlineAlarmRepository
            )
            alarmReportManager.start()

            observeGps()
            observePendingAlarms()
            observeWsConnection()
            observeVideoRecorderEvents()

            _detectionState.value = _detectionState.value.copy(detectorReady = true)
        }
    }

    private fun observeGps() {
        gpsTracker.gpsEnabled
            .onEach { enabled ->
                _detectionState.value = _detectionState.value.copy(gpsConnected = enabled)
            }
            .launchIn(serviceScope)
    }

    private fun observePendingAlarms() {
        offlineAlarmRepository.observePendingAlarms()
            .onEach { list ->
                _detectionState.value = _detectionState.value.copy(pendingAlarms = list.size)
            }
            .launchIn(serviceScope)
    }

    private fun observeWsConnection() {
        alarmReportManager.wsConnected
            .onEach { connected ->
                _detectionState.value = _detectionState.value.copy(wsConnected = connected)
                if (connected) {
                    uploadPendingAlarms()
                }
            }
            .launchIn(serviceScope)
    }

    private fun observeVideoRecorderEvents() {
        serviceScope.launch {
            for (event in videoRecorder.recordingEvents) {
                when (event) {
                    is LocalVideoRecorder.RecordingEvent.Saved -> {
                        _lastAlarm.value?.let { alarm ->
                            val updatedAlarm = alarm.copy(videoPath = event.file.absolutePath)
                            _lastAlarm.value = updatedAlarm
                            serviceScope.launch {
                                offlineAlarmRepository.saveAlarm(updatedAlarm)
                            }
                        }
                    }
                    is LocalVideoRecorder.RecordingEvent.Error -> {
                    }
                    else -> {}
                }
            }
        }
    }

    private suspend fun startDetection() {
        withContext(Dispatchers.Main) {
            setupCameraX()
            setupFloatWindow()
        }
        _detectionState.value = _detectionState.value.copy(isRunning = true)
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
        val selector = CameraSelector.Builder()
            .requireLensFacing(CameraSelector.LENS_FACING_FRONT)
            .build()

        val resolutionSelector = CameraSizeSelector.createResolutionSelector(640, 480)

        val preview = Preview.Builder()
            .setResolutionSelector(resolutionSelector)
            .build()

        val imageAnalysis = ImageAnalysis.Builder()
            .setResolutionSelector(resolutionSelector)
            .setBackpressureStrategy(ImageAnalysis.STRATEGY_KEEP_ONLY_LATEST)
            .build()
            .also {
                it.setAnalyzer(cameraExecutor) { imageProxy ->
                    analyzeFrame(imageProxy)
                }
            }

        provider.unbindAll()
        provider.bindToLifecycle(this, selector, preview, imageAnalysis)
    }

    private fun analyzeFrame(imageProxy: androidx.camera.core.ImageProxy) {
        serviceScope.launch {
            try {
                val timestamp = System.currentTimeMillis()
                val frameId = ++frameIdCounter

                val bitmap = withContext(Dispatchers.Default) {
                    toBitmap(imageProxy).let {
                        FrameUtils.rotateBitmap(it, 0, true)
                    }
                }

                val mpImage = FrameUtils.bitmapToMPImage(
                    FrameUtils.resizeBitmap(bitmap, 640, 480)
                )

                val landmarks = faceLandmarkDetector.detectAsync(mpImage, timestamp)

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
                    GazeDirection.UNKNOWN, 0f, 0f, 0f, false
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

                _lastResult.value = result

                val state = _detectionState.value
                _detectionState.value = state.copy(
                    score = aggregation.score,
                    level = aggregation.level,
                    framesProcessed = frameId,
                    lastFrameTimestamp = timestamp
                )

                updateFloatWindow(aggregation.score, aggregation.level)
                updateNotification(_detectionState.value)

                if (aggregation.level >= AlarmLevel.LEVEL_1) {
                    handleAlarm(result, aggregation)
                }

                if (frameId % 10 == 0L) {
                    reportFrame(result)
                }

                if (frameId % 600 == 0L) {
                    scheduleOfflineUpload()
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

        _lastAlarm.value = alarm

        if (now - lastAlarmTime >= alarmCooldownMs) {
            lastAlarmTime = now

            voiceAlertManager.speakForAlarmLevel(
                level = alarm.level,
                isYawning = alarm.isYawning,
                isHeadDown = result.isHeadDown,
                isGazeAway = result.isGazeAway
            )

            serviceScope.launch {
                vibrationController.vibrateForAlarmLevel(alarm.level)
            }

            if (now - lastVideoTriggerTime >= videoCooldownMs) {
                lastVideoTriggerTime = now
                serviceScope.launch {
                    if (!videoRecorder.isRecordingActive()) {
                        videoRecorder.startRecording()
                        delay(100)
                    }
                    videoRecorder.triggerAlarm()
                }
            }

            alarmReportManager.reportAlarm(alarm)
            serviceScope.launch { offlineAlarmRepository.saveAlarm(alarm) }
        }
    }

    private fun reportFrame(result: DetectionFrameResult) {
        val prefs = getSharedPreferences("device_prefs", Context.MODE_PRIVATE)
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

    private fun scheduleOfflineUpload() {
        val constraints = Constraints.Builder()
            .setRequiredNetworkType(NetworkType.CONNECTED)
            .build()

        val uploadRequest = OneTimeWorkRequestBuilder<AlarmUploadWorker>()
            .setConstraints(constraints)
            .build()

        WorkManager.getInstance(this).enqueueUniqueWork(
            "OfflineAlarmUpload_${System.currentTimeMillis()}",
            ExistingWorkPolicy.REPLACE,
            uploadRequest
        )
    }

    private suspend fun uploadPendingAlarms() {
        val delegate = AlarmUploadWorkerDelegate(
            context = this,
            alarmRepository = offlineAlarmRepository,
            apiService = apiService
        )
        delegate.executeUpload()
    }

    private fun setupFloatWindow() {
        try {
            windowManager = getSystemService(WINDOW_SERVICE) as WindowManager

            val layoutInflater = getSystemService(LAYOUT_INFLATER_SERVICE) as LayoutInflater
            floatWindow = layoutInflater.inflate(R.layout.float_window_score, null)

            val type = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY
            } else {
                @Suppress("DEPRECATION")
                WindowManager.LayoutParams.TYPE_PHONE
            }

            val params = WindowManager.LayoutParams(
                WindowManager.LayoutParams.WRAP_CONTENT,
                WindowManager.LayoutParams.WRAP_CONTENT,
                type,
                WindowManager.LayoutParams.FLAG_NOT_FOCUSABLE or
                    WindowManager.LayoutParams.FLAG_NOT_TOUCHABLE or
                    WindowManager.LayoutParams.FLAG_LAYOUT_IN_SCREEN,
                PixelFormat.TRANSLUCENT
            )

            params.gravity = Gravity.TOP or Gravity.END
            params.x = 16
            params.y = 16

            windowManager?.addView(floatWindow, params)
        } catch (e: Exception) {
        }
    }

    private fun updateFloatWindow(score: Int, level: AlarmLevel) {
        val view = floatWindow ?: return
        uiHandler.post {
            runCatching {
                val progressBar = view.findViewById<ProgressBar>(R.id.floatScoreProgress)
                val scoreText = view.findViewById<TextView>(R.id.floatScoreText)
                progressBar?.progress = score.coerceIn(0, 100)
                scoreText?.text = score.toString()

                val color = when (level) {
                    AlarmLevel.LEVEL_1 -> R.color.fatigue_warning
                    AlarmLevel.LEVEL_2 -> R.color.fatigue_alert
                    AlarmLevel.LEVEL_3 -> R.color.fatigue_danger
                    else -> R.color.fatigue_normal
                }
                val colorValue = ContextCompat.getColor(this, color)
                progressBar?.progressTintList =
                    android.content.res.ColorStateList.valueOf(colorValue)
                scoreText?.setTextColor(colorValue)
            }
        }
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID,
                getString(R.string.notification_channel_detection),
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = getString(R.string.notification_detection_running)
                setShowBadge(false)
            }
            val manager = getSystemService(NotificationManager::class.java)
            manager.createNotificationChannel(channel)
        }
    }

    private fun buildNotification(state: DetectionState): Notification {
        val intent = Intent(this, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_SINGLE_TOP
        }
        val pendingIntent = PendingIntent.getActivity(
            this, 0, intent,
            PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
        )

        val statusText = when (state.level) {
            AlarmLevel.LEVEL_1 -> getString(R.string.fatigue_warning)
            AlarmLevel.LEVEL_2 -> getString(R.string.fatigue_alert)
            AlarmLevel.LEVEL_3 -> getString(R.string.fatigue_danger)
            else -> getString(R.string.fatigue_normal)
        }

        return NotificationCompat.Builder(this, CHANNEL_ID)
            .setContentTitle("疲劳检测防护")
            .setContentText("$statusText | 指数: ${state.score}")
            .setSmallIcon(android.R.drawable.ic_menu_view)
            .setContentIntent(pendingIntent)
            .setOngoing(true)
            .setPriority(NotificationCompat.PRIORITY_LOW)
            .setShowWhen(true)
            .build()
    }

    private fun updateNotification(state: DetectionState) {
        val notification = buildNotification(state)
        val manager = getSystemService(NotificationManager::class.java)
        manager.notify(NOTIFICATION_ID, notification)
    }

    private fun toBitmap(imageProxy: androidx.camera.core.ImageProxy): Bitmap {
        val buffer = imageProxy.planes[0].buffer
        val pixelStride = imageProxy.planes[0].pixelStride
        val rowStride = imageProxy.planes[0].rowStride
        val rowPadding = rowStride - pixelStride * imageProxy.width

        val bitmap = Bitmap.createBitmap(
            imageProxy.width + rowPadding / pixelStride,
            imageProxy.height,
            Bitmap.Config.ARGB_8888
        )
        buffer.rewind()
        bitmap.copyPixelsFromBuffer(buffer)
        return Bitmap.createBitmap(bitmap, 0, 0, imageProxy.width, imageProxy.height)
    }

    override fun onDestroy() {
        lifecycleRegistry.currentState = Lifecycle.State.DESTROYED
        serviceScope.coroutineContext.cancelChildren()
        cameraExecutor.shutdown()
        cameraProvider?.unbindAll()
        runCatching { faceLandmarkDetector.close() }
        runCatching { voiceAlertManager.release() }
        runCatching { vibrationController.setEnabled(false) }
        runCatching { gpsTracker.stopPeriodicReports() }
        runCatching { alarmReportManager.release() }
        runCatching { windowManager?.removeView(floatWindow) }
        super.onDestroy()
    }

    companion object {
        private const val CHANNEL_ID = "fatigue_detection_channel"
        private const val NOTIFICATION_ID = 1001
    }
}
