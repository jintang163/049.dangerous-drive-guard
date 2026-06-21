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
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.AlertDialog
import androidx.compose.material.Button
import androidx.compose.material.ButtonDefaults
import androidx.compose.material.Card
import androidx.compose.material.CircularProgressIndicator
import androidx.compose.material.Icon
import androidx.compose.material.MaterialTheme
import androidx.compose.material.OutlinedButton
import androidx.compose.material.Text
import androidx.compose.material.TextButton
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.ArrowForward
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Directions
import androidx.compose.material.icons.filled.LocationOn
import androidx.compose.material.icons.filled.LocalGasStation
import androidx.compose.material.icons.filled.MyLocation
import androidx.compose.material.icons.filled.Place
import androidx.compose.material.icons.filled.Speed
import androidx.compose.material.icons.filled.Timer
import androidx.compose.material.icons.filled.Warning
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.ddg.driver.data.model.FatigueLevel
import com.ddg.driver.data.model.ReplanCandidate
import com.ddg.driver.data.model.ReplanSuggestion
import com.ddg.driver.data.model.RoutePlan
import com.ddg.driver.data.model.ServiceArea
import com.ddg.driver.data.repository.ReplanRepository
import com.ddg.driver.ui.theme.DDGGray
import com.ddg.driver.ui.theme.DDGInfo
import com.ddg.driver.ui.theme.DDGPrimary
import com.ddg.driver.ui.theme.DDGRed
import com.ddg.driver.ui.theme.DDGSuccess
import com.ddg.driver.ui.theme.DDGSurface
import com.ddg.driver.ui.theme.DDGTextPrimary
import com.ddg.driver.ui.theme.DDGTextSecondary
import com.ddg.driver.ui.theme.DDGWarning
import com.ddg.driver.ui.components.FatigueIndicator
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.launch
import org.koin.core.context.GlobalContext

