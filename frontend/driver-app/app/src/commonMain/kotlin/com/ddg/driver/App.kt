package com.ddg.driver

import androidx.compose.material.MaterialTheme
import androidx.compose.runtime.Composable
import com.ddg.driver.di.appModule
import com.ddg.driver.ui.navigation.NavGraph
import com.ddg.driver.ui.theme.DDGDarkTheme
import org.koin.compose.KoinApplication

@Composable
fun App() {
    KoinApplication(application = {
        modules(appModule)
    }) {
        DDGDarkTheme {
            NavGraph()
        }
    }
}
