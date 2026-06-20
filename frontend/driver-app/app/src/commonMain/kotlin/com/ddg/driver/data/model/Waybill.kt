package com.ddg.driver.data.model

import kotlinx.serialization.Serializable

@Serializable
data class Waybill(
    val id: String,
    val waybillNo: String,
    val cargoName: String,
    val cargoType: String,
    val cargoWeight: Double,
    val originAddress: String,
    val originLat: Double,
    val originLng: Double,
    val destAddress: String,
    val destLat: Double,
    val destLng: Double,
    val vehiclePlate: String,
    val driverName: String,
    val driverPhone: String,
    val escortName: String,
    val escortPhone: String,
    val status: WaybillStatus,
    val progress: Int,
    val distanceTotal: Double,
    val distanceCompleted: Double,
    val startTime: Long,
    val estimatedArrivalTime: Long
)

@Serializable
enum class WaybillStatus {
    PENDING,
    IN_TRANSIT,
    STOPPED,
    COMPLETED,
    CANCELLED
}