@Composable
fun NavigationScreen(
    onBack: () -> Unit,
    waybillId: Long? = null,
    vehicleId: Long? = null,
    driverId: Long? = null,
    initialRoutePlan: RoutePlan? = null
) {
    val replanRepo: ReplanRepository = remember { GlobalContext.get().get() }
    val scope = rememberCoroutineScope()

    var currentSpeed by remember { mutableStateOf(62.5) }
    var remainingDistance by remember { mutableStateOf(initialRoutePlan?.totalDistance ?: 158.3) }
    var remainingTime by remember { mutableStateOf((initialRoutePlan?.estimatedDuration ?: 126) / 60.0) }
    var fatigueProgress by remember { mutableStateOf(0.35f) }
    var nextTurn by remember { mutableStateOf("前方3.2公里右转进入G5高速") }

    var currentRoutePlan by remember { mutableStateOf<RoutePlan?>(initialRoutePlan) }
    var currentSuggestion by remember { mutableStateOf<ReplanSuggestion?>(null) }
    var confirmLoading by remember { mutableStateOf(false) }
    var toastMessage by remember { mutableStateOf<String?>(null) }

    val routeRedrawSignal = remember { mutableStateOf(0) }
    val eventLog = remember { mutableStateListOf<String>() }

    val nearbyServiceAreas = remember {
        listOf(
            ServiceArea(
                id = "1",
                name = "济南服务区",
                lat = 36.65,
                lng = 117.12,
                distanceFromOrigin = 45.0,
                hasRestRoom = true,
                hasFuelStation = true,
                hasRestaurant = true,
                hasParking = true,
                openHours = "24小时"
            ),
            ServiceArea(
                id = "2",
                name = "泰安服务区",
                lat = 36.20,
                lng = 117.10,
                distanceFromOrigin = 89.5,
                hasRestRoom = true,
                hasFuelStation = true,
                hasRestaurant = true,
                hasParking = true,
                openHours = "24小时"
            )
        )
    }

    LaunchedEffect(Unit) {
        replanRepo.connectWs(vehicleId, driverId)
        eventLog.add("✅ WebSocket 已连接")
    }

    LaunchedEffect(Unit) {
        replanRepo.observeReplanSuggestions().collectLatest { suggestion ->
            currentSuggestion = suggestion
            eventLog.add("⚠️ 收到重规划建议：${suggestion.replan_no} 原因：${suggestion.trigger_reason.take(20)}")
        }
    }

    LaunchedEffect(Unit) {
        replanRepo.observeRouteApplied().collectLatest { applied ->
            val result = replanRepo.getRoutePlan(applied.new_route_plan_id)
            result.onSuccess { plan ->
                currentRoutePlan = plan
                remainingDistance = plan.totalDistance
                remainingTime = plan.estimatedDuration / 60.0
                routeRedrawSignal.value++
                eventLog.add("✅ 新路线已应用，route_plan_id=${plan.id}，重绘导航地图")
                toastMessage = "路线已更新"
            }
            result.onFailure {
                eventLog.add("❌ 加载新路线失败：${it.message}")
            }
        }
    }

    LaunchedEffect(Unit) {
        replanRepo.observeTrafficEvents().collectLatest { payload ->
            eventLog.add("🛣️ 路况推送：${payload.take(40)}...")
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(DDGSurface)
    ) {
        Column(modifier = Modifier.fillMaxSize()) {
            TopNavBar(
                title = "导航中",
                onBack = onBack
            )

            MapPlaceholder(
                routePlan = currentRoutePlan,
                redrawSignal = routeRedrawSignal.value,
                suggestion = currentSuggestion
            )

            NavigationInfoBar(
                currentSpeed = currentSpeed,
                remainingDistance = remainingDistance,
                remainingTime = remainingTime,
                nextTurn = nextTurn
            )

            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                item {
                    FatigueMonitorCard(
                        progress = fatigueProgress,
                        level = FatigueLevel.WARNING
                    )
                }

                if (eventLog.isNotEmpty()) {
                    item {
                        EventLogCard(events = eventLog.takeLast(6))
                    }
                }

                item {
                    ServiceAreaRecommendations(
                        areas = nearbyServiceAreas
                    )
                }
            }
        }

        toastMessage?.let { msg ->
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 80.dp),
                contentAlignment = Alignment.TopCenter
            ) {
                Card(
                    backgroundColor = DDGSuccess,
                    shape = RoundedCornerShape(24.dp),
                    modifier = Modifier.padding(16.dp)
                ) {
                    Text(
                        text = msg,
                        color = Color.White,
                        modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp)
                    )
                }
                LaunchedEffect(msg) {
                    kotlinx.coroutines.delay(2000)
                    toastMessage = null
                }
            }
        }

        currentSuggestion?.let { sug ->
            ReplanSuggestDialog(
                suggestion = sug,
                loading = confirmLoading,
                onConfirm = {
                    scope.launch {
                        confirmLoading = true
                        val result = replanRepo.confirmReplan(sug.replan_id, "confirm", "司机确认重规划")
                        confirmLoading = false
                        result.onSuccess {
                            toastMessage = "已提交重规划确认"
                            currentSuggestion = null
                        }
                        result.onFailure {
                            toastMessage = "确认失败：${it.message}"
                        }
                    }
                },
                onReject = {
                    scope.launch {
                        confirmLoading = true
                        val result = replanRepo.confirmReplan(sug.replan_id, "reject", "司机选择不重规划")
                        confirmLoading = false
                        result.onSuccess {
                            currentSuggestion = null
                        }
                        result.onFailure {
                            toastMessage = "提交失败：${it.message}"
                        }
                    }
                },
                onDismiss = { currentSuggestion = null }
            )
        }
    }
}

@Composable
private fun TopNavBar(
    title: String,
    onBack: () -> Unit
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
        Spacer(modifier = Modifier.width(16.dp))
        Text(
            text = title,
            style = MaterialTheme.typography.h4,
            fontWeight = FontWeight.Bold
        )
    }
}

@Composable
private fun MapPlaceholder(
    routePlan: RoutePlan?,
    redrawSignal: Int,
    suggestion: ReplanSuggestion?
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(220.dp)
            .background(DDGGray)
            .padding(16.dp),
        contentAlignment = Alignment.Center
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Icon(
                imageVector = Icons.Default.MyLocation,
                contentDescription = null,
                tint = DDGRed,
                modifier = Modifier.size(64.dp)
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "高德/腾讯地图组件 [重绘次数: $redrawSignal]",
                style = MaterialTheme.typography.body2,
                color = DDGTextSecondary
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = routePlan?.let { "路线ID: ${it.id} · 策略: ${it.strategy}" }
                    ?: "起点: 青岛化工园  →  终点: 济南物流中心",
                style = MaterialTheme.typography.caption,
                color = DDGTextPrimary
            )
            if (suggestion != null) {
                Spacer(modifier = Modifier.height(8.dp))
                Card(
                    backgroundColor = DDGWarning.copy(alpha = 0.15f),
                    shape = RoundedCornerShape(8.dp),
                    modifier = Modifier.padding(horizontal = 16.dp)
                ) {
                    Text(
                        text = "⚠️ 待确认：${suggestion.trigger_reason.take(24)}",
                        color = DDGWarning,
                        style = MaterialTheme.typography.caption,
                        modifier = Modifier.padding(8.dp)
                    )
                }
            }
        }
    }
}

