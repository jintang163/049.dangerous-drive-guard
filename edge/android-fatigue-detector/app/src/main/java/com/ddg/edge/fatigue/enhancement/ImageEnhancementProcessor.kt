package com.ddg.edge.fatigue.enhancement

import android.graphics.Bitmap
import android.graphics.Color
import android.util.Log
import kotlin.math.pow
import kotlin.math.sqrt
import kotlin.math.exp
import kotlin.math.abs

data class EnhanceConfig(
    val mode: EnhanceMode = EnhanceMode.AUTO,
    val gamma: Float = 1.2f,
    val brightnessDelta: Int = 30,
    val contrastDelta: Int = 20,
    val applyHistogramEq: Boolean = true,
    val applyClahe: Boolean = true,
    val applyDenoise: Boolean = true,
    val denoiseStrength: Int = 3,
    val applySharpen: Boolean = false,
    val sharpenStrength: Int = 2
)

enum class EnhanceMode {
    AUTO,
    NIGHT,
    INFRARED,
    LOW_LIGHT,
    MANUAL
}

data class EnhanceResult(
    val enhancedBitmap: Bitmap,
    val qualityScoreBefore: Float,
    val qualityScoreAfter: Float,
    val qualityImprovement: Float,
    val avgBrightnessBefore: Int,
    val avgBrightnessAfter: Int,
    val processingTimeMs: Long,
    val appliedGamma: Float,
    val appliedBrightnessDelta: Int,
    val appliedContrastDelta: Int,
    val histogramEqApplied: Boolean,
    val denoiseApplied: Boolean,
    val sharpenApplied: Boolean
)

class ImageEnhancementProcessor {

    companion object {
        private const val TAG = "ImageEnhancement"

        fun getConfigForMode(mode: EnhanceMode, lightLevelLux: Float): EnhanceConfig {
            return when (mode) {
                EnhanceMode.NIGHT -> EnhanceConfig(
                    mode = EnhanceMode.NIGHT,
                    gamma = 1.2f,
                    brightnessDelta = 30,
                    contrastDelta = 20,
                    applyHistogramEq = true,
                    applyDenoise = true,
                    denoiseStrength = 3,
                    applySharpen = false
                )
                EnhanceMode.INFRARED -> EnhanceConfig(
                    mode = EnhanceMode.INFRARED,
                    gamma = 1.1f,
                    brightnessDelta = 15,
                    contrastDelta = 25,
                    applyHistogramEq = true,
                    applyDenoise = true,
                    denoiseStrength = 2,
                    applySharpen = true,
                    sharpenStrength = 2
                )
                EnhanceMode.LOW_LIGHT -> EnhanceConfig(
                    mode = EnhanceMode.LOW_LIGHT,
                    gamma = 1.3f,
                    brightnessDelta = 40,
                    contrastDelta = 25,
                    applyHistogramEq = true,
                    applyDenoise = true,
                    denoiseStrength = 3,
                    applySharpen = true,
                    sharpenStrength = 2
                )
                EnhanceMode.MANUAL -> EnhanceConfig(
                    mode = EnhanceMode.MANUAL
                )
                EnhanceMode.AUTO -> {
                    when {
                        lightLevelLux < 10f -> EnhanceConfig(
                            mode = EnhanceMode.LOW_LIGHT,
                            gamma = 1.35f,
                            brightnessDelta = 45,
                            contrastDelta = 25,
                            applyHistogramEq = true,
                            applyDenoise = true,
                            denoiseStrength = 4,
                            applySharpen = true
                        )
                        lightLevelLux < 50f -> EnhanceConfig(
                            mode = EnhanceMode.NIGHT,
                            gamma = 1.2f,
                            brightnessDelta = 30,
                            contrastDelta = 20,
                            applyHistogramEq = true,
                            applyDenoise = true,
                            denoiseStrength = 3,
                            applySharpen = false
                        )
                        lightLevelLux < 200f -> EnhanceConfig(
                            mode = EnhanceMode.NIGHT,
                            gamma = 1.1f,
                            brightnessDelta = 15,
                            contrastDelta = 15,
                            applyHistogramEq = false,
                            applyDenoise = true,
                            denoiseStrength = 2,
                            applySharpen = false
                        )
                        else -> EnhanceConfig(
                            mode = EnhanceMode.AUTO,
                            gamma = 1.0f,
                            brightnessDelta = 0,
                            contrastDelta = 0,
                            applyHistogramEq = false,
                            applyDenoise = false,
                            applySharpen = false
                        )
                    }
                }
            }
        }
    }

