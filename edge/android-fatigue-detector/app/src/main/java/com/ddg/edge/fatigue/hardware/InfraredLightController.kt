package com.ddg.edge.fatigue.hardware

import android.content.Context
import android.hardware.Sensor
import android.hardware.SensorEvent
import android.hardware.SensorEventListener
import android.hardware.SensorManager
import android.util.Log
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlin.math.max
import kotlin.math.min

data class InfraredLightState(
    val isOn: Boolean = false,
    val intensity: Int = 50,
    val isAutoMode: Boolean = true,
    val lightLevelLux: Float = 0f,
    val triggerReason: String = "",
    val lastChangedAt: Long = System.currentTimeMillis()
)

class InfraredLightController(
    private val context: Context,
    private val coroutineScope: CoroutineScope = CoroutineScope(Dispatchers.Default + SupervisorJob())
) {
    companion object {
        private const val TAG = "InfraredLightCtrl"
        private const val LOW_LIGHT_THRESHOLD_LUX = 50f
        private const val HIGH_LIGHT_THRESHOLD_LUX = 200f
        private const val MIN_INTENSITY = 20
        private const val MAX_INTENSITY = 100
        private const val AUTO_INTENSITY_MIN = 30
        private const val AUTO_INTENSITY_MAX = 80
    }

    private val _state = MutableStateFlow(InfraredLightState())
    val state: StateFlow<InfraredLightState> = _state.asStateFlow()

    private var sensorManager: SensorManager? = null
    private var lightSensor: Sensor? = null
    private var isHardwareAvailable = false
    private var autoIntensityEnabled = true
    private var lowLightThreshold = LOW_LIGHT_THRESHOLD_LUX
    private var highLightThreshold = HIGH_LIGHT_THRESHOLD_LUX

    private val sensorListener = object : SensorEventListener {
        override fun onSensorChanged(event: SensorEvent?) {
            event ?: return
            if (event.sensor.type == Sensor.TYPE_LIGHT) {
                val lux = event.values[0]
                handleLightLevelChange(lux)
            }
        }

        override fun onAccuracyChanged(sensor: Sensor?, accuracy: Int) {}
    }

    private var stateChangeListener: ((InfraredLightState) -> Unit)? = null

    fun initialize(): Result<Unit> {
        return runCatching {
            sensorManager = context.getSystemService(Context.SENSOR_SERVICE) as? SensorManager
            lightSensor = sensorManager?.getDefaultSensor(Sensor.TYPE_LIGHT)

            if (lightSensor != null) {
                sensorManager?.registerListener(
                    sensorListener,
                    lightSensor,
                    SensorManager.SENSOR_DELAY_NORMAL
                )
                Log.d(TAG, "Light sensor registered")
            } else {
                Log.w(TAG, "No light sensor available, using manual mode")
            }

            isHardwareAvailable = checkHardwareAvailability()
            if (isHardwareAvailable) {
                Log.d(TAG, "Infrared light hardware available")
            } else {
                Log.w(TAG, "Infrared light hardware not available (simulated mode)")
            }

            Unit
        }
    }

    private fun checkHardwareAvailability(): Boolean {
        return try {
            val pm = context.packageManager
            pm.hasSystemFeature("android.hardware.camera.flash") ||
                    pm.hasSystemFeature("com.ddg.hardware.infrared")
        } catch (e: Exception) {
            false
        }
    }

    fun setAutoMode(enabled: Boolean) {
        val current = _state.value
        if (current.isAutoMode == enabled) return

        _state.value = current.copy(
            isAutoMode = enabled,
            lastChangedAt = System.currentTimeMillis()
        )
        Log.d(TAG, "Auto mode set to: $enabled")

        if (enabled) {
            applyAutoControl(current.lightLevelLux)
        }
    }

    fun turnOn(reason: String = "manual") {
        if (_state.value.isOn) return

        val success = controlHardware(true, _state.value.intensity)
        if (success || !isHardwareAvailable) {
            val newState = _state.value.copy(
                isOn = true,
                triggerReason = reason,
                lastChangedAt = System.currentTimeMillis()
            )
            _state.value = newState
            stateChangeListener?.invoke(newState)
            Log.d(TAG, "Infrared light turned on: $reason")
        }
    }

    fun turnOff(reason: String = "manual") {
        if (!_state.value.isOn) return

        val success = controlHardware(false, 0)
        if (success || !isHardwareAvailable) {
            val newState = _state.value.copy(
                isOn = false,
                triggerReason = reason,
                lastChangedAt = System.currentTimeMillis()
            )
            _state.value = newState
            stateChangeListener?.invoke(newState)
            Log.d(TAG, "Infrared light turned off: $reason")
        }
    }

    fun setIntensity(intensity: Int, reason: String = "manual") {
        val clamped = intensity.coerceIn(MIN_INTENSITY, MAX_INTENSITY)
        val current = _state.value
        if (current.intensity == clamped && current.isOn) return

        if (current.isOn) {
            val success = controlHardware(true, clamped)
            if (success || !isHardwareAvailable) {
                val newState = current.copy(
                    intensity = clamped,
                    triggerReason = reason,
                    lastChangedAt = System.currentTimeMillis()
                )
                _state.value = newState
                stateChangeListener?.invoke(newState)
                Log.d(TAG, "Infrared intensity set to: $clamped")
            }
        } else {
            _state.value = current.copy(intensity = clamped)
        }
    }

    fun setThresholds(low: Float, high: Float) {
        lowLightThreshold = low
        highLightThreshold = high
    }

    fun setAutoIntensityEnabled(enabled: Boolean) {
        autoIntensityEnabled = enabled
    }

    private fun handleLightLevelChange(lux: Float) {
        val current = _state.value
        _state.value = current.copy(lightLevelLux = lux)

        if (current.isAutoMode) {
            applyAutoControl(lux)
        }
    }

    private fun applyAutoControl(lux: Float) {
        val current = _state.value

        when {
            lux <= lowLightThreshold && !current.isOn -> {
                turnOn("auto_low_light")
                if (autoIntensityEnabled) {
                    val autoIntensity = calculateAutoIntensity(lux)
                    setIntensity(autoIntensity, "auto_adjust")
                }
            }
            lux >= highLightThreshold && current.isOn -> {
                turnOff("auto_high_light")
            }
            current.isOn && autoIntensityEnabled -> {
                val targetIntensity = calculateAutoIntensity(lux)
                if (kotlin.math.abs(targetIntensity - current.intensity) >= 5) {
                    setIntensity(targetIntensity, "auto_adjust")
                }
            }
        }
    }

    private fun calculateAutoIntensity(lux: Float): Int {
        if (lux <= 0f) return AUTO_INTENSITY_MAX

        val ratio = 1f - (lux / highLightThreshold).coerceIn(0f, 1f)
        val intensity = AUTO_INTENSITY_MIN + ratio * (AUTO_INTENSITY_MAX - AUTO_INTENSITY_MIN)
        return intensity.toInt().coerceIn(AUTO_INTENSITY_MIN, AUTO_INTENSITY_MAX)
    }

    private fun controlHardware(on: Boolean, intensity: Int): Boolean {
        return try {
            if (isHardwareAvailable) {
                true
            } else {
                true
            }
        } catch (e: Exception) {
            Log.e(TAG, "Hardware control failed", e)
            false
        }
    }

    fun isNightTime(hour: Int, nightStartHour: Int = 19, nightEndHour: Int = 6): Boolean {
        return if (nightStartHour < nightEndHour) {
            hour in nightStartHour until nightEndHour
        } else {
            hour >= nightStartHour || hour < nightEndHour
        }
    }

    fun shouldUseNightEnhancement(lux: Float, hour: Int): Boolean {
        return lux < lowLightThreshold * 2 || isNightTime(hour)
    }

    fun setStateChangeListener(listener: (InfraredLightState) -> Unit) {
        this.stateChangeListener = listener
    }

    fun getCurrentIntensity(): Int = _state.value.intensity
    fun isLightOn(): Boolean = _state.value.isOn
    fun getLightLevelLux(): Float = _state.value.lightLevelLux
    fun isAutoMode(): Boolean = _state.value.isAutoMode

    fun release() {
        try {
            sensorManager?.unregisterListener(sensorListener)
            turnOff("release")
            coroutineScope.cancel()
        } catch (e: Exception) {
            Log.e(TAG, "Release failed", e)
        }
    }
}
