package com.ddg.driver.data.repository

import com.ddg.driver.data.model.AlarmInfo
import com.ddg.driver.data.model.FatigueMetrics
import com.ddg.driver.data.remote.ApiService

class FatigueRepository(private val apiService: ApiService) {

    suspend fun uploadFatigueData(metrics: FatigueMetrics): Result<Unit> {
        return try {
            apiService.uploadFatigueData(metrics)
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun reportSOS(
        lat: Double,
        lng: Double,
        description: String?
    ): Result<AlarmInfo> {
        return try {
            val alarm = apiService.sosRequest(lat, lng, description)
            Result.success(alarm)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getAlarmHistory(limit: Int = 50): Result<List<AlarmInfo>> {
        return try {
            val alarms = apiService.getAlarmHistory(limit)
            Result.success(alarms)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