    fun enhance(bitmap: Bitmap, config: EnhanceConfig): EnhanceResult {
        val startTime = System.currentTimeMillis()

        val qualityBefore = calculateImageQuality(bitmap)
        val brightnessBefore = calculateAvgBrightness(bitmap)

        var result = bitmap.copy(bitmap.config ?: Bitmap.Config.ARGB_8888, true)

        val appliedGamma = if (config.gamma != 1.0f && config.gamma > 0) {
            result = applyGammaCorrection(result, config.gamma)
            config.gamma
        } else 1.0f

        val histoApplied = if (config.applyHistogramEq) {
            result = applyHistogramEqualization(result)
            true
        } else false

        val brightnessApplied = if (config.brightnessDelta != 0) {
            result = applyBrightness(result, config.brightnessDelta)
            config.brightnessDelta
        } else 0

        val contrastApplied = if (config.contrastDelta != 0) {
            result = applyContrast(result, config.contrastDelta)
            config.contrastDelta
        } else 0

        val denoiseApplied = if (config.applyDenoise) {
            result = applyDenoise(result, config.denoiseStrength)
            true
        } else false

        val sharpenApplied = if (config.applySharpen) {
            result = applySharpen(result, config.sharpenStrength)
            true
        } else false

        val qualityAfter = calculateImageQuality(result)
        val brightnessAfter = calculateAvgBrightness(result)

        val improvement = if (qualityBefore > 0f) {
            (qualityAfter - qualityBefore) / qualityBefore * 100f
        } else 0f

        val processingTime = System.currentTimeMillis() - startTime

        return EnhanceResult(
            enhancedBitmap = result,
            qualityScoreBefore = qualityBefore,
            qualityScoreAfter = qualityAfter,
            qualityImprovement = improvement,
            avgBrightnessBefore = brightnessBefore,
            avgBrightnessAfter = brightnessAfter,
            processingTimeMs = processingTime,
            appliedGamma = appliedGamma,
            appliedBrightnessDelta = brightnessApplied,
            appliedContrastDelta = contrastApplied,
            histogramEqApplied = histoApplied,
            denoiseApplied = denoiseApplied,
            sharpenApplied = sharpenApplied
        )
    }

    fun applyGammaCorrection(bitmap: Bitmap, gamma: Float): Bitmap {
        val width = bitmap.width
        val height = bitmap.height
        val result = Bitmap.createBitmap(width, height, bitmap.config ?: Bitmap.Config.ARGB_8888)

        val invGamma = 1.0 / gamma
        val lut = IntArray(256) { i ->
            ((i / 255.0).pow(invGamma) * 255.0).toInt().coerceIn(0, 255)
        }

        val pixels = IntArray(width * height)
        bitmap.getPixels(pixels, 0, width, 0, 0, width, height)

        for (i in pixels.indices) {
            val pixel = pixels[i]
            val r = lut[Color.red(pixel)]
            val g = lut[Color.green(pixel)]
            val b = lut[Color.blue(pixel)]
            pixels[i] = Color.argb(Color.alpha(pixel), r, g, b)
        }

        result.setPixels(pixels, 0, width, 0, 0, width, height)
        return result
    }

    fun applyBrightness(bitmap: Bitmap, delta: Int): Bitmap {
        val width = bitmap.width
        val height = bitmap.height
        val result = Bitmap.createBitmap(width, height, bitmap.config ?: Bitmap.Config.ARGB_8888)

        val pixels = IntArray(width * height)
        bitmap.getPixels(pixels, 0, width, 0, 0, width, height)

        for (i in pixels.indices) {
            val pixel = pixels[i]
            val r = (Color.red(pixel) + delta).coerceIn(0, 255)
            val g = (Color.green(pixel) + delta).coerceIn(0, 255)
            val b = (Color.blue(pixel) + delta).coerceIn(0, 255)
            pixels[i] = Color.argb(Color.alpha(pixel), r, g, b)
        }

        result.setPixels(pixels, 0, width, 0, 0, width, height)
        return result
    }