@Composable
private fun NavigationInfoBar(
    currentSpeed: Double,
    remainingDistance: Double,
    remainingTime: Double,
    nextTurn: String
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp),
        backgroundColor = DDGGray,
        shape = RoundedCornerShape(bottomStart = 12.dp, bottomEnd = 12.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    modifier = Modifier.weight(1f)
                ) {
                    Icon(
                        imageVector = Icons.Default.Directions,
                        contentDescription = null,
                        tint = DDGInfo,
                        modifier = Modifier.size(24.dp)
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = nextTurn,
                        style = MaterialTheme.typography.body1,
                        fontWeight = FontWeight.Medium
                    )
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceAround
            ) {
                NavInfoItem(
                    icon = Icons.Default.Speed,
                    value = "${currentSpeed.toInt()}",
                    unit = "km/h",
                    label = "当前速度",
                    color = DDGSuccess
                )
                NavInfoItem(
                    icon = Icons.Default.Place,
                    value = "${"%.1f".format(remainingDistance)}",
                    unit = "km",
                    label = "剩余里程",
                    color = DDGInfo
                )
                NavInfoItem(
                    icon = Icons.Default.Timer,
                    value = "${"%.1f".format(remainingTime)}",
                    unit = "h",
                    label = "预计时间",
                    color = DDGWarning
                )
            }
        }
    }
}

@Composable
private fun NavInfoItem(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    value: String,
    unit: String,
    label: String,
    color: Color
) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Icon(
            imageVector = icon,
            contentDescription = label,
            tint = color,
            modifier = Modifier.size(24.dp)
        )
        Spacer(modifier = Modifier.height(4.dp))
        Row(verticalAlignment = Alignment.Bottom) {
            Text(
                text = value,
                style = MaterialTheme.typography.h3,
                fontWeight = FontWeight.Bold,
                color = DDGTextPrimary
            )
            Spacer(modifier = Modifier.width(2.dp))
            Text(
                text = unit,
                style = MaterialTheme.typography.caption,
                color = DDGTextSecondary
            )
        }
        Spacer(modifier = Modifier.height(2.dp))
        Text(
            text = label,
            style = MaterialTheme.typography.caption,
            color = DDGTextSecondary
        )
    }
}

@Composable
private fun FatigueMonitorCard(
    progress: Float,
    level: FatigueLevel
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        backgroundColor = DDGGray,
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(
                text = "实时疲劳监测",
                style = MaterialTheme.typography.h5,
                fontWeight = FontWeight.Bold,
                modifier = Modifier.padding(bottom = 12.dp)
            )
            Row(
                verticalAlignment = Alignment.CenterVertically
            ) {
                FatigueIndicator(
                    progress = progress,
                    level = level,
                    size = 90.dp
                )
                Spacer(modifier = Modifier.width(20.dp))
                Column(modifier = Modifier.weight(1f)) {
                    val levelText = when (level) {
                        FatigueLevel.NORMAL -> "状态正常"
                        FatigueLevel.WARNING -> "轻度疲劳"
                        FatigueLevel.DANGEROUS -> "重度疲劳"
                    }
                    val levelColor = when (level) {
                        FatigueLevel.NORMAL -> DDGSuccess
                        FatigueLevel.WARNING -> DDGWarning
                        FatigueLevel.DANGEROUS -> DDGRed
                    }
                    Text(
                        text = levelText,
                        style = MaterialTheme.typography.h4,
                        color = levelColor,
                        fontWeight = FontWeight.Bold
                    )
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(
                        text = "注意力评分: ${(100 - progress * 100).toInt()}/100",
                        style = MaterialTheme.typography.body2
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = if (level == FatigueLevel.DANGEROUS)
                            "⚠️ 建议立即停车休息20分钟以上"
                        else
                            "提示: 注意保持良好驾驶状态",
                        style = MaterialTheme.typography.caption,
                        color = levelColor
                    )
                }
            }
        }
    }
}

