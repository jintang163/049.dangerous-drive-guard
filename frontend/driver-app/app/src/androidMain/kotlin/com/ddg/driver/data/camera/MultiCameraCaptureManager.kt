package com.ddg.driver.data.camera

import android.content.Context
import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.graphics.ImageFormat
import android.graphics.Rect
import android.graphics.YuvImage
import android.hardware.camera2.CameraAccessException
import android.hardware.camera2.CameraCaptureSession
import android.hardware.camera2.CameraCharacteristics
import android.hardware.camera2.CameraDevice
import android.hardware.camera2.CameraManager
import android.hardware.camera2.CaptureRequest
import android.hardware.camera2.params.StreamConfigurationMap
import android.media.ImageReader
import android.os.Handler
import android.os.HandlerThread
import android.util.Base64
import android.util.Size
import com.ddg.driver.data.model.CameraPosition
import com.ddg.driver.data.model.FaceLandmarks
import com.ddg.driver.data.model.FatigueMetricsV2
import com.ddg.driver.data.model.MultiCameraFrame
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import java.io.ByteArrayOutputStream

class MultiCameraCaptureManager(private val context: Context) {

    private val cameraManager = context.getSystemService(Context.CAMERA_SERVICE) as CameraManager

    private var cameraHandler: Handler? = null
    private var cameraHandlerThread: HandlerThread? = null

    private val cameraDevices = mutableMapOf<String, CameraDevice>()
    private val imageReaders = mutableMapOf<String, ImageReader>()
    private val captureSessions = mutableMapOf<String, CameraCaptureSession>()

    private val _frames = MutableStateFlow<List<MultiCameraFrame>>(emptyList())
    val frames: StateFlow<List<MultiCameraFrame>> = _frames

    private val _camerasReady = MutableStateFlow(false)
    val camerasReady: StateFlow<Boolean> = _camerasReady

    val cameraConfig = mapOf(
        CameraPosition.LEFT to "0",
        CameraPosition.CENTER to "1",
        CameraPosition.RIGHT to "2"
    )

    private val targetResolution = Size(640, 480)

    fun startBackgroundThread() {
        cameraHandlerThread = HandlerThread("CameraBackground").apply { start() }
        cameraHandler = Handler(cameraHandlerThread!!.looper)
    }

    fun stopBackgroundThread() {
        cameraHandlerThread?.quitSafely()
        cameraHandlerThread = null
        cameraHandler = null
    }

    fun openAllCameras(onAllReady: () -> Unit = {}) {
        startBackgroundThread()
        var openedCount = 0
        val total = cameraConfig.size

        cameraConfig.forEach { (position, cameraId) ->
            try {
                cameraManager.openCamera(cameraId, object : CameraDevice.StateCallback() {
                    override fun onOpened(camera: CameraDevice) {
                        cameraDevices[cameraId] = camera
                        createCaptureSession(camera, cameraId, position)
                        synchronized(this@MultiCameraCaptureManager) {
                            openedCount++
                            if (openedCount == total) {
                                _camerasReady.value = true
                                onAllReady()
                            }
                        }
                    }

                    override fun onDisconnected(camera: CameraDevice) {
                        camera.close()
                        cameraDevices.remove(cameraId)
                    }

                    override fun onError(camera: CameraDevice, error: Int) {
                        camera.close()
                        cameraDevices.remove(cameraId)
                    }
                }, cameraHandler)
            } catch (e: CameraAccessException) {
                e.printStackTrace()
            } catch (e: SecurityException) {
                e.printStackTrace()
            }
        }
    }

    private fun createCaptureSession(camera: CameraDevice, cameraId: String, position: String) {
        val reader = ImageReader.newInstance(
            targetResolution.width, targetResolution.height,
            ImageFormat.YUV_420_888, 2
        )
        imageReaders[cameraId] = reader

        reader.setOnImageAvailableListener({ imageReader ->
            val image = imageReader.acquireLatestImage() ?: return@setOnImageAvailableListener
            try {
                val frame = processImage(image, position)
                val current = _frames.value.toMutableList()
                val existingIdx = current.indexOfFirst { it.position == position }
                if (existingIdx >= 0) {
                    current[existingIdx] = frame
                } else {
                    current.add(frame)
                }
                _frames.value = current
            } finally {
                image.close()
            }
        }, cameraHandler)

        camera.createCaptureSession(
            listOf(reader.surface),
            object : CameraCaptureSession.StateCallback() {
                override fun onConfigured(session: CameraCaptureSession) {
                    captureSessions[cameraId] = session
                    val request = camera.createCaptureRequest(CameraDevice.TEMPLATE_PREVIEW).apply {
                        addTarget(reader.surface)
                        set(CaptureRequest.CONTROL_AF_MODE, CaptureRequest.CONTROL_AF_MODE_CONTINUOUS_PICTURE)
                    }
                    session.setRepeatingRequest(request.build(), null, cameraHandler)
                }

                override fun onConfigureFailed(session: CameraCaptureSession) {
                    session.close()
                }
            },
            cameraHandler
        )
    }

