package com.ddg.edge.fatigue.detector

import com.ddg.edge.fatigue.model.GazeDirection
import com.google.mediapipe.tasks.components.containers.NormalizedLandmark
import kotlin.math.atan2
import kotlin.math.sqrt

class GazeDirectionDetector(
    private val angleThresholdDeg: Float = 15f
) {

    private val LEFT_IRIS = intArrayOf(468, 469, 470, 471, 472)
    private val RIGHT_IRIS = intArrayOf(473, 474, 475, 476, 477)

    private val LEFT_EYE_CORNERS = intArrayOf(33, 133)
    private val RIGHT_EYE_CORNERS = intArrayOf(362, 263)

    private val LEFT_EYE_TOP_BOTTOM = intArrayOf(159, 145)
    private val RIGHT_EYE_TOP_BOTTOM = intArrayOf(386, 374)

    data class GazeResult(
        val direction: GazeDirection,
        val horizontalAngle: Float,
        val verticalAngle: Float,
        val angleDeviation: Float,
        val isGazeAway: Boolean
    )

    fun detect(landmarks: List<NormalizedLandmark>): GazeResult {
        if (landmarks.size < 478) {
            return GazeResult(
                direction = GazeDirection.UNKNOWN,
                horizontalAngle = 0f,
                verticalAngle = 0f,
                angleDeviation = 0f,
                isGazeAway = false
            )
        }

        val leftIrisCenter = getIrisCenter(landmarks, LEFT_IRIS)
        val rightIrisCenter = getIrisCenter(landmarks, RIGHT_IRIS)

        val leftEyeLeftCorner = landmarks[LEFT_EYE_CORNERS[0]]
        val leftEyeRightCorner = landmarks[LEFT_EYE_CORNERS[1]]
        val rightEyeLeftCorner = landmarks[RIGHT_EYE_CORNERS[0]]
        val rightEyeRightCorner = landmarks[RIGHT_EYE_CORNERS[1]]

        val leftEyeTop = landmarks[LEFT_EYE_TOP_BOTTOM[0]]
        val leftEyeBottom = landmarks[LEFT_EYE_TOP_BOTTOM[1]]
        val rightEyeTop = landmarks[RIGHT_EYE_TOP_BOTTOM[0]]
        val rightEyeBottom = landmarks[RIGHT_EYE_TOP_BOTTOM[1]]

        val leftHorizontalRatio = calculateHorizontalRatio(
            leftIrisCenter, leftEyeLeftCorner, leftEyeRightCorner
        )
        val rightHorizontalRatio = calculateHorizontalRatio(
            rightIrisCenter, rightEyeLeftCorner, rightEyeRightCorner
        )

        val leftVerticalRatio = calculateVerticalRatio(
            leftIrisCenter, leftEyeTop, leftEyeBottom
        )
        val rightVerticalRatio = calculateVerticalRatio(
            rightIrisCenter, rightEyeTop, rightEyeBottom
        )

        val avgHorizontalRatio = (leftHorizontalRatio + rightHorizontalRatio) / 2f
        val avgVerticalRatio = (leftVerticalRatio + rightVerticalRatio) / 2f

        val horizontalAngle = ratioToAngle(avgHorizontalRatio)
        val verticalAngle = ratioToAngle(avgVerticalRatio)

        val angleDeviation = sqrt(
            horizontalAngle * horizontalAngle + verticalAngle * verticalAngle
        )

        val direction = determineDirection(
            avgHorizontalRatio, avgVerticalRatio, angleThresholdDeg
        )

        val isGazeAway = angleDeviation > angleThresholdDeg

        return GazeResult(
            direction = direction,
            horizontalAngle = horizontalAngle,
            verticalAngle = verticalAngle,
            angleDeviation = angleDeviation,
            isGazeAway = isGazeAway
        )
    }

    private fun getIrisCenter(
        landmarks: List<NormalizedLandmark>,
        irisIndices: IntArray
    ): NormalizedLandmark {
        var sumX = 0f
        var sumY = 0f
        var sumZ = 0f

        for (idx in irisIndices) {
            val lm = landmarks[idx]
            sumX += lm.x()
            sumY += lm.y()
            sumZ += lm.z()
        }

        val count = irisIndices.size
        return NormalizedLandmark.create(
            sumX / count, sumY / count, sumZ / count
        )
    }

    private fun calculateHorizontalRatio(
        irisCenter: NormalizedLandmark,
        leftCorner: NormalizedLandmark,
        rightCorner: NormalizedLandmark
    ): Float {
        val eyeWidth = distance(leftCorner, rightCorner)
        if (eyeWidth < 1e-6f) return 0f

        val midX = (leftCorner.x() + rightCorner.x()) / 2f
        val midY = (leftCorner.y() + rightCorner.y()) / 2f

        val offsetX = irisCenter.x() - midX
        val normalizedOffset = (offsetX * 2f) / eyeWidth

        return normalizedOffset.coerceIn(-1f, 1f)
    }

    private fun calculateVerticalRatio(
        irisCenter: NormalizedLandmark,
        top: NormalizedLandmark,
        bottom: NormalizedLandmark
    ): Float {
        val eyeHeight = distance(top, bottom)
        if (eyeHeight < 1e-6f) return 0f

        val midY = (top.y() + bottom.y()) / 2f

        val offsetY = irisCenter.y() - midY
        val normalizedOffset = (offsetY * 2f) / eyeHeight

        return normalizedOffset.coerceIn(-1f, 1f)
    }

    private fun ratioToAngle(ratio: Float): Float {
        val clamped = ratio.coerceIn(-1f, 1f)
        return Math.toDegrees(atan2(clamped.toDouble(), 1.0)).toFloat()
    }

    private fun determineDirection(
        horizontalRatio: Float,
        verticalRatio: Float,
        threshold: Float
    ): GazeDirection {
        val thresholdRatio = Math.tan(Math.toRadians(threshold.toDouble())).toFloat()

        val absH = Math.abs(horizontalRatio)
        val absV = Math.abs(verticalRatio)

        if (absH < thresholdRatio && absV < thresholdRatio) {
            return GazeDirection.CENTER
        }

        return if (absH >= absV) {
            if (horizontalRatio > 0) GazeDirection.RIGHT else GazeDirection.LEFT
        } else {
            if (verticalRatio > 0) GazeDirection.DOWN else GazeDirection.UP
        }
    }

    private fun distance(p1: NormalizedLandmark, p2: NormalizedLandmark): Float {
        val dx = p1.x() - p2.x()
        val dy = p1.y() - p2.y()
        return sqrt(dx * dx + dy * dy)
    }
}
