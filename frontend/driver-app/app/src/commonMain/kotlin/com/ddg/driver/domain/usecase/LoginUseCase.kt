package com.ddg.driver.domain.usecase

import com.ddg.driver.data.model.LoginResponse
import com.ddg.driver.data.repository.AuthRepository

class LoginUseCase(
    private val authRepository: AuthRepository
) {
    suspend operator fun invoke(phone: String, password: String): Result<LoginResponse> {
        require(phone.isNotBlank()) { "手机号不能为空" }
        require(password.isNotBlank()) { "密码不能为空" }
        return authRepository.login(phone, password)
    }
}
