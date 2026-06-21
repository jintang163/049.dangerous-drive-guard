package com.ddg.driver.data.model

import kotlinx.serialization.Serializable

@Serializable
data class RestCountdown(
    val driver_id: Long = 0,
    val vehicle_id: Long = 0,
    val waybill_id: Long = 0,
    val status: String = "driving",
    val continuous_drive_minutes: Int = 0,
    val remaining_drive_minutes: Int = 240,
    val max_continuous_drive: Int = 240,
    val is_overtime: Boolean = false,
    val overtime_minutes: Int = 0,
    val min_rest_required: Int = 20,
    val current_rest_minutes: Int = 0,
    val rest_progress_percent: Double = 0.0,
    val can_continue_driving: Boolean = true,
    val current_service_area_id: Long = 0,
    val current_service_area_name: String = ""
)

@Serializable
data class CheckInRequest(
    val driver_id: Long,
    val vehicle_id: Long,
    val service_area_id: Long,
    val latitude: Double,
    val longitude: Double,
    val waybill_id: Long = 0
)

@Serializable
data class CheckOutRequest(
    val driver_id: Long,
    val vehicle_id: Long,
    val latitude: Double = 0.0,
    val longitude: Double = 0.0
)

@Serializable
data class DrivingRestRecord(
    val id: Long = 0,
    val driver_id: Long = 0,
    val vehicle_id: Long = 0,
    val waybill_id: Long = 0,
    val record_date: String = "",
    val drive_start_time: String = "",
    val drive_end_time: String? = null,
    val continuous_drive_minutes: Int = 0,
    val rest_start_time: String? = null,
    val rest_end_time: String? = null,
    val rest_duration_minutes: Int = 0,
    val rest_service_area_id: Long = 0,
    val rest_service_area_name: String = "",
    val status: String = "driving",
    val is_overtime: Boolean = false,
    val overtime_minutes: Int = 0,
    val min_rest_required: Int = 20,
    val max_continuous_drive: Int = 240
)

@Serializable
data class SubmitReviewRequest(
    val service_area_id: Long,
    val driver_id: Long,
    val security_score: Int,
    val environment_score: Int = 0,
    val food_score: Int = 0,
    val service_score: Int = 0,
    val comment_text: String = "",
    val tags: List<String> = emptyList(),
    val is_anonymous: Boolean = false,
    val waybill_id: Long = 0,
    val vehicle_id: Long = 0
)

@Serializable
data class ServiceAreaReview(
    val id: Long = 0,
    val service_area_id: Long = 0,
    val driver_id: Long = 0,
    val driver_name: String = "",
    val security_score: Int = 0,
    val environment_score: Int = 0,
    val food_score: Int = 0,
    val service_score: Int = 0,
    val overall_score: Double = 0.0,
    val comment_text: String = "",
    val tags_array: List<String> = emptyList(),
    val is_anonymous: Boolean = false,
    val created_at: String = ""
)

@Serializable
data class RecommendRequest(
    val driver_id: Long,
    val vehicle_id: Long,
    val waybill_id: Long = 0,
    val latitude: Double,
    val longitude: Double,
    val hazard_class: String = "",
    val fatigue_score: Double = 0.0,
    val radius_km: Double = 100.0
)

@Serializable
data class ServiceAreaRecommendation(
    val id: Long = 0,
    val recommend_no: String = "",
    val driver_id: Long = 0,
    val vehicle_id: Long = 0,
    val continuous_drive_minutes: Int = 0,
    val remaining_drive_minutes: Int = 0,
    val recommend_reason: String = "",
    val recommended_service_area_id: Long = 0,
    val recommended_service_area_name: String = "",
    val distance_km: Double = 0.0,
    val estimated_arrival_minutes: Int = 0,
    val alternatives_array: List<RecommendedServiceAreaItem> = emptyList(),
    val status: String = "pending"
)

@Serializable
data class RecommendedServiceAreaItem(
    val service_area_id: Long = 0,
    val service_area_name: String = "",
    val distance_km: Double = 0.0,
    val estimated_arrival_minutes: Int = 0,
    val available_danger_spaces: Int = 0,
    val security_level: Int = 0,
    val restaurant_rating: Double = 0.0,
    val has_fuel: Boolean = false,
    val has_charging: Boolean = false,
    val match_score: Double = 0.0
)

@Serializable
data class ServiceAreaRealtimeStatus(
    val id: Long = 0,
    val service_area_id: Long = 0,
    val total_parking_spaces: Int = 0,
    val available_parking_spaces: Int = 0,
    val total_danger_spaces: Int = 0,
    val available_danger_spaces: Int = 0,
    val has_fuel: Boolean = false,
    val fuel_price_diesel: Double = 0.0,
    val has_charging: Boolean = false,
    val charging_piles_available: Int = 0,
    val charging_piles_total: Int = 0,
    val has_restaurant: Boolean = false,
    val restaurant_rating: Double = 0.0,
    val restaurant_wait_minutes: Int = 0,
    val security_level: Int = 3,
    val security_patrol_interval: Int = 30,
    val crowd_level: Int = 2,
    val weather_condition: String = "",
    val update_time: String = "",
    val data_source: String = "manual"
)
