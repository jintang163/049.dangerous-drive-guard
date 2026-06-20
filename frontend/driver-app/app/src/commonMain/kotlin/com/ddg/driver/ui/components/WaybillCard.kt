package com.ddg.driver.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.Card
import androidx.compose.material.Icon
import androidx.compose.material.LinearProgressIndicator
import androidx.compose.material.MaterialTheme
import androidx.compose.material.Text
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.LocalShipping
import androidx.compose.material.icons.filled.LocationOn
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.Route
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.ddg.driver.data.model.Waybill
import com.ddg.driver.data.model.WaybillStatus
import com.ddg.driver.ui.theme.DDGInfo
import com.ddg.driver.ui.theme.DDGGray
import com.ddg.driver.ui.theme.DDGRed
import com.ddg.driver.ui.theme.DDGSuccess
import com.ddg.driver.ui.theme.DDGTextHint
import com.ddg.driver.ui.theme.DDGTextPrimary
import com.ddg.driver.ui.theme.DDGTextSecondary
import com.ddg.driver.ui.theme.DDGWarning

@Composable
fun WaybillCard(
    waybill: Waybill?
) {
    if (waybill == null) {
        EmptyWaybillCard()
        return
    }

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
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        imageVector = Icons.Default.LocalShipping,
                        contentDescription = null,
                        tint = DDGRed,
                        modifier = Modifier.size(22.dp)
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "当前运单",
                        style = MaterialTheme.typography.h5,
                        fontWeight = FontWeight.Bold
                    )
                }
                StatusBadge(status = waybill.status)
            }

            Spacer(modifier = Modifier.height(12.dp))

            Text(
                text = "运单号: ${waybill.waybillNo}",
                style = MaterialTheme.typography.caption,
                color = DDGTextHint
            )

            Spacer(modifier = Modifier.height(8.dp))

            CargoInfoRow(
                cargoName = waybill.cargoName,
                cargoType = waybill.cargoType,
                cargoWeight = waybill.cargoWeight
            )

            Spacer(modifier = Modifier.height(12.dp))

            RouteInfo(
                origin = waybill.originAddress,
                dest = waybill.destAddress
            )

            Spacer(modifier = Modifier.height(12.dp))

            CrewInfo(
                driverName = waybill.driverName,
                driverPhone = waybill.driverPhone,
                escortName = waybill.escortName,
                escortPhone = waybill.escortPhone,
                plate = waybill.vehiclePlate
            )

            Spacer(modifier = Modifier.height(16.dp))

            ProgressSection(
                progress = waybill.progress,
                distanceCompleted = waybill.distanceCompleted,
                distanceTotal = waybill.distanceTotal
            )
        }
    }
}

@Composable
private fun EmptyWaybillCard() {
    Card(
        modifier = Modifier.fillMaxWidth(),
        backgroundColor = DDGGray,
        shape = RoundedCornerShape(12.dp)
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(180.dp),
            contentAlignment = Alignment.Center
        ) {
            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                Icon(
                    imageVector = Icons.Default.Route,
                    contentDescription = null,
                    tint = DDGTextHint,
                    modifier = Modifier.size(48.dp)
                )
                Spacer(modifier = Modifier.height(12.dp))
                Text(
                    text = "暂无进行中的运单",
                    style = MaterialTheme.typography.body1,
                    color = DDGTextHint
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = "请联系调度分配任务",
                    style = MaterialTheme.typography.caption,
                    color = DDGTextSecondary
                )
            }
        }
    }
}

@Composable
private fun StatusBadge(status: WaybillStatus) {
    val (text, color) = when (status) {
        WaybillStatus.PENDING -> "待发车" to DDGInfo
        WaybillStatus.IN_TRANSIT -> "运输中" to DDGSuccess
        WaybillStatus.STOPPED -> "临时停车" to DDGWarning
        WaybillStatus.COMPLETED -> "已完成" to DDGTextSecondary
        WaybillStatus.CANCELLED -> "已取消" to DDGTextHint
    }
    Box(
        modifier = Modifier
            .background(color.copy(alpha = 0.2f), RoundedCornerShape(6.dp))
            .padding(horizontal = 10.dp, vertical = 4.dp)
    ) {
        Text(
            text = text,
            style = MaterialTheme.typography.caption,
            color = color,
            fontWeight = FontWeight.Bold
        )
    }
}

