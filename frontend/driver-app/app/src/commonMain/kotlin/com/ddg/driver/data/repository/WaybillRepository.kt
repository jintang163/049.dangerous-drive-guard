package com.ddg.driver.data.repository

import com.ddg.driver.data.model.RoutePlan
import com.ddg.driver.data.model.ServiceArea
import com.ddg.driver.data.model.Waybill
import com.ddg.driver.data.remote.ApiService

class WaybillRepository(private val apiService: ApiService) {

    suspend fun getCurrentWaybill(): Result<Waybill?> {
        return try {
            val waybill = apiService.getCurrentWaybill()
            Result.success(waybill)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun startNavigation(waybillId: String): Result<RoutePlan> {
        return try {
            val routePlan = apiService.startNavigation(waybillId)
            Result.success(routePlan)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun listServiceAreas(
        lat: Double,
        lng: Double,
        radiusKm: Double
    ): Result<List<ServiceArea>> {
        return try {
            val areas = apiService.listServiceAreas(lat, lng, radiusKm)
            Result.success(areas)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
