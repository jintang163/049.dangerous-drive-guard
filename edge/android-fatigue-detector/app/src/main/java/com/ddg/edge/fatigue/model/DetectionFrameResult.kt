package com.ddg.edge.fatigue.model

import com.google.mediapipe.tasks.components.containers.NormalizedLandmark

data class DetectionFrameResult(
    val timestamp: Long,
    val frameId: Long,
    val landmarks: List<NormalizedLandmark>?,
    val ear: Float,
    val mar: Float,
    val perclos: Float,
    val headPitch: Float,
    val headYaw: Float,
    val headRoll: Float,
    val gazeDirection: GazeDirection,
    val gazeAngleDeviation: Float,
    val isYawning: Boolean,
    val isHeadDown: Boolean,
    val isGazeAway: Boolean,
    val fatigueScore: Int,
    val alarmLevel: AlarmLevel,
    val gpsLatitude: Double?,
    val gpsLongitude: Double?,
    val gpsSpeed: Float?
)

enum class GazeDirection {
    CENTER, LEFT, RIGHT, UP, DOWN, UNKNOWN
}

enum class AlarmLevel(val value: Int, val description: String) {
    NONE(0, "正常"),
    LEVEL_1(1, "轻度疲劳"),
    LEVEL_2(2, "中度疲劳"),
    LEVEL_3(3, "重度疲劳");

    companion object {
        fun fromValue(value: Int): AlarmLevel = values().firstOrNull { it.value == value } ?: NONE
    }
}