    fun applyContrast(bitmap: Bitmap, delta: Int): Bitmap {
        val width = bitmap.width
        val height = bitmap.height
        val result = Bitmap.createBitmap(width, height, bitmap.config ?: Bitmap.Config.ARGB_8888)

        val factor = (259.0 * (delta + 255)) / (255.0 * (259 - delta))

        val pixels = IntArray(width * height)
        bitmap.getPixels(pixels, 0, width, 0, 0, width, height)

        for (i in pixels.indices) {
            val pixel = pixels[i]
            val r = (factor * (Color.red(pixel) - 128) + 128).toInt().coerceIn(0, 255)
            val g = (factor * (Color.green(pixel) - 128) + 128).toInt().coerceIn(0, 255)
            val b = (factor * (Color.blue(pixel) - 128) + 128).toInt().coerceIn(0, 255)
            pixels[i] = Color.argb(Color.alpha(pixel), r, g, b)
        }

        result.setPixels(pixels, 0, width, 0, 0, width, height)
        return result
    }

    fun applyHistogramEqualization(bitmap: Bitmap): Bitmap {
        val width = bitmap.width
        val height = bitmap.height
        val result = Bitmap.createBitmap(width, height, bitmap.config ?: Bitmap.Config.ARGB_8888)

        val pixels = IntArray(width * height)
        bitmap.getPixels(pixels, 0, width, 0, 0, width, height)

        val lumHistogram = IntArray(256)
        val totalPixels = width * height

        for (pixel in pixels) {
            val lum = (0.299 * Color.red(pixel) + 0.587 * Color.green(pixel) + 0.114 * Color.blue(pixel)).toInt()
            lumHistogram[lum.coerceIn(0, 255)]++
        }

        val cdf = IntArray(256)
        cdf[0] = lumHistogram[0]
        for (i in 1 until 256) {
            cdf[i] = cdf[i - 1] + lumHistogram[i]
        }

        var cdfMin = 0
        for (i in 0 until 256) {
            if (cdf[i] > 0) {
                cdfMin = cdf[i]
                break
            }
        }

        val lut = IntArray(256)
        for (i in 0 until 256) {
            lut[i] = if (totalPixels - cdfMin > 0) {
                ((cdf[i] - cdfMin).toFloat() / (totalPixels - cdfMin) * 255f).toInt().coerceIn(0, 255)
            } else i
        }

        for (i in pixels.indices) {
            val pixel = pixels[i]
            val lum = (0.299 * Color.red(pixel) + 0.587 * Color.green(pixel) + 0.114 * Color.blue(pixel)).toInt()
            val newLum = lut[lum.coerceIn(0, 255)].toFloat()

            val ratio = if (lum > 0) newLum / lum.toFloat() else 1f
            val r = (Color.red(pixel) * ratio).toInt().coerceIn(0, 255)
            val g = (Color.green(pixel) * ratio).toInt().coerceIn(0, 255)
            val b = (Color.blue(pixel) * ratio).toInt().coerceIn(0, 255)

            pixels[i] = Color.argb(Color.alpha(pixel), r, g, b)
        }

        result.setPixels(pixels, 0, width, 0, 0, width, height)
        return result
    }

    fun applyDenoise(bitmap: Bitmap, strength: Int): Bitmap {
        val radius = strength.coerceIn(1, 5)
        val width = bitmap.width
        val height = bitmap.height
        val result = Bitmap.createBitmap(width, height, bitmap.config ?: Bitmap.Config.ARGB_8888)

        val pixels = IntArray(width * height)
        bitmap.getPixels(pixels, 0, width, 0, 0, width, height)
        val outPixels = IntArray(width * height)

        for (y in 0 until height) {
            for (x in 0 until width) {
                var sumR = 0
                var sumG = 0
                var sumB = 0
                var count = 0

                for (dy in -radius..radius) {
                    for (dx in -radius..radius) {
                        val nx = x + dx
                        val ny = y + dy
                        if (nx in 0 until width && ny in 0 until height) {
                            val idx = ny * width + nx
                            sumR += Color.red(pixels[idx])
                            sumG += Color.green(pixels[idx])
                            sumB += Color.blue(pixels[idx])
                            count++
                        }
                    }
                }

                val idx = y * width + x
                val alpha = Color.alpha(pixels[idx])
                outPixels[idx] = Color.argb(
                    alpha,
                    (sumR / count).coerceIn(0, 255),
                    (sumG / count).coerceIn(0, 255),
                    (sumB / count).coerceIn(0, 255)
                )
            }
        }

        result.setPixels(outPixels, 0, width, 0, 0, width, height)
        return result
    }

