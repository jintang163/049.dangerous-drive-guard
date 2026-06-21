package com.ddg.driver.di

import com.ddg.driver.data.local.AppDataStore
import com.ddg.driver.data.remote.ApiClient
import com.ddg.driver.data.remote.ApiService
import com.ddg.driver.data.remote.WebSocketClient
import com.ddg.driver.data.repository.AuthRepository
import com.ddg.driver.data.repository.FatigueRepository
import com.ddg.driver.data.repository.ReplanRepository
import com.ddg.driver.data.repository.ServiceAreaRepository
import com.ddg.driver.data.repository.TrackRepository
import com.ddg.driver.data.repository.WaybillRepository
import com.ddg.driver.domain.usecase.CheckInUseCase
import com.ddg.driver.domain.usecase.CheckOutUseCase
import com.ddg.driver.domain.usecase.GetCurrentWaybillUseCase
import com.ddg.driver.domain.usecase.GetRestCountdownUseCase
import com.ddg.driver.domain.usecase.LoginUseCase
import com.ddg.driver.domain.usecase.RecommendServiceAreaUseCase
import com.ddg.driver.domain.usecase.ReportFatigueUseCase
import com.ddg.driver.domain.usecase.ReportMultiCameraFatigueUseCase
import com.ddg.driver.domain.usecase.ReportSOSUseCase
import com.ddg.driver.domain.usecase.SubmitServiceAreaReviewUseCase
import org.koin.dsl.module

val appModule = module {
    single { AppDataStore() }
    single { ApiClient(get()) }
    single { ApiService(get<ApiClient>().client) }
    single { WebSocketClient(get<ApiClient>().wsClient, get()) }

    single { AuthRepository(get(), get()) }
    single { WaybillRepository(get()) }
    single { FatigueRepository(get()) }
    single { TrackRepository(get()) }
    single { ReplanRepository(get(), get()) }
    single { ServiceAreaRepository(get()) }

    factory { LoginUseCase(get()) }
    factory { GetCurrentWaybillUseCase(get()) }
    factory { ReportFatigueUseCase(get()) }
    factory { ReportMultiCameraFatigueUseCase(get()) }
    factory { ReportSOSUseCase(get()) }
    factory { GetRestCountdownUseCase(get()) }
    factory { CheckInUseCase(get()) }
    factory { CheckOutUseCase(get()) }
    factory { RecommendServiceAreaUseCase(get()) }
    factory { SubmitServiceAreaReviewUseCase(get()) }
}
