package com.ddg.driver.ui.screen

import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.Button
import androidx.compose.material.ButtonDefaults
import androidx.compose.material.Card
import androidx.compose.material.CircularProgressIndicator
import androidx.compose.material.Icon
import androidx.compose.material.MaterialTheme
import androidx.compose.material.Text
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AccessTime
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.DirectionsCar
import androidx.compose.material.icons.filled.LocalParking
import androidx.compose.material.icons.filled.Warning
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.ddg.driver.data.model.RestCountdown
import com.ddg.driver.data.model.ServiceAreaRecommendation
import com.ddg.driver.domain.usecase.GetRestCountdownUseCase
import com.ddg.driver.domain.usecase.RecommendServiceAreaUseCase
import com.ddg.driver.ui.theme.DDGDarkGray
import com.ddg.driver.ui.theme.DDGGray
import com.ddg.driver.ui.theme.DDGRed
import com.ddg.driver.ui.theme.DDGSurface
import com.ddg.driver.ui.theme.DDGSuccess
import com.ddg.driver.ui.theme.DDGTextPrimary
import com.ddg.driver.ui.theme.DDGTextSecondary
import com.ddg.driver.ui.theme.DDGWarning
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import org.koin.compose.getKoin

@Composable
fun RestCountdownScreen(
    driverId: Long,
    vehicleId: Long,
    waybillId: Long,
    latitude: Double,
    longitude: Double,
    onNavigateToCheckIn: (Long, Long, Long, Long) -> Unit,
    onBack: () -> Unit
) {
    val getRestCountdownUseCase: GetRestCountdownUseCase = getKoin().get()
    val recommendServiceAreaUseCase: RecommendServiceAreaUseCase = getKoin().get()
    val scope = rememberCoroutineScope()

    var restCountdown by remember { mutableStateOf<RestCountdown?>(null) }
    var recommendation by remember { mutableStateOf<ServiceAreaRecommendation?>(null) }
    var isLoading by remember { mutableStateOf(true) }
    var errorMessage by remember { mutableStateOf<String?>(null) }
    var showOvertimeWarning by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) {
        scope.launch {
            val result = getRestCountdownUseCase(driverId, vehicleId)
            result.onSuccess {
                restCountdown = it
                isLoading = false
                if (it.is_overtime) {
                    showOvertimeWarning = true
                }
                if (it.remaining_drive_minutes <= 60 && it.remaining_drive_minutes > 0) {
                    val recResult = recommendServiceAreaUseCase(
                        com.ddg.driver.data.model.RecommendRequest(
                            driver_id = driverId,
                            vehicle_id = vehicleId,
                            waybill_id = waybillId,
                            latitude = latitude,
                            longitude = longitude,
                            fatigue_score = it.continuous_drive_minutes.toDouble()
                        )
                    )
                    recResult.onSuccess { rec -> recommendation = rec as? ServiceAreaRecommendation }
                }
            }.onFailure {
                errorMessage = it.message
                isLoading = false
            }
        }
    }

    LaunchedEffect(restCountdown) {
        while (restCountdown?.status == "driving" || restCountdown?.status == "resting") {
            delay(60_000)
            scope.launch {
                val result = getRestCountdownUseCase(driverId, vehicleId)
                result.onSuccess {
                    restCountdown = it
                    if (it.is_overtime && !showOvertimeWarning) {
                        showOvertimeWarning = true
                    }
                }
            }
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(DDGSurface)
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            TopBar(onBack = onBack)

            Spacer(modifier = Modifier.height(24.dp))

            if (isLoading) {
                CircularProgressIndicator(color = DDGRed)
            } else if (errorMessage != null) {
                Text(text = errorMessage ?: "", color = DDGRed)
            } else {
                restCountdown?.let { countdown ->
                    CountdownCircle(countdown = countdown)

                    Spacer(modifier = Modifier.height(24.dp))

                    StatusInfoCard(countdown = countdown)

                    Spacer(modifier = Modifier.height(16.dp))

                    if (showOvertimeWarning && countdown.is_overtime) {
                        OvertimeWarningCard(overtimeMinutes = countdown.overtime_minutes)
                        Spacer(modifier = Modifier.height(16.dp))
                    }

                    recommendation?.let { rec ->
                        RecommendationCard(recommendation = rec)
                        Spacer(modifier = Modifier.height(16.dp))
                    }

                    if (countdown.status == "resting") {
                        RestProgressCard(countdown = countdown)

                        Spacer(modifier = Modifier.height(16.dp))

                        if (countdown.can_continue_driving) {
                            Button(
                                onClick = { onBack() },
                                modifier = Modifier.fillMaxWidth(),
                                colors = ButtonDefaults.buttonColors(backgroundColor = DDGSuccess)
                            ) {
                                Icon(Icons.Default.CheckCircle, contentDescription = null)
                                Spacer(modifier = Modifier.width(8.dp))
                                Text("休息完毕，可继续驾驶", color = Color.White, fontWeight = FontWeight.Bold)
                            }
                        } else {
                            Card(
                                modifier = Modifier.fillMaxWidth(),
                                backgroundColor = DDGWarning.copy(alpha = 0.2f),
                                shape = RoundedCornerShape(12.dp)
                            ) {
                                Text(
                                    text = "休息不足 ${countdown.min_rest_required} 分钟，请继续等待",
                                    modifier = Modifier.padding(16.dp),
                                    color = DDGWarning,
                                    fontWeight = FontWeight.Bold
                                )
                            }
                        }
                    } else if (countdown.status == "driving" && countdown.remaining_drive_minutes <= 60) {
                        Button(
                            onClick = {
                                recommendation?.let { rec ->
                                    onNavigateToCheckIn(driverId, vehicleId, waybillId, rec.recommended_service_area_id)
                                }
                            },
                            modifier = Modifier.fillMaxWidth(),
                            colors = ButtonDefaults.buttonColors(backgroundColor = DDGWarning)
                        ) {
                            Icon(Icons.Default.LocalParking, contentDescription = null)
                            Spacer(modifier = Modifier.width(8.dp))
                            Text("前往推荐服务区", color = Color.White, fontWeight = FontWeight.Bold)
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun TopBar(onBack: () -> Unit) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(
            text = "←",
            style = MaterialTheme.typography.h4,
            color = DDGTextPrimary,
            modifier = Modifier.clickable { onBack() }
        )
        Spacer(modifier = Modifier.width(16.dp))
        Text(
            text = "休息倒计时",
            style = MaterialTheme.typography.h4,
            fontWeight = FontWeight.Bold,
            color = DDGTextPrimary
        )
    }
}

@Composable
private fun CountdownCircle(countdown: RestCountdown) {
    val progress = if (countdown.status == "driving") {
        countdown.continuous_drive_minutes.toFloat() / countdown.max_continuous_drive.toFloat()
    } else {
        countdown.rest_progress_percent.toFloat() / 100f
    }

    val animatedProgress by animateFloatAsState(
        targetValue = progress.coerceIn(0f, 1f),
        animationSpec = tween(durationMillis = 1000, easing = LinearEasing)
    )

    val color = when {
        countdown.is_overtime -> DDGRed
        countdown.status == "resting" -> DDGWarning
        countdown.remaining_drive_minutes <= 60 -> DDGWarning
        else -> DDGSuccess
    }

    Box(
        modifier = Modifier.size(220.dp),
        contentAlignment = Alignment.Center
    ) {
        Canvas(modifier = Modifier.size(220.dp)) {
            val canvasSize = size.width
            val stroke = 14.dp.toPx()

            drawCircle(
                color = DDGDarkGray,
                style = Stroke(width = stroke),
                radius = (canvasSize - stroke) / 2,
                center = Offset(canvasSize / 2, canvasSize / 2)
            )

            drawArc(
                color = color,
                startAngle = -90f,
                sweepAngle = animatedProgress * 360f,
                useCenter = false,
                style = Stroke(width = stroke, cap = StrokeCap.Round),
                topLeft = Offset(stroke / 2, stroke / 2),
                size = Size(canvasSize - stroke, canvasSize - stroke)
            )
        }

        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Text(
                text = if (countdown.status == "driving") {
                    "${countdown.remaining_drive_minutes}"
                } else {
                    "${countdown.current_rest_minutes}"
                },
                style = MaterialTheme.typography.h2,
                fontWeight = FontWeight.Bold,
                color = color
            )
            Text(
                text = if (countdown.status == "driving") "剩余可驾驶(分钟)" else "已休息(分钟)",
                style = MaterialTheme.typography.caption,
                color = DDGTextSecondary
            )
        }
    }
}

@Composable
private fun StatusInfoCard(countdown: RestCountdown) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        backgroundColor = DDGGray,
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                InfoItem(
                    icon = Icons.Default.DirectionsCar,
                    label = "驾驶状态",
                    value = when (countdown.status) {
                        "driving" -> "驾驶中"
                        "resting" -> "休息中"
                        else -> "已完成"
                    },
                    color = if (countdown.status == "driving") DDGSuccess else DDGWarning
                )
                InfoItem(
                    icon = Icons.Default.AccessTime,
                    label = "连续驾驶",
                    value = "${countdown.continuous_drive_minutes}分钟",
                    color = if (countdown.continuous_drive_minutes >= 240) DDGRed else DDGTextPrimary
                )
                InfoItem(
                    icon = Icons.Default.Warning,
                    label = "超时",
                    value = if (countdown.is_overtime) "${countdown.overtime_minutes}分钟" else "无",
                    color = if (countdown.is_overtime) DDGRed else DDGSuccess
                )
            }
        }
    }
}

