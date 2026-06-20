package com.ddg.driver.di

import com.ddg.driver.data.local.AppDataStore
import com.ddg.driver.data.remote.ApiClient
import com.ddg.driver.data.remote.ApiService
import com.ddg.driver.data.repository.AuthRepository
import com.ddg.driver.data.repository.FatigueRepository
import com.ddg.driver.data.repository.TrackRepository
import com.ddg.driver.data.repository.WaybillRepository
import com.ddg.driver.domain.usecase.GetCurrentWaybillUseCase
import com.ddg.driver.domain.usecase.LoginUseCase
import com.ddg.driver.domain.usecase.ReportFatigueUseCase
import com.ddg.driver.domain.usecase.ReportSOSUseCase
import org.koin.dsl.module

val appModule = module {
    single { AppDataStore() }
    single { ApiClient(get()) }
    single { ApiService(get()) }

    single { AuthRepository(get(), get()) }
    single { WaybillRepository(get()) }
    single { FatigueRepository(get()) }
    single { TrackRepository(get()) }

    factory { LoginUseCase(get()) }
    factory { GetCurrentWaybillUseCase(get()) }
    factory { ReportFatigueUseCase(get()) }
    factory { ReportSOSUseCase(get()) }
}
