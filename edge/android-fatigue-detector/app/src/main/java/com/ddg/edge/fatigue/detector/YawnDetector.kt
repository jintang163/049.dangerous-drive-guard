package com.ddg.edge.fatigue.detector

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

class YawnDetector(
    private val marThreshold: Float = 0.6f,
    private val minDurationMs: Long = 1000L,
    private val cooldownMs: Long = 3000L
) {

    private var yawnStartTimestamp: Long? = null
    private var lastYawnEndTimestamp: Long = 0L
    private var consecutiveFramesOverThreshold = 0

    private val _isYawning = MutableStateFlow(false)
    val isYawning: StateFlow<Boolean> = _isYawning.asStateFlow()

    private val _yawnCount = MutableStateFlow(0)
    val yawnCount: StateFlow<Int> = _yawnCount.asStateFlow()

    private val _yawnEvents = MutableStateFlow<List<YawnEvent>>(emptyList())
    val yawnEvents: StateFlow<List<YawnEvent>> = _yawnEvents.asStateFlow()

    private val yawnHistory = mutableListOf<YawnEvent>()
    private val maxHistorySize = 100

    data class YawnEvent(
        val startTime: Long,
        val endTime: Long,
        val durationMs: Long,
        val peakMAR: Float
    )

    fun processFrame(timestamp: Long, mar: Float): Boolean {
        val inCooldown = timestamp - lastYawnEndTimestamp < cooldownMs

        if (mar >= marThreshold) {
            consecutiveFramesOverThreshold++

            if (yawnStartTimestamp == null && !inCooldown) {
                yawnStartTimestamp = timestamp
            }

            if (yawnStartTimestamp != null) {
                val elapsed = timestamp - yawnStartTimestamp!!
                if (elapsed >= minDurationMs) {
                    if (!_isYawning.value) {
                        _isYawning.value = true
                    }
                }
            }
        } else {
            if (yawnStartTimestamp != null && _isYawning.value) {
                val duration = timestamp - yawnStartTimestamp!!
                if (duration >= minDurationMs) {
                    val event = YawnEvent(
                        startTime = yawnStartTimestamp!!,
                        endTime = timestamp,
                        durationMs = duration,
                        peakMAR = marThreshold
                    )
                    recordYawnEvent(event)
                    lastYawnEndTimestamp = timestamp
                }
            }
            yawnStartTimestamp = null
            consecutiveFramesOverThreshold = 0
            _isYawning.value = false
        }

        return _isYawning.value
    }

    private fun recordYawnEvent(event: YawnEvent) {
        yawnHistory.add(event)
        if (yawnHistory.size > maxHistorySize) {
            yawnHistory.removeAt(0)
        }
        _yawnCount.value++
        _yawnEvents.value = yawnHistory.toList()
    }

    fun getYawnCountInWindow(windowMs: Long, currentTime: Long): Int {
        val cutoff = currentTime - windowMs
        return yawnHistory.count { it.startTime >= cutoff }
    }

    fun reset() {
        yawnStartTimestamp = null
        lastYawnEndTimestamp = 0L
        consecutiveFramesOverThreshold = 0
        _isYawning.value = false
        _yawnCount.value = 0
        yawnHistory.clear()
        _yawnEvents.value = emptyList()
    }
}
