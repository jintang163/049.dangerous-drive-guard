package com.ddg.edge.fatigue.model

data class FatigueAlarm(
    val alarmId: String,
    val timestamp: Long,
    val level: AlarmLevel,
    val score: Int,
    val triggers: Set<AlarmTrigger>,
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
    val videoPath: String? = null,
    val isUploaded: Boolean = false
)

enum class AlarmTrigger {
    PERCLOS_EXCEEDED,
    MAR_EXCEEDED,
    HEAD_DOWN,
    GAZE_AWAY,
    YAWNING,
    SCORE_LOW
}
