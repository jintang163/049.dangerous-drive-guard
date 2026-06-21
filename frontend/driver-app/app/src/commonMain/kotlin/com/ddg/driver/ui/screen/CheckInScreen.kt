package com.ddg.driver.ui.screen

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
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.Button
import androidx.compose.material.ButtonDefaults
import androidx.compose.material.Card
import androidx.compose.material.CircularProgressIndicator
import androidx.compose.material.Icon
import androidx.compose.material.MaterialTheme
import androidx.compose.material.OutlinedButton
import androidx.compose.material.Text
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.ExitToApp
import androidx.compose.material.icons.filled.LocationOn
import androidx.compose.material.icons.filled.LocalParking
import androidx.compose.material.icons.filled.Security
import androidx.compose.material.icons.filled.Star
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.ddg.driver.data.model.CheckInRequest
import com.ddg.driver.data.model.CheckOutRequest
import com.ddg.driver.data.model.RestCountdown
import com.ddg.driver.data.model.ServiceArea
import com.ddg.driver.data.model.ServiceAreaRealtimeStatus
import com.ddg.driver.domain.usecase.CheckInUseCase
import com.ddg.driver.domain.usecase.CheckOutUseCase
import com.ddg.driver.domain.usecase.GetRestCountdownUseCase
import com.ddg.driver.ui.theme.DDGGray
import com.ddg.driver.ui.theme.DDGRed
import com.ddg.driver.ui.theme.DDGSurface
import com.ddg.driver.ui.theme.DDGSuccess
import com.ddg.driver.ui.theme.DDGTextPrimary
import com.ddg.driver.ui.theme.DDGTextSecondary
import com.ddg.driver.ui.theme.DDGWarning
import kotlinx.coroutines.launch
import org.koin.compose.getKoin

