package com.ddg.edge.fatigue.manager

import android.content.Context
import android.graphics.Bitmap
import android.util.Log
import com.ddg.edge.fatigue.enhancement.EnhanceConfig
import com.ddg.edge.fatigue.enhancement.EnhanceMode
import com.ddg.edge.fatigue.enhancement.EnhanceResult
import com.ddg.edge.fatigue.enhancement.ImageEnhancementProcessor
import com.ddg.edge.fatigue.hardware.InfraredLightController
import com.ddg.edge.fatigue.hardware.InfraredLightState
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.launchIn
import kotlinx.coroutines.flow.onEach
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import java.util.Calendar
import kotlin.math.abs

data class NightVisionConfig(
    val infraredEnabled: Boolean = true,
    val infraredAutoMode: Boolean = true,
    val infraredManualOn: Boolean = false,
    val infraredIntensity: Int = 50,
    val infraredIntensityAuto: Boolean = true,
    val lowLightThresholdLux: Float = 50f,
    val highLightThresholdLux: Float = 200f,

    val enhancementEnabled: Boolean = true,
    val enhanceMode: EnhanceMode = EnhanceMode.AUTO,
    val gammaValue: Float = 1.2f,
    val brightnessBoost: Int = 30,
    val contrastBoost: Int = 20,
    val histogramEqualization: Boolean = true,
    val claheEnabled: Boolean = true,
    val denoiseEnabled: Boolean = true,
    val denoiseStrength: Int = 3,
    val sharpenEnabled: Boolean = false,
    val sharpenStrength: Int = 2,

    val nightModeAuto: Boolean = true,
    val nightStartHour: Int = 19,
    val nightEndHour: Int = 6,
    val lowLightFaceDetect: Boolean = true,
    val minFaceConfidenceNight: Float = 0.4f
)

data class NightVisionState(
    val isNightTime: Boolean = false,
    val isLowLight: Boolean = false,
    val lightLevelLux: Float = 0f,
    val infraredOn: Boolean = false,
    val infraredIntensity: Int = 50,
    val enhancementActive: Boolean = false,
    val enhanceMode: EnhanceMode = EnhanceMode.AUTO,
    val qualityScoreBefore: Float = 0f,
    val qualityScoreAfter: Float = 0f,
    val faceDetectedBefore: Boolean = false,
    val faceDetectedAfter: Boolean = false,
    val faceConfidenceBefore: Float = 0f,
    val faceConfidenceAfter: Float = 0f
)