@Composable
private fun InfoItem(icon: ImageVector, label: String, value: String, color: Color) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Icon(icon, contentDescription = label, tint = color, modifier = Modifier.size(24.dp))
        Spacer(modifier = Modifier.height(4.dp))
        Text(text = value, style = MaterialTheme.typography.body1, fontWeight = FontWeight.Bold, color = color)
        Text(text = label, style = MaterialTheme.typography.caption, color = DDGTextSecondary)
    }
}

@Composable
private fun OvertimeWarningCard(overtimeMinutes: Int) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        backgroundColor = DDGRed.copy(alpha = 0.15f),
        shape = RoundedCornerShape(12.dp)
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(Icons.Default.Warning, contentDescription = null, tint = DDGRed, modifier = Modifier.size(32.dp))
            Spacer(modifier = Modifier.width(12.dp))
            Column {
                Text("超时驾驶警告！", fontWeight = FontWeight.Bold, color = DDGRed, fontSize = 18.sp)
                Text("已超时 $overtimeMinutes 分钟，请立即寻找服务区休息", color = DDGRed)
            }
        }
    }
}

@Composable
private fun RecommendationCard(recommendation: ServiceAreaRecommendation) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        backgroundColor = DDGGray,
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text("智能推荐", style = MaterialTheme.typography.h6, fontWeight = FontWeight.Bold, color = DDGSuccess)
            Spacer(modifier = Modifier.height(8.dp))

            Text(
                recommendation.recommended_service_area_name,
                style = MaterialTheme.typography.h5,
                fontWeight = FontWeight.Bold,
                color = DDGTextPrimary
            )

            Spacer(modifier = Modifier.height(4.dp))
            Text(
                "距离 ${recommendation.distance_km}km · 预计 ${recommendation.estimated_arrival_minutes} 分钟到达",
                color = DDGTextSecondary
            )
            Text(
                "推荐理由：${recommendation.recommend_reason}",
                color = DDGWarning,
                style = MaterialTheme.typography.body2
            )

            if (recommendation.alternatives_array.isNotEmpty()) {
                Spacer(modifier = Modifier.height(8.dp))
                Text("备选：", style = MaterialTheme.typography.caption, color = DDGTextSecondary)
                recommendation.alternatives_array.forEach { alt ->
                    Text(
                        "  · ${alt.service_area_name} (${alt.distance_km}km)",
                        style = MaterialTheme.typography.caption,
                        color = DDGTextSecondary
                    )
                }
            }
        }
    }
}

