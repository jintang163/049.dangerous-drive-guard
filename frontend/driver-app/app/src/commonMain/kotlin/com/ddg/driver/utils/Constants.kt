package com.ddg.driver.utils

object Constants {
    const val BASE_URL = "https://api.ddg.example.com/v1"

    const val API_TIMEOUT_CONNECT = 30_000L
    const val API_TIMEOUT_REQUEST = 60_000L

    const val MAX_CONTINUOUS_DRIVING_HOURS = 4
    const val MIN_REST_MINUTES = 20

    const val HIGHWAY_SPEED_LIMIT = 100.0
    const val URBAN_SPEED_LIMIT = 60.0
    const val DANGEROUS_AREA_SPEED_LIMIT = 40.0

    const val FATIGUE_LEVEL_NORMAL_THRESHOLD = 0.3f
    const val FATIGUE_LEVEL_WARNING_THRESHOLD = 0.6f

    const val TRACK_UPLOAD_INTERVAL_SECONDS = 30L
    const val FATIGUE_DETECT_INTERVAL_SECONDS = 60L

    const val SOS_CONFIRM_DELAY_MS = 3000L

    const val MAP_ZOOM_LEVEL_DEFAULT = 14.0f

    const val DATASTORE_KEY_TOKEN = "auth_token"
    const val DATASTORE_KEY_USER = "user_info"
    const val DATASTORE_KEY_SETTINGS = "app_settings"
}
