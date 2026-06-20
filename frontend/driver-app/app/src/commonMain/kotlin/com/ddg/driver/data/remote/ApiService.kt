package com.ddg.driver.data.remote

import com.ddg.driver.data.model.AlarmInfo
import com.ddg.driver.data.model.FatigueMetrics
import com.ddg.driver.data.model.LoginRequest
import com.ddg.driver.data.model.LoginResponse
import com.ddg.driver.data.model.RoutePlan
import com.ddg.driver.data.model.ServiceArea
import com.ddg.driver.data.model.Waybill
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

    suspend fun uploadFatigueData(metrics: FatigueMetrics) {
        client.post<Unit> {
            url("${ApiClient.BASE_URL}/fatigue/upload")
            setBody(metrics)
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
}
