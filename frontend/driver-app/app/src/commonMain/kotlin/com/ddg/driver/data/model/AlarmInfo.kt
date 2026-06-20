package com.ddg.driver.data.model

import kotlinx.serialization.Serializable

@Serializable
data class AlarmInfo(
    val id: String,
    val type: AlarmType,
    val level: AlarmLevel,
    val title: String,
    val description: String,
    val timestamp: Long,
    val status: AlarmStatus,
    val locationLat: Double,
    val locationLng: Double,
    val locationAddress: String,
    val waybillId: String?,
    val imageUrl: String?
)

@Serializable
enum class AlarmType {
    FATIGUE,
    OVERSPEED,
    LANE_DEPARTURE,
    COLLISION_WARNING,
    ABNORMAL_STOP,
    SOS,
    VEHICLE_ABNORMAL,
    CARGO_ABNORMAL
}

@Serializable
enum class AlarmLevel {
    LOW,
    MEDIUM,
    HIGH,
    CRITICAL
}

@Serializable
enum class AlarmStatus {
    UNHANDLED,
    PROCESSING,
    RESOLVED,
    IGNORED
}
