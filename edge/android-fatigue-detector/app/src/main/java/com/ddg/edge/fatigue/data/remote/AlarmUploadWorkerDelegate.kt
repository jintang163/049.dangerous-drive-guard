package com.ddg.edge.fatigue.data.remote

import android.content.Context
import android.os.Build
import com.ddg.edge.fatigue.data.local.OfflineAlarmRepository
import com.ddg.edge.fatigue.model.StoredAlarm
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import okhttp3.OkHttpClient
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import java.util.UUID
import java.util.concurrent.TimeUnit

class AlarmUploadWorkerDelegate(
    private val context: Context,
    private val alarmRepository: OfflineAlarmRepository,
    private val apiService: PlatformApiService,
    private val batchSize: Int = 50,
    private val maxRetryCount: Int = 5
) {

    suspend fun executeUpload(): UploadResult = withContext(Dispatchers.IO) {
        val deviceId = getDeviceId()
        val pendingAlarms = alarmRepository.observePendingAlarms().first()

        if (pendingAlarms.isEmpty()) {
            return@withContext UploadResult(success = true, total = 0, uploaded = 0, failed = 0)
        }

        val alarmsToUpload = pendingAlarms.filter { it.uploadRetryCount < maxRetryCount }
        val batches = alarmsToUpload.chunked(batchSize)

        var totalUploaded = 0
        var totalFailed = 0
        val uploadedIds = mutableListOf<Long>()
        val failedIds = mutableListOf<Long>()

        coroutineScope {
            batches.forEach { batch ->
                launch {
                    val result = uploadBatch(batch, deviceId)
                    if (result.success) {
                        totalUploaded += result.uploaded
                        uploadedIds.addAll(result.uploadedDbIds)
                    } else {
                        totalFailed += result.failed
                        failedIds.addAll(result.failedDbIds)
                    }
                }
            }
        }

        if (uploadedIds.isNotEmpty()) {
            alarmRepository.markAsUploaded(uploadedIds)
        }

        failedIds.forEach { id ->
            alarmRepository.incrementRetryCount(id)
        }

        UploadResult(
            success = totalFailed == 0,
            total = alarmsToUpload.size,
            uploaded = totalUploaded,
            failed = totalFailed
        )
    }

    private suspend fun uploadBatch(
        batch: List<StoredAlarm>,
        deviceId: String
    ): BatchResult {
        val uploadedIds = mutableListOf<Long>()
        val failedIds = mutableListOf<Long>()

        try {
            val requests = batch.map { stored ->
                val triggers = alarmRepository.jsonToTriggers(stored.triggersJson)
                    .map { it.name }
                AlarmUploadRequest.fromStoredAlarm(stored, deviceId, triggers)
            }

            val response = apiService.uploadAlarmsBatch(
                BatchAlarmUploadRequest(deviceId = deviceId, alarms = requests)
            )

            if (response.isSuccessful && response.body()?.success == true) {
                val data = response.body()?.data
                val failedServerIds = data?.failedIds ?: emptyList()

                batch.forEach { stored ->
                    if (stored.alarmId in failedServerIds) {
                        failedIds.add(stored.id)
                    } else {
                        uploadedIds.add(stored.id)
                    }
                }
            } else {
                failedIds.addAll(batch.map { it.id })
            }
        } catch (e: Exception) {
            failedIds.addAll(batch.map { it.id })
        }

        return BatchResult(
            success = failedIds.isEmpty(),
            uploaded = uploadedIds.size,
            failed = failedIds.size,
            uploadedDbIds = uploadedIds,
            failedDbIds = failedIds
        )
    }

    private fun getDeviceId(): String {
        val prefs = context.getSharedPreferences("device_prefs", Context.MODE_PRIVATE)
        var deviceId = prefs.getString("device_id", null)
        if (deviceId == null) {
            deviceId = UUID.randomUUID().toString()
            prefs.edit().putString("device_id", deviceId).apply()
        }
        return deviceId
    }

    data class UploadResult(
        val success: Boolean,
        val total: Int,
        val uploaded: Int,
        val failed: Int
    )

    private data class BatchResult(
        val success: Boolean,
        val uploaded: Int,
        val failed: Int,
        val uploadedDbIds: List<Long>,
        val failedDbIds: List<Long>
    )

    companion object {
        fun createApiService(baseUrl: String): PlatformApiService {
            val client = OkHttpClient.Builder()
                .connectTimeout(30, TimeUnit.SECONDS)
                .readTimeout(30, TimeUnit.SECONDS)
                .writeTimeout(30, TimeUnit.SECONDS)
                .retryOnConnectionFailure(true)
                .addInterceptor { chain ->
                    val request = chain.request().newBuilder()
                        .header("User-Agent", "DDG-Edge/${Build.VERSION.RELEASE}")
                        .header("X-Client-Type", "ANDROID_EDGE")
                        .build()
                    chain.proceed(request)
                }
                .build()

            return Retrofit.Builder()
                .baseUrl(baseUrl)
                .client(client)
                .addConverterFactory(GsonConverterFactory.create())
                .build()
                .create(PlatformApiService::class.java)
        }
    }
}
