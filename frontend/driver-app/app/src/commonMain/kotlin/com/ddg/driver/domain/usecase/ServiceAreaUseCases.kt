package com.ddg.driver.domain.usecase

import com.ddg.driver.data.model.CheckInRequest
import com.ddg.driver.data.model.CheckOutRequest
import com.ddg.driver.data.model.RecommendRequest
import com.ddg.driver.data.model.RestCountdown
import com.ddg.driver.data.model.SubmitReviewRequest
import com.ddg.driver.data.repository.ServiceAreaRepository

class GetRestCountdownUseCase(private val repository: ServiceAreaRepository) {
    suspend operator fun invoke(driverId: Long, vehicleId: Long): Result<RestCountdown> {
        return repository.getRestCountdown(driverId, vehicleId)
    }
}

class CheckInUseCase(private val repository: ServiceAreaRepository) {
    suspend operator fun invoke(request: CheckInRequest): Result<Any> {
        return repository.checkIn(request)
    }
}

class CheckOutUseCase(private val repository: ServiceAreaRepository) {
    suspend operator fun invoke(request: CheckOutRequest): Result<Any> {
        return repository.checkOut(request)
    }
}

class RecommendServiceAreaUseCase(private val repository: ServiceAreaRepository) {
    suspend operator fun invoke(request: RecommendRequest): Result<Any> {
        return repository.recommend(request)
    }
}

class SubmitServiceAreaReviewUseCase(private val repository: ServiceAreaRepository) {
    suspend operator fun invoke(request: SubmitReviewRequest): Result<Any> {
        return repository.submitReview(request)
    }
}
