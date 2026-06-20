package com.ddg.edge.fatigue.detector

import com.ddg.edge.fatigue.model.AlarmLevel
import com.ddg.edge.fatigue.model.AlarmTrigger
import kotlin.math.max
import kotlin.math.min

class FatigueScoreAggregator(
    private val perclosWeight: Float = 300f,
    private val marPenaltyWeight: Float = 2f,
    private val headDownPenaltyWeight: Float = 1.5f,
    private val gazeAwayPenaltyWeight: Float = 2f,
    private val marThreshold: Float = 0.6f,
    private val headDownPitchThreshold: Float = 15f,
    private val gazeAngleThreshold: Float = 15f
) {

    data class AggregationResult(
        val score: Int,
        val level: AlarmLevel,
        val triggers: Set<AlarmTrigger>,
        val perclosPenalty: Float,
        val marPenalty: Float,
        val headDownPenalty: Float,
        val gazeAwayPenalty: Float
    )

    fun aggregate(
        perclos: Float,
        mar: Float,
        headPitch: Float,
        gazeAngleDeviation: Float,
        isYawning: Boolean
    ): AggregationResult {
        val baseScore = 100f

        val perclosPenalty = calculatePERCLOSPenalty(perclos)
        val marPenalty = calculateMARPENALTY(mar, isYawning)
        val headDownPenalty = calculateHeadDownPenalty(headPitch)
        val gazeAwayPenalty = calculateGazeAwayPenalty(gazeAngleDeviation)

        var rawScore = baseScore - perclosPenalty - marPenalty - headDownPenalty - gazeAwayPenalty

        if (isYawning) {
            rawScore -= 5f
        }

        val score = min(100, max(0, rawScore.toInt()))

        val triggers = mutableSetOf<AlarmTrigger>()

        if (perclos > 0.15f) triggers.add(AlarmTrigger.PERCLOS_EXCEEDED)
        if (mar > marThreshold) triggers.add(AlarmTrigger.MAR_EXCEEDED)
        if (isYawning) triggers.add(AlarmTrigger.YAWNING)
        if (headPitch > headDownPitchThreshold) triggers.add(AlarmTrigger.HEAD_DOWN)
        if (gazeAngleDeviation > gazeAngleThreshold) triggers.add(AlarmTrigger.GAZE_AWAY)
        if (score < 60) triggers.add(AlarmTrigger.SCORE_LOW)

        val level = when {
            score >= 80 -> AlarmLevel.NONE
            score >= 60 -> AlarmLevel.LEVEL_1
            score >= 40 -> AlarmLevel.LEVEL_2
            else -> AlarmLevel.LEVEL_3
        }

        return AggregationResult(
            score = score,
            level = level,
            triggers = triggers,
            perclosPenalty = perclosPenalty,
            marPenalty = marPenalty,
            headDownPenalty = headDownPenalty,
            gazeAwayPenalty = gazeAwayPenalty
        )
    }

    private fun calculatePERCLOSPenalty(perclos: Float): Float {
        val clamped = min(1f, max(0f, perclos))
        return clamped * perclosWeight
    }

    private fun calculateMARPENALTY(mar: Float, isYawning: Boolean): Float {
        if (mar <= 0.3f) return 0f

        val excess = mar - 0.3f
        var penalty = excess * 50f * marPenaltyWeight

        if (isYawning) {
            penalty += 10f
        }

        return min(penalty, 30f)
    }

    private fun calculateHeadDownPenalty(pitch: Float): Float {
        if (pitch <= headDownPitchThreshold) return 0f

        val excess = pitch - headDownPitchThreshold
        val penalty = excess * headDownPenaltyWeight
        return min(penalty, 25f)
    }

    private fun calculateGazeAwayPenalty(gazeAngle: Float): Float {
        if (gazeAngle <= gazeAngleThreshold) return 0f

        val excess = gazeAngle - gazeAngleThreshold
        val penalty = excess * gazeAwayPenaltyWeight
        return min(penalty, 30f)
    }

    data class ScoreThresholds(
        val perclosWarning: Float = 0.10f,
        val perclosDanger: Float = 0.20f,
        val marWarning: Float = 0.5f,
        val marDanger: Float = 0.65f,
        val headPitchWarning: Float = 10f,
        val headPitchDanger: Float = 20f,
        val gazeAngleWarning: Float = 10f,
        val gazeAngleDanger: Float = 20f
    )
}
