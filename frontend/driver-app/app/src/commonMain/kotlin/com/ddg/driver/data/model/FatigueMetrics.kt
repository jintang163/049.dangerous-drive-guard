package com.ddg.driver.data.model

import kotlinx.serialization.Serializable

@Serializable
data class FatigueMetrics(
    val driverId: String,
    val eyeClosureRate: Double,
    val blinkFrequency: Int,
    val yawnCount: Int,
    val headTiltAngle: Double,
    val focusScore: Double,
    val fatigueLevel: FatigueLevel,
    val drivingDuration: Long,
    val continuousDrivingMinutes: Int,
    val restBreakMinutes: Int,
    val lastDetectTime: Long
)

@Serializable
enum class FatigueLevel {
    NORMAL,
    WARNING,
    DANGEROUS
}
