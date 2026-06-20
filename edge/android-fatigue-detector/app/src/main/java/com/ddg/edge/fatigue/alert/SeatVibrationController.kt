package com.ddg.edge.fatigue.alert

import android.hardware.usb.UsbDevice
import android.hardware.usb.UsbDeviceConnection
import android.hardware.usb.UsbEndpoint
import android.hardware.usb.UsbInterface
import android.hardware.usb.UsbManager
import com.ddg.edge.fatigue.model.AlarmLevel
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.receiveAsFlow
import kotlinx.coroutines.withContext
import java.io.IOException

class SeatVibrationController(
    private val usbManager: UsbManager,
    private val targetVendorId: Int = 0x1A86,
    private val targetProductId: Int = 0x7523,
    private val baudRate: Int = 9600
) {

    private var usbConnection: UsbDeviceConnection? = null
    private var usbInterface: UsbInterface? = null
    private var endpointOut: UsbEndpoint? = null
    private var endpointIn: UsbEndpoint? = null
    private var isConnected = false

    private val _isVibrating = MutableStateFlow(false)
    val isVibrating: StateFlow<Boolean> = _isVibrating.asStateFlow()

    private val _isEnabled = MutableStateFlow(true)
    val isEnabled: StateFlow<Boolean> = _isEnabled.asStateFlow()

    private val _connectionState = MutableStateFlow(ConnectionState.DISCONNECTED)
    val connectionState: StateFlow<ConnectionState> = _connectionState.asStateFlow()

    private val vibrationChannel = Channel<VibrationEvent>(Channel.CONFLATED)
    val vibrationFlow = vibrationChannel.receiveAsFlow()

    enum class ConnectionState {
        DISCONNECTED, CONNECTING, CONNECTED, ERROR
    }

    data class VibrationEvent(
        val strength: Int,
        val durationMs: Long,
        val level: AlarmLevel,
        val timestamp: Long
    )

    data class VibrationPattern(
        val intensity: Int,
        val durationMs: Long,
        val pattern: IntArray,
        val repeat: Boolean = false
    )

    companion object {
        private const val CMD_START_VIBRATION: Byte = 0x01
        private const val CMD_STOP_VIBRATION: Byte = 0x02
        private const val CMD_SET_PWM: Byte = 0x03
        private const val CMD_GET_STATUS: Byte = 0x04

        private const val FRAME_HEADER: Byte = 0xAA.toByte()
        private const val FRAME_TAIL: Byte = 0x55
    }

    suspend fun connect(): Boolean = withContext(Dispatchers.IO) {
        if (isConnected) return@withContext true

        _connectionState.value = ConnectionState.CONNECTING

        runCatching {
            val device = findTargetDevice()
            if (device == null) {
                _connectionState.value = ConnectionState.DISCONNECTED
                return@runCatching false
            }

            val connection = usbManager.openDevice(device)
            if (connection == null) {
                _connectionState.value = ConnectionState.ERROR
                return@runCatching false
            }

            usbInterface = device.getInterface(0)
            if (!connection.claimInterface(usbInterface, true)) {
                connection.close()
                _connectionState.value = ConnectionState.ERROR
                return@runCatching false
            }

            for (i in 0 until usbInterface!!.endpointCount) {
                val endpoint = usbInterface!!.getEndpoint(i)
                if (endpoint.type == 2) {
                    if (endpoint.direction == 0) {
                        endpointOut = endpoint
                    } else {
                        endpointIn = endpoint
                    }
                }
            }

            if (endpointOut == null) {
                connection.releaseInterface(usbInterface)
                connection.close()
                _connectionState.value = ConnectionState.ERROR
                return@runCatching false
            }

            usbConnection = connection
            isConnected = true
            _connectionState.value = ConnectionState.CONNECTED
            initializeSerialPort()
            return@runCatching true
        }.getOrDefault(false)
    }

    private fun findTargetDevice(): UsbDevice? {
        val deviceList = usbManager.deviceList
        for (device in deviceList.values) {
            if (device.vendorId == targetVendorId && device.productId == targetProductId) {
                return device
            }
        }
        return deviceList.values.firstOrNull()
    }

    private fun initializeSerialPort() {
        val baudRateBytes = intArrayOf(
            baudRate and 0xFF,
            (baudRate shr 8) and 0xFF,
            (baudRate shr 16) and 0xFF,
            (baudRate shr 24) and 0xFF
        )
    }

    suspend fun vibrateForAlarmLevel(level: AlarmLevel): Boolean {
        if (!_isEnabled.value || !isConnected) return false

        val pattern = when (level) {
            AlarmLevel.LEVEL_1 -> VibrationPattern(
                intensity = 50,
                durationMs = 500,
                pattern = intArrayOf(200, 200, 200),
                repeat = false
            )
            AlarmLevel.LEVEL_2 -> VibrationPattern(
                intensity = 75,
                durationMs = 1000,
                pattern = intArrayOf(150, 100, 150, 100, 150),
                repeat = false
            )
            AlarmLevel.LEVEL_3 -> VibrationPattern(
                intensity = 100,
                durationMs = 2000,
                pattern = intArrayOf(100, 50, 100, 50, 100, 50, 100),
                repeat = true
            )
            else -> return false
        }

        return startVibration(pattern, level)
    }

    private suspend fun startVibration(
        pattern: VibrationPattern,
        level: AlarmLevel
    ): Boolean = withContext(Dispatchers.IO) {
        if (!isConnected || usbConnection == null || endpointOut == null) {
            return@withContext false
        }

        runCatching {
            setPwmDutyCycle(pattern.intensity)

            for ((index, duration) in pattern.pattern.withIndex()) {
                val cmd = if (index % 2 == 0) CMD_START_VIBRATION else CMD_STOP_VIBRATION
                sendCommand(cmd, duration)
                Thread.sleep(duration.toLong())
            }

            if (pattern.repeat) {
                vibrationChannel.trySend(
                    VibrationEvent(pattern.intensity, pattern.durationMs, level, System.currentTimeMillis())
                )
                _isVibrating.value = true
            } else {
                stopVibrationInternal()
            }

            return@runCatching true
        }.getOrDefault(false)
    }

    private fun setPwmDutyCycle(dutyCyclePercent: Int) {
        val duty = dutyCyclePercent.coerceIn(0, 100)
        sendCommand(CMD_SET_PWM, duty)
    }

    private fun sendCommand(command: Byte, data: Int): Boolean {
        if (!isConnected || usbConnection == null || endpointOut == null) return false

        return try {
            val dataLow = (data and 0xFF).toByte()
            val dataHigh = ((data shr 8) and 0xFF).toByte()
            val checksum = (FRAME_HEADER + command + dataLow + dataHigh).toByte()

            val frame = byteArrayOf(
                FRAME_HEADER,
                command,
                dataLow,
                dataHigh,
                checksum,
                FRAME_TAIL
            )

            val result = usbConnection!!.bulkTransfer(endpointOut, frame, frame.size, 1000)
            result == frame.size
        } catch (e: Exception) {
            false
        }
    }

    private suspend fun stopVibrationInternal() = withContext(Dispatchers.IO) {
        sendCommand(CMD_STOP_VIBRATION, 0)
        _isVibrating.value = false
    }

    suspend fun stopVibration() {
        stopVibrationInternal()
    }

    fun setEnabled(enabled: Boolean) {
        _isEnabled.value = enabled
        if (!enabled) {
            kotlinx.coroutines.runBlocking { stopVibration() }
        }
    }

    suspend fun disconnect() = withContext(Dispatchers.IO) {
        stopVibrationInternal()
        runCatching {
            if (usbConnection != null && usbInterface != null) {
                usbConnection!!.releaseInterface(usbInterface)
            }
            usbConnection?.close()
        }
        usbConnection = null
        usbInterface = null
        endpointOut = null
        endpointIn = null
        isConnected = false
        _connectionState.value = ConnectionState.DISCONNECTED
    }
}
