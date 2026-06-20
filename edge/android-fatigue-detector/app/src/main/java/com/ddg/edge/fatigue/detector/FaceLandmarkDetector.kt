package com.ddg.edge.fatigue.detector

import android.content.Context
import android.graphics.Bitmap
import com.google.mediapipe.framework.image.MPImage
import com.google.mediapipe.tasks.components.containers.NormalizedLandmark
import com.google.mediapipe.tasks.core.BaseOptions
import com.google.mediapipe.tasks.core.Delegate
import com.google.mediapipe.tasks.vision.core.RunningMode
import com.google.mediapipe.tasks.vision.facelandmarker.FaceLandmarker
import com.google.mediapipe.tasks.vision.facelandmarker.FaceLandmarkerResult
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.channels.awaitClose
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.callbackFlow
import kotlinx.coroutines.withContext

class FaceLandmarkDetector(
    private val context: Context,
    private val modelAssetPath: String = "face_landmarker_v2_with_blendshapes.task",
    private val numFaces: Int = 1,
    private val minFaceDetectionConfidence: Float = 0.5f,
    private val minFacePresenceConfidence: Float = 0.5f,
    private val minTrackingConfidence: Float = 0.5f,
    private val useGpu: Boolean = true
) {

    private var faceLandmarker: FaceLandmarker? = null
    private var isInitialized = false

    suspend fun initialize(): Result<Unit> = withContext(Dispatchers.IO) {
        runCatching {
            if (isInitialized) return@runCatching

            val baseOptionsBuilder = BaseOptions.builder()
                .setModelAssetPath(modelAssetPath)

            if (useGpu) {
                try {
                    baseOptionsBuilder.setDelegate(Delegate.GPU)
                } catch (e: Exception) {
                    baseOptionsBuilder.setDelegate(Delegate.CPU)
                }
            }

            val options = FaceLandmarker.FaceLandmarkerOptions.builder()
                .setBaseOptions(baseOptionsBuilder.build())
                .setRunningMode(RunningMode.LIVE_STREAM)
                .setNumFaces(numFaces)
                .setMinFaceDetectionConfidence(minFaceDetectionConfidence)
                .setMinFacePresenceConfidence(minFacePresenceConfidence)
                .setMinTrackingConfidence(minTrackingConfidence)
                .setResultListener { result, mpImage ->
                    lastResult = result
                }
                .setErrorListener { errorCode, message ->
                    lastError = "ErrorCode: $errorCode, Msg: $message"
                }
                .build()

            faceLandmarker = FaceLandmarker.createFromOptions(context, options)
            isInitialized = true
        }
    }

    private var lastResult: FaceLandmarkerResult? = null
    private var lastError: String? = null

    suspend fun detectAsync(
        mpImage: MPImage,
        timestampMs: Long
    ): List<NormalizedLandmark>? = withContext(Dispatchers.Default) {
        val detector = faceLandmarker ?: return@withContext null
        lastResult = null
        lastError = null

        runCatching {
            detector.detectAsync(mpImage, timestampMs)
        }.getOrElse {
            lastError = it.message
            null
        }

        Thread.sleep(2)
        extractLandmarks(lastResult)
    }

    suspend fun detectBitmap(bitmap: Bitmap): List<NormalizedLandmark>? = withContext(Dispatchers.Default) {
        val detector = faceLandmarker ?: return@withContext null
        runCatching {
            val result = detector.detect(bitmap)
            extractLandmarks(result)
        }.getOrNull()
    }

    private fun extractLandmarks(result: FaceLandmarkerResult?): List<NormalizedLandmark>? {
        if (result == null) return null
        val faceLandmarks = result.faceLandmarks()
        if (faceLandmarks.isEmpty()) return null
        return faceLandmarks[0]
    }

    fun getLastError(): String? = lastError

    fun close() {
        faceLandmarker?.close()
        faceLandmarker = null
        isInitialized = false
    }
}
