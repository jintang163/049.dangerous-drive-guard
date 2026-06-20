package com.ddg.edge.fatigue.data.remote

import com.ddg.edge.fatigue.model.AlarmLevel
import com.ddg.edge.fatigue.model.DetectionFrameResult
import com.ddg.edge.fatigue.model.FatigueAlarm
import com.ddg.edge.fatigue.model.StoredAlarm
import okhttp3.MultipartBody
import okhttp3.RequestBody
import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.Multipart
import retrofit2.http.POST
import retrofit2.http.Part
import retrofit2.http.Query

interface PlatformApiService {

    @POST("api/v1/detection/frames")
    suspend fun uploadDetectionFrame(
        @Body frameRequest: DetectionFrameUploadRequest
    ): Response<ApiResponse<FrameUploadResult>>

    @POST("api/v1/alarms")
    suspend fun uploadAlarm(
        @Body alarmRequest: AlarmUploadRequest
    ): Response<ApiResponse<AlarmUploadResult>>

    @POST("api/v1/alarms/batch")
    suspend fun uploadAlarmsBatch(
        @Body batchRequest: BatchAlarmUploadRequest
    ): Response<ApiResponse<BatchUploadResult>>

    @Multipart
    @POST("api/v1/alarms/video")
    suspend fun uploadAlarmVideo(
        @Part("alarmId") alarmId: RequestBody,
        @Part video: MultipartBody.Part
    ): Response<ApiResponse<VideoUploadResult>>

    @POST("api/v1/device/location")
    suspend fun reportLocation(
        @Body locationRequest: LocationReportRequest
    ): Response<ApiResponse<LocationReportResult>>

    @POST("api/v1/device/locations/batch")
    suspend fun reportLocationsBatch(
        @Body batchRequest: BatchLocationRequest
    ): Response<ApiResponse<BatchUploadResult>>

    @GET("api/v1/config/device")
    suspend fun fetchDeviceConfig(
        @Query("deviceId") deviceId: String
    ): Response<ApiResponse<DeviceConfig>>

    @POST("api/v1/device/heartbeat")
    suspend fun sendHeartbeat(
        @Body heartbeat: HeartbeatRequest
    ): Response<ApiResponse<HeartbeatResult>>
}

data class DetectionFrameUploadRequest(
    val deviceId: String,
    val timestamp: Long,
    val frameId: Long,
    val ear: Float,
    val mar: Float,
    val perclos: Float,
    val headPitch: Float,
    val headYaw: Float,
    val headRoll: Float,
    val gazeDirection: String,
    val gazeAngleDeviation: Float,
    val isYawning: Boolean,
    val isHeadDown: Boolean,
    val isGazeAway: Boolean,
    val fatigueScore: Int,
    val alarmLevel: Int,
    val gpsLatitude: Double?,
    val gpsLongitude: Double?,
    val gpsSpeed: Float?
) {
    companion object {
        fun fromResult(result: DetectionFrameResult, deviceId: String) = DetectionFrameUploadRequest(
            deviceId = deviceId,
            timestamp = result.timestamp,
            frameId = result.frameId,
            ear = result.ear,
            mar = result.mar,
            perclos = result.perclos,
            headPitch = result.headPitch,
            headYaw = result.headYaw,
            headRoll = result.headRoll,
            gazeDirection = result.gazeDirection.name,
            gazeAngleDeviation = result.gazeAngleDeviation,
            isYawning = result.isYawning,
            isHeadDown = result.isHeadDown,
            isGazeAway = result.isGazeAway,
            fatigueScore = result.fatigueScore,
            alarmLevel = result.alarmLevel.value,
            gpsLatitude = result.gpsLatitude,
            gpsLongitude = result.gpsLongitude,
            gpsSpeed = result.gpsSpeed
        )
    }
}

