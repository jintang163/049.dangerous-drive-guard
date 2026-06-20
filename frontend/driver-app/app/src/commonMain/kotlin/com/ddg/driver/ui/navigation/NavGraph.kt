package com.ddg.driver.ui.navigation

import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import com.ddg.driver.data.repository.AuthRepository
import com.ddg.driver.ui.screen.AlarmScreen
import com.ddg.driver.ui.screen.HomeScreen
import com.ddg.driver.ui.screen.LoginScreen
import com.ddg.driver.ui.screen.NavigationScreen
import com.ddg.driver.ui.screen.ProfileScreen
import org.koin.compose.getKoin

sealed class Screen(val route: String) {
    object Login : Screen("login")
    object Home : Screen("home")
    object Navigation : Screen("navigation")
    object Alarm : Screen("alarm")
    object Profile : Screen("profile")
}

@Composable
fun NavGraph() {
    val authRepository: AuthRepository = getKoin().get()
    val isLoggedIn by authRepository.isLoggedIn.collectAsState(initial = false)

    var currentScreen by remember { mutableStateOf<Screen>(if (isLoggedIn) Screen.Home else Screen.Login) }

    when (currentScreen) {
        is Screen.Login -> LoginScreen(
            onLoginSuccess = { currentScreen = Screen.Home }
        )
        is Screen.Home -> HomeScreen(
            onStartNavigation = { currentScreen = Screen.Navigation },
            onViewAlarms = { currentScreen = Screen.Alarm },
            onViewProfile = { currentScreen = Screen.Profile }
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
    }
}