@Composable
private fun CargoInfoRow(
    cargoName: String,
    cargoType: String,
    cargoWeight: Double
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(DDGRed.copy(alpha = 0.1f), RoundedCornerShape(8.dp))
            .padding(12.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Box(
            modifier = Modifier
                .size(36.dp)
                .background(DDGRed.copy(alpha = 0.2f), RoundedCornerShape(8.dp)),
            contentAlignment = Alignment.Center
        ) {
            Text(
                text = "⚠",
                color = DDGRed,
                style = MaterialTheme.typography.h5
            )
        }
        Spacer(modifier = Modifier.width(12.dp))
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = cargoName,
                style = MaterialTheme.typography.h6,
                fontWeight = FontWeight.SemiBold
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = "$cargoType · $cargoWeight 吨",
                style = MaterialTheme.typography.caption,
                color = DDGTextSecondary
            )
        }
    }
}

@Composable
private fun RouteInfo(
    origin: String,
    dest: String
) {
    Row(
        modifier = Modifier.fillMaxWidth()
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Box(
                modifier = Modifier
                    .size(12.dp)
                    .background(DDGSuccess, androidx.compose.foundation.shape.CircleShape)
            )
            Box(
                modifier = Modifier
                    .width(2.dp)
                    .height(32.dp)
                    .background(DDGInfo)
            )
            Box(
                modifier = Modifier
                    .size(12.dp)
                    .background(DDGRed, androidx.compose.foundation.shape.CircleShape)
            )
        }
        Spacer(modifier = Modifier.width(12.dp))
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = origin,
                style = MaterialTheme.typography.body1,
                fontWeight = FontWeight.Medium,
                modifier = Modifier.padding(top = 0.dp)
            )
            Spacer(modifier = Modifier.height(20.dp))
            Text(
                text = dest,
                style = MaterialTheme.typography.body1,
                fontWeight = FontWeight.Medium
            )
        }
        Spacer(modifier = Modifier.width(8.dp))
        Column {
            Text(
                text = "装货",
                style = MaterialTheme.typography.caption,
                color = DDGSuccess
            )
            Spacer(modifier = Modifier.height(20.dp))
            Text(
                text = "卸货",
                style = MaterialTheme.typography.caption,
                color = DDGRed
            )
        }
    }
}

@Composable
private fun CrewInfo(
    driverName: String,
    driverPhone: String,
    escortName: String,
    escortPhone: String,
    plate: String
) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        CrewItem(
            icon = Icons.Default.Person,
            title = "驾驶员",
            name = driverName,
            phone = driverPhone,
            color = DDGInfo,
            modifier = Modifier.weight(1f)
        )
        CrewItem(
            icon = Icons.Default.Person,
            title = "押运员",
            name = escortName,
            phone = escortPhone,
            color = DDGWarning,
            modifier = Modifier.weight(1f)
        )
    }
    Spacer(modifier = Modifier.height(8.dp))
    Row(verticalAlignment = Alignment.CenterVertically) {
        Icon(
            imageVector = Icons.Default.LocalShipping,
            contentDescription = null,
            tint = DDGTextSecondary,
            modifier = Modifier.size(16.dp)
        )
        Spacer(modifier = Modifier.width(6.dp))
        Text(
            text = "车牌: $plate",
            style = MaterialTheme.typography.body2,
            color = DDGTextSecondary
        )
    }
}

@Composable
private fun CrewItem(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    title: String,
    name: String,
    phone: String,
    color: Color,
    modifier: Modifier = Modifier
) {
    Column(modifier = modifier) {
        Text(
            text = title,
            style = MaterialTheme.typography.caption,
            color = DDGTextHint
        )
        Spacer(modifier = Modifier.height(4.dp))
        Text(
            text = name,
            style = MaterialTheme.typography.body2,
            fontWeight = FontWeight.Medium
        )
        Spacer(modifier = Modifier.height(2.dp))
        Text(
            text = phone,
            style = MaterialTheme.typography.caption,
            color = DDGTextSecondary
        )
    }
}

@Composable
private fun ProgressSection(
    progress: Int,
    distanceCompleted: Double,
    distanceTotal: Double
) {
    Column {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = "运输进度",
                style = MaterialTheme.typography.h6
            )
            Text(
                text = "${"%.1f".format(distanceCompleted)} / ${"%.1f".format(distanceTotal)} km",
                style = MaterialTheme.typography.caption,
                color = DDGTextSecondary
            )
        }
        Spacer(modifier = Modifier.height(8.dp))
        LinearProgressIndicator(
            progress = progress / 100f,
            modifier = Modifier
                .fillMaxWidth()
                .height(8.dp),
            backgroundColor = DDGTextHint.copy(alpha = 0.3f),
            color = DDGSuccess
        )
        Spacer(modifier = Modifier.height(6.dp))
        Text(
            text = "已完成 $progress%",
            style = MaterialTheme.typography.caption,
            color = DDGSuccess,
            fontWeight = FontWeight.Bold
        )
    }
}
