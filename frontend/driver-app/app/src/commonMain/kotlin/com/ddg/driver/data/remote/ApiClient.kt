package com.ddg.driver.data.remote

import com.ddg.driver.data.local.AppDataStore
import io.ktor.client.HttpClient
import io.ktor.client.plugins.HttpTimeout
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.plugins.logging.LogLevel
import io.ktor.client.plugins.logging.Logger
import io.ktor.client.plugins.logging.Logging
import io.ktor.client.plugins.logging.SIMPLE
import io.ktor.client.request.HttpRequestBuilder
import io.ktor.client.request.header
import io.ktor.client.statement.HttpReceivePipeline
import io.ktor.http.HttpHeaders
import io.ktor.serialization.kotlinx.json.json
import kotlinx.coroutines.flow.first
import kotlinx.serialization.json.Json

class ApiClient(private val dataStore: AppDataStore) {

    companion object {
        const val BASE_URL = "https://api.ddg.example.com/v1"
        const val WS_BASE_URL = "wss://api.ddg.example.com"
        const val CONNECT_TIMEOUT = 30_000L
        const val REQUEST_TIMEOUT = 60_000L
    }

    val client: HttpClient = HttpClient {
        install(HttpTimeout) {
            connectTimeoutMillis = CONNECT_TIMEOUT
            requestTimeoutMillis = REQUEST_TIMEOUT
        }

        install(ContentNegotiation) {
            json(Json {
                prettyPrint = true
                isLenient = true
                ignoreUnknownKeys = true
                encodeDefaults = true
            })
        }

        install(Logging) {
            logger = Logger.SIMPLE
            level = LogLevel.BODY
        }

        sendPipeline.intercept(HttpSendPipeline.Before) {
            context.addAuthHeader(context.builder)
        }
    }

    val wsClient: HttpClient = HttpClient {
        install(io.ktor.client.plugins.websocket.WebSockets) {
            pingInterval = 30_000
        }
        install(HttpTimeout) {
            connectTimeoutMillis = CONNECT_TIMEOUT
            requestTimeoutMillis = REQUEST_TIMEOUT
        }
        install(Logging) {
            logger = Logger.SIMPLE
            level = LogLevel.HEADERS
        }
    }

    private suspend fun addAuthHeader(builder: HttpRequestBuilder) {
        val token = dataStore.token.first()
        if (!token.isNullOrEmpty()) {
            builder.header(HttpHeaders.Authorization, "Bearer $token")
        }
    }

    private object HttpSendPipeline {
        val Before = io.ktor.client.plugins.api.sendPipeline.Before
    }
}