data class AlarmUploadRequest(
    val deviceId: String,
    val alarmId: String,
    val timestamp: Long,
    val level: Int,
    val score: Int,
    val triggers: List<String>,
    val ear: Float,
    val mar: Float,
    val perclos: Float,
    val headPitch: Float,
    val headYaw: Float,
    val gazeAngle: Float,
    val isYawning: Boolean,
    val gpsLatitude: Double?,
    val gpsLongitude: Double?,
    val gpsSpeed: Float?,
    val videoExists: Boolean
) {
    companion object {
        fun fromAlarm(alarm: FatigueAlarm, deviceId: String) = AlarmUploadRequest(
            deviceId = deviceId,
            alarmId = alarm.alarmId,
            timestamp = alarm.timestamp,
            level = alarm.level.value,
            score = alarm.score,
            triggers = alarm.triggers.map { it.name },
            ear = alarm.ear,
            mar = alarm.mar,
            perclos = alarm.perclos,
            headPitch = alarm.headPitch,
            headYaw = alarm.headYaw,
            gazeAngle = alarm.gazeAngle,
            isYawning = alarm.isYawning,
            gpsLatitude = alarm.gpsLatitude,
            gpsLongitude = alarm.gpsLongitude,
            gpsSpeed = alarm.gpsSpeed,
            videoExists = alarm.videoPath != null
        )

        fun fromStoredAlarm(stored: StoredAlarm, deviceId: String, triggers: List<String>) = AlarmUploadRequest(
            deviceId = deviceId,
            alarmId = stored.alarmId,
            timestamp = stored.timestamp,
            level = stored.level,
            score = stored.score,
            triggers = triggers,
            ear = stored.ear,
            mar = stored.mar,
            perclos = stored.perclos,
            headPitch = stored.headPitch,
            headYaw = stored.headYaw,
            gazeAngle = stored.gazeAngle,
            isYawning = stored.isYawning,
            gpsLatitude = stored.gpsLatitude,
            gpsLongitude = stored.gpsLongitude,
            gpsSpeed = stored.gpsSpeed,
            videoExists = stored.videoPath != null
        )
    }
}

data class BatchAlarmUploadRequest(
    val deviceId: String,
    val alarms: List<AlarmUploadRequest>
)

data class LocationReportRequest(
    val deviceId: String,
    val timestamp: Long,
    val latitude: Double,
    val longitude: Double,
    val altitude: Double?,
    val speed: Float?,
    val bearing: Float?,
    val accuracy: Float?,
    val provider: String?
)

data class BatchLocationRequest(
    val deviceId: String,
    val locations: List<LocationReportRequest>
)

data class DeviceConfig(
    val deviceId: String,
    val detectionIntervalMs: Int = 100,
    val earThreshold: Float = 0.25f,
    val marThreshold: Float = 0.6f,
    val perclosWarning: Float = 0.10f,
    val perclosDanger: Float = 0.20f,
    val headDownThreshold: Float = 15f,
    val gazeThreshold: Float = 15f,
    val alarmLevel1Score: Int = 80,
    val alarmLevel2Score: Int = 60,
    val alarmLevel3Score: Int = 40,
    val gpsReportIntervalMs: Int = 3000,
    val frameUploadEnabled: Boolean = false,
    val videoRecordEnabled: Boolean = true,
    val voiceAlertEnabled: Boolean = true,
    val vibrationEnabled: Boolean = true,
    val serverUrl: String = "https://api.ddg.example.com"
)

data class HeartbeatRequest(
    val deviceId: String,
    val timestamp: Long,
    val batteryLevel: Int?,
    val networkType: String?,
    val gpsStatus: Boolean,
    val detectionRunning: Boolean,
    val pendingAlarmCount: Int
)

data class ApiResponse<T>(
    val code: Int,
    val message: String,
    val data: T?,
    val success: Boolean = code == 200
)

data class FrameUploadResult(
    val received: Boolean,
    val frameId: Long
)

data class AlarmUploadResult(
    val alarmId: String,
    val received: Boolean,
    val serverAlarmId: String?
)

data class BatchUploadResult(
    val totalCount: Int,
    val successCount: Int,
    val failedIds: List<String>
)

data class VideoUploadResult(
    val alarmId: String,
    val videoUrl: String?,
    val received: Boolean
)

data class LocationReportResult(
    val received: Boolean
)

data class HeartbeatResult(
    val nextHeartbeatMs: Int,
    val configUpdated: Boolean,
    val updatedConfig: DeviceConfig?
)