@Composable
private fun RestProgressCard(countdown: RestCountdown) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        backgroundColor = DDGGray,
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text("休息进度", style = MaterialTheme.typography.h6, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(8.dp))

            val progress = (countdown.rest_progress_percent / 100f).coerceIn(0f, 1f)
            val animatedProgress by animateFloatAsState(
                targetValue = progress,
                animationSpec = tween(500)
            )

            Canvas(modifier = Modifier.fillMaxWidth().height(12.dp)) {
                drawRoundRect(
                    color = DDGDarkGray,
                    cornerRadius = androidx.compose.ui.geometry.CornerRadius(6.dp.toPx())
                )
                drawRoundRect(
                    color = if (countdown.can_continue_driving) DDGSuccess else DDGWarning,
                    cornerRadius = androidx.compose.ui.geometry.CornerRadius(6.dp.toPx()),
                    size = Size(size.width * animatedProgress, size.height)
                )
            }

            Spacer(modifier = Modifier.height(8.dp))
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    "已休息 ${countdown.current_rest_minutes} / ${countdown.min_rest_required} 分钟",
                    style = MaterialTheme.typography.caption,
                    color = DDGTextSecondary
                )
                Text(
                    if (countdown.can_continue_driving) "✓ 可以继续驾驶" else "请继续休息",
                    style = MaterialTheme.typography.caption,
                    color = if (countdown.can_continue_driving) DDGSuccess else DDGWarning
                )
            }

            if (countdown.current_service_area_name.isNotEmpty()) {
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    "当前服务区：${countdown.current_service_area_name}",
                    style = MaterialTheme.typography.caption,
                    color = DDGTextSecondary
                )
            }
        }
    }
}
