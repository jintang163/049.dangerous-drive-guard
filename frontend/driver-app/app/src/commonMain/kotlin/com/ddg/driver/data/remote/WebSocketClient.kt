package com.ddg.driver.data.remote

import com.ddg.driver.data.local.AppDataStore
import com.ddg.driver.data.model.ReplanAppliedInfo
import com.ddg.driver.data.model.ReplanSuggestion
import com.ddg.driver.utils.Constants
import io.ktor.client.HttpClient
import io.ktor.client.plugins.websocket.webSocketSession
import io.ktor.client.request.header
import io.ktor.http.ContentType
import io.ktor.http.HttpMethod
import io.ktor.websocket.Frame
import io.ktor.websocket.WebSocketSession
import io.ktor.websocket.close
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.channels.consumeEach
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.jsonPrimitive
import kotlinx.serialization.json.contentOrNull

class WebSocketClient(
    private val client: HttpClient,
    private val dataStore: AppDataStore
) {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Default)
    private var session: WebSocketSession? = null
    private val json = Json { ignoreUnknownKeys = true; isLenient = true }

    private val _connected = MutableStateFlow(false)
    val connected: StateFlow<Boolean> = _connected.asStateFlow()

    private val _replanSuggestions = MutableSharedFlow<ReplanSuggestion>(extraBufferCapacity = 8)
    val replanSuggestions: SharedFlow<ReplanSuggestion> = _replanSuggestions.asSharedFlow()

    private val _routeApplied = MutableSharedFlow<ReplanAppliedInfo>(extraBufferCapacity = 4)
    val routeApplied: SharedFlow<ReplanAppliedInfo> = _routeApplied.asSharedFlow()

    private val _trafficEvents = MutableSharedFlow<String>(extraBufferCapacity = 8)
    val trafficEvents: SharedFlow<String> = _trafficEvents.asSharedFlow()

    fun connect(vehicleId: Long? = null, driverId: Long? = null) {
        scope.launch {
            try {
                val token = dataStore.getToken() ?: ""
                val url = buildString {
                    append(ApiClient.WS_BASE_URL)
                    append("?token=").append(token)
                    if (vehicleId != null) append("&vehicle_id=").append(vehicleId)
                    if (driverId != null) append("&driver_id=").append(driverId)
                }

                session = client.webSocketSession(
                    method = HttpMethod.Get,
                    host = "localhost",
                    port = 8080,
                    path = "/ws"
                ) {
                    url("${ApiClient.WS_BASE_URL}/ws?token=$token&vehicle_id=${vehicleId ?: 0}&driver_id=${driverId ?: 0}")
                    header("Authorization", "Bearer $token")
                }

                _connected.value = true
                listenMessages()
            } catch (e: Exception) {
                _connected.value = false
            }
        }
    }

    private suspend fun listenMessages() {
        scope.launch {
            try {
                session?.incoming?.consumeEach { frame ->
                    if (frame is Frame.Text) {
                        val text = frame.readText()
                        dispatchMessage(text)
                    }
                }
            } catch (e: Exception) {
                _connected.value = false
            }
        }
    }

    private suspend fun dispatchMessage(raw: String) {
        runCatching {
            val obj = json.parseToJsonElement(raw) as JsonObject
            val type = obj["type"]?.jsonPrimitive?.contentOrNull ?: return
            val payload = obj["payload"]?.toString() ?: return

            when (type) {
                "route_replan_suggest" -> {
                    val suggestion = json.decodeFromString<ReplanSuggestion>(payload)
                    _replanSuggestions.emit(suggestion)
                }
                "route_applied" -> {
                    val applied = json.decodeFromString<ReplanAppliedInfo>(payload)
                    _routeApplied.emit(applied)
                }
                "traffic_event" -> {
                    _trafficEvents.emit(payload)
                }
                else -> {}
            }
        }
    }

    fun send(message: String) {
        scope.launch {
            runCatching {
                session?.send(Frame.Text(message))
            }
        }
    }

    fun disconnect() {
        scope.launch {
            runCatching {
                session?.close()
                session = null
                _connected.value = false
            }
        }
    }

    fun destroy() {
        disconnect()
        scope.cancel()
    }
}
