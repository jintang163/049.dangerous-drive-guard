package com.ddg.driver.data.local

import com.ddg.driver.data.model.UserInfo
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.map

class AppDataStore {

    private val _token = MutableStateFlow<String?>(null)
    val token: Flow<String?> = _token

    private val _user = MutableStateFlow<UserInfo?>(null)
    val user: Flow<UserInfo?> = _user

    val isLoggedIn: Flow<Boolean> = _token.map { !it.isNullOrEmpty() }

    suspend fun saveToken(token: String) {
        _token.value = token
    }

    suspend fun saveUser(user: UserInfo) {
        _user.value = user
    }

    suspend fun clearAll() {
        _token.value = null
        _user.value = null
    }
}
