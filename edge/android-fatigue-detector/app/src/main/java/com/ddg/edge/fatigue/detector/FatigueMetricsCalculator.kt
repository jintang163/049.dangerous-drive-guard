package com.ddg.edge.fatigue.detector

import com.google.mediapipe.tasks.components.containers.NormalizedLandmark
import kotlin.math.atan2
import kotlin.math.cos
import kotlin.math.sin
import kotlin.math.sqrt

object FatigueMetricsCalculator {

    private const val LEFT_EYE_INDICES = intArrayOf(33, 160, 158, 133, 153, 144)
    private const val RIGHT_EYE_INDICES = intArrayOf(362, 385, 387, 263, 373, 380)

    private const val MOUTH_INDICES = intArrayOf(61, 146, 91, 181, 84, 17, 314, 405, 321, 375)
    private const val MOUTH_WIDTH_INDICES = intArrayOf(61, 291)
    private const val MOUTH_HEIGHT_INDICES = intArrayOf(13, 14)

    data class HeadPose(
        val pitch: Float,
        val yaw: Float,
        val roll: Float
    )

    fun calculateEAR(landmarks: List<NormalizedLandmark>): Float {
        if (landmarks.size < 468) return 0f

        val leftEAR = calculateSingleEyeEAR(landmarks, LEFT_EYE_INDICES)
        val rightEAR = calculateSingleEyeEAR(landmarks, RIGHT_EYE_INDICES)

        return (leftEAR + rightEAR) / 2.0f
    }

    private fun calculateSingleEyeEAR(
        landmarks: List<NormalizedLandmark>,
        indices: IntArray
    ): Float {
        val p1 = landmarks[indices[0]]
        val p2 = landmarks[indices[1]]
        val p3 = landmarks[indices[2]]
        val p4 = landmarks[indices[3]]
        val p5 = landmarks[indices[4]]
        val p6 = landmarks[indices[5]]

        val v1 = distance(p2, p6)
        val v2 = distance(p3, p5)
        val h = distance(p1, p4)

        if (h < 1e-6f) return 0f

        return (v1 + v2) / (2.0f * h)
    }

    fun calculateMAR(landmarks: List<NormalizedLandmark>): Float {
        if (landmarks.size < 468) return 0f

        val mouthTop = landmarks[MOUTH_HEIGHT_INDICES[0]]
        val mouthBottom = landmarks[MOUTH_HEIGHT_INDICES[1]]
        val mouthLeft = landmarks[MOUTH_WIDTH_INDICES[0]]
        val mouthRight = landmarks[MOUTH_WIDTH_INDICES[1]]

        val verticalDist = distance(mouthTop, mouthBottom)
        val horizontalDist = distance(mouthLeft, mouthRight)

        if (horizontalDist < 1e-6f) return 0f

        return verticalDist / horizontalDist
    }

    fun calculateMAR10Point(landmarks: List<NormalizedLandmark>): Float {
        if (landmarks.size < 468) return 0f

        val p1 = landmarks[MOUTH_INDICES[0]]
        val p2 = landmarks[MOUTH_INDICES[1]]
        val p3 = landmarks[MOUTH_INDICES[2]]
        val p4 = landmarks[MOUTH_INDICES[3]]
        val p5 = landmarks[MOUTH_INDICES[4]]
        val p6 = landmarks[MOUTH_INDICES[5]]
        val p7 = landmarks[MOUTH_INDICES[6]]
        val p8 = landmarks[MOUTH_INDICES[7]]
        val p9 = landmarks[MOUTH_INDICES[8]]
        val p10 = landmarks[MOUTH_INDICES[9]]

        val v1 = distance(p2, p10)
        val v2 = distance(p3, p9)
        val v3 = distance(p4, p8)
        val v4 = distance(p5, p7)
        val h = distance(p1, p6)

        if (h < 1e-6f) return 0f

        return (v1 + v2 + v3 + v4) / (4.0f * h)
    }

