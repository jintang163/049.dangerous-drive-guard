package com.ddg.driver.data.model

import kotlinx.serialization.Serializable

@Serializable
data class VehicleStatus(
    val vehicleId: String,
    val plate: String,
    val speed: Double,
    val rpm: Int,
    val fuelLevel: Double,
    val engineTemp: Double,
    val tirePressure: TirePressure,
    val brakeStatus: BrakeStatus,
    val locationLat: Double,
    val locationLng: Double,
    val heading: Double,
    val lastUpdateTime: Long
)

@Serializable
data class TirePressure(
    val frontLeft: Double,
    val frontRight: Double,
    val rearLeft: Double,
    val rearRight: Double
)

@Serializable
enum class BrakeStatus {
    NORMAL,
    ABNORMAL,
    WARNING
}
