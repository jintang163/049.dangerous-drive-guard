package com.ddg.edge.fatigue.data.remote

import android.content.Context
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import com.ddg.edge.fatigue.data.local.AlarmDatabase
import com.ddg.edge.fatigue.data.local.OfflineAlarmRepository
import kotlinx.coroutines.Dispatchers

class AlarmUploadWorker(
    context: Context,
    params: WorkerParameters
) : CoroutineWorker(context, params) {

    override suspend fun doWork(): Result {
        return runCatching {
            val database = AlarmDatabase.getDatabase(applicationContext)
            val repository = OfflineAlarmRepository(database.alarmDao(), Dispatchers.IO)

            val baseUrl = inputData.getString(KEY_BASE_URL) ?: DEFAULT_BASE_URL
            val apiService = AlarmUploadWorkerDelegate.createApiService(baseUrl)

            val delegate = AlarmUploadWorkerDelegate(
                context = applicationContext,
                alarmRepository = repository,
                apiService = apiService
            )

            val result = delegate.executeUpload()

            if (result.success || result.uploaded > 0) {
                Result.success()
            } else {
                Result.retry()
            }
        }.getOrElse {
            Result.retry()
        }
    }

    companion object {
        const val KEY_BASE_URL = "base_url"
        const val DEFAULT_BASE_URL = "https://api.ddg.example.com/"
        const val WORK_NAME = "AlarmUploadWork"
    }
}
