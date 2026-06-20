package com.ddg.edge.fatigue.detector

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import java.util.LinkedList

class PERCLOSCalculator(
    private val windowDurationMs: Long = 60_000L,
    private val earThreshold: Float = 0.25f
) {

    private data class FrameRecord(
        val timestamp: Long,
        val ear: Float
    )

    private val frameWindow = LinkedList<FrameRecord>()

    private val _perclos = MutableStateFlow(0f)
    val perclos: StateFlow<Float> = _perclos.asStateFlow()

    private val _closedFramesInWindow = MutableStateFlow(0)
    val closedFramesInWindow: StateFlow<Int> = _closedFramesInWindow.asStateFlow()

    private val _totalFramesInWindow = MutableStateFlow(0)
    val totalFramesInWindow: StateFlow<Int> = _totalFramesInWindow.asStateFlow()

    @Synchronized
    fun addFrame(timestamp: Long, ear: Float): Float {
        frameWindow.addLast(FrameRecord(timestamp, ear))

        val cutoffTime = timestamp - windowDurationMs
        while (frameWindow.isNotEmpty() && frameWindow.first.timestamp < cutoffTime) {
            frameWindow.removeFirst()
        }

        val total = frameWindow.size
        val closed = frameWindow.count { it.ear < earThreshold }

        _totalFramesInWindow.value = total
        _closedFramesInWindow.value = closed

        val perclosValue = if (total > 0) {
            closed.toFloat() / total.toFloat()
        } else {
            0f
        }

        _perclos.value = perclosValue
        return perclosValue
    }

    @Synchronized
    fun reset() {
        frameWindow.clear()
        _perclos.value = 0f
        _closedFramesInWindow.value = 0
        _totalFramesInWindow.value = 0
    }

    @Synchronized
    fun getCurrentStats(): PERCLOSStats {
        return PERCLOSStats(
            perclos = _perclos.value,
            closedFrames = _closedFramesInWindow.value,
            totalFrames = _totalFramesInWindow.value,
            windowDurationMs = windowDurationMs,
            earThreshold = earThreshold
        )
    }

    data class PERCLOSStats(
        val perclos: Float,
        val closedFrames: Int,
        val totalFrames: Int,
        val windowDurationMs: Long,
        val earThreshold: Float
    )
}
