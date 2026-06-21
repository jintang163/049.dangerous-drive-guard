package com.ddg.driver.data.model

import kotlinx.serialization.Serializable

@Serializable
data class GeoPoint(
    val lat: Double = 0.0,
    val lng: Double = 0.0
)

@Serializable
data class GeoFenceCheckRequest(
    val vehicle_id: Long,
    val driver_id: Long? = null,
    val waybill_id: Long? = null,
    val latitude: Double,
    val longitude: Double,
    val address: String? = null,
    val threshold_meters: Int = 500
)

@Serializable
data class GeoFenceCheckResult(
    val alert_id: Long = 0,
    val alert_no: String? = null,
    val is_deviated: Boolean = false,
    val distance_from_route_meters: Double = 0.0,
    val threshold_meters: Int = 500,
    val alert_level: Int = 0,
    val daily_deviate_count: Int = 0,
    val auto_reported: Boolean = false,
    val status: String? = null,
    val nearest_route_point: GeoPoint? = null,
    val message: String? = null
)

@Serializable
data class GeoFenceConfirmRequest(
    val alert_id: Long,
    val confirm_type: String,
    val reason_detail: String? = null,
    val note: String? = null,
    val latitude: Double? = null,
    val longitude: Double? = null
)

@Serializable
data class GeoFenceAlertItem(
    val id: Long = 0,
    val alert_no: String? = null,
    val vehicle_id: Long = 0,
    val plate_number: String? = null,
    val driver_id: Long = 0,
    val driver_name: String? = null,
    val escort_id: Long = 0,
    val escort_name: String? = null,
    val waybill_id: Long = 0,
    val waybill_no: String? = null,
    val latitude: Double = 0.0,
    val longitude: Double = 0.0,
    val address: String? = null,
    val distance_from_route_meters: Double = 0.0,
    val threshold_meters: Int = 500,
    val alert_level: Int = 0,
    val status: String? = null,
    val deviate_reason: String? = null,
    val confirm_note: String? = null,
    val confirmed_at: String? = null,
    val reported_to_dispatch: Boolean = false,
    val reported_at: String? = null,
    val daily_deviate_count: Int = 0,
    val nearest_route_point: GeoPoint? = null,
    val auto_reported: Boolean = false,
    val created_at: String? = null
)
