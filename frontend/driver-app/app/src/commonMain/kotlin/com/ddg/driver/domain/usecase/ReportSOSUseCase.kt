package com.ddg.driver.domain.usecase

import com.ddg.driver.data.model.AlarmInfo
import com.ddg.driver.data.repository.FatigueRepository

class ReportSOSUseCase(
    private val fatigueRepository: FatigueRepository
) {
    suspend operator fun invoke(
        lat: Double,
        lng: Double,
        description: String? = null
    ): Result<AlarmInfo> {
        return fatigueRepository.reportSOS(lat, lng, description)
    }
}