class NightVisionManager(
    private val context: Context,
    private val coroutineScope: CoroutineScope = CoroutineScope(Dispatchers.Default + SupervisorJob())
) {
    companion object {
        private const val TAG = "NightVisionManager"
        private const val MIN_ENHANCE_INTERVAL_MS = 500L
    }

    private var infraredController: InfraredLightController? = null
    private var enhancementProcessor: ImageEnhancementProcessor? = null
    private var config = NightVisionConfig()

    private val stateMutex = Mutex()
    private var currentState = NightVisionState()

    private var lastEnhanceTime = 0L
    private var lastEnhanceResult: EnhanceResult? = null

    private var stateChangeListener: ((NightVisionState) -> Unit)? = null
    private var infraredStateListener: ((InfraredLightState) -> Unit)? = null

    fun initialize(): Result<Unit> {
        return runCatching {
            infraredController = InfraredLightController(context, coroutineScope).apply {
                initialize().onFailure {
                    Log.w(TAG, "Infrared controller init failed: ${it.message}")
                }
                setStateChangeListener { state ->
                    onInfraredStateChanged(state)
                }
            }

            enhancementProcessor = ImageEnhancementProcessor()

            updateConfig(config)
            observeInfraredState()

            Log.d(TAG, "Night vision manager initialized")
            Unit
        }
    }

    fun updateConfig(newConfig: NightVisionConfig) {
        config = newConfig

        infraredController?.apply {
            setAutoMode(newConfig.infraredAutoMode)
            setThresholds(newConfig.lowLightThresholdLux, newConfig.highLightThresholdLux)
            setAutoIntensityEnabled(newConfig.infraredIntensityAuto)

            if (!newConfig.infraredEnabled) {
                turnOff("disabled")
            } else if (!newConfig.infraredAutoMode && newConfig.infraredManualOn) {
                turnOn("manual")
                setIntensity(newConfig.infraredIntensity, "manual")
            }
        }

        coroutineScope.launch {
            updateState { state ->
                state.copy(enhanceMode = newConfig.enhanceMode)
            }
        }

        Log.d(TAG, "Config updated: enhanceMode=${newConfig.enhanceMode}, " +
                "infraredEnabled=${newConfig.infraredEnabled}, " +
                "infraredAuto=${newConfig.infraredAutoMode}")
    }

    private fun observeInfraredState() {
        infraredController?.state?.onEach { irState ->
            updateState { state ->
                state.copy(
                    lightLevelLux = irState.lightLevelLux,
                    infraredOn = irState.isOn,
                    infraredIntensity = irState.intensity,
                    isLowLight = irState.lightLevelLux < config.lowLightThresholdLux
                )
            }
            infraredStateListener?.invoke(irState)
        }?.launchIn(coroutineScope)
    }

    private fun onInfraredStateChanged(state: InfraredLightState) {
        Log.d(TAG, "Infrared state changed: on=${state.isOn}, " +
                "intensity=${state.intensity}, reason=${state.triggerReason}")
    }

    suspend fun preprocessForDetection(bitmap: Bitmap): Bitmap {
        if (!config.enhancementEnabled) {
            return bitmap
        }

        val now = System.currentTimeMillis()
        if (now - lastEnhanceTime < MIN_ENHANCE_INTERVAL_MS && lastEnhanceResult != null) {
            return lastEnhanceResult!!.enhancedBitmap
        }

        val lightLevel = infraredController?.getLightLevelLux()
            ?: enhancementProcessor?.estimateLightLevel(bitmap) ?: 100f

        val currentHour = Calendar.getInstance().get(Calendar.HOUR_OF_DAY)
        val isNight = isNightTime(currentHour)

        val enhanceConfig = determineEnhanceConfig(lightLevel, isNight)

        val result = withContext(Dispatchers.Default) {
            enhancementProcessor?.enhance(bitmap, enhanceConfig)
        }

        if (result != null) {
            lastEnhanceResult = result
            lastEnhanceTime = now

            updateState { state ->
                state.copy(
                    enhancementActive = true,
                    enhanceMode = enhanceConfig.mode,
                    qualityScoreBefore = result.qualityScoreBefore,
                    qualityScoreAfter = result.qualityScoreAfter,
                    isLowLight = lightLevel < config.lowLightThresholdLux,
                    isNightTime = isNight
                )
            }

            Log.d(TAG, "Enhancement done: quality ${result.qualityScoreBefore.format(2)} -> " +
                    "${result.qualityScoreAfter.format(2)} " +
                    "(${if (result.qualityImprovement >= 0) "+" else ""}${result.qualityImprovement.format(1)}%), " +
                    "mode=${enhanceConfig.mode}, time=${result.processingTimeMs}ms")
        }

        return result?.enhancedBitmap ?: bitmap
    }

    private fun determineEnhanceConfig(lightLevel: Float, isNight: Boolean): EnhanceConfig {
        return when {
            config.enhanceMode != EnhanceMode.AUTO -> {
                buildManualConfig(config.enhanceMode)
            }
            isNight || lightLevel < config.lowLightThresholdLux * 2 -> {
                val mode = when {
                    infraredController?.isLightOn() == true -> EnhanceMode.INFRARED
                    lightLevel < 30f -> EnhanceMode.LOW_LIGHT
                    else -> EnhanceMode.NIGHT
                }
                ImageEnhancementProcessor.getConfigForMode(mode, lightLevel)
            }
            else -> {
                EnhanceConfig(
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

    private fun buildManualConfig(mode: EnhanceMode): EnhanceConfig {
        return EnhanceConfig(
            mode = mode,
            gamma = config.gammaValue,
            brightnessDelta = config.brightnessBoost,
            contrastDelta = config.contrastBoost,
            applyHistogramEq = config.histogramEqualization,
            applyDenoise = config.denoiseEnabled,
            denoiseStrength = config.denoiseStrength,
            applySharpen = config.sharpenEnabled,
            sharpenStrength = config.sharpenStrength
        )
    }

    fun isNightTime(hour: Int): Boolean {
        if (!config.nightModeAuto) return false
        return infraredController?.isNightTime(hour, config.nightStartHour, config.nightEndHour) ?: false
    }

    fun shouldEnhance(lightLevelLux: Float, hour: Int): Boolean {
        if (!config.enhancementEnabled) return false
        return lightLevelLux < config.lowLightThresholdLux * 2 || isNightTime(hour)
    }

    fun getMinFaceConfidence(): Float {
        val lightLevel = infraredController?.getLightLevelLux() ?: 500f
        return if (lightLevel < config.lowLightThresholdLux * 2 && config.lowLightFaceDetect) {
            config.minFaceConfidenceNight
        } else {
            0.5f
        }
    }

    fun onFaceDetectionResult(
        detectedBefore: Boolean,
        confidenceBefore: Float,
        detectedAfter: Boolean,
        confidenceAfter: Float
    ) {
        coroutineScope.launch {
            updateState { state ->
                state.copy(
                    faceDetectedBefore = detectedBefore,
                    faceConfidenceBefore = confidenceBefore,
                    faceDetectedAfter = detectedAfter,
                    faceConfidenceAfter = confidenceAfter
                )
            }
        }
    }

    fun turnOnInfrared(reason: String = "manual") {
        infraredController?.turnOn(reason)
    }

    fun turnOffInfrared(reason: String = "manual") {
        infraredController?.turnOff(reason)
    }

    fun setInfraredIntensity(intensity: Int, reason: String = "manual") {
        infraredController?.setIntensity(intensity, reason)
    }

    fun setInfraredAutoMode(enabled: Boolean) {
        infraredController?.setAutoMode(enabled)
    }

    fun getCurrentState(): NightVisionState = currentState

    fun getLightLevelLux(): Float = infraredController?.getLightLevelLux() ?: 0f

    fun isInfraredOn(): Boolean = infraredController?.isLightOn() ?: false

    fun setStateChangeListener(listener: (NightVisionState) -> Unit) {
        this.stateChangeListener = listener
    }

    fun setInfraredStateListener(listener: (InfraredLightState) -> Unit) {
        this.infraredStateListener = listener
    }

    fun getConfig(): NightVisionConfig = config

    private suspend fun updateState(transform: (NightVisionState) -> NightVisionState) {
        stateMutex.withLock {
            val newState = transform(currentState)
            if (newState != currentState) {
                currentState = newState
                stateChangeListener?.invoke(newState)
            }
        }
    }

    private fun Float.format(digits: Int): String {
        return "%.${digits}f".format(this)
    }

    fun release() {
        try {
            infraredController?.release()
            coroutineScope.cancel()
            Log.d(TAG, "Night vision manager released")
        } catch (e: Exception) {
            Log.e(TAG, "Release failed", e)
        }
    }
}
