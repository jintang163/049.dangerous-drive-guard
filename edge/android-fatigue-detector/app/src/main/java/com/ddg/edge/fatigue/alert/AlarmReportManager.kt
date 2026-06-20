package com.ddg.edge.fatigue.alert

import android.content.Context
import com.ddg.edge.fatigue.data.local.OfflineAlarmRepository
import com.ddg.edge.fatigue.data.remote.AlarmUploadRequest
import com.ddg.edge.fatigue.data.remote.BatchAlarmUploadRequest
import com.ddg.edge.fatigue.data.remote.PlatformApiService
import com.ddg.edge.fatigue.model.FatigueAlarm
import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.launchIn
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.onEach
import kotlinx.coroutines.flow.receiveAsFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import java.util.UUID
import java.util.concurrent.TimeUnit

class AlarmReportManager(
    private val context: Context,
    private val apiService: PlatformApiService,
    private val alarmRepository: OfflineAlarmRepository,
    private val ioDispatcher: CoroutineDispatcher = Dispatchers.IO
) {

    private val scope = CoroutineScope(SupervisorJob() + ioDispatcher)

    private var webSocket: WebSocket? = null
    private var wsClient: OkHttpClient? = null
    private var reconnectJob: Job? = null

    private val alarmChannel = Channel<FatigueAlarm>(Channel.UNLIMITED)
    private val frameChannel = Channel<FrameReport>(Channel.CONFLATED)

    private val _wsConnected = MutableStateFlow(false)
    val wsConnected: StateFlow<Boolean> = _wsConnected.asStateFlow()

    private val _reportEvents = Channel<ReportEvent>(Channel.UNLIMITED)
    val reportFlow = _reportEvents.receiveAsFlow()

    private val pendingAlarmIds = mutableSetOf<String>()

    data class FrameReport(
        val deviceId: String,
        val timestamp: Long,
        val score: Int,
        val alarmLevel: Int,
        val ear: Float,
        val mar: Float,
        val perclos: Float,
        val gpsLatitude: Double?,
        val gpsLongitude: Double?
    )

    sealed class ReportEvent {
        data class AlarmReported(val alarmId: String, val via: ReportChannel, val success: Boolean) : ReportEvent()
        data class WsMessageReceived(val message: String) : ReportEvent()
        data class WsStatusChanged(val connected: Boolean) : ReportEvent()
    }

    enum class ReportChannel { WS, HTTP, OFFLINE }

    private val deviceId: String by lazy {
        val prefs = context.getSharedPreferences("device_prefs", Context.MODE_PRIVATE)
        prefs.getString("device_id", null) ?: UUID.randomUUID().toString().also {
            prefs.edit().putString("device_id", it).apply()
        }
    }

    fun start() {
        scope.launch {
            processAlarmQueue()
        }
        scope.launch {
            processFrameQueue()
        }
    }

    suspend fun connectWebSocket(wsUrl: String) = withContext(ioDispatcher) {
        runCatching {
            wsClient = OkHttpClient.Builder()
                .pingInterval(30, TimeUnit.SECONDS)
                .connectTimeout(30, TimeUnit.SECONDS)
                .readTimeout(0, TimeUnit.SECONDS)
                .build()

            val request = Request.Builder()
                .url(wsUrl)
                .header("X-Device-ID", deviceId)
                .header("X-Client-Type", "ANDROID_EDGE")
                .build()

            webSocket = wsClient!!.newWebSocket(request, object : WebSocketListener() {
                override fun onOpen(webSocket: WebSocket, response: Response) {
                    _wsConnected.value = true
                    scope.launch { _reportEvents.send(ReportEvent.WsStatusChanged(true)) }
                }

                override fun onMessage(webSocket: WebSocket, text: String) {
                    scope.launch { _reportEvents.send(ReportEvent.WsMessageReceived(text)) }
                }

                override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
                    webSocket.close(code, reason)
                }

                override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                    _wsConnected.value = false
                    scope.launch {
                        _reportEvents.send(ReportEvent.WsStatusChanged(false))
                        scheduleReconnect(wsUrl)
                    }
                }

                override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                    _wsConnected.value = false
                    scope.launch {
                        _reportEvents.send(ReportEvent.WsStatusChanged(false))
                        scheduleReconnect(wsUrl)
                    }
                }
            })
        }
    }

    private fun scheduleReconnect(wsUrl: String) {
        reconnectJob?.cancel()
        reconnectJob = scope.launch {
            delay(5000)
            connectWebSocket(wsUrl)
        }
    }

    fun disconnectWebSocket() {
        reconnectJob?.cancel()
        webSocket?.close(1000, "Client closing")
        webSocket = null
        wsClient?.dispatcher?.executorService?.shutdown()
        wsClient = null
        _wsConnected.value = false
    }

    fun reportAlarm(alarm: FatigueAlarm) {
        alarmChannel.trySend(alarm)
    }

    fun reportFrame(frame: FrameReport) {
        frameChannel.trySend(frame)
    }

    private suspend fun processAlarmQueue() = withContext(ioDispatcher) {
        alarmChannel.receiveAsFlow()
            .onEach { alarm ->
                val reported = tryReportAlarmViaWs(alarm) || tryReportAlarmViaHttp(alarm)

                if (!reported) {
                    alarmRepository.saveAlarm(alarm)
                    _reportEvents.send(ReportEvent.AlarmReported(alarm.alarmId, ReportChannel.OFFLINE, true))
                }
            }
            .launchIn(scope)
    }

    private suspend fun processFrameQueue() = withContext(ioDispatcher) {
        frameChannel.receiveAsFlow()
            .onEach { frame ->
                if (_wsConnected.value) {
                    val frameJson = buildWsFrameMessage(frame)
                    webSocket?.send(frameJson)
                }
            }
            .launchIn(scope)
    }

    private suspend fun tryReportAlarmViaWs(alarm: FatigueAlarm): Boolean {
        if (!_wsConnected.value || webSocket == null) return false

        return runCatching {
            val wsMessage = buildWsAlarmMessage(alarm)
            val sent = webSocket!!.send(wsMessage)
            if (sent) {
                pendingAlarmIds.add(alarm.alarmId)
                _reportEvents.send(ReportEvent.AlarmReported(alarm.alarmId, ReportChannel.WS, true))
                alarmRepository.saveAlarm(alarm.copy(isUploaded = true))
            }
            sent
        }.getOrDefault(false)
    }

    private suspend fun tryReportAlarmViaHttp(alarm: FatigueAlarm): Boolean {
        return runCatching {
            val request = AlarmUploadRequest.fromAlarm(alarm, deviceId)
            val response = apiService.uploadAlarm(request)

            val success = response.isSuccessful && response.body()?.success == true
            if (success) {
                _reportEvents.send(ReportEvent.AlarmReported(alarm.alarmId, ReportChannel.HTTP, true))
                alarmRepository.saveAlarm(alarm.copy(isUploaded = true))
            }
            success
        }.getOrDefault(false)
    }

    suspend fun uploadOfflineAlarms(): Int = withContext(ioDispatcher) {
        val pending = alarmRepository.getPendingAlarms()
        if (pending.isEmpty()) return@withContext 0

        val triggersMap = pending.associate { stored ->
            stored.alarmId to alarmRepository.jsonToTriggers(stored.triggersJson).map { it.name }
        }

        val requests = pending.map { stored ->
            AlarmUploadRequest.fromStoredAlarm(
                stored, deviceId, triggersMap[stored.alarmId] ?: emptyList()
            )
        }

        runCatching {
            val response = apiService.uploadAlarmsBatch(
                BatchAlarmUploadRequest(deviceId = deviceId, alarms = requests)
            )

            if (response.isSuccessful && response.body()?.success == true) {
                val ids = pending.map { it.id }
                alarmRepository.markAsUploaded(ids)
                return@runCatching ids.size
            }
            return@runCatching 0
        }.getOrDefault(0)
    }

    private fun buildWsAlarmMessage(alarm: FatigueAlarm): String {
        val triggers = alarm.triggers.joinToString(",") { it.name }
        return """
            {
              "type":"alarm",
              "deviceId":"$deviceId",
              "alarmId":"${alarm.alarmId}",
              "timestamp":${alarm.timestamp},
              "level":${alarm.level.value},
              "score":${alarm.score},
              "triggers":"$triggers",
              "ear":${alarm.ear},
              "mar":${alarm.mar},
              "perclos":${alarm.perclos},
              "headPitch":${alarm.headPitch},
              "headYaw":${alarm.headYaw},
              "gazeAngle":${alarm.gazeAngle},
              "isYawning":${alarm.isYawning},
              "gpsLatitude":${alarm.gpsLatitude},
              "gpsLongitude":${alarm.gpsLongitude},
              "gpsSpeed":${alarm.gpsSpeed}
            }
        """.trimIndent()
    }

    private fun buildWsFrameMessage(frame: FrameReport): String {
        return """
            {
              "type":"frame",
              "deviceId":"${frame.deviceId}",
              "timestamp":${frame.timestamp},
              "score":${frame.score},
              "alarmLevel":${frame.alarmLevel},
              "ear":${frame.ear},
              "mar":${frame.mar},
              "perclos":${frame.perclos},
              "gpsLatitude":${frame.gpsLatitude},
              "gpsLongitude":${frame.gpsLongitude}
            }
        """.trimIndent()
    }

    fun release() {
        disconnectWebSocket()
    }
}