    fun calculateHeadPose(landmarks: List<NormalizedLandmark>): HeadPose {
        if (landmarks.size < 468) return HeadPose(0f, 0f, 0f)

        val noseTip = landmarks[1]
        val chin = landmarks[152]
        val leftEyeOuter = landmarks[33]
        val rightEyeOuter = landmarks[263]
        val leftMouth = landmarks[61]
        val rightMouth = landmarks[291]
        val forehead = landmarks[10]

        val modelPoints = arrayOf(
            doubleArrayOf(0.0, 0.0, 0.0),
            doubleArrayOf(0.0, -330.0, -65.0),
            doubleArrayOf(-225.0, 170.0, -135.0),
            doubleArrayOf(225.0, 170.0, -135.0),
            doubleArrayOf(-150.0, -150.0, -125.0),
            doubleArrayOf(150.0, -150.0, -125.0),
            doubleArrayOf(0.0, 250.0, -150.0)
        )

        val imagePoints = arrayOf(
            doubleArrayOf(noseTip.x().toDouble(), noseTip.y().toDouble()),
            doubleArrayOf(chin.x().toDouble(), chin.y().toDouble()),
            doubleArrayOf(leftEyeOuter.x().toDouble(), leftEyeOuter.y().toDouble()),
            doubleArrayOf(rightEyeOuter.x().toDouble(), rightEyeOuter.y().toDouble()),
            doubleArrayOf(leftMouth.x().toDouble(), leftMouth.y().toDouble()),
            doubleArrayOf(rightMouth.x().toDouble(), rightMouth.y().toDouble()),
            doubleArrayOf(forehead.x().toDouble(), forehead.y().toDouble())
        )

        return solveHeadPoseEuler(modelPoints, imagePoints)
    }

    private fun solveHeadPoseEuler(
        modelPoints: Array<DoubleArray>,
        imagePoints: Array<DoubleArray>
    ): HeadPose {
        val centerModel = centroid(modelPoints)
        val centerImage = centroid2D(imagePoints)

        val centeredModel = modelPoints.map {
            doubleArrayOf(it[0] - centerModel[0], it[1] - centerModel[1], it[2] - centerModel[2])
        }.toTypedArray()

        val centeredImage = imagePoints.map {
            doubleArrayOf(it[0] - centerImage[0], it[1] - centerImage[1])
        }.toTypedArray()

        val sxx = centeredModel.zip(centeredImage).sumOf { it.first[0] * it.second[0] }
        val sxy = centeredModel.zip(centeredImage).sumOf { it.first[0] * it.second[1] }
        val syx = centeredModel.zip(centeredImage).sumOf { it.first[1] * it.second[0] }
        val syy = centeredModel.zip(centeredImage).sumOf { it.first[1] * it.second[1] }
        val szx = centeredModel.zip(centeredImage).sumOf { it.first[2] * it.second[0] }
        val szy = centeredModel.zip(centeredImage).sumOf { it.first[2] * it.second[1] }

        val pitch = atan2(-szy, syy).toFloat()
        val yaw = atan2(szx, sxx).toFloat()
        val roll = atan2(-sxy, sxx).toFloat()

        val pitchDeg = Math.toDegrees(pitch.toDouble()).toFloat()
        val yawDeg = Math.toDegrees(yaw.toDouble()).toFloat()
        val rollDeg = Math.toDegrees(roll.toDouble()).toFloat()

        return HeadPose(
            pitch = pitchDeg.coerceIn(-90f, 90f),
            yaw = yawDeg.coerceIn(-90f, 90f),
            roll = rollDeg.coerceIn(-45f, 45f)
        )
    }

    private fun centroid(points: Array<DoubleArray>): DoubleArray {
        val n = points.size
        val sum = points.fold(doubleArrayOf(0.0, 0.0, 0.0)) { acc, p ->
            doubleArrayOf(acc[0] + p[0], acc[1] + p[1], acc[2] + p[2])
        }
        return doubleArrayOf(sum[0] / n, sum[1] / n, sum[2] / n)
    }

    private fun centroid2D(points: Array<DoubleArray>): DoubleArray {
        val n = points.size
        val sum = points.fold(doubleArrayOf(0.0, 0.0)) { acc, p ->
            doubleArrayOf(acc[0] + p[0], acc[1] + p[1])
        }
        return doubleArrayOf(sum[0] / n, sum[1] / n)
    }

    private fun distance(p1: NormalizedLandmark, p2: NormalizedLandmark): Float {
        val dx = p1.x() - p2.x()
        val dy = p1.y() - p2.y()
        return sqrt(dx * dx + dy * dy)
    }
}
