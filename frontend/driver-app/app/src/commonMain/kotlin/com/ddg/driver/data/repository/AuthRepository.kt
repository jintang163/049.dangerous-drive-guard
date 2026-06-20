package com.ddg.driver.data.repository

import com.ddg.driver.data.local.AppDataStore
import com.ddg.driver.data.model.LoginRequest
import com.ddg.driver.data.model.LoginResponse
import com.ddg.driver.data.remote.ApiService
import kotlinx.coroutines.flow.Flow

class AuthRepository(
    private val apiService: ApiService,
    private val dataStore: AppDataStore
) {

    val isLoggedIn: Flow<Boolean> = dataStore.isLoggedIn

    suspend fun login(phone: String, password: String): Result<LoginResponse> {
        return try {
            val request = LoginRequest(phone, password)
            val response = apiService.login(request)
            dataStore.saveToken(response.token)
            dataStore.saveUser(response.user)
            Result.success(response)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun logout() {
        dataStore.clearAll()
    }
}
