package com.ddg.driver.ui.navigation

import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import com.ddg.driver.data.repository.AuthRepository
import com.ddg.driver.ui.screen.AlarmScreen
import com.ddg.driver.ui.screen.CheckInScreen
import com.ddg.driver.ui.screen.HomeScreen
import com.ddg.driver.ui.screen.LoginScreen
import com.ddg.driver.ui.screen.NavigationScreen
import com.ddg.driver.ui.screen.ProfileScreen
import com.ddg.driver.ui.screen.RestCountdownScreen
import com.ddg.driver.ui.screen.ReviewScreen
import org.koin.compose.getKoin

sealed class Screen(val route: String) {
    object Login : Screen("login")
    object Home : Screen("home")
    object Navigation : Screen("navigation")
    object Alarm : Screen("alarm")
    object Profile : Screen("profile")
    object RestCountdown : Screen("rest_countdown")
    object CheckIn : Screen("check_in")
    object Review : Screen("review")
}

data class DriverContext(
    val driverId: Long = 1,
    val vehicleId: Long = 1,
    val waybillId: Long = 0,
    val latitude: Double = 0.0,
    val longitude: Double = 0.0,
    val hazardClass: String = "",
    val selectedServiceAreaId: Long = 0
)

@Composable
fun NavGraph() {
    val authRepository: AuthRepository = getKoin().get()
    val isLoggedIn by authRepository.isLoggedIn.collectAsState(initial = false)

    var currentScreen by remember { mutableStateOf<Screen>(if (isLoggedIn) Screen.Home else Screen.Login) }
    var driverContext by remember { mutableStateOf(DriverContext()) }

    when (currentScreen) {
        is Screen.Login -> LoginScreen(
            onLoginSuccess = { currentScreen = Screen.Home }
        )
        is Screen.Home -> HomeScreen(
            onStartNavigation = { currentScreen = Screen.Navigation },
            onViewAlarms = { currentScreen = Screen.Alarm },
            onViewProfile = { currentScreen = Screen.Profile },
            onViewRestCountdown = {
                currentScreen = Screen.RestCountdown
            },
            onAutoRecommend = { ctx ->
                driverContext = ctx
                currentScreen = Screen.RestCountdown
            }
        )
        is Screen.Navigation -> NavigationScreen(
            onBack = { currentScreen = Screen.Home }
        )
        is Screen.Alarm -> AlarmScreen(
            onBack = { currentScreen = Screen.Home }
        )
        is Screen.Profile -> ProfileScreen(
            onBack = { currentScreen = Screen.Home },
            onLogout = { currentScreen = Screen.Login }
        )
        is Screen.RestCountdown -> RestCountdownScreen(
            driverId = driverContext.driverId,
            vehicleId = driverContext.vehicleId,
            waybillId = driverContext.waybillId,
            latitude = driverContext.latitude,
            longitude = driverContext.longitude,
            onNavigateToCheckIn = { dId, vId, wId, saId ->
                driverContext = driverContext.copy(selectedServiceAreaId = saId)
                currentScreen = Screen.CheckIn
            },
            onBack = { currentScreen = Screen.Home }
        )
        is Screen.CheckIn -> CheckInScreen(
            driverId = driverContext.driverId,
            vehicleId = driverContext.vehicleId,
            waybillId = driverContext.waybillId,
            latitude = driverContext.latitude,
            longitude = driverContext.longitude,
            preselectedServiceAreaId = driverContext.selectedServiceAreaId,
            onCheckInSuccess = {},
            onCheckOutSuccess = { currentScreen = Screen.Home },
            onNavigateToReview = { saId ->
                driverContext = driverContext.copy(selectedServiceAreaId = saId)
                currentScreen = Screen.Review
            },
            onBack = { currentScreen = Screen.Home }
        )
        is Screen.Review -> ReviewScreen(
            serviceAreaId = driverContext.selectedServiceAreaId,
            driverId = driverContext.driverId,
            vehicleId = driverContext.vehicleId,
            waybillId = driverContext.waybillId,
            onSubmitSuccess = { currentScreen = Screen.Home },
            onBack = { currentScreen = Screen.CheckIn }
        )
    }
}
