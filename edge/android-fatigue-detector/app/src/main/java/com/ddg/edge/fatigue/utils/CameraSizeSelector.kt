package com.ddg.edge.fatigue.utils

import android.util.Size
import androidx.camera.core.CameraSelector
import androidx.camera.core.resolutionselector.ResolutionFilter
import androidx.camera.core.resolutionselector.ResolutionSelector
import androidx.camera.core.resolutionselector.ResolutionStrategy
import java.util.Collections
import kotlin.math.abs
import kotlin.math.max
import kotlin.math.min

object CameraSizeSelector {

    private val TARGET_RESOLUTIONS = listOf(
        Size(640, 480),
        Size(800, 600),
        Size(1280, 720),
        Size(1024, 768),
        Size(1920, 1080)
    )

    fun createResolutionSelector(
        targetWidth: Int = 640,
        targetHeight: Int = 480
    ): ResolutionSelector {
        val targetAspectRatio = targetWidth.toFloat() / targetHeight.toFloat()

        val filter = ResolutionFilter { supportedSizes ->
            val scoredSizes = supportedSizes.map { size ->
                val aspectRatio = size.width.toFloat() / size.height.toFloat()
                val aspectDiff = abs(aspectRatio - targetAspectRatio)
                val widthDiff = abs(size.width - targetWidth)
                val heightDiff = abs(size.height - targetHeight)
                val areaDiff = abs(size.width * size.height - targetWidth * targetHeight)
                val score = aspectDiff * 1000 + areaDiff.toFloat() * 0.00001f + (widthDiff + heightDiff).toFloat()
                ScoredSize(size, score)
            }

            scoredSizes.sortedBy { it.score }.map { it.size }
        }

        return ResolutionSelector.Builder()
            .setResolutionFilter(filter)
            .setResolutionStrategy(
                ResolutionStrategy(
                    Size(targetWidth, targetHeight),
                    ResolutionStrategy.FALLBACK_RULE_CLOSEST_HIGHER_THEN_LOWER
                )
            )
            .build()
    }

    fun selectOptimalPreviewSize(
        supportedSizes: List<Size>,
        targetWidth: Int = 640,
        targetHeight: Int = 480,
        maxWidth: Int = 1920,
        maxHeight: Int = 1080
    ): Size {
        if (supportedSizes.isEmpty()) return Size(targetWidth, targetHeight)

        val targetArea = targetWidth * targetHeight
        val targetRatio = targetWidth.toDouble() / targetHeight

        val candidates = supportedSizes.filter {
            it.width <= maxWidth && it.height <= maxHeight
        }.ifEmpty { supportedSizes }

        val scored = candidates.map { size ->
            val area = size.width * size.height
            val ratio = size.width.toDouble() / size.height
            val ratioDiff = abs(ratio - targetRatio)
            val areaDiff = abs(area - targetArea)

            val score = when {
                ratioDiff < 0.01 -> areaDiff.toDouble()
                else -> ratioDiff * 1_000_000.0 + areaDiff.toDouble()
            }
            ScoredSize(size, score.toFloat())
        }

        val exactMatch = candidates.firstOrNull {
            it.width == targetWidth && it.height == targetHeight
        }
        if (exactMatch != null) return exactMatch

        val targetAreaMatch = candidates.firstOrNull {
            it.width * it.height == targetArea
        }
        if (targetAreaMatch != null) return targetAreaMatch

        return scored.minByOrNull { it.score }?.size ?: candidates.first()
    }

    fun findNearestHigherSize(
        sizes: List<Size>,
        minWidth: Int,
        minHeight: Int
    ): Size? {
        return sizes.filter {
            it.width >= minWidth && it.height >= minHeight
        }.minByOrNull {
            it.width * it.height
        }
    }

    fun findBestAspectMatch(
        sizes: List<Size>,
        aspectWidth: Int,
        aspectHeight: Int
    ): Size? {
        if (sizes.isEmpty()) return null

        val targetRatio = aspectWidth.toDouble() / aspectHeight

        return sizes.minByOrNull {
            val ratio = it.width.toDouble() / it.height
            abs(ratio - targetRatio)
        }
    }

    fun getSupportedResolutionsString(sizes: List<Size>): String {
        return sizes.joinToString(", ") { "${it.width}x${it.height}" }
    }

    fun rotateSize(size: Size, rotationDegrees: Int): Size {
        return if (rotationDegrees == 90 || rotationDegrees == 270) {
            Size(size.height, size.width)
        } else {
            size
        }
    }

    fun isLandscape(size: Size): Boolean = size.width >= size.height

    fun isPortrait(size: Size): Boolean = size.height >= size.width

    private data class ScoredSize(
        val size: Size,
        val score: Float
    )
}
