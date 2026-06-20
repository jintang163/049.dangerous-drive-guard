package com.ddg.edge.fatigue.data.local

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Update
import com.ddg.edge.fatigue.model.StoredAlarm
import kotlinx.coroutines.flow.Flow

@Dao
interface AlarmDao {

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertAlarm(alarm: StoredAlarm): Long

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertAlarms(alarms: List<StoredAlarm>)

    @Update
    suspend fun updateAlarm(alarm: StoredAlarm)

    @Update
    suspend fun updateAlarms(alarms: List<StoredAlarm>)

    @Query("SELECT * FROM stored_alarms WHERE isUploaded = 0 ORDER BY timestamp ASC")
    suspend fun getPendingAlarms(): List<StoredAlarm>

    @Query("SELECT * FROM stored_alarms WHERE isUploaded = 0 ORDER BY timestamp ASC")
    fun observePendingAlarms(): Flow<List<StoredAlarm>>

    @Query("SELECT * FROM stored_alarms ORDER BY timestamp DESC LIMIT :limit")
    suspend fun getRecentAlarms(limit: Int = 100): List<StoredAlarm>

    @Query("SELECT * FROM stored_alarms ORDER BY timestamp DESC")
    fun observeAllAlarms(): Flow<List<StoredAlarm>>

    @Query("SELECT * FROM stored_alarms WHERE timestamp >= :startTime AND timestamp <= :endTime ORDER BY timestamp DESC")
    suspend fun getAlarmsInTimeRange(startTime: Long, endTime: Long): List<StoredAlarm>

    @Query("SELECT COUNT(*) FROM stored_alarms WHERE isUploaded = 0")
    suspend fun getPendingAlarmCount(): Int

    @Query("DELETE FROM stored_alarms WHERE id = :id")
    suspend fun deleteAlarmById(id: Long)

    @Query("DELETE FROM stored_alarms WHERE isUploaded = 1 AND timestamp < :beforeTime")
    suspend fun deleteOldUploadedAlarms(beforeTime: Long)

    @Query("UPDATE stored_alarms SET isUploaded = 1 WHERE id IN (:ids)")
    suspend fun markAsUploaded(ids: List<Long>)

    @Query("UPDATE stored_alarms SET uploadRetryCount = uploadRetryCount + 1, lastUploadAttempt = :attemptTime WHERE id = :id")
    suspend fun incrementRetryCount(id: Long, attemptTime: Long)
}
