package com.ddg.driver.data.repository

import com.ddg.driver.data.model.CheckInRequest
import com.ddg.driver.data.model.CheckOutRequest
import com.ddg.driver.data.model.RecommendRequest
import com.ddg.driver.data.model.RestCountdown
import com.ddg.driver.data.model.ServiceAreaRecommendation
import com.ddg.driver.data.model.ServiceAreaReview
import com.ddg.driver.data.model.SubmitReviewRequest
import com.ddg.driver.data.remote.ApiService

class ServiceAreaRepository(private val apiService: ApiService) {

    suspend fun getRestCountdown(driverId: Long, vehicleId: Long): Result<RestCountdown> {
        return try {
            Result.success(apiService.getRestCountdown(driverId, vehicleId))
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun startDriving(driverId: Long, vehicleId: Long, waybillId: Long = 0): Result<Any> {
        return try {
            Result.success(apiService.startDriving(driverId, vehicleId, waybillId))
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun checkIn(request: CheckInRequest): Result<Any> {
        return try {
            Result.success(apiService.checkInServiceArea(request))
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun checkOut(request: CheckOutRequest): Result<Any> {
        return try {
            Result.success(apiService.checkOutServiceArea(request))
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun recommend(request: RecommendRequest): Result<ServiceAreaRecommendation> {
        return try {
            Result.success(apiService.recommendServiceAreas(request))
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun submitReview(request: SubmitReviewRequest): Result<ServiceAreaReview> {
        return try {
            Result.success(apiService.submitReview(request))
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun acceptRecommendation(id: Long): Result<Any> {
        return try {
            Result.success(apiService.acceptRecommendation(id))
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun rejectRecommendation(id: Long): Result<Any> {
        return try {
            Result.success(apiService.rejectRecommendation(id))
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
