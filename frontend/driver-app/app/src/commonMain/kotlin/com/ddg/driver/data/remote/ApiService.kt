package com.ddg.driver.data.remote

import com.ddg.driver.data.model.AlarmInfo
import com.ddg.driver.data.model.FatigueMetrics
import com.ddg.driver.data.model.LoginRequest
import com.ddg.driver.data.model.LoginResponse
import com.ddg.driver.data.model.MultiCameraDetectResponse
import com.ddg.driver.data.model.MultiCameraUploadRequest
import com.ddg.driver.data.model.ReplanConfirmRequest
import com.ddg.driver.data.model.RoutePlan
import com.ddg.driver.data.model.ServiceArea
import com.ddg.driver.data.model.Waybill
import com.ddg.driver.data.model.RestCountdown
import com.ddg.driver.data.model.CheckInRequest
import com.ddg.driver.data.model.CheckOutRequest
import com.ddg.driver.data.model.DrivingRestRecord
import com.ddg.driver.data.model.SubmitReviewRequest
import com.ddg.driver.data.model.ServiceAreaReview
import com.ddg.driver.data.model.RecommendRequest
import com.ddg.driver.data.model.ServiceAreaRecommendation
import com.ddg.driver.data.model.ServiceAreaRealtimeStatus
import io.ktor.client.HttpClient
import io.ktor.client.request.get
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.request.url

class ApiService(private val client: HttpClient) {

    suspend fun login(request: LoginRequest): LoginResponse {
        return client.post {
            url("${ApiClient.BASE_URL}/auth/login")
            setBody(request)
        }
    }

    suspend fun getCurrentWaybill(): Waybill? {
        return client.get {
            url("${ApiClient.BASE_URL}/waybills/current")
        }
    }

    suspend fun startNavigation(waybillId: String): RoutePlan {
        return client.post {
            url("${ApiClient.BASE_URL}/navigation/start/$waybillId")
        }
    }

    suspend fun getRoutePlan(routePlanId: Long): RoutePlan {
        return client.get {
            url("${ApiClient.BASE_URL}/route-plans/$routePlanId")
        }
    }

    suspend fun confirmReplan(replanId: Long, action: String, note: String? = null): Any {
        return client.post {
            url("${ApiClient.BASE_URL}/replans/$replanId/confirm")
            setBody(ReplanConfirmRequest(action = action, note = note))
        }
    }

    suspend fun listReplanHistory(waybillId: Long? = null, status: String? = null, page: Int = 1, pageSize: Int = 20): Any {
        return client.get {
            url("${ApiClient.BASE_URL}/replans")
            url {
                waybillId?.let { parameters.append("waybill_id", it.toString()) }
                status?.let { parameters.append("status", it) }
                parameters.append("page", page.toString())
                parameters.append("page_size", pageSize.toString())
            }
        }
    }

    suspend fun pullActiveRestrictedAreas(sinceVersion: Long? = null, hazardClass: String? = null): Any {
        return client.get {
            url("${ApiClient.BASE_URL}/restricted-areas/sync/pull")
            url {
                sinceVersion?.let { parameters.append("since_version", it.toString()) }
                hazardClass?.let { parameters.append("hazard_class", it) }
            }
        }
    }

    suspend fun uploadFatigueData(metrics: FatigueMetrics) {
        client.post<Unit> {
            url("${ApiClient.BASE_URL}/fatigue/upload")
            setBody(metrics)
        }
    }

    suspend fun uploadMultiCameraFrames(request: MultiCameraUploadRequest): MultiCameraDetectResponse {
        return client.post {
            url("${ApiClient.BASE_URL}/fatigue/upload/multi-camera")
            setBody(request)
        }
    }

    suspend fun uploadTrack(
        waybillId: String,
        lat: Double,
        lng: Double,
        speed: Double,
        timestamp: Long
    ) {
        client.post<Unit> {
            url("${ApiClient.BASE_URL}/track/upload")
            setBody(
                mapOf(
                    "waybillId" to waybillId,
                    "lat" to lat,
                    "lng" to lng,
                    "speed" to speed,
                    "timestamp" to timestamp
                )
            )
        }
    }

    suspend fun sosRequest(
        lat: Double,
        lng: Double,
        description: String?
    ): AlarmInfo {
        return client.post {
            url("${ApiClient.BASE_URL}/sos/request")
            setBody(
                mapOf(
                    "lat" to lat,
                    "lng" to lng,
                    "description" to (description ?: "")
                )
            )
        }
    }

    suspend fun listServiceAreas(
        lat: Double,
        lng: Double,
        radiusKm: Double = 50.0
    ): List<ServiceArea> {
        return client.get {
            url("${ApiClient.BASE_URL}/service-areas")
            url {
                parameters.append("lat", lat.toString())
                parameters.append("lng", lng.toString())
                parameters.append("radius", radiusKm.toString())
            }
        }
    }

    suspend fun getAlarmHistory(limit: Int = 50): List<AlarmInfo> {
        return client.get {
            url("${ApiClient.BASE_URL}/alarms/history")
            url {
                parameters.append("limit", limit.toString())
            }
        }
    }

    suspend fun getRestCountdown(driverId: Long, vehicleId: Long): RestCountdown {
        return client.get {
            url("${ApiClient.BASE_URL}/service-areas/rest/countdown")
            url {
                parameters.append("driver_id", driverId.toString())
                parameters.append("vehicle_id", vehicleId.toString())
            }
        }
    }

    suspend fun startDriving(driverId: Long, vehicleId: Long, waybillId: Long = 0): DrivingRestRecord {
        return client.post {
            url("${ApiClient.BASE_URL}/service-areas/rest/start")
            setBody(mapOf("driver_id" to driverId, "vehicle_id" to vehicleId, "waybill_id" to waybillId))
        }
    }

    suspend fun checkInServiceArea(request: CheckInRequest): DrivingRestRecord {
        return client.post {
            url("${ApiClient.BASE_URL}/service-areas/rest/check-in")
            setBody(request)
        }
    }

    suspend fun checkOutServiceArea(request: CheckOutRequest): DrivingRestRecord {
        return client.post {
            url("${ApiClient.BASE_URL}/service-areas/rest/check-out")
            setBody(request)
        }
    }

    suspend fun recommendServiceAreas(request: RecommendRequest): ServiceAreaRecommendation {
        return client.post {
            url("${ApiClient.BASE_URL}/service-areas/recommend")
            setBody(request)
        }
    }

    suspend fun submitReview(request: SubmitReviewRequest): ServiceAreaReview {
        return client.post {
            url("${ApiClient.BASE_URL}/service-areas/reviews")
            setBody(request)
        }
    }

    suspend fun getServiceAreaDetail(id: Long): Map<String, Any> {
        return client.get {
            url("${ApiClient.BASE_URL}/service-areas/$id")
        }
    }

    suspend fun acceptRecommendation(id: Long): Map<String, Any> {
        return client.post {
            url("${ApiClient.BASE_URL}/service-areas/recommendations/$id/accept")
        }
    }

    suspend fun rejectRecommendation(id: Long): Map<String, Any> {
        return client.post {
            url("${ApiClient.BASE_URL}/service-areas/recommendations/$id/reject")
        }
    }
}
