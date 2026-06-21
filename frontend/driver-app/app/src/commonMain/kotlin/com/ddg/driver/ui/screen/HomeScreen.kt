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
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.Card
import androidx.compose.material.CircularProgressIndicator
import androidx.compose.material.ExperimentalMaterialApi
import androidx.compose.material.Icon
import androidx.compose.material.MaterialTheme
import androidx.compose.material.Text
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AccessTime
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.DirectionsCar
import androidx.compose.material.icons.filled.ExitToApp
import androidx.compose.material.icons.filled.LocalGasStation
import androidx.compose.material.icons.filled.LocationOn
import androidx.compose.material.icons.filled.Menu
import androidx.compose.material.icons.filled.Navigation
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.Report
import androidx.compose.material.icons.filled.Route
import androidx.compose.material.icons.filled.Timer
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material.AlertDialog
import androidx.compose.material.Button
import androidx.compose.material.ButtonDefaults
import androidx.compose.material.TextField
import androidx.compose.material.TextFieldDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.ddg.driver.data.local.AppDataStore
import com.ddg.driver.data.model.RestCountdown
import com.ddg.driver.data.model.Waybill
import com.ddg.driver.data.model.GeoFenceCheckResult
import com.ddg.driver.data.model.GeoFenceCheckRequest
import com.ddg.driver.data.model.GeoFenceConfirmRequest
import com.ddg.driver.data.remote.ApiService
import com.ddg.driver.domain.usecase.GetCurrentWaybillUseCase
import com.ddg.driver.domain.usecase.GetRestCountdownUseCase
import com.ddg.driver.ui.navigation.DriverContext
import com.ddg.driver.ui.theme.DDGDanger
import com.ddg.driver.ui.theme.DDGGray
import com.ddg.driver.ui.theme.DDGRed
import com.ddg.driver.ui.theme.DDGSuccess
import com.ddg.driver.ui.theme.DDGSurface
import com.ddg.driver.ui.theme.DDGTextPrimary
import com.ddg.driver.ui.theme.DDGTextSecondary
import com.ddg.driver.ui.theme.DDGWarning
import com.ddg.driver.ui.components.FatigueIndicator
import com.ddg.driver.ui.components.SOSButton
import com.ddg.driver.ui.components.WaybillCard
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import org.koin.compose.getKoin

