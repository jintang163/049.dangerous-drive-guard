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
import androidx.compose.material.Card
import androidx.compose.material.Icon
import androidx.compose.material.MaterialTheme
import androidx.compose.material.Text
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.ArrowForward
import androidx.compose.material.icons.filled.Directions
import androidx.compose.material.icons.filled.LocationOn
import androidx.compose.material.icons.filled.LocalGasStation
import androidx.compose.material.icons.filled.MyLocation
import androidx.compose.material.icons.filled.Place
import androidx.compose.material.icons.filled.Speed
import androidx.compose.material.icons.filled.Timer
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.ddg.driver.data.model.FatigueLevel
import com.ddg.driver.data.model.ServiceArea
import com.ddg.driver.ui.theme.DDGGray
import com.ddg.driver.ui.theme.DDGInfo
import com.ddg.driver.ui.theme.DDGRed
import com.ddg.driver.ui.theme.DDGSuccess
import com.ddg.driver.ui.theme.DDGSurface
import com.ddg.driver.ui.theme.DDGTextPrimary
import com.ddg.driver.ui.theme.DDGTextSecondary
import com.ddg.driver.ui.theme.DDGWarning
import com.ddg.driver.ui.components.FatigueIndicator

@Composable
fun NavigationScreen(
    onBack: () -> Unit
) {
    var currentSpeed by remember { mutableStateOf(62.5) }
    var remainingDistance by remember { mutableStateOf(158.3) }
    var remainingTime by remember { mutableStateOf(2.1) }
    var fatigueProgress by remember { mutableStateOf(0.35f) }
    var nextTurn by remember { mutableStateOf("前方3.2公里右转进入G5高速") }
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

            MapPlaceholder()

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

                item {
                    ServiceAreaRecommendations(
                        areas = nearbyServiceAreas
                    )
                }
            }
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
private fun MapPlaceholder() {
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
                text = "高德/腾讯地图组件",
                style = MaterialTheme.typography.body2,
                color = DDGTextSecondary
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = "起点: 青岛化工园  →  终点: 济南物流中心",
                style = MaterialTheme.typography.caption,
                color = DDGTextPrimary
            )
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
                    value = "${remainingDistance}",
                    unit = "km",
                    label = "剩余里程",
                    color = DDGInfo
                )
                NavInfoItem(
                    icon = Icons.Default.Timer,
                    value = "${remainingTime}",
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
