package com.ddg.driver.domain.usecase

import com.ddg.driver.data.model.MultiCameraDetectResponse
import com.ddg.driver.data.model.MultiCameraUploadRequest
import com.ddg.driver.data.repository.FatigueRepository

class ReportMultiCameraFatigueUseCase(
    private val fatigueRepository: FatigueRepository
) {
    suspend operator fun invoke(request: MultiCameraUploadRequest): Result<MultiCameraDetectResponse> {
        return fatigueRepository.uploadMultiCameraFrames(request)
    }
}
