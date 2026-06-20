package com.ddg.driver.ui.screen

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
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
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.AlertDialog
import androidx.compose.material.Card
import androidx.compose.material.Icon
import androidx.compose.material.MaterialTheme
import androidx.compose.material.OutlinedButton
import androidx.compose.material.Text
import androidx.compose.material.Button
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Alarm
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.BrokenImage
import androidx.compose.material.icons.filled.CarCrash
import androidx.compose.material.icons.filled.CloudOff
import androidx.compose.material.icons.filled.Error
import androidx.compose.material.icons.filled.LocalFireDepartment
import androidx.compose.material.icons.filled.Report
import androidx.compose.material.icons.filled.Route
import androidx.compose.material.icons.filled.Speed
import androidx.compose.material.icons.filled.VisibilityOff
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.ddg.driver.data.model.AlarmInfo
import com.ddg.driver.data.model.AlarmLevel
import com.ddg.driver.data.model.AlarmStatus
import com.ddg.driver.data.model.AlarmType
import com.ddg.driver.data.repository.FatigueRepository
import com.ddg.driver.domain.usecase.ReportFatigueUseCase
import com.ddg.driver.ui.theme.DDGInfo
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
fun AlarmScreen(
    onBack: () -> Unit
) {
    val fatigueRepository: FatigueRepository = getKoin().get()

    var alarms by remember { mutableStateOf<List<AlarmInfo>?>(null) }
    var showFatigueAlert by remember { mutableStateOf(true) }

    LaunchedEffect(Unit) {
        kotlinx.coroutines.GlobalScope.launch {
            val result = fatigueRepository.getAlarmHistory()
            result.onSuccess { alarms = it }
        }
    }

    val alarmList = alarms ?: listOf(
        AlarmInfo(
            id = "1",
            type = AlarmType.FATIGUE,
            level = AlarmLevel.HIGH,
            title = "疲劳驾驶预警",
            description = "检测到连续驾驶超过4小时，建议立即休息",
            timestamp = System.currentTimeMillis() - 3600000,
            status = AlarmStatus.UNHANDLED,
            locationLat = 36.06,
            locationLng = 120.38,
            locationAddress = "山东省青岛市",
            waybillId = "W20240101001",
            imageUrl = null
        ),
        AlarmInfo(
            id = "2",
            type = AlarmType.OVERSPEED,
            level = AlarmLevel.MEDIUM,
            title = "超速提醒",
            description = "在G20高速路段超速10%",
            timestamp = System.currentTimeMillis() - 7200000,
            status = AlarmStatus.RESOLVED,
            locationLat = 36.50,
            locationLng = 120.50,
            locationAddress = "山东省潍坊市",
            waybillId = "W20240101001",
            imageUrl = null
        ),
        AlarmInfo(
            id = "3",
            type = AlarmType.LANE_DEPARTURE,
            level = AlarmLevel.LOW,
            title = "车道偏离",
            description = "检测到未打转向灯变道",
            timestamp = System.currentTimeMillis() - 10800000,
            status = AlarmStatus.RESOLVED,
            locationLat = 36.65,
            locationLng = 120.00,
            locationAddress = "山东省济南市",
            waybillId = "W20240101001",
            imageUrl = null
        )
    )

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(DDGSurface)
    ) {
        Column(modifier = Modifier.fillMaxSize() {
            TopBar(
                title = "报警记录",
                onBack = onBack,
                unhandledCount = alarmList.count { it.status == AlarmStatus.UNHANDLED }
            )

            AnimatedVisibility(
                visible = showFatigueAlert,
                enter = fadeIn(animationSpec = tween(300),
                exit = fadeOut(animationSpec = tween(300))
            ) {
                FatigueAlertBanner(
                    onDismiss = { showFatigueAlert = false }
                )
            }

            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                items(alarmList) { alarm ->
                    AlarmItem(alarm = alarm)
                }
            }
        }
    }
}

@Composable
private fun TopBar(
    title: String,
    onBack: () -> Unit,
    unhandledCount: Int
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Icon(
            imageVector = Icons.Default.ArrowBack,
            contentDescription = "返回",
            modifier = Modifier
                .size(28.dp)
                .clickable { onBack() },
            tint = DDGTextPrimary
        )
        Spacer(modifier = Modifier.width(16.dp)
        Text(
            text = title,
            style = MaterialTheme.typography.h4,
            fontWeight = FontWeight.Bold,
            modifier = Modifier.weight(1f)
        )
        if (unhandledCount > 0) {
            Box(
                modifier = Modifier
                    .clip(CircleShape)
                    .background(DDGRed)
                    .padding(horizontal = 10.dp, vertical = 4.dp)
            ) {
                Text(
                    text = "$unhandledCount 待处理",
                    style = MaterialTheme.typography.caption,
                    color = DDGTextPrimary,
                    fontWeight = FontWeight.Bold
                )
            }
        }
    }
}

