package com.ddg.driver.domain.usecase

import com.ddg.driver.data.model.FatigueMetrics
import com.ddg.driver.data.repository.FatigueRepository

class ReportFatigueUseCase(
    private val fatigueRepository: FatigueRepository
) {
    suspend operator fun invoke(metrics: FatigueMetrics): Result<Unit> {
        return fatigueRepository.uploadFatigueData(metrics)
    }
}
