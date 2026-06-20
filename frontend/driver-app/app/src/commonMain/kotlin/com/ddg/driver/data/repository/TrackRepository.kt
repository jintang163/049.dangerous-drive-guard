package com.ddg.driver.data.repository

import com.ddg.driver.data.remote.ApiService

class TrackRepository(private val apiService: ApiService) {

    suspend fun uploadTrack(
        waybillId: String,
        lat: Double,
        lng: Double,
        speed: Double,
        timestamp: Long
    ): Result<Unit> {
        return try {
            apiService.uploadTrack(waybillId, lat, lng, speed, timestamp)
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