@Composable
private fun FatigueAlertBanner(
    onDismiss: () -> Unit
) {
    var pulseScale by remember { mutableStateOf(1f) }
    val scale by animateFloatAsState(
        targetValue = pulseScale,
        animationSpec = tween(durationMillis = 600)
    )

    LaunchedEffect(Unit) {
        while (true) {
            pulseScale = 1.05f
            kotlinx.coroutines.delay(600)
            pulseScale = 1f
            kotlinx.coroutines.delay(600)
        }
    }

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp)
            .scale(scale),
        backgroundColor = DDGRed,
        shape = RoundedCornerShape(12.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(
                imageVector = Icons.Default.Alarm,
                contentDescription = null,
                tint = DDGTextPrimary,
                modifier = Modifier.size(36.dp)
            )
            Spacer(modifier = Modifier.width(12.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = "疲劳预警！",
                    style = MaterialTheme.typography.h5,
                    fontWeight = FontWeight.Bold,
                    color = DDGTextPrimary
                )
                Text(
                    text = "检测到疲劳特征，请立即停车休息",
                    style = MaterialTheme.typography.body2,
                    color = DDGTextPrimary
                )
            }
            Text(
                text = "知道了",
                style = MaterialTheme.typography.button,
                color = DDGTextPrimary,
                modifier = Modifier
                    .clickable { onDismiss() }
                    .background(
                        color = DDGRed,
                        shape = RoundedCornerShape(8.dp)
                    )
                    .padding(horizontal = 12.dp, vertical = 6.dp)
            )
        }
    }
}

@Composable
private fun AlarmItem(
    alarm: AlarmInfo
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        backgroundColor = DDGGray,
        shape = RoundedCornerShape(12.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            AlarmIcon(
                type = alarm.type,
                level = alarm.level
            )
            Spacer(modifier = Modifier.width(12.dp)
            Column(modifier = Modifier.weight(1f)) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Text(
                        text = alarm.title,
                        style = MaterialTheme.typography.h6,
                        fontWeight = FontWeight.SemiBold
                    )
                    AlarmStatusBadge(status = alarm.status)
                }
                Spacer(modifier = Modifier.height(6.dp))
                Text(
                    text = alarm.description,
                    style = MaterialTheme.typography.body2,
                    color = DDGTextSecondary
                )
                Spacer(modifier = Modifier.height(8.dp)) {
                    Icon(
                        imageVector = Icons.Default.Route,
                        contentDescription = null,
                        tint = DDGTextHint,
                        modifier = Modifier.size(14.dp)
                    )
                    Spacer(modifier = Modifier.width(4.dp)
                    Text(
                        text = alarm.locationAddress,
                        style = MaterialTheme.typography.caption,
                        color = DDGTextSecondary
                    )
                    Spacer(modifier = Modifier.width(12.dp)
                    Icon(
                        imageVector = Icons.Default.Report,
                        contentDescription = null,
                        tint = DDGTextHint,
                        modifier = Modifier.size(14.dp)
                    )
                    Spacer(modifier = Modifier.width(4.dp)
                    val timeStr = formatTime(alarm.timestamp)
                    Text(
                        text = timeStr,
                        style = MaterialTheme.typography.caption,
                        color = DDGTextSecondary
                    )
                }
            }
        }
    }
}

@Composable
private fun RowScope.AlarmIcon(
    type: AlarmType,
    level: AlarmLevel
) {
    val (icon, color) = when (type) {
        AlarmType.FATIGUE -> Icons.Default.VisibilityOff to DDGWarning
        AlarmType.OVERSPEED -> Icons.Default.Speed to DDGInfo
        AlarmType.LANE_DEPARTURE -> Icons.Default.Route to DDGWarning
        AlarmType.COLLISION_WARNING -> Icons.Default.CarCrash to DDGRed
        AlarmType.ABNORMAL_STOP -> Icons.Default.CloudOff to DDGInfo
        AlarmType.SOS -> Icons.Default.Error to DDGRed
        AlarmType.VEHICLE_ABNORMAL -> Icons.Default.BrokenImage to DDGWarning
        AlarmType.CARGO_ABNORMAL -> Icons.Default.LocalFireDepartment to DDGWarning
    }
    Box(
        modifier = Modifier
            .size(48.dp)
            .clip(CircleShape)
            .background(color.copy(alpha = 0.2f),
        contentAlignment = Alignment.Center
    ) {
        Icon(
            imageVector = icon,
            contentDescription = null,
            tint = color,
            modifier = Modifier.size(28.dp)
        )
    }
}

@Composable
private fun AlarmStatusBadge(
    status: AlarmStatus
) {
    val (text: String
    val color: Color
    when (status) {
        AlarmStatus.UNHANDLED -> {
            text = "待处理"
            color = DDGRed
        }
        AlarmStatus.PROCESSING -> {
            text = "处理中"
            color = DDGWarning
        }
        AlarmStatus.RESOLVED -> {
            text = "已解决"
            color = DDGSuccess
        }
        AlarmStatus.IGNORED -> {
            text = "已忽略"
            color = DDGTextSecondary
        }
    }
    Box(
        modifier = Modifier
            .clip(RoundedCornerShape(6.dp)
            .background(color.copy(alpha = 0.2f))
            .padding(horizontal = 8.dp, vertical = 3.dp)
    ) {
        Text(
            text = text,
            style = MaterialTheme.typography.overline,
            color = color,
            fontWeight = FontWeight.Bold
        )
    }
}

private fun formatTime(timestamp: Long): String {
    val current = System.currentTimeMillis()
    val diff = current - timestamp
    return when {
        diff < 3600000 -> "${diff / 60000}分钟前"
        diff < 86400000 -> "${diff / 3600000}小时前"
        else -> "${diff / 86400000}天前"
    }
}
