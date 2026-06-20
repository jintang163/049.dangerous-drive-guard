package com.ddg.driver.domain.usecase

import com.ddg.driver.data.model.Waybill
import com.ddg.driver.data.repository.WaybillRepository

class GetCurrentWaybillUseCase(
    private val waybillRepository: WaybillRepository
) {
    suspend operator fun invoke(): Result<Waybill?> {
        return waybillRepository.getCurrentWaybill()
    }
}
