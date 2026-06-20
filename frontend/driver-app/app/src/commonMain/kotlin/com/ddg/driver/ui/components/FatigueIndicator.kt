package com.ddg.driver.ui.components

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.size
import androidx.compose.material.MaterialTheme
import androidx.compose.material.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.ddg.driver.data.model.FatigueLevel
import com.ddg.driver.ui.theme.DDGRed
import com.ddg.driver.ui.theme.DDGSuccess
import com.ddg.driver.ui.theme.DDGTextHint
import com.ddg.driver.ui.theme.DDGTextPrimary
import com.ddg.driver.ui.theme.DDGWarning

@Composable
fun FatigueIndicator(
    progress: Float,
    level: FatigueLevel,
    size: Dp = 80.dp,
    strokeWidth: Float = 8f
) {
    val indicatorColor = when (level) {
        FatigueLevel.NORMAL -> DDGSuccess
        FatigueLevel.WARNING -> DDGWarning
        FatigueLevel.DANGEROUS -> DDGRed
    }

    val percentageText = "${(progress * 100).toInt()}%"

    Box(
        modifier = Modifier.size(size),
        contentAlignment = Alignment.Center
    ) {
        Canvas(modifier = Modifier.size(size)) {
            val canvasSize = size.toPx()
            val stroke = strokeWidth.dp.toPx()

            drawCircle(
                color = DDGTextHint.copy(alpha = 0.2f),
                style = Stroke(width = stroke),
                radius = (canvasSize - stroke) / 2,
                center = Offset(canvasSize / 2, canvasSize / 2)
            )

            drawArc(
                color = indicatorColor,
                startAngle = -90f,
                sweepAngle = progress * 360f,
                useCenter = false,
                style = Stroke(width = stroke, cap = StrokeCap.Round),
                topLeft = Offset(stroke / 2, stroke / 2),
                size = Size(canvasSize - stroke, canvasSize - stroke)
            )
        }
        Text(
            text = percentageText,
            style = MaterialTheme.typography.h6,
            fontWeight = FontWeight.Bold,
            color = indicatorColor
        )
    }
}
