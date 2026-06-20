package com.ddg.driver.data.model

import kotlinx.serialization.Serializable

@Serializable
data class RoutePlan(
    val id: String,
    val waybillId: String,
    val originLat: Double,
    val originLng: Double,
    val destLat: Double,
    val destLng: Double,
    val waypoints: List<Waypoint>,
    val totalDistance: Double,
    val estimatedDuration: Long,
    val serviceAreas: List<ServiceArea>,
    val riskPoints: List<RiskPoint>
)

@Serializable
data class Waypoint(
    val lat: Double,
    val lng: Double,
    val instruction: String,
    val distanceFromStart: Double
)

@Serializable
data class ServiceArea(
    val id: String,
    val name: String,
    val lat: Double,
    val lng: Double,
    val distanceFromOrigin: Double,
    val hasRestRoom: Boolean,
    val hasFuelStation: Boolean,
    val hasRestaurant: Boolean,
    val hasParking: Boolean,
    val openHours: String
)

@Serializable
data class RiskPoint(
    val lat: Double,
    val lng: Double,
    val type: RiskType,
    val description: String,
    val suggestedSpeed: Double
)

@Serializable
enum class RiskType {
    CONGESTION,
    ACCIDENT_PRONE,
    CONSTRUCTION,
    WEATHER,
    WEIGH_STATION,
    INSPECTION
}
