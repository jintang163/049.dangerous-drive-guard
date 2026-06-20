package com.ddg.driver.ui.theme

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material.MaterialTheme
import androidx.compose.material.darkColors
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

private val DarkColorPalette = darkColors(
    primary = DDGRed,
    primaryVariant = DDGRedDark,
    secondary = DDGRedLight,
    background = DDGBackground,
    surface = DDGSurface,
    error = DDGDanger,
    onPrimary = DDGTextPrimary,
    onSecondary = DDGTextPrimary,
    onBackground = DDGTextPrimary,
    onSurface = DDGTextPrimary,
    onError = DDGTextPrimary
)

@Composable
fun DDGDarkTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit
) {
    val colors = DarkColorPalette

    MaterialTheme(
        colors = colors,
        typography = DDGTypography,
        content = content
    )
}