@Composable
fun HomeScreen(
    onStartNavigation: () -> Unit,
    onViewAlarms: () -> Unit,
    onViewProfile: () -> Unit,
    onViewRestCountdown: () -> Unit = {},
    onAutoRecommend: (DriverContext) -> Unit = {}
) {
    val dataStore: AppDataStore = getKoin().get()
    val getCurrentWaybillUseCase: GetCurrentWaybillUseCase = getKoin().get()
    val getRestCountdownUseCase: GetRestCountdownUseCase = getKoin().get()
    val apiService: ApiService = getKoin().get()
    val scope = rememberCoroutineScope()

    var waybill by remember { mutableStateOf<Waybill?>(null) }
    var restCountdown by remember { mutableStateOf<RestCountdown?>(null) }
    var isLoading by remember { mutableStateOf(true) }
    var errorMessage by remember { mutableStateOf<String?>(null) }

    var geoFenceLastResult by remember { mutableStateOf<GeoFenceCheckResult?>(null) }
    var pendingGeoFenceAlert by remember { mutableStateOf<GeoFenceCheckResult?>(null) }
    var showGeoFenceConfirmDialog by remember { mutableStateOf(false) }
    var geoFenceConfirmType by remember { mutableStateOf<String?>(null) }
    var geoFenceReasonDetail by remember { mutableStateOf("") }
    var geoFenceConfirming by remember { mutableStateOf(false) }
    var geoFenceErrorMsg by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(Unit) {
        scope.launch {
            val result = getCurrentWaybillUseCase()
            isLoading = false
            result.onSuccess { waybill = it }
                .onFailure { errorMessage = it.message }
        }
        scope.launch {
            val result = getRestCountdownUseCase(1, 1)
            result.onSuccess { restCountdown = it }
        }
    }

    LaunchedEffect(restCountdown) {
        while (true) {
            delay(60_000)
            scope.launch {
                val result = getRestCountdownUseCase(1, 1)
                result.onSuccess { restCountdown = it }
            }
        }
    }

    LaunchedEffect(waybill) {
        while (true) {
            delay(30_000)
            val wb = waybill ?: continue
            val vehicleId = wb.vehicle_id ?: continue
            val waybillId = wb.id ?: continue
            val lat = wb.current_lat ?: wb.start_lat ?: continue
            val lng = wb.current_lng ?: wb.start_lng ?: continue
            scope.launch {
                runCatching {
                    apiService.checkGeoFence(
                        GeoFenceCheckRequest(
                            vehicle_id = vehicleId,
                            driver_id = 1,
                            waybill_id = waybillId,
                            latitude = lat,
                            longitude = lng,
                            address = wb.current_location,
                            threshold_meters = 500
                        )
                    )
                }.onSuccess { result ->
                    geoFenceLastResult = result
                    if (result.is_deviated && result.status == "pending" && result.alert_id > 0) {
                        if (pendingGeoFenceAlert?.alert_id != result.alert_id) {
                            pendingGeoFenceAlert = result
                            showGeoFenceConfirmDialog = true
                        }
                    }
                }.onFailure {
                    geoFenceErrorMsg = it.message
                }
            }
        }
    }

    fun submitGeoFenceConfirm(type: String) {
        val alert = pendingGeoFenceAlert ?: return
        geoFenceConfirming = true
        scope.launch {
            runCatching {
                apiService.confirmGeoFenceAlert(
                    GeoFenceConfirmRequest(
                        alert_id = alert.alert_id,
                        confirm_type = type,
                        reason_detail = geoFenceReasonDetail.ifBlank { null }
                    )
                )
            }.onSuccess {
                showGeoFenceConfirmDialog = false
                pendingGeoFenceAlert = null
                geoFenceConfirmType = null
                geoFenceReasonDetail = ""
            }.onFailure {
                geoFenceErrorMsg = it.message
            }
            geoFenceConfirming = false
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
                HeaderBar(
                    onViewProfile = onViewProfile,
                    onViewAlarms = onViewAlarms
                )
            }

            item {
                if (isLoading) {
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(200.dp),
                        contentAlignment = Alignment.Center
                    ) {
                        CircularProgressIndicator(color = DDGRed)
                    }
                } else {
                    WaybillCard(waybill = waybill)
                }
            }

            item {
                QuickActions(
                    onStartNavigation = onStartNavigation,
                    onSOS = { },
                    onViewVehicleStatus = { },
                    onViewEscortRecord = { }
                )
            }

            item {
                TodayStats(
                    distance = 125.6,
                    drivingHours = 3.5f,
                    fatigueAlerts = 1
                )
            }

            item {
                FatigueSection(
                    restCountdown = restCountdown,
                    onViewRestCountdown = onViewRestCountdown
                )
            }

            item {
                restCountdown?.let { countdown ->
                    if (countdown.is_overtime) {
                        OvertimeWarningBanner(
                            overtimeMinutes = countdown.overtime_minutes,
                            onNavigateToRest = {
                                onAutoRecommend(DriverContext(
                                    driverId = countdown.driver_id,
                                    vehicleId = countdown.vehicle_id,
                                    waybillId = countdown.waybill_id
                                ))
                            }
                        )
                    } else if (countdown.remaining_drive_minutes <= 60 && countdown.remaining_drive_minutes > 0 && countdown.status == "driving") {
                        ApproachingLimitBanner(
                            remainingMinutes = countdown.remaining_drive_minutes,
                            onNavigateToRest = onViewRestCountdown
                        )
                    }
                }
            }

            item {
                GeoFenceSection(
                    lastResult = geoFenceLastResult,
                    pendingAlert = pendingGeoFenceAlert,
                    onConfirmClick = { showGeoFenceConfirmDialog = true }
                )
            }

            item {
                SOSButton(onClick = { })
            }
        }

        if (showGeoFenceConfirmDialog && pendingGeoFenceAlert != null) {
            GeoFenceConfirmDialog(
                alert = pendingGeoFenceAlert!!,
                confirmType = geoFenceConfirmType,
                reasonDetail = geoFenceReasonDetail,
                confirming = geoFenceConfirming,
                onSelectType = { geoFenceConfirmType = it },
                onReasonChange = { geoFenceReasonDetail = it },
                onSubmit = { type -> submitGeoFenceConfirm(type) },
                onDismiss = {
                    if (!geoFenceConfirming) {
                        showGeoFenceConfirmDialog = false
                    }
                }
            )
        }
    }
}

