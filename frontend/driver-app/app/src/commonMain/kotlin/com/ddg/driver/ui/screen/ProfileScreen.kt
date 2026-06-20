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
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.AlertDialog
import androidx.compose.material.Button
import androidx.compose.material.ButtonDefaults
import androidx.compose.material.Card
import androidx.compose.material.Icon
import androidx.compose.material.MaterialTheme
import androidx.compose.material.OutlinedButton
import androidx.compose.material.Text
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Assignment
import androidx.compose.material.icons.filled.Badge
import androidx.compose.material.icons.filled.CarRental
import androidx.compose.material.icons.filled.ContactEmergency
import androidx.compose.material.icons.filled.ContactPhone
import androidx.compose.material.icons.filled.CreditCard
import androidx.compose.material.icons.filled.DateRange
import androidx.compose.material.icons.filled.DirectionsCar
import androidx.compose.material.icons.filled.ExitToApp
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.LocalOffer
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.Route
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material.icons.filled.Star
import androidx.compose.material.icons.filled.Timer
import androidx.compose.material.icons.filled.WorkspacePremium
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.ddg.driver.data.local.AppDataStore
import com.ddg.driver.data.model.DriverStatus
import com.ddg.driver.data.model.UserInfo
import com.ddg.driver.data.repository.AuthRepository
import com.ddg.driver.ui.theme.DDGInfo
import com.ddg.driver.ui.theme.DDGGray
import com.ddg.driver.ui.theme.DDGRed
import com.ddg.driver.ui.theme.DDGSuccess
import com.ddg.driver.ui.theme.DDGSurface
import com.ddg.driver.ui.theme.DDGTextPrimary
import com.ddg.driver.ui.theme.DDGTextSecondary
import com.ddg.driver.ui.theme.DDGWarning
import kotlinx.coroutines.launch
import org.koin.compose.getKoin

@Composable
fun ProfileScreen(
    onBack: () -> Unit,
    onLogout: () -> Unit
) {
    val dataStore: AppDataStore = getKoin().get()
    val authRepository: AuthRepository = getKoin().get()
    val scope = rememberCoroutineScope()
    val user by dataStore.user.collectAsState(initial = null)

    var showLogoutDialog by remember { mutableStateOf(false) }

    val currentUser = user ?: UserInfo(
        id = "1001",
        phone = "138****5678",
        name = "张三",
        avatar = "",
        idCard = "370***********1234",
        driverLicense = "A1A2",
        licenseExpireDate = 1900000000000L,
        qualificationCert = "道路危险品运输",
        rating = 4.8,
        totalTrips = 356,
        totalDistance = 185600.0,
        safetyScore = 95,
        companyName = "青岛安达危险品运输有限公司",
        department = "运输一部",
        emergencyContact = "李经理",
        emergencyPhone = "13900001111",
        joinDate = 1609459200000L,
        status = DriverStatus.ON_DUTY
    )

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(DDGSurface)
    ) {
        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            item {
                ProfileHeader(
                    user = currentUser,
                    onBack = onBack
                )
            }

            item {
                DriverStatsCard(user = currentUser)
            }

            item {
                ProfileSectionTitle("个人信息")
                ProfileMenuItem(
                    icon = Icons.Default.Badge,
                    title = "姓名",
                    value = currentUser.name
                )
                ProfileMenuItem(
                    icon = Icons.Default.ContactPhone,
                    title = "手机号",
                    value = currentUser.phone
                )
                ProfileMenuItem(
                    icon = Icons.Default.CreditCard,
                    title = "身份证号",
                    value = currentUser.idCard
                )
                ProfileMenuItem(
                    icon = Icons.Default.DirectionsCar,
                    title = "驾驶证类型",
                    value = currentUser.driverLicense
                )
                ProfileMenuItem(
                    icon = Icons.Default.WorkspacePremium,
                    title = "从业资格证",
                    value = currentUser.qualificationCert,
                    showDivider = false
                )
            }

            item {
                ProfileSectionTitle("单位信息")
                ProfileMenuItem(
                    icon = Icons.Default.CarRental,
                    title = "所属公司",
                    value = currentUser.companyName
                )
                ProfileMenuItem(
                    icon = Icons.Default.Badge,
                    title = "所属部门",
                    value = currentUser.department
                )
                ProfileMenuItem(
                    icon = Icons.Default.ContactEmergency,
                    title = "紧急联系人",
                    value = "${currentUser.emergencyContact} (${currentUser.emergencyPhone)",
                    showDivider = false
                )
            }

            item {
                ProfileSectionTitle("其他")
                ProfileMenuItem(
                    icon = Icons.Default.Info,
                    title = "入职日期",
                    value = formatDate(currentUser.joinDate)
                )
                ProfileMenuItem(
                    icon = Icons.Default.Settings,
                    title = "系统设置",
                    value = ""
                )
                ProfileMenuItem(
                    icon = Icons.Default.LocalOffer,
                    title = "帮助中心",
                    value = "",
                    showDivider = false
                )
            }

            item {
                Spacer(modifier = Modifier.height(8.dp))
                Button(
                    onClick = { showLogoutDialog = true },
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp)
                        .height(48.dp),
                    colors = ButtonDefaults.buttonColors(
                        backgroundColor = DDGRed
                    ),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Icon(
                        imageVector = Icons.Default.ExitToApp,
                        contentDescription = null,
                        tint = DDGTextPrimary
                    )
                    Spacer(modifier = Modifier.width(8.dp)
                    Text(
                        text = "退出登录",
                        style = MaterialTheme.typography.button
                    )
                }
                Spacer(modifier = Modifier.height(24.dp))
            }
        }
    }

    if (showLogoutDialog) {
        AlertDialog(
            onDismissRequest = { showLogoutDialog = false },
            title = { Text("确认退出") },
            text = { Text("确定要退出登录吗？") },
            confirmButton = {
                Button(
                    onClick = {
                        scope.launch {
                            authRepository.logout()
                            showLogoutDialog = false
                            onLogout()
                        }
                    },
                    colors = ButtonDefaults.buttonColors(backgroundColor = DDGRed)
                ) {
                    Text("确定")
                }
            },
            dismissButton = {
                OutlinedButton(
                    onClick = { showLogoutDialog = false }
                ) {
                    Text("取消")
                }
            }
        )
    }
}