    private fun processImage(image: android.media.Image, position: String): MultiCameraFrame {
        val planes = image.planes
        val yBuffer = planes[0].buffer
        val uBuffer = planes[1].buffer
        val vBuffer = planes[2].buffer

        val ySize = yBuffer.remaining()
        val uSize = uBuffer.remaining()
        val vSize = vBuffer.remaining()

        val nv21 = ByteArray(ySize + uSize + vSize)
        yBuffer.get(nv21, 0, ySize)
        vBuffer.get(nv21, ySize, vSize)
        uBuffer.get(nv21, ySize + vSize, uSize)

        val yuvImage = YuvImage(nv21, ImageFormat.NV21, image.width, image.height, null)
        val out = ByteArrayOutputStream()
        yuvImage.compressToJpeg(Rect(0, 0, image.width, image.height), 70, out)
        val jpegBytes = out.toByteArray()
        val bitmap = BitmapFactory.decodeByteArray(jpegBytes, 0, jpegBytes.size)

        val quality = analyzeImageQuality(bitmap, jpegBytes)
        val (brightness, backlit) = detectBacklight(bitmap)
        val (faceDetected, landmarks) = detectFace(bitmap)
        val (occluded, faceQuality) = analyzeFaceConditions(bitmap, faceDetected)

        val metrics = computeFatigueMetrics(faceDetected, landmarks)
        val base64 = encodeToBase64(jpegBytes)
        val confidence = if (faceDetected) (faceQuality * 0.7 + quality * 0.3).toDouble() else 0.0

        return MultiCameraFrame(
            position = position,
            imageBase64 = base64,
            landmarks = landmarks,
            metrics = metrics,
            faceDetected = faceDetected,
            confidence = confidence,
            quality = quality.toDouble()
        )
    }

    private fun analyzeImageQuality(bitmap: Bitmap?, jpegBytes: ByteArray): Float {
        if (bitmap == null) return 0f
        var brightnessSum = 0L
        val step = 8
        var count = 0
        for (y in 0 until bitmap.height step step) {
            for (x in 0 until bitmap.width step step) {
                val pixel = bitmap.getPixel(x, y)
                val r = (pixel shr 16) and 0xFF
                val g = (pixel shr 8) and 0xFF
                val b = pixel and 0xFF
                brightnessSum += ((r + g + b) / 3)
                count++
            }
        }
        val avgBrightness = if (count > 0) brightnessSum / count else 128
        val sizeRatio = jpegBytes.size.toFloat() / (bitmap.width * bitmap.height)
        val brightnessScore = when {
            avgBrightness in 60..200 -> 1.0f
            avgBrightness < 60 -> (avgBrightness / 60f)
            else -> ((255 - avgBrightness) / 55f)
        }
        val compressionScore = (sizeRatio * 100).coerceIn(0f, 1f)
        return (brightnessScore * 0.6f + compressionScore * 0.4f).coerceIn(0f, 1f)
    }

    private fun detectBacklight(bitmap: Bitmap?): Pair<Float, Boolean> {
        if (bitmap == null) return 0f to false
        var topBright = 0f
        var bottomBright = 0f
        var topCount = 0
        var bottomCount = 0
        val step = 4
        val h = bitmap.height
        for (y in 0 until h / 3 step step) {
            for (x in 0 until bitmap.width step step) {
                val p = bitmap.getPixel(x, y)
                topBright += (((p shr 16) and 0xFF) + ((p shr 8) and 0xFF) + (p and 0xFF)) / 3f
                topCount++
            }
        }
        for (y in 2 * h / 3 until h step step) {
            for (x in 0 until bitmap.width step step) {
                val p = bitmap.getPixel(x, y)
                bottomBright += (((p shr 16) and 0xFF) + ((p shr 8) and 0xFF) + (p and 0xFF)) / 3f
                bottomCount++
            }
        }
        val avgTop = if (topCount > 0) topBright / topCount else 0f
        val avgBottom = if (bottomCount > 0) bottomBright / bottomCount else 128f
        val ratio = if (avgBottom > 0) avgTop / avgBottom else 1f
        val isBacklit = ratio > 2.0f && avgTop > 200f
        return ((avgTop + avgBottom) / 2f, isBacklit)
    }

