package com.ddg.driver.data.model

import kotlinx.serialization.Serializable

@Serializable
data class ReplanCandidate(
    val id: Long? = null,
    val strategy: String,
    val total_distance: Double,
    val estimated_duration: Int,
    val toll_fee: Double? = 0.0,
    val safety_score: Int? = 0,
    val is_recommended: Int = 0
)

@Serializable
data class TrafficEventInfo(
    val id: Long,
    val type: String,
    val level: Int,
    val title: String,
    val road_name: String? = null,
    val center_lat: Double,
    val center_lng: Double
)

@Serializable
data class ReplanSuggestion(
    val replan_id: Long,
    val replan_no: String,
    val waybill_id: Long,
    val waybill_no: String,
    val vehicle_id: Long? = null,
    val vehicle_plate: String? = null,
    val driver_id: Long? = null,
    val driver_name: String? = null,
    val trigger_type: String,
    val trigger_reason: String,
    val original_distance_remaining: Double,
    val original_duration_remaining: Int,
    val new_distance_remaining: Double,
    val new_duration_remaining: Int,
    val distance_delta: Double,
    val duration_delta: Int,
    val current_lat: Double? = null,
    val current_lng: Double? = null,
    val candidates: List<ReplanCandidate> = emptyList(),
    val status: String,
    val created_at: String,
    val traffic_event: TrafficEventInfo? = null
)

@Serializable
data class ReplanAppliedInfo(
    val replan_id: Long,
    val replan_no: String,
    val new_route_plan_id: Long,
    val waybill_id: Long,
    val applied_at: String
)

@Serializable
data class ReplanConfirmRequest(
    val action: String,
    val note: String? = null
)
