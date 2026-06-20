package com.ddg.edge.fatigue.alert

import android.content.Context
import android.media.AudioManager
import android.speech.tts.TextToSpeech
import com.ddg.edge.fatigue.R
import com.ddg.edge.fatigue.model.AlarmLevel
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.receiveAsFlow
import kotlinx.coroutines.withContext
import java.util.Locale

class VoiceAlertManager(
    private val context: Context,
    private val cooldownMs: Long = 3000L
) : TextToSpeech.OnInitListener {

    private var tts: TextToSpeech? = null
    private var isInitialized = false
    private var lastAlertTime: Long = 0L

    private val _isSpeaking = MutableStateFlow(false)
    val isSpeaking: StateFlow<Boolean> = _isSpeaking.asStateFlow()

    private val _isEnabled = MutableStateFlow(true)
    val isEnabled: StateFlow<Boolean> = _isEnabled.asStateFlow()

    private val alertChannel = Channel<String>(Channel.CONFLATED)
    val alertFlow = alertChannel.receiveAsFlow()

    enum class AlertType {
        REST, FATIGUE, POSTURE, GAZE
    }

    fun initialize() {
        tts = TextToSpeech(context.applicationContext, this)
    }

    override fun onInit(status: Int) {
        if (status == TextToSpeech.SUCCESS) {
            val result = tts?.setLanguage(Locale.SIMPLIFIED_CHINESE)
            if (result == TextToSpeech.LANG_MISSING_DATA || result == TextToSpeech.LANG_NOT_SUPPORTED) {
                tts?.setLanguage(Locale.CHINESE)
            }
            tts?.setPitch(1.0f)
            tts?.setSpeechRate(1.1f)
            isInitialized = true
        }
    }

    suspend fun speakAlert(type: AlertType, force: Boolean = false): Boolean {
        val text = when (type) {
            AlertType.REST -> context.getString(R.string.tts_rest)
            AlertType.FATIGUE -> context.getString(R.string.tts_fatigue)
            AlertType.POSTURE -> context.getString(R.string.tts_posture)
            AlertType.GAZE -> context.getString(R.string.tts_gaze)
        }
        return speak(text, force)
    }

    suspend fun speakForAlarmLevel(
        level: AlarmLevel,
        isYawning: Boolean,
        isHeadDown: Boolean,
        isGazeAway: Boolean
    ): Boolean {
        if (!_isEnabled.value) return false

        return when {
            level == AlarmLevel.LEVEL_3 -> speakAlert(AlertType.REST)
            level == AlarmLevel.LEVEL_2 -> speakAlert(AlertType.FATIGUE)
            isYawning && level >= AlarmLevel.LEVEL_1 -> speakAlert(AlertType.FATIGUE)
            isHeadDown -> speakAlert(AlertType.POSTURE)
            isGazeAway -> speakAlert(AlertType.GAZE)
            else -> false
        }
    }

    suspend fun speak(text: String, force: Boolean = false): Boolean = withContext(Dispatchers.Main) {
        if (!isInitialized || tts == null) return@withContext false
        if (!_isEnabled.value && !force) return@withContext false

        val now = System.currentTimeMillis()
        if (!force && now - lastAlertTime < cooldownMs) {
            return@withContext false
        }

        val params = hashMapOf<String, String>(
            TextToSpeech.Engine.KEY_PARAM_VOLUME to "1.0",
            TextToSpeech.Engine.KEY_PARAM_PAN to "0.0"
        )

        val result = tts!!.speak(text, TextToSpeech.QUEUE_FLUSH, params)
        if (result == TextToSpeech.SUCCESS) {
            lastAlertTime = now
            _isSpeaking.value = true
            alertChannel.trySend(text)
            monitorSpeech()
            return@withContext true
        }
        return@withContext false
    }

    private fun monitorSpeech() {
        Thread {
            try {
                while (tts?.isSpeaking == true) {
                    Thread.sleep(100)
                }
            } catch (e: Exception) {
            } finally {
                _isSpeaking.value = false
            }
        }.start()
    }

    fun stop() {
        tts?.stop()
        _isSpeaking.value = false
    }

    fun setEnabled(enabled: Boolean) {
        _isEnabled.value = enabled
        if (!enabled) stop()
    }

    fun setVolume(volume: Float) {
        val audioManager = context.getSystemService(Context.AUDIO_SERVICE) as AudioManager
        val maxVolume = audioManager.getStreamMaxVolume(AudioManager.STREAM_MUSIC)
        val targetVolume = (volume * maxVolume).toInt()
        audioManager.setStreamVolume(
            AudioManager.STREAM_MUSIC,
            targetVolume,
            0
        )
    }

    fun release() {
        stop()
        tts?.shutdown()
        tts = null
        isInitialized = false
    }
}
