package com.ddg.driver.data.model

import kotlinx.serialization.Serializable

@Serializable
data class UserInfo(
    val id: String,
    val phone: String,
    val name: String,
    val avatar: String,
    val idCard: String,
    val driverLicense: String,
    val licenseExpireDate: Long,
    val qualificationCert: String,
    val rating: Double,
    val totalTrips: Int,
    val totalDistance: Double,
    val safetyScore: Int,
    val companyName: String,
    val department: String,
    val emergencyContact: String,
    val emergencyPhone: String,
    val joinDate: Long,
    val status: DriverStatus
)

@Serializable
enum class DriverStatus {
    ON_DUTY,
    OFF_DUTY,
    RESTING,
    SUSPENDED
}

@Serializable
data class LoginRequest(
    val phone: String,
    val password: String
)

@Serializable
data class LoginResponse(
    val token: String,
    val refreshToken: String,
    val expiresIn: Long,
    val user: UserInfo
)