@Composable
private fun rememberCoroutineScope(): kotlinx.coroutines.CoroutineScope {
    return kotlinx.coroutines.GlobalScope
}

@Composable
private fun ProfileHeader(
    user: UserInfo,
    onBack: () -> Unit
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .background(DDGGray)
            .padding(16.dp)
    ) {
        Column {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.fillMaxWidth()
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
                    text = "个人中心",
                    style = MaterialTheme.typography.h4,
                    fontWeight = FontWeight.Bold
                )
            }

            Spacer(modifier = Modifier.height(24.dp))

            Row(
                verticalAlignment = Alignment.CenterVertically
            ) {
                Box(
                    modifier = Modifier
                        .size(72.dp)
                        .clip(CircleShape)
                        .background(DDGRed.copy(alpha = 0.2f),
                    contentAlignment = Alignment.Center
                ) {
                    Icon(
                        imageVector = Icons.Default.Person,
                        contentDescription = null,
                        tint = DDGRed,
                        modifier = Modifier.size(40.dp)
                    )
                }
                Spacer(modifier = Modifier.width(16.dp))
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = user.name,
                        style = MaterialTheme.typography.h3,
                        fontWeight = FontWeight.Bold
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        StatusBadge(status = user.status)
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = user.department,
                            style = MaterialTheme.typography.body2
                        )
                    }
                }
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        imageVector = Icons.Default.Star,
                        contentDescription = null,
                        tint = DDGWarning,
                        modifier = Modifier.size(18.dp)
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = user.rating.toString(),
                        style = MaterialTheme.typography.h6,
                        fontWeight = FontWeight.Bold,
                        color = DDGWarning
                    )
                }
            }
        }
    }
}

@Composable
private fun StatusBadge(
    status: DriverStatus
) {
    val (text, color) = when (status) {
        DriverStatus.ON_DUTY -> "在岗" to DDGSuccess
        DriverStatus.OFF_DUTY -> "离岗" to DDGTextSecondary
        DriverStatus.RESTING -> "休息中" to DDGInfo
        DriverStatus.SUSPENDED -> "停职" to DDGWarning
    }
    Box(
        modifier = Modifier
            .clip(RoundedCornerShape(6.dp)
            .background(color.copy(alpha = 0.2f))
            .padding(horizontal = 8.dp, vertical = 2.dp)
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
private fun DriverStatsCard(
    user: UserInfo
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp),
        backgroundColor = DDGGray,
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceAround
            ) {
                StatItem(
                    icon = Icons.Default.Assignment,
                    value = "${user.totalTrips,
                    label = "总运单数",
                    color = DDGSuccess
                )
                StatItem(
                    icon = Icons.Default.Route,
                    value = "${"%.1f".format(user.totalDistance / 1000) + "万",
                    label = "总里程(km)",
                    color = DDGInfo
                )
                StatItem(
                    icon = Icons.Default.Favorite,
                    value = "${user.safetyScore}",
                    label = "安全评分",
                    color = DDGWarning
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
            modifier = Modifier.size(24.dp)
        )
        Spacer(modifier = Modifier.height(6.dp))
        Text(
            text = value,
            style = MaterialTheme.typography.h5,
            fontWeight = FontWeight.Bold
        )
        Spacer(modifier = Modifier.height(2.dp))
        Text(
            text = label,
            style = MaterialTheme.typography.caption,
            color = DDGTextSecondary
        )
    }
}

@Composable
private fun ProfileSectionTitle(title: String) {
    Text(
        text = title,
        style = MaterialTheme.typography.h6,
        color = DDGTextSecondary,
        modifier = Modifier.padding(start = 16.dp, top = 16.dp, bottom = 8.dp)
    )
}

@Composable
private fun ProfileMenuItem(
    icon: ImageVector,
    title: String,
    value: String,
    showDivider: Boolean = true
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(DDGGray)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 14.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier
                    .size(36.dp)
                    .clip(CircleShape)
                    .background(DDGSurface),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    imageVector = icon,
                    contentDescription = null,
                    tint = DDGRed,
                    modifier = Modifier.size(20.dp)
                )
            }
            Spacer(modifier = Modifier.width(12.dp))
            Text(
                text = title,
                style = MaterialTheme.typography.body1
            )
            Spacer(modifier = Modifier.weight(1f))
            if (value.isNotEmpty()) {
                Text(
                    text = value,
                    style = MaterialTheme.typography.body2,
                    color = DDGTextSecondary
                )
            }
        }
        if (showDivider) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(0.5.dp)
                    .padding(start = 64.dp)
                    .background(DDGSurface)
            )
        }
    }
}

private fun formatDate(timestamp: Long): String {
    val sdf = java.text.SimpleDateFormat("yyyy年MM月dd日")
    return sdf.format(java.util.Date(timestamp))
}