@Composable
private fun HeaderBar(
    onViewProfile: () -> Unit,
    onViewAlarms: () -> Unit
) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(
            text = "危险品运输护航",
            style = MaterialTheme.typography.h3,
            fontWeight = FontWeight.Bold
        )
        Row {
            Icon(
                imageVector = Icons.Default.Notifications,
                contentDescription = "报警",
                modifier = Modifier
                    .size(28.dp)
                    .clickable { onViewAlarms() },
                tint = DDGRed
            )
            Spacer(modifier = Modifier.width(16.dp))
            Icon(
                imageVector = Icons.Default.Person,
                contentDescription = "个人中心",
                modifier = Modifier
                    .size(28.dp)
                    .clickable { onViewProfile() },
                tint = DDGTextPrimary
            )
        }
    }
}

@OptIn(ExperimentalMaterialApi::class)
@Composable
private fun QuickActions(
    onStartNavigation: () -> Unit,
    onSOS: () -> Unit,
    onViewVehicleStatus: () -> Unit,
    onViewEscortRecord: () -> Unit
) {
    Column {
        Text(
            text = "快捷功能",
            style = MaterialTheme.typography.h5,
            modifier = Modifier.padding(bottom = 12.dp)
        )
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            QuickActionItem(
                icon = Icons.Default.Navigation,
                label = "开始导航",
                color = DDGSuccess,
                modifier = Modifier.weight(1f),
                onClick = onStartNavigation
            )
            QuickActionItem(
                icon = Icons.Default.Report,
                label = "紧急求助",
                color = DDGDanger,
                modifier = Modifier.weight(1f),
                onClick = onSOS
            )
        }
        Spacer(modifier = Modifier.height(12.dp))
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            QuickActionItem(
                icon = Icons.Default.DirectionsCar,
                label = "车辆状态",
                color = DDGWarning,
                modifier = Modifier.weight(1f),
                onClick = onViewVehicleStatus
            )
            QuickActionItem(
                icon = Icons.Default.Menu,
                label = "押运记录",
                color = DDGRed,
                modifier = Modifier.weight(1f),
                onClick = onViewEscortRecord
            )
        }
    }
}

@OptIn(ExperimentalMaterialApi::class)
@Composable
private fun QuickActionItem(
    icon: ImageVector,
    label: String,
    color: androidx.compose.ui.graphics.Color,
    modifier: Modifier = Modifier,
    onClick: () -> Unit
) {
    Card(
        modifier = modifier
            .height(100.dp)
            .clickable { onClick() },
        backgroundColor = DDGGray,
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(
            modifier = Modifier.fillMaxSize(),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Icon(
                imageVector = icon,
                contentDescription = label,
                tint = color,
                modifier = Modifier.size(36.dp)
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = label,
                style = MaterialTheme.typography.body2,
                color = DDGTextPrimary
            )
        }
    }
}