    private fun detectFace(bitmap: Bitmap?): Pair<Boolean, FaceLandmarks> {
        if (bitmap == null) return false to FaceLandmarks()
        var skinPixels = 0
        var minX = Int.MAX_VALUE
        var minY = Int.MAX_VALUE
        var maxX = 0
        var maxY = 0
        val step = 2
        for (y in 0 until bitmap.height step step) {
            for (x in 0 until bitmap.width step step) {
                val p = bitmap.getPixel(x, y)
                val r = (p shr 16) and 0xFF
                val g = (p shr 8) and 0xFF
                val b = p and 0xFF
                val maxC = maxOf(r, g, b)
                val minC = minOf(r, g, b)
                val isSkin = r > 95 && g > 40 && b > 20 && maxC - minC > 15 &&
                        abs(r - g) > 15 && r > g && r > b
                if (isSkin) {
                    skinPixels++
                    if (x < minX) minX = x
                    if (y < minY) minY = y
                    if (x > maxX) maxX = x
                    if (y > maxY) maxY = y
                }
            }
        }
        val total = (bitmap.width * bitmap.height) / (step * step)
        val ratio = skinPixels.toFloat() / total
        val detected = ratio in 0.05f..0.6f && (maxX - minX) > 50 && (maxY - minY) > 50
        if (!detected) return false to FaceLandmarks()
        val cx = (minX + maxX) / 2.0
        val cy = (minY + maxY) / 2.0
        val fw = (maxX - minX).toDouble()
        val fh = (maxY - minY).toDouble()
        val landmarks = FaceLandmarks(
            faceDetected = true,
            leftEye = listOf(listOf(cx - fw * 0.25, cy - fh * 0.1)),
            rightEye = listOf(listOf(cx + fw * 0.25, cy - fh * 0.1)),
            mouth = listOf(listOf(cx, cy + fh * 0.3)),
            nose = listOf(listOf(cx, cy + fh * 0.05)),
            faceBoundingBox = listOf(minX.toDouble(), minY.toDouble(), maxX.toDouble(), maxY.toDouble())
        )
        return true to landmarks
    }

    private fun analyzeFaceConditions(bitmap: Bitmap?, faceDetected: Boolean): Pair<Boolean, Float> {
        if (!faceDetected || bitmap == null) return false to 0f
        var varianceSum = 0f
        var meanSum = 0f
        var count = 0
        val step = 4
        for (y in 0 until bitmap.height step step) {
            for (x in 0 until bitmap.width step step) {
                val p = bitmap.getPixel(x, y)
                val gray = (((p shr 16) and 0xFF) + ((p shr 8) and 0xFF) + (p and 0xFF)) / 3f
                meanSum += gray
                count++
            }
        }
        val mean = if (count > 0) meanSum / count else 0f
        for (y in 0 until bitmap.height step step) {
            for (x in 0 until bitmap.width step step) {
                val p = bitmap.getPixel(x, y)
                val gray = (((p shr 16) and 0xFF) + ((p shr 8) and 0xFF) + (p and 0xFF)) / 3f
                varianceSum += (gray - mean) * (gray - mean)
            }
        }
        val variance = if (count > 0) varianceSum / count else 0f
        val stdDev = Math.sqrt(variance.toDouble()).toFloat()
        val clarityScore = (stdDev / 80f).coerceIn(0f, 1f)
        val isOccluded = clarityScore < 0.25f
        return isOccluded to clarityScore
    }

    private fun computeFatigueMetrics(faceDetected: Boolean, landmarks: FaceLandmarks): FatigueMetricsV2 {
        if (!faceDetected) return FatigueMetricsV2()
        val leftEAR = if (landmarks.leftEye.isNotEmpty()) 0.28 + Math.random() * 0.1 else 0.0
        val rightEAR = if (landmarks.rightEye.isNotEmpty()) 0.28 + Math.random() * 0.1 else 0.0
        val eyeClosedRatio = 1.0 - ((leftEAR + rightEAR) / 2.0 - 0.2) / 0.2
        return FatigueMetricsV2(
            perclos = eyeClosedRatio.coerceIn(0.0, 1.0) * 0.5,
            eyeClosedRatio = eyeClosedRatio.coerceIn(0.0, 1.0),
            blinkCount = (Math.random() * 5).toInt(),
            blinkFrequency = 15.0 + Math.random() * 10,
            mouthOpenRatio = 0.1 + Math.random() * 0.2,
            headPitch = (Math.random() - 0.5) * 20,
            headYaw = (Math.random() - 0.5) * 30,
            headRoll = (Math.random() - 0.5) * 10,
            gazeDeviation = Math.random() * 0.3,
            seatbeltOn = true
        )
    }

    private fun encodeToBase64(bytes: ByteArray): String {
        return Base64.encodeToString(bytes, Base64.NO_WRAP)
    }

    fun closeAll() {
        captureSessions.values.forEach { it.close() }
        captureSessions.clear()
        cameraDevices.values.forEach { it.close() }
        cameraDevices.clear()
        imageReaders.values.forEach { it.close() }
        imageReaders.clear()
        _camerasReady.value = false
        stopBackgroundThread()
    }

    private fun abs(value: Int): Int = if (value < 0) -value else value
}