@Composable
fun CheckInScreen(
    driverId: Long,
    vehicleId: Long,
    waybillId: Long,
    latitude: Double,
    longitude: Double,
    preselectedServiceAreaId: Long,
    onCheckInSuccess: () -> Unit,
    onCheckOutSuccess: () -> Unit,
    onNavigateToReview: (Long) -> Unit,
    onBack: () -> Unit
) {
    val checkInUseCase: CheckInUseCase = getKoin().get()
    val checkOutUseCase: CheckOutUseCase = getKoin().get()
    val getRestCountdownUseCase: GetRestCountdownUseCase = getKoin().get()
    val scope = rememberCoroutineScope()

    var restCountdown by remember { mutableStateOf<RestCountdown?>(null) }
    var isLoading by remember { mutableStateOf(true) }
    var isProcessing by remember { mutableStateOf(false) }
    var errorMessage by remember { mutableStateOf<String?>(null) }
    var successMessage by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(Unit) {
        scope.launch {
            val result = getRestCountdownUseCase(driverId, vehicleId)
            result.onSuccess { restCountdown = it }
            isLoading = false
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(DDGSurface)
    ) {
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            item {
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
                        text = "服务区签到",
                        style = MaterialTheme.typography.h4,
                        fontWeight = FontWeight.Bold,
                        color = DDGTextPrimary
                    )
                }
            }

            item {
                if (isLoading) {
                    Box(modifier = Modifier.fillMaxWidth().height(100.dp), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator(color = DDGRed)
                    }
                } else if (errorMessage != null) {
                    Text(text = errorMessage ?: "", color = DDGRed)
                }
            }

            item {
                restCountdown?.let { countdown ->
                    CurrentStatusCard(countdown)
                }
            }

            item {
                if (restCountdown?.status == "driving") {
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        backgroundColor = DDGGray,
                        shape = RoundedCornerShape(12.dp)
                    ) {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Icon(
                                    Icons.Default.LocalParking,
                                    contentDescription = null,
                                    tint = DDGSuccess,
                                    modifier = Modifier.size(24.dp)
                                )
                                Spacer(modifier = Modifier.width(8.dp))
                                Text("签到打卡", style = MaterialTheme.typography.h6, fontWeight = FontWeight.Bold)
                            }

                            Spacer(modifier = Modifier.height(12.dp))

                            Text(
                                "确认已到达服务区？签到后将开始计算休息时间。",
                                style = MaterialTheme.typography.body2,
                                color = DDGTextSecondary
                            )

                            Spacer(modifier = Modifier.height(16.dp))

                            Button(
                                onClick = {
                                    scope.launch {
                                        isProcessing = true
                                        errorMessage = null
                                        val result = checkInUseCase(
                                            CheckInRequest(
                                                driver_id = driverId,
                                                vehicle_id = vehicleId,
                                                service_area_id = preselectedServiceAreaId,
                                                latitude = latitude,
                                                longitude = longitude,
                                                waybill_id = waybillId
                                            )
                                        )
                                        isProcessing = false
                                        result.onSuccess {
                                            successMessage = "签到成功！开始休息"
                                            scope.launch {
                                                val updated = getRestCountdownUseCase(driverId, vehicleId)
                                                updated.onSuccess { restCountdown = it }
                                            }
                                        }.onFailure { errorMessage = it.message }
                                    }
                                },
                                modifier = Modifier.fillMaxWidth(),
                                colors = ButtonDefaults.buttonColors(backgroundColor = DDGSuccess),
                                enabled = !isProcessing
                            ) {
                                if (isProcessing) {
                                    CircularProgressIndicator(color = Color.White, modifier = Modifier.size(20.dp))
                                } else {
                                    Icon(Icons.Default.CheckCircle, contentDescription = null)
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text("确认签到", color = Color.White, fontWeight = FontWeight.Bold)
                                }
                            }
                        }
                    }
                }
            }

            item {
                if (restCountdown?.status == "resting") {
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        backgroundColor = DDGGray,
                        shape = RoundedCornerShape(12.dp)
                    ) {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Icon(
                                    Icons.Default.ExitToApp,
                                    contentDescription = null,
                                    tint = DDGWarning,
                                    modifier = Modifier.size(24.dp)
                                )
                                Spacer(modifier = Modifier.width(8.dp))
                                Text("签退", style = MaterialTheme.typography.h6, fontWeight = FontWeight.Bold)
                            }

                            Spacer(modifier = Modifier.height(12.dp))

                            val canContinue = restCountdown?.can_continue_driving == true
                            val currentRestMinutes = restCountdown?.current_rest_minutes ?: 0
                            val minRequired = restCountdown?.min_rest_required ?: 20

                            Text(
                                if (canContinue) {
                                    "已休息 $currentRestMinutes 分钟，满足最低 ${minRequired} 分钟要求，可以继续驾驶。"
                                } else {
                                    "已休息 $currentRestMinutes 分钟，还需 ${minRequired - currentRestMinutes} 分钟才能继续驾驶。"
                                },
                                style = MaterialTheme.typography.body2,
                                color = if (canContinue) DDGSuccess else DDGWarning
                            )

                            Spacer(modifier = Modifier.height(16.dp))

                            Button(
                                onClick = {
                                    scope.launch {
                                        isProcessing = true
                                        errorMessage = null
                                        val result = checkOutUseCase(
                                            CheckOutRequest(
                                                driver_id = driverId,
                                                vehicle_id = vehicleId,
                                                latitude = latitude,
                                                longitude = longitude
                                            )
                                        )
                                        isProcessing = false
                                        result.onSuccess {
                                            successMessage = "签退成功！"
                                            onCheckOutSuccess()
                                        }.onFailure { errorMessage = it.message }
                                    }
                                },
                                modifier = Modifier.fillMaxWidth(),
                                colors = ButtonDefaults.buttonColors(
                                    backgroundColor = if (canContinue) DDGSuccess else DDGGray
                                ),
                                enabled = canContinue && !isProcessing
                            ) {
                                if (isProcessing) {
                                    CircularProgressIndicator(color = Color.White, modifier = Modifier.size(20.dp))
                                } else {
                                    Icon(Icons.Default.ExitToApp, contentDescription = null)
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text(
                                        if (canContinue) "确认签退" else "休息不足，无法签退",
                                        color = Color.White,
                                        fontWeight = FontWeight.Bold
                                    )
                                }
                            }

                            Spacer(modifier = Modifier.height(12.dp))

                            OutlinedButton(
                                onClick = {
                                    onNavigateToReview(preselectedServiceAreaId)
                                },
                                modifier = Modifier.fillMaxWidth()
                            ) {
                                Icon(Icons.Default.Star, contentDescription = null, tint = DDGWarning)
                                Spacer(modifier = Modifier.width(8.dp))
                                Text("评价此服务区", color = DDGWarning)
                            }
                        }
                    }
                }
            }

            item {
                successMessage?.let { msg ->
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        backgroundColor = DDGSuccess.copy(alpha = 0.15f),
                        shape = RoundedCornerShape(12.dp)
                    ) {
                        Row(
                            modifier = Modifier.padding(16.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Icon(Icons.Default.CheckCircle, null, tint = DDGSuccess)
                            Spacer(modifier = Modifier.width(8.dp))
                            Text(msg, color = DDGSuccess, fontWeight = FontWeight.Bold)
                        }
                    }
                }
                errorMessage?.let { msg ->
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        backgroundColor = DDGRed.copy(alpha = 0.15f),
                        shape = RoundedCornerShape(12.dp)
                    ) {
                        Row(
                            modifier = Modifier.padding(16.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Icon(Icons.Default.ExitToApp, null, tint = DDGRed)
                            Spacer(modifier = Modifier.width(8.dp))
                            Text(msg, color = DDGRed, fontWeight = FontWeight.Bold)
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun CurrentStatusCard(countdown: RestCountdown) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        backgroundColor = DDGGray,
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text("当前状态", style = MaterialTheme.typography.h6, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(8.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Column {
                    Text("状态", style = MaterialTheme.typography.caption, color = DDGTextSecondary)
                    Text(
                        when (countdown.status) {
                            "driving" -> "驾驶中"
                            "resting" -> "休息中"
                            else -> "已完成"
                        },
                        fontWeight = FontWeight.Bold,
                        color = if (countdown.status == "driving") DDGSuccess else DDGWarning
                    )
                }
                Column {
                    Text("连续驾驶", style = MaterialTheme.typography.caption, color = DDGTextSecondary)
                    Text(
                        "${countdown.continuous_drive_minutes}分钟",
                        fontWeight = FontWeight.Bold,
                        color = if (countdown.is_overtime) DDGRed else DDGTextPrimary
                    )
                }
                Column {
                    Text("剩余可驾", style = MaterialTheme.typography.caption, color = DDGTextSecondary)
                    Text(
                        "${countdown.remaining_drive_minutes}分钟",
                        fontWeight = FontWeight.Bold,
                        color = if (countdown.remaining_drive_minutes <= 60) DDGWarning else DDGTextPrimary
                    )
                }
            }

            if (countdown.status == "resting") {
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    "已休息 ${countdown.current_rest_minutes} / ${countdown.min_rest_required} 分钟",
                    style = MaterialTheme.typography.caption,
                    color = DDGTextSecondary
                )
                if (countdown.current_service_area_name.isNotEmpty()) {
                    Text(
                        "当前服务区：${countdown.current_service_area_name}",
                        style = MaterialTheme.typography.caption,
                        color = DDGTextSecondary
                    )
                }
            }
        }
    }
}