@Composable
private fun EventLogCard(events: List<String>) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        backgroundColor = DDGGray,
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(
                text = "实时事件日志",
                style = MaterialTheme.typography.h5,
                fontWeight = FontWeight.Bold,
                modifier = Modifier.padding(bottom = 8.dp)
            )
            events.forEach { line ->
                Text(
                    text = "· $line",
                    style = MaterialTheme.typography.caption,
                    color = DDGTextSecondary,
                    modifier = Modifier.padding(vertical = 2.dp)
                )
            }
        }
    }
}

@Composable
private fun ServiceAreaRecommendations(
    areas: List<ServiceArea>
) {
    Column {
        Row(
            modifier = Modifier.padding(bottom = 12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(
                imageVector = Icons.Default.LocalGasStation,
                contentDescription = null,
                tint = DDGSuccess,
                modifier = Modifier.size(20.dp)
            )
            Spacer(modifier = Modifier.width(8.dp))
            Text(
                text = "附近服务区建议",
                style = MaterialTheme.typography.h5,
                fontWeight = FontWeight.Bold
            )
        }

        areas.forEach { area ->
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 8.dp),
                backgroundColor = DDGSurface,
                shape = RoundedCornerShape(8.dp)
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(12.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Box(
                        modifier = Modifier
                            .size(40.dp)
                            .clip(CircleShape)
                            .background(DDGGray),
                        contentAlignment = Alignment.Center
                    ) {
                        Icon(
                            imageVector = Icons.Default.LocationOn,
                            contentDescription = null,
                            tint = DDGRed,
                            modifier = Modifier.size(22.dp)
                        )
                    }
                    Spacer(modifier = Modifier.width(12.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Text(
                            text = area.name,
                            style = MaterialTheme.typography.body1,
                            fontWeight = FontWeight.SemiBold
                        )
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(
                            text = "距离 ${area.distanceFromOrigin}km · 营业: ${area.openHours}",
                            style = MaterialTheme.typography.caption,
                            color = DDGTextSecondary
                        )
                        Spacer(modifier = Modifier.height(4.dp))
                        Row {
                            if (area.hasRestRoom) Tag("卫生间")
                            if (area.hasFuelStation) Tag("加油")
                            if (area.hasRestaurant) Tag("餐饮")
                            if (area.hasParking) Tag("停车")
                        }
                    }
                    Icon(
                        imageVector = Icons.Default.ArrowForward,
                        contentDescription = null,
                        tint = DDGTextSecondary,
                        modifier = Modifier.size(20.dp)
                    )
                }
            }
        }
    }
}

@Composable
private fun Tag(text: String) {
    Box(
        modifier = Modifier
            .padding(end = 6.dp)
            .background(DDGGray, RoundedCornerShape(4.dp))
            .padding(horizontal = 8.dp, vertical = 2.dp)
    ) {
        Text(
            text = text,
            style = MaterialTheme.typography.caption,
            fontSize = 10.sp
        )
    }
}

