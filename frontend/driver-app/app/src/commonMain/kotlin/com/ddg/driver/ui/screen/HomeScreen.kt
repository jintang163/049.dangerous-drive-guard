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
import androidx.compose.material.icons.filled.DirectionsCar
import androidx.compose.material.icons.filled.ExitToApp
import androidx.compose.material.icons.filled.LocalGasStation
import androidx.compose.material.icons.filled.Menu
import androidx.compose.material.icons.filled.Navigation
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.Report
import androidx.compose.material.icons.filled.Route
import androidx.compose.material.icons.filled.Timer
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
import com.ddg.driver.data.model.Waybill
import com.ddg.driver.domain.usecase.GetCurrentWaybillUseCase
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
import kotlinx.coroutines.launch
import org.koin.compose.getKoin

@Composable
fun HomeScreen(
    onStartNavigation: () -> Unit,
    onViewAlarms: () -> Unit,
    onViewProfile: () -> Unit
) {
    val dataStore: AppDataStore = getKoin().get()
    val getCurrentWaybillUseCase: GetCurrentWaybillUseCase = getKoin().get()
    val scope = rememberCoroutineScope()

    var waybill by remember { mutableStateOf<Waybill?>(null) }
    var isLoading by remember { mutableStateOf(true) }
    var errorMessage by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(Unit) {
        scope.launch {
            val result = getCurrentWaybillUseCase()
            isLoading = false
            result.onSuccess { waybill = it }
                .onFailure { errorMessage = it.message }
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
                FatigueSection()
            }

            item {
                SOSButton(onClick = { })
            }
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
private fun FatigueSection() {
    Card(
        modifier = Modifier.fillMaxWidth(),
        backgroundColor = DDGGray,
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(
                text = "疲劳监测",
                style = MaterialTheme.typography.h5,
                modifier = Modifier.padding(bottom = 12.dp)
            )
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically
            ) {
                FatigueIndicator(
                    progress = 0.25f,
                    level = com.ddg.driver.data.model.FatigueLevel.NORMAL,
                    size = 80.dp
                )
                Spacer(modifier = Modifier.width(16.dp))
                Column {
                    Text(
                        text = "状态良好",
                        style = MaterialTheme.typography.h5,
                        color = DDGSuccess,
                        fontWeight = FontWeight.Bold
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = "已连续驾驶 2小时15分钟",
                        style = MaterialTheme.typography.body2
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = "建议：再行驶1小时后停车休息",
                        style = MaterialTheme.typography.caption,
                        color = DDGWarning
                    )
                }
            }
        }
    }
}
