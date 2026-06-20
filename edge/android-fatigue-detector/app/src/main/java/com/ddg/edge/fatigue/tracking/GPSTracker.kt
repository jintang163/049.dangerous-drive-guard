package com.ddg.edge.fatigue.tracking

import android.Manifest
import android.annotation.SuppressLint
import android.content.Context
import android.content.pm.PackageManager
import android.location.GnssStatus
import android.location.Location
import android.location.LocationListener
import android.location.LocationManager
import android.os.Build
import android.os.Bundle
import androidx.core.app.ActivityCompat
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

class GPSTracker(
    private val context: Context,
    private val minTimeMs: Long = 3000L,
    private val minDistanceM: Float = 1f
) {

    private val locationManager: LocationManager by lazy {
        context.getSystemService(Context.LOCATION_SERVICE) as LocationManager
    }

    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Default)
    private var reportJob: Job? = null

    private val _currentLocation = MutableStateFlow<LocationData?>(null)
    val currentLocation: StateFlow<LocationData?> = _currentLocation.asStateFlow()

    private val _locationUpdates = Channel<LocationData>(Channel.CONFLATED)
    val locationFlow = _locationUpdates.receiveAsFlow()

    private val _satelliteCount = MutableStateFlow(0)
    val satelliteCount: StateFlow<Int> = _satelliteCount.asStateFlow()

    private val _isTracking = MutableStateFlow(false)
    val isTracking: StateFlow<Boolean> = _isTracking.asStateFlow()

    private val _gpsEnabled = MutableStateFlow(false)
    val gpsEnabled: StateFlow<Boolean> = _gpsEnabled.asStateFlow()

    private val _usedProviders = MutableStateFlow<Set<String>>(emptySet())
    val usedProviders: StateFlow<Set<String>> = _usedProviders.asStateFlow()

    data class LocationData(
        val latitude: Double,
        val longitude: Double,
        val altitude: Double?,
        val speed: Float?,
        val bearing: Float?,
        val accuracy: Float?,
        val timestamp: Long,
        val provider: String
    )

    private val gpsLocationListener = object : LocationListener {
        override fun onLocationChanged(location: Location) {
            handleLocationUpdate(location, LocationManager.GPS_PROVIDER)
        }

        override fun onProviderEnabled(provider: String) {
            updateProviders()
        }

        override fun onProviderDisabled(provider: String) {
            updateProviders()
        }

        @Deprecated("Deprecated in Java")
        override fun onStatusChanged(provider: String?, status: Int, extras: Bundle?) {
        }
    }

    private val networkLocationListener = object : LocationListener {
        override fun onLocationChanged(location: Location) {
            handleLocationUpdate(location, LocationManager.NETWORK_PROVIDER)
        }

        override fun onProviderEnabled(provider: String) {
            updateProviders()
        }

        override fun onProviderDisabled(provider: String) {
            updateProviders()
        }

        @Deprecated("Deprecated in Java")
        override fun onStatusChanged(provider: String?, status: Int, extras: Bundle?) {
        }
    }

    private val gnssStatusCallback = object : GnssStatus.Callback() {
        override fun onSatelliteStatusChanged(status: GnssStatus) {
            var count = 0
            for (i in 0 until status.satelliteCount) {
                if (status.usedInFix(i)) count++
            }
            _satelliteCount.value = count
        }

        override fun onFirstFix(ttffMillis: Int) {
            super.onFirstFix(ttffMillis)
        }

        override fun onStarted() {
            super.onStarted()
        }

        override fun onStopped() {
            super.onStopped()
            _satelliteCount.value = 0
        }
    }

    @SuppressLint("MissingPermission")
    suspend fun startTracking(): Boolean = withContext(Dispatchers.IO) {
        if (!hasPermissions()) return@withContext false
        if (_isTracking.value) return@withContext true

        _gpsEnabled.value = isGpsProviderEnabled()

        updateProviders()

        runCatching {
            if (locationManager.isProviderEnabled(LocationManager.GPS_PROVIDER)) {
                locationManager.requestLocationUpdates(
                    LocationManager.GPS_PROVIDER,
                    minTimeMs,
                    minDistanceM,
                    gpsLocationListener,
                    android.os.Looper.getMainLooper()
                )
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N) {
                    locationManager.registerGnssStatusCallback(gnssStatusCallback)
                }
            }

            if (locationManager.isProviderEnabled(LocationManager.NETWORK_PROVIDER)) {
                locationManager.requestLocationUpdates(
                    LocationManager.NETWORK_PROVIDER,
                    minTimeMs,
                    minDistanceM,
                    networkLocationListener,
                    android.os.Looper.getMainLooper()
                )
            }

            getLastKnownLocation()?.let { last ->
                handleLocationUpdate(last, last.provider ?: "last")
            }

            _isTracking.value = true
            return@runCatching true
        }.getOrDefault(false)
    }

    @SuppressLint("MissingPermission")
    suspend fun stopTracking() = withContext(Dispatchers.IO) {
        runCatching {
            locationManager.removeUpdates(gpsLocationListener)
            locationManager.removeUpdates(networkLocationListener)
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N) {
                locationManager.unregisterGnssStatusCallback(gnssStatusCallback)
            }
        }
        _isTracking.value = false
        _satelliteCount.value = 0
    }

    fun startPeriodicReports(intervalMs: Long = 3000L) {
        stopPeriodicReports()
        reportJob = scope.launch {
            while (true) {
                val loc = _currentLocation.value
                if (loc != null) {
                    _locationUpdates.trySend(loc)
                }
                delay(intervalMs)
            }
        }
    }

    fun stopPeriodicReports() {
        reportJob?.cancel()
        reportJob = null
    }

    private fun handleLocationUpdate(location: Location, sourceProvider: String) {
        val data = LocationData(
            latitude = location.latitude,
            longitude = location.longitude,
            altitude = if (location.hasAltitude()) location.altitude else null,
            speed = if (location.hasSpeed()) location.speed else null,
            bearing = if (location.hasBearing()) location.bearing else null,
            accuracy = if (location.hasAccuracy()) location.accuracy else null,
            timestamp = location.time,
            provider = sourceProvider
        )
        _currentLocation.value = data
    }

    @SuppressLint("MissingPermission")
    private fun getLastKnownLocation(): Location? {
        val providers = listOf(
            LocationManager.GPS_PROVIDER,
            LocationManager.NETWORK_PROVIDER,
            LocationManager.PASSIVE_PROVIDER
        )

        var bestLocation: Location? = null
        for (provider in providers) {
            runCatching {
                val loc = locationManager.getLastKnownLocation(provider)
                if (loc != null && (bestLocation == null || loc.time > bestLocation!!.time)) {
                    bestLocation = loc
                }
            }
        }
        return bestLocation
    }

    private fun updateProviders() {
        val enabled = mutableSetOf<String>()
        runCatching {
            if (locationManager.isProviderEnabled(LocationManager.GPS_PROVIDER)) {
                enabled.add("GPS")
            }
            if (locationManager.isProviderEnabled(LocationManager.NETWORK_PROVIDER)) {
                enabled.add("NETWORK")
            }
        }
        _usedProviders.value = enabled
        _gpsEnabled.value = enabled.contains("GPS")
    }

    private fun isGpsProviderEnabled(): Boolean {
        return runCatching {
            locationManager.isProviderEnabled(LocationManager.GPS_PROVIDER)
        }.getOrDefault(false)
    }

    fun hasPermissions(): Boolean {
        return ActivityCompat.checkSelfPermission(
            context,
            Manifest.permission.ACCESS_FINE_LOCATION
        ) == PackageManager.PERMISSION_GRANTED ||
        ActivityCompat.checkSelfPermission(
            context,
            Manifest.permission.ACCESS_COARSE_LOCATION
        ) == PackageManager.PERMISSION_GRANTED
    }

    fun getCurrentLocationSnapshot(): LocationData? = _currentLocation.value
}
