package com.ddg.edge.fatigue.model

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "stored_alarms")
data class StoredAlarm(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    val alarmId: String,
    val timestamp: Long,
    val level: Int,
    val score: Int,
    val triggersJson: String,
    val ear: Float,
    val mar: Float,
    val perclos: Float,
    val headPitch: Float,
    val headYaw: Float,
    val gazeAngle: Float,
    val isYawning: Boolean,
    val gpsLatitude: Double?,
    val gpsLongitude: Double?,
    val gpsSpeed: Float?,
    val videoPath: String?,
    var isUploaded: Boolean = false,
    var uploadRetryCount: Int = 0,
    var lastUploadAttempt: Long? = null
)
