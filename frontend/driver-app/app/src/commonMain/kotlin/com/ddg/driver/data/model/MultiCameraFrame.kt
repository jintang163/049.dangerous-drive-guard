package com.ddg.driver.data.model

import kotlinx.serialization.Serializable

@Serializable
data class CameraPosition {
    companion object {
        const val LEFT = "left"
        const val CENTER = "center"
        const val RIGHT = "right"
    }
}

@Serializable
data class FaceLandmarks(
    val faceDetected: Boolean = false,
    val leftEye: List<List<Double>> = emptyList(),
    val rightEye: List<List<Double>> = emptyList(),
    val mouth: List<List<Double>> = emptyList(),
    val nose: List<List<Double>> = emptyList(),
    val faceBoundingBox: List<Double> = emptyList()
)

@Serializable
data class FatigueMetricsV2(
    val perclos: Double = 0.0,
    val eyeClosedRatio: Double = 0.0,
    val blinkCount: Int = 0,
    val blinkFrequency: Double = 0.0,
    val yawnCount: Int = 0,
    val mouthOpenRatio: Double = 0.0,
    val headPitch: Double = 0.0,
    val headYaw: Double = 0.0,
    val headRoll: Double = 0.0,
    val gazeDeviation: Double = 0.0,
    val phoneUsageDetected: Boolean = false,
    val smokingDetected: Boolean = false,
    val seatbeltOn: Boolean = true
)

@Serializable
data class MultiCameraFrame(
    val position: String,
    val imageUrl: String? = null,
    val imageBase64: String? = null,
    val landmarks: FaceLandmarks = FaceLandmarks(),
    val metrics: FatigueMetricsV2 = FatigueMetricsV2(),
    val faceDetected: Boolean = false,
    val confidence: Double = 0.0,
    val quality: Double = 0.0
)

@Serializable
data class MultiCameraUploadRequest(
    val vehicleId: Long,
    val driverId: Long,
    val waybillId: Long? = null,
    val frames: List<MultiCameraFrame> = emptyList(),
    val latitude: Double? = null,
    val longitude: Double? = null,
    val vehicleSpeed: Double? = null,
    val timestamp: Long = System.currentTimeMillis(),
    val edgeComputed: Boolean = true,
    val networkStatus: String = "online"
)

@Serializable
data class FusionResult(
    val fatigueScore: Double,
    val fatigueLevel: String,
    val fusionMethod: String,
    val usedCameras: List<String>,
    val primaryCamera: String,
    val fusionConfidence: Double,
    val leftScore: Double,
    val centerScore: Double,
    val rightScore: Double,
    val occlusionDetected: Boolean,
    val backlitDetected: Boolean
)

@Serializable
data class MultiCameraDetectResponse(
    val fatigueScore: Double,
    val fatigueLevel: String,
    val needAlarm: Boolean,
    val alarmType: String? = null,
    val alarmMessage: String? = null,
    val fusionResult: FusionResult? = null,
    val seatbeltAlert: Boolean = false,
    val phoneAlert: Boolean = false,
    val smokingAlert: Boolean = false,
    val recommendRestMinutes: Int = 0
)