@Composable
private fun TodayStats(
    distance: Double,
    drivingHours: Float,
    fatigueAlerts: Int
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        backgroundColor = DDGGray,
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(
                text = "今日驾驶统计",
                style = MaterialTheme.typography.h5,
                modifier = Modifier.padding(bottom = 16.dp)
            )
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceAround
            ) {
                StatItem(
                    icon = Icons.Default.Route,
                    value = "${distance}km",
                    label = "行驶里程",
                    color = DDGSuccess
                )
                StatItem(
                    icon = Icons.Default.Timer,
                    value = "${drivingHours}h",
                    label = "驾驶时长",
                    color = DDGWarning
                )
                StatItem(
                    icon = Icons.Default.LocalGasStation,
                    value = "$fatigueAlerts",
                    label = "疲劳次数",
                    color = DDGDanger
                )
            }
        }
    }
}

@Composable
private fun StatItem(
    icon: ImageVector,
    value: String,
    label: String,
    color: androidx.compose.ui.graphics.Color
) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Icon(
            imageVector = icon,
            contentDescription = label,
            tint = color,
            modifier = Modifier.size(28.dp)
        )
        Spacer(modifier = Modifier.height(6.dp))
        Text(
            text = value,
            style = MaterialTheme.typography.h4,
            fontWeight = FontWeight.Bold,
            color = DDGTextPrimary
        )
        Spacer(modifier = Modifier.height(4.dp))
        Text(
            text = label,
            style = MaterialTheme.typography.caption,
            color = DDGTextSecondary
        )
    }
}

@Composable
private fun FatigueSection(
    restCountdown: RestCountdown?,
    onViewRestCountdown: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        backgroundColor = DDGGray,
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "疲劳监测",
                    style = MaterialTheme.typography.h5,
                    modifier = Modifier.padding(bottom = 12.dp)
                )
                if (restCountdown != null) {
                    Card(
                        modifier = Modifier.clickable { onViewRestCountdown() },
                        backgroundColor = DDGWarning.copy(alpha = 0.2f),
                        shape = RoundedCornerShape(8.dp)
                    ) {
                        Text(
                            "查看倒计时 →",
                            modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp),
                            style = MaterialTheme.typography.caption,
                            color = DDGWarning
                        )
                    }
                }
            }
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically
            ) {
                val progress = restCountdown?.let {
                    it.continuous_drive_minutes.toFloat() / it.max_continuous_drive.toFloat()
                } ?: 0.25f
                val level = restCountdown?.let {
                    when {
                        it.is_overtime -> com.ddg.driver.data.model.FatigueLevel.DANGEROUS
                        it.remaining_drive_minutes <= 60 -> com.ddg.driver.data.model.FatigueLevel.WARNING
                        else -> com.ddg.driver.data.model.FatigueLevel.NORMAL
                    }
                } ?: com.ddg.driver.data.model.FatigueLevel.NORMAL

                FatigueIndicator(
                    progress = progress.coerceIn(0f, 1f),
                    level = level,
                    size = 80.dp
                )
                Spacer(modifier = Modifier.width(16.dp))
                Column {
                    val statusText = restCountdown?.let {
                        when {
                            it.is_overtime -> "超时驾驶！"
                            it.remaining_drive_minutes <= 60 -> "即将到达上限"
                            it.status == "resting" -> "休息中"
                            else -> "状态良好"
                        }
                    } ?: "状态良好"
                    val statusColor = restCountdown?.let {
                        when {
                            it.is_overtime -> DDGRed
                            it.remaining_drive_minutes <= 60 -> DDGWarning
                            it.status == "resting" -> DDGWarning
                            else -> DDGSuccess
                        }
                    } ?: DDGSuccess

                    Text(
                        text = statusText,
                        style = MaterialTheme.typography.h5,
                        color = statusColor,
                        fontWeight = FontWeight.Bold
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = restCountdown?.let {
                            if (it.status == "resting") {
                                "已休息 ${it.current_rest_minutes} 分钟"
                            } else {
                                "已连续驾驶 ${it.continuous_drive_minutes / 60}小时${it.continuous_drive_minutes % 60}分钟"
                            }
                        } ?: "已连续驾驶 2小时15分钟",
                        style = MaterialTheme.typography.body2
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = restCountdown?.let {
                            if (it.is_overtime) {
                                "⚠ 超时 ${it.overtime_minutes} 分钟，请立即休息！"
                            } else if (it.remaining_drive_minutes <= 60) {
                                "建议：再行驶 ${it.remaining_drive_minutes} 分钟后停车休息"
                            } else {
                                "建议：再行驶1小时后停车休息"
                            }
                        } ?: "建议：再行驶1小时后停车休息",
                        style = MaterialTheme.typography.caption,
                        color = if (restCountdown?.is_overtime == true) DDGRed else DDGWarning
                    )
                }
            }
        }
    }
}