@Composable
private fun ReplanSuggestDialog(
    suggestion: ReplanSuggestion,
    loading: Boolean,
    onConfirm: () -> Unit,
    onReject: () -> Unit,
    onDismiss: () -> Unit
) {
    AlertDialog(
        onDismissRequest = onDismiss,
        backgroundColor = DDGSurface,
        title = {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(
                    imageVector = Icons.Default.Warning,
                    tint = DDGWarning,
                    modifier = Modifier.size(28.dp)
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text("前方路况变化，建议重规划")
            }
        },
        text = {
            Column(modifier = Modifier.fillMaxWidth()) {
                Card(
                    backgroundColor = DDGWarning.copy(alpha = 0.12f),
                    shape = RoundedCornerShape(8.dp),
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Column(modifier = Modifier.padding(12.dp)) {
                        Text(
                            text = suggestion.trigger_reason,
                            style = MaterialTheme.typography.body1,
                            fontWeight = FontWeight.SemiBold,
                            color = DDGWarning
                        )
                        if (suggestion.traffic_event != null) {
                            Spacer(modifier = Modifier.height(6.dp))
                            Text(
                                text = "🛣️ ${suggestion.traffic_event.title} · ${suggestion.traffic_event.road_name.orEmpty()}",
                                style = MaterialTheme.typography.caption,
                                color = DDGTextSecondary
                            )
                        }
                    }
                }

                Spacer(modifier = Modifier.height(12.dp))

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    DeltaBox(
                        label = "里程变化",
                        current = suggestion.original_distance_remaining,
                        new = suggestion.new_distance_remaining,
                        suffix = " km"
                    )
                    DeltaBox(
                        label = "时长变化",
                        current = suggestion.original_duration_remaining.toDouble(),
                        new = suggestion.new_duration_remaining.toDouble(),
                        suffix = " 分"
                    )
                }

                if (suggestion.candidates.isNotEmpty()) {
                    Spacer(modifier = Modifier.height(12.dp))
                    Text(
                        "候选路线方案（${suggestion.candidates.size}）:",
                        style = MaterialTheme.typography.body2,
                        fontWeight = FontWeight.SemiBold,
                        modifier = Modifier.padding(bottom = 8.dp)
                    )
                    suggestion.candidates.forEach { c ->
                        CandidateCard(c)
                        Spacer(modifier = Modifier.height(6.dp))
                    }
                }
            }
        },
        buttons = {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(end = 16.dp, bottom = 16.dp),
                horizontalArrangement = Arrangement.spacedBy(8.dp, Alignment.End)
            ) {
                if (loading) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(24.dp),
                        strokeWidth = 2.dp,
                        color = DDGPrimary
                    )
                } else {
                    OutlinedButton(onClick = onReject) {
                        Icon(Icons.Default.Close, null, modifier = Modifier.size(16.dp))
                        Spacer(Modifier.width(4.dp))
                        Text("拒绝")
                    }
                    Button(
                        onClick = onConfirm,
                        colors = ButtonDefaults.buttonColors(backgroundColor = DDGSuccess)
                    ) {
                        Icon(Icons.Default.CheckCircle, null, modifier = Modifier.size(16.dp))
                        Spacer(Modifier.width(4.dp))
                        Text("确认新路线", color = Color.White)
                    }
                }
            }
        }
    )
}

@Composable
private fun DeltaBox(
    label: String,
    current: Double,
    new: Double,
    suffix: String
) {
    val delta = new - current
    val color = if (delta > 0) DDGRed else DDGSuccess
    Column {
        Text(
            text = label,
            style = MaterialTheme.typography.caption,
            color = DDGTextSecondary
        )
        Text(
            text = "${"%.1f".format(current)} → ${"%.1f".format(new)}$suffix",
            style = MaterialTheme.typography.body2,
            fontWeight = FontWeight.Medium
        )
        Text(
            text = "${if (delta >= 0) "+" else ""}${"%.1f".format(delta)}$suffix",
            color = color,
            style = MaterialTheme.typography.caption,
            fontWeight = FontWeight.SemiBold
        )
    }
}

@Composable
private fun CandidateCard(c: ReplanCandidate) {
    val bgColor = if (c.is_recommended == 1) DDGSuccess.copy(alpha = 0.12f) else DDGGray
    Card(
        backgroundColor = bgColor,
        shape = RoundedCornerShape(8.dp),
        modifier = Modifier.fillMaxWidth(),
        border = null
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(10.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Column(modifier = Modifier.weight(1f)) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(
                        text = c.strategy.uppercase(),
                        style = MaterialTheme.typography.body2,
                        fontWeight = FontWeight.SemiBold
                    )
                    if (c.is_recommended == 1) {
                        Spacer(Modifier.width(6.dp))
                        Box(
                            modifier = Modifier
                                .background(DDGSuccess, RoundedCornerShape(4.dp))
                                .padding(horizontal = 6.dp, vertical = 1.dp)
                        ) {
                            Text("推荐", color = Color.White, fontSize = 10.sp)
                        }
                    }
                }
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    "📏 ${"%.1f".format(c.total_distance)}km  ⏱️ ${c.estimated_duration}分  🛡️ 安全${c.safety_score ?: 0}",
                    style = MaterialTheme.typography.caption,
                    color = DDGTextSecondary
                )
            }
        }
    }
}