    fun applySharpen(bitmap: Bitmap, strength: Int): Bitmap {
        val width = bitmap.width
        val height = bitmap.height
        val result = Bitmap.createBitmap(width, height, bitmap.config ?: Bitmap.Config.ARGB_8888)

        val pixels = IntArray(width * height)
        bitmap.getPixels(pixels, 0, width, 0, 0, width, height)
        val outPixels = IntArray(width * height)

        val kernelScale = strength.coerceIn(1, 5).toFloat()
        val kernelCenter = 1f + 4f * kernelScale
        val kernelNeighbor = -kernelScale

        for (y in 1 until height - 1) {
            for (x in 1 until width - 1) {
                var sumR = 0f
                var sumG = 0f
                var sumB = 0f

                for (ky in -1..1) {
                    for (kx in -1..1) {
                        val idx = (y + ky) * width + (x + kx)
                        val weight = if (kx == 0 && ky == 0) kernelCenter else kernelNeighbor
                        sumR += Color.red(pixels[idx]) * weight
                        sumG += Color.green(pixels[idx]) * weight
                        sumB += Color.blue(pixels[idx]) * weight
                    }
                }

                val idx = y * width + x
                val alpha = Color.alpha(pixels[idx])
                outPixels[idx] = Color.argb(
                    alpha,
                    sumR.toInt().coerceIn(0, 255),
                    sumG.toInt().coerceIn(0, 255),
                    sumB.toInt().coerceIn(0, 255)
                )
            }
        }

        for (y in 0 until height) {
            for (x in 0 until width) {
                if (y == 0 || y == height - 1 || x == 0 || x == width - 1) {
                    outPixels[y * width + x] = pixels[y * width + x]
                }
            }
        }

        result.setPixels(outPixels, 0, width, 0, 0, width, height)
        return result
    }

    fun calculateAvgBrightness(bitmap: Bitmap): Int {
        val width = bitmap.width
        val height = bitmap.height
        if (width == 0 || height == 0) return 0

        val sampleStep = 4
        var totalBrightness = 0.0
        var samples = 0

        for (y in 0 until height step sampleStep) {
            for (x in 0 until width step sampleStep) {
                val pixel = bitmap.getPixel(x, y)
                val lum = 0.299 * Color.red(pixel) + 0.587 * Color.green(pixel) + 0.114 * Color.blue(pixel)
                totalBrightness += lum
                samples++
            }
        }

        return if (samples > 0) (totalBrightness / samples).toInt() else 0
    }

    fun calculateImageQuality(bitmap: Bitmap): Float {
        val brightness = calculateAvgBrightness(bitmap).toFloat()

        val brightnessScore = when {
            brightness < 50f -> brightness / 50f
            brightness > 200f -> (255f - brightness) / 55f
            else -> 1.0f
        }

        val contrastScore = calculateContrastScore(bitmap)

        val quality = 0.4f * brightnessScore + 0.6f * contrastScore
        return quality.coerceIn(0.05f, 1.0f)
    }

    private fun calculateContrastScore(bitmap: Bitmap): Float {
        val width = bitmap.width
        val height = bitmap.height
        if (width < 10 || height < 10) return 0.5f

        val sampleStep = 8
        var edgeCount = 0
        var totalSamples = 0

        val pixels = IntArray(width * height)
        bitmap.getPixels(pixels, 0, width, 0, 0, width, height)

        for (y in sampleStep until height - sampleStep step sampleStep) {
            for (x in sampleStep until width - sampleStep step sampleStep) {
                val idx = y * width + x
                val idxLeft = idx - sampleStep
                val idxUp = idx - sampleStep * width

                val lum1 = luminance(pixels[idx])
                val lum2 = luminance(pixels[idxLeft])
                val lum3 = luminance(pixels[idxUp])

                val diff1 = abs(lum1 - lum2)
                val diff2 = abs(lum1 - lum3)

                if (diff1 > 20 || diff2 > 20) {
                    edgeCount++
                }
                totalSamples++
            }
        }

        return if (totalSamples > 0) {
            (edgeCount.toFloat() / totalSamples * 3f).coerceAtMost(1.0f)
        } else 0.5f
    }

    private fun luminance(pixel: Int): Float {
        return (0.299f * Color.red(pixel) + 0.587f * Color.green(pixel) + 0.114f * Color.blue(pixel))
    }

    fun estimateLightLevel(bitmap: Bitmap): Float {
        val avgBrightness = calculateAvgBrightness(bitmap)
        return when {
            avgBrightness < 20 -> 5f
            avgBrightness < 50 -> 20f
            avgBrightness < 80 -> 40f
            avgBrightness < 120 -> 80f
            avgBrightness < 160 -> 150f
            avgBrightness < 200 -> 250f
            else -> 500f
        }
    }
}