@Composable
private fun OvertimeWarningBanner(
    overtimeMinutes: Int,
    onNavigateToRest: () -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onNavigateToRest() },
        backgroundColor = DDGRed.copy(alpha = 0.15f),
        shape = RoundedCornerShape(12.dp)
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(
                Icons.Default.Warning,
                contentDescription = null,
                tint = DDGRed,
                modifier = Modifier.size(32.dp)
            )
            Spacer(modifier = Modifier.width(12.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text("超时驾驶警告！", fontWeight = FontWeight.Bold, color = DDGRed)
                Text("已超时 $overtimeMinutes 分钟，请立即寻找服务区休息", style = MaterialTheme.typography.caption, color = DDGRed)
            }
            Icon(
                Icons.Default.DirectionsCar,
                contentDescription = null,
                tint = DDGRed,
                modifier = Modifier.size(20.dp)
            )
        }
    }
}

@Composable
private fun ApproachingLimitBanner(
    remainingMinutes: Int,
    onNavigateToRest: () -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onNavigateToRest() },
        backgroundColor = DDGWarning.copy(alpha = 0.12f),
        shape = RoundedCornerShape(12.dp)
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(
                Icons.Default.AccessTime,
                contentDescription = null,
                tint = DDGWarning,
                modifier = Modifier.size(28.dp)
            )
            Spacer(modifier = Modifier.width(12.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text("即将达到驾驶时限", fontWeight = FontWeight.Bold, color = DDGWarning)
                Text("剩余 $remainingMinutes 分钟，建议提前规划休息", style = MaterialTheme.typography.caption, color = DDGWarning)
            }
        }
    }
}

