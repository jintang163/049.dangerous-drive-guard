package com.ddg.edge.fatigue.data.local

import com.ddg.edge.fatigue.model.AlarmTrigger
import com.ddg.edge.fatigue.model.FatigueAlarm
import com.ddg.edge.fatigue.model.StoredAlarm
import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flowOn
import kotlinx.coroutines.withContext

class OfflineAlarmRepository(
    private val alarmDao: AlarmDao,
    private val ioDispatcher: CoroutineDispatcher
) {

    suspend fun saveAlarm(alarm: FatigueAlarm): Long = withContext(ioDispatcher) {
        val stored = toStoredAlarm(alarm)
        alarmDao.insertAlarm(stored)
    }

    suspend fun getPendingAlarms(): List<StoredAlarm> = withContext(ioDispatcher) {
        alarmDao.getPendingAlarms()
    }

    fun observePendingAlarms(): Flow<List<StoredAlarm>> {
        return alarmDao.observePendingAlarms().flowOn(ioDispatcher)
    }

    suspend fun getRecentAlarms(limit: Int = 100): List<StoredAlarm> = withContext(ioDispatcher) {
        alarmDao.getRecentAlarms(limit)
    }

    fun observeAllAlarms(): Flow<List<StoredAlarm>> {
        return alarmDao.observeAllAlarms().flowOn(ioDispatcher)
    }

    suspend fun getPendingAlarmCount(): Int = withContext(ioDispatcher) {
        alarmDao.getPendingAlarmCount()
    }

    suspend fun markAsUploaded(ids: List<Long>) = withContext(ioDispatcher) {
        alarmDao.markAsUploaded(ids)
    }

    suspend fun markAsUploaded(alarms: List<StoredAlarm>) = withContext(ioDispatcher) {
        val ids = alarms.map { it.id }
        if (ids.isNotEmpty()) {
            alarmDao.markAsUploaded(ids)
        }
    }

    suspend fun incrementRetryCount(alarmId: Long) = withContext(ioDispatcher) {
        alarmDao.incrementRetryCount(alarmId, System.currentTimeMillis())
    }

    suspend fun cleanupOldAlarms(maxAgeMs: Long = 7 * 24 * 60 * 60 * 1000L) = withContext(ioDispatcher) {
        val cutoffTime = System.currentTimeMillis() - maxAgeMs
        alarmDao.deleteOldUploadedAlarms(cutoffTime)
    }

    private fun toStoredAlarm(alarm: FatigueAlarm): StoredAlarm {
        return StoredAlarm(
            alarmId = alarm.alarmId,
            timestamp = alarm.timestamp,
            level = alarm.level.value,
            score = alarm.score,
            triggersJson = triggersToJson(alarm.triggers),
            ear = alarm.ear,
            mar = alarm.mar,
            perclos = alarm.perclos,
            headPitch = alarm.headPitch,
            headYaw = alarm.headYaw,
            gazeAngle = alarm.gazeAngle,
            isYawning = alarm.isYawning,
            gpsLatitude = alarm.gpsLatitude,
            gpsLongitude = alarm.gpsLongitude,
            gpsSpeed = alarm.gpsSpeed,
            videoPath = alarm.videoPath,
            isUploaded = alarm.isUploaded
        )
    }

    private fun triggersToJson(triggers: Set<AlarmTrigger>): String {
        return triggers.joinToString(",") { it.name }
    }

    fun jsonToTriggers(json: String): Set<AlarmTrigger> {
        if (json.isBlank()) return emptySet()
        return json.split(",")
            .mapNotNull { runCatching { AlarmTrigger.valueOf(it) }.getOrNull() }
            .toSet()
    }

    fun storedToFatigueAlarm(stored: StoredAlarm): FatigueAlarm {
        return FatigueAlarm(
            alarmId = stored.alarmId,
            timestamp = stored.timestamp,
            level = com.ddg.edge.fatigue.model.AlarmLevel.fromValue(stored.level),
            score = stored.score,
            triggers = jsonToTriggers(stored.triggersJson),
            ear = stored.ear,
            mar = stored.mar,
            perclos = stored.perclos,
            headPitch = stored.headPitch,
            headYaw = stored.headYaw,
            gazeAngle = stored.gazeAngle,
            isYawning = stored.isYawning,
            gpsLatitude = stored.gpsLatitude,
            gpsLongitude = stored.gpsLongitude,
            gpsSpeed = stored.gpsSpeed,
            videoPath = stored.videoPath,
            isUploaded = stored.isUploaded
        )
    }
}
