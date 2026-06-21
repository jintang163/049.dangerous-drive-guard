package com.ddg.driver.data.repository

import com.ddg.driver.data.model.ReplanAppliedInfo
import com.ddg.driver.data.model.ReplanSuggestion
import com.ddg.driver.data.remote.ApiService
import com.ddg.driver.data.remote.WebSocketClient
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.SharedFlow

class ReplanRepository(
    private val apiService: ApiService,
    private val wsClient: WebSocketClient
) {
    fun connectWs(vehicleId: Long? = null, driverId: Long? = null) {
        wsClient.connect(vehicleId, driverId)
    }

    fun disconnectWs() {
        wsClient.disconnect()
    }

    fun observeReplanSuggestions(): Flow<ReplanSuggestion> {
        return wsClient.replanSuggestions
    }

    fun observeRouteApplied(): Flow<ReplanAppliedInfo> {
        return wsClient.routeApplied
    }

    fun observeTrafficEvents(): SharedFlow<String> {
        return wsClient.trafficEvents
    }

    suspend fun confirmReplan(replanId: Long, action: String, note: String? = null): Result<Unit> {
        return runCatching {
            apiService.confirmReplan(replanId, action, note)
            Unit
        }
    }

    suspend fun getRoutePlan(routePlanId: Long) = runCatching {
        apiService.getRoutePlan(routePlanId)
    }
}