@Composable
private fun GeoFenceSection(
    lastResult: GeoFenceCheckResult?,
    pendingAlert: GeoFenceCheckResult?,
    onConfirmClick: () -> Unit
) {
    val isDeviated = lastResult?.is_deviated == true || pendingAlert != null
    val dist = lastResult?.distance_from_route_meters ?: pendingAlert?.distance_from_route_meters ?: 0.0
    val threshold = lastResult?.threshold_meters ?: pendingAlert?.threshold_meters ?: 500
    val dailyCount = lastResult?.daily_deviate_count ?: pendingAlert?.daily_deviate_count ?: 0
    val autoReported = lastResult?.auto_reported == true || pendingAlert?.auto_reported == true

    val bgColor = when {
        autoReported -> DDGRed.copy(alpha = 0.12f)
        isDeviated -> DDGWarning.copy(alpha = 0.15f)
        else -> DDGGray
    }
    val titleColor = when {
        autoReported -> DDGRed
        isDeviated -> DDGDanger
        else -> DDGSuccess
    }

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .then(if (pendingAlert != null) Modifier.clickable { onConfirmClick() } else Modifier),
        backgroundColor = bgColor,
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        imageVector = Icons.Default.LocationOn,
                        contentDescription = null,
                        tint = titleColor,
                        modifier = Modifier.size(28.dp)
                    )
                    Spacer(modifier = Modifier.width(10.dp))
                    Text(
                        text = "电子围栏监控",
                        style = MaterialTheme.typography.h5,
                        color = titleColor,
                        fontWeight = FontWeight.Bold
                    )
                }
                if (pendingAlert != null) {
                    Card(
                        backgroundColor = DDGRed.copy(alpha = 0.25f),
                        shape = RoundedCornerShape(8.dp)
                    ) {
                        Text(
                            text = "待确认 →",
                            modifier = Modifier.padding(horizontal = 10.dp, vertical = 5.dp),
                            style = MaterialTheme.typography.caption,
                            color = DDGRed,
                            fontWeight = FontWeight.Bold
                        )
                    }
                } else if (autoReported) {
                    Card(
                        backgroundColor = DDGRed.copy(alpha = 0.25f),
                        shape = RoundedCornerShape(8.dp)
                    ) {
                        Text(
                            text = "已上报调度",
                            modifier = Modifier.padding(horizontal = 10.dp, vertical = 5.dp),
                            style = MaterialTheme.typography.caption,
                            color = DDGRed,
                            fontWeight = FontWeight.Bold
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(12.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceAround
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(
                        text = String.format("%.0f", dist),
                        style = MaterialTheme.typography.h4,
                        fontWeight = FontWeight.Bold,
                        color = if (dist > threshold) DDGDanger else DDGTextPrimary
                    )
                    Text(
                        text = "偏离距离(米)",
                        style = MaterialTheme.typography.caption,
                        color = DDGTextSecondary
                    )
                }
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(
                        text = "$threshold",
                        style = MaterialTheme.typography.h4,
                        fontWeight = FontWeight.Bold,
                        color = DDGTextPrimary
                    )
                    Text(
                        text = "告警阈值(米)",
                        style = MaterialTheme.typography.caption,
                        color = DDGTextSecondary
                    )
                }
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(
                        text = "$dailyCount",
                        style = MaterialTheme.typography.h4,
                        fontWeight = FontWeight.Bold,
                        color = if (dailyCount >= 3) DDGRed else if (dailyCount >= 2) DDGWarning else DDGTextPrimary
                    )
                    Text(
                        text = "当日累计(次)",
                        style = MaterialTheme.typography.caption,
                        color = DDGTextSecondary
                    )
                }
            }

            if (isDeviated) {
                Spacer(modifier = Modifier.height(12.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        Icons.Default.Warning,
                        contentDescription = null,
                        tint = if (autoReported) DDGRed else DDGDanger,
                        modifier = Modifier.size(18.dp)
                    )
                    Spacer(modifier = Modifier.width(6.dp))
                    Text(
                        text = when {
                            autoReported -> "⚠️ 当日累计偏航已达${dailyCount}次，系统已自动上报调度中心！"
                            dailyCount >= 2 -> "⚠️ 当日累计偏航${dailyCount}次，再偏航${3 - dailyCount}次将自动上报调度！"
                            else -> "⚠️ 当前偏离预设路线 ${String.format("%.0f", dist)} 米，请押运员及时确认原因"
                        },
                        style = MaterialTheme.typography.caption,
                        color = if (autoReported) DDGRed else DDGDanger
                    )
                }
            } else {
                Spacer(modifier = Modifier.height(12.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        Icons.Default.CheckCircle,
                        contentDescription = null,
                        tint = DDGSuccess,
                        modifier = Modifier.size(18.dp)
                    )
                    Spacer(modifier = Modifier.width(6.dp))
                    Text(
                        text = "路线正常，车辆正在沿预设路线行驶",
                        style = MaterialTheme.typography.caption,
                        color = DDGSuccess
                    )
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterialApi::class)
@Composable
private fun GeoFenceConfirmDialog(
    alert: GeoFenceCheckResult,
    confirmType: String?,
    reasonDetail: String,
    confirming: Boolean,
    onSelectType: (String) -> Unit,
    onReasonChange: (String) -> Unit,
    onSubmit: (String) -> Unit,
    onDismiss: () -> Unit
) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(
                    Icons.Default.Warning,
                    contentDescription = null,
                    tint = DDGRed,
                    modifier = Modifier.size(28.dp)
                )
                Spacer(modifier = Modifier.width(10.dp))
                Text(
                    text = "偏航告警确认",
                    fontWeight = FontWeight.Bold,
                    color = DDGRed
                )
            }
        },
        text = {
            Column(modifier = Modifier.fillMaxWidth()) {
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    backgroundColor = DDGWarning.copy(alpha = 0.12f),
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Column(modifier = Modifier.padding(12.dp)) {
                        Text(
                            text = "偏离距离：${String.format("%.0f", alert.distance_from_route_meters)} 米 / ${alert.threshold_meters}米阈值",
                            style = MaterialTheme.typography.body2,
                            color = DDGDanger,
                            fontWeight = FontWeight.Bold
                        )
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(
                            text = "告警编号：${alert.alert_no ?: "(无)"}",
                            style = MaterialTheme.typography.caption,
                            color = DDGTextSecondary
                        )
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(
                            text = "当日累计偏航：${alert.daily_deviate_count} 次" + if (alert.daily_deviate_count >= 2) "（已接近自动上报阈值3次）" else "",
                            style = MaterialTheme.typography.caption,
                            color = if (alert.daily_deviate_count >= 2) DDGRed else DDGWarning
                        )
                    }
                }

                Spacer(modifier = Modifier.height(14.dp))
                Text(
                    text = "请选择偏航原因：",
                    style = MaterialTheme.typography.body2,
                    fontWeight = FontWeight.Bold
                )
                Spacer(modifier = Modifier.height(8.dp))

                Row(modifier = Modifier.fillMaxWidth()) {
                    Card(
                        modifier = Modifier
                            .weight(1f)
                            .height(80.dp)
                            .clickable(enabled = !confirming) { onSelectType("detour") },
                        backgroundColor = if (confirmType == "detour") DDGSuccess.copy(alpha = 0.15f) else DDGGray,
                        shape = RoundedCornerShape(10.dp),
                        border = androidx.compose.foundation.BorderStroke(
                            width = if (confirmType == "detour") 2.dp else 1.dp,
                            color = if (confirmType == "detour") DDGSuccess else DDGGray
                        )
                    ) {
                        Column(
                            modifier = Modifier.fillMaxSize().padding(10.dp),
                            horizontalAlignment = Alignment.CenterHorizontally,
                            verticalArrangement = Arrangement.Center
                        ) {
                            Icon(
                                Icons.Default.Route,
                                contentDescription = null,
                                tint = DDGSuccess,
                                modifier = Modifier.size(24.dp)
                            )
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(
                                text = "合理绕路",
                                style = MaterialTheme.typography.body2,
                                color = DDGSuccess,
                                fontWeight = FontWeight.Bold
                            )
                            Text(
                                text = "施工/拥堵/检查",
                                style = MaterialTheme.typography.overline,
                                color = DDGTextSecondary
                            )
                        }
                    }

                    Spacer(modifier = Modifier.width(12.dp))

                    Card(
                        modifier = Modifier
                            .weight(1f)
                            .height(80.dp)
                            .clickable(enabled = !confirming) { onSelectType("deviate") },
                        backgroundColor = if (confirmType == "deviate") DDGRed.copy(alpha = 0.15f) else DDGGray,
                        shape = RoundedCornerShape(10.dp),
                        border = androidx.compose.foundation.BorderStroke(
                            width = if (confirmType == "deviate") 2.dp else 1.dp,
                            color = if (confirmType == "deviate") DDGRed else DDGGray
                        )
                    ) {
                        Column(
                            modifier = Modifier.fillMaxSize().padding(10.dp),
                            horizontalAlignment = Alignment.CenterHorizontally,
                            verticalArrangement = Arrangement.Center
                        ) {
                            Icon(
                                Icons.Default.Warning,
                                contentDescription = null,
                                tint = DDGRed,
                                modifier = Modifier.size(24.dp)
                            )
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(
                                text = "异常偏航",
                                style = MaterialTheme.typography.body2,
                                color = DDGRed,
                                fontWeight = FontWeight.Bold
                            )
                            Text(
                                text = "不明原因偏离",
                                style = MaterialTheme.typography.overline,
                                color = DDGTextSecondary
                            )
                        }
                    }
                }

                Spacer(modifier = Modifier.height(14.dp))
                Text(
                    text = "具体说明（必填）",
                    style = MaterialTheme.typography.body2,
                    fontWeight = FontWeight.Bold
                )
                Spacer(modifier = Modifier.height(6.dp))
                TextField(
                    value = reasonDetail,
                    onValueChange = onReasonChange,
                    enabled = !confirming,
                    modifier = Modifier.fillMaxWidth().height(100.dp),
                    placeholder = {
                        Text(
                            text = "请填写具体原因，例如：前方G5高速事故封路，绕行G108国道",
                            style = MaterialTheme.typography.caption,
                            color = DDGTextSecondary
                        )
                    },
                    colors = TextFieldDefaults.textFieldColors(
                        backgroundColor = DDGGray,
                        focusedIndicatorColor = DDGWarning,
                        unfocusedIndicatorColor = androidx.compose.ui.graphics.Color.Transparent
                    ),
                    shape = RoundedCornerShape(8.dp)
                )
                if (alert.daily_deviate_count >= 2) {
                    Spacer(modifier = Modifier.height(10.dp))
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        backgroundColor = DDGRed.copy(alpha = 0.1f),
                        shape = RoundedCornerShape(6.dp)
                    ) {
                        Row(modifier = Modifier.padding(10.dp), verticalAlignment = Alignment.CenterVertically) {
                            Icon(
                                Icons.Default.Warning,
                                contentDescription = null,
                                tint = DDGRed,
                                modifier = Modifier.size(18.dp)
                            )
                            Spacer(modifier = Modifier.width(6.dp))
                            Text(
                                text = "当日累计偏航已达${alert.daily_deviate_count}次，若再偏航${3 - alert.daily_deviate_count}次，系统将自动上报调度中心！",
                                style = MaterialTheme.typography.caption,
                                color = DDGRed
                            )
                        }
                    }
                }
            }
        },
        confirmButton = {
            Button(
                onClick = {
                    val type = confirmType ?: return@Button
                    if (reasonDetail.isBlank()) return@Button
                    onSubmit(type)
                },
                enabled = !confirming && confirmType != null && reasonDetail.isNotBlank(),
                colors = ButtonDefaults.buttonColors(
                    backgroundColor = if (confirmType == "detour") DDGSuccess else DDGRed
                )
            ) {
                if (confirming) {
                    CircularProgressIndicator(
                        color = androidx.compose.ui.graphics.Color.White,
                        modifier = Modifier.size(18.dp),
                        strokeWidth = 2.dp
                    )
                    Spacer(modifier = Modifier.width(6.dp))
                    Text("提交中...", color = androidx.compose.ui.graphics.Color.White)
                } else {
                    Text(
                        text = "确认提交",
                        color = androidx.compose.ui.graphics.Color.White,
                        fontWeight = FontWeight.Bold
                    )
                }
            }
        },
        dismissButton = {
            Button(
                onClick = onDismiss,
                enabled = !confirming,
                colors = ButtonDefaults.buttonColors(backgroundColor = DDGGray)
            ) {
                Text("稍后处理", color = DDGTextPrimary)
            }
        },
        shape = RoundedCornerShape(14.dp)
    )
}
