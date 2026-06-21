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
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.Button
import androidx.compose.material.ButtonDefaults
import androidx.compose.material.Card
import androidx.compose.material.CircularProgressIndicator
import androidx.compose.material.Icon
import androidx.compose.material.MaterialTheme
import androidx.compose.material.OutlinedTextField
import androidx.compose.material.Text
import androidx.compose.material.TextFieldDefaults
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Security
import androidx.compose.material.icons.filled.Star
import androidx.compose.material.icons.filled.StarBorder
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
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
import com.ddg.driver.data.model.SubmitReviewRequest
import com.ddg.driver.domain.usecase.SubmitServiceAreaReviewUseCase
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
fun ReviewScreen(
    serviceAreaId: Long,
    driverId: Long,
    vehicleId: Long,
    waybillId: Long,
    onSubmitSuccess: () -> Unit,
    onBack: () -> Unit
) {
    val submitReviewUseCase: SubmitServiceAreaReviewUseCase = getKoin().get()
    val scope = rememberCoroutineScope()

    var securityScore by remember { mutableStateOf(0) }
    var environmentScore by remember { mutableStateOf(0) }
    var foodScore by remember { mutableStateOf(0) }
    var serviceScore by remember { mutableStateOf(0) }
    var commentText by remember { mutableStateOf("") }
    var selectedTags by remember { mutableStateOf(setOf<String>()) }
    var isAnonymous by remember { mutableStateOf(false) }
    var isSubmitting by remember { mutableStateOf(false) }
    var successMessage by remember { mutableStateOf<String?>(null) }
    var errorMessage by remember { mutableStateOf<String?>(null) }

    val availableTags = listOf("安保好", "车位充足", "餐饮棒", "环境整洁", "服务好", "充电方便", "加油方便", "巡逻频繁")

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(DDGSurface)
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(16.dp)
        ) {
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
                    text = "服务区评价",
                    style = MaterialTheme.typography.h4,
                    fontWeight = FontWeight.Bold,
                    color = DDGTextPrimary
                )
            }

            Spacer(modifier = Modifier.height(24.dp))

            Text("安全性评分 *", style = MaterialTheme.typography.h6, color = DDGTextPrimary)
            Spacer(modifier = Modifier.height(8.dp))
            StarRating(
                score = securityScore,
                onScoreChange = { securityScore = it },
                label = "安全性"
            )

            Spacer(modifier = Modifier.height(20.dp))

            Text("环境评分", style = MaterialTheme.typography.subtitle1, color = DDGTextPrimary)
            Spacer(modifier = Modifier.height(8.dp))
            StarRating(
                score = environmentScore,
                onScoreChange = { environmentScore = it },
                label = "环境"
            )

            Spacer(modifier = Modifier.height(20.dp))

            Text("餐饮评分", style = MaterialTheme.typography.subtitle1, color = DDGTextPrimary)
            Spacer(modifier = Modifier.height(8.dp))
            StarRating(
                score = foodScore,
                onScoreChange = { foodScore = it },
                label = "餐饮"
            )

            Spacer(modifier = Modifier.height(20.dp))

            Text("服务评分", style = MaterialTheme.typography.subtitle1, color = DDGTextPrimary)
            Spacer(modifier = Modifier.height(8.dp))
            StarRating(
                score = serviceScore,
                onScoreChange = { serviceScore = it },
                label = "服务"
            )

            Spacer(modifier = Modifier.height(20.dp))

            Text("评价标签", style = MaterialTheme.typography.subtitle1, color = DDGTextPrimary)
            Spacer(modifier = Modifier.height(8.dp))
            LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                items(availableTags) { tag ->
                    val isSelected = tag in selectedTags
                    Card(
                        modifier = Modifier.clickable {
                            selectedTags = if (isSelected) selectedTags - tag else selectedTags + tag
                        },
                        backgroundColor = if (isSelected) DDGSuccess.copy(alpha = 0.2f) else DDGGray,
                        shape = RoundedCornerShape(16.dp)
                    ) {
                        Text(
                            tag,
                            modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp),
                            color = if (isSelected) DDGSuccess else DDGTextSecondary,
                            style = MaterialTheme.typography.caption
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(20.dp))

            Text("文字评价", style = MaterialTheme.typography.subtitle1, color = DDGTextPrimary)
            Spacer(modifier = Modifier.height(8.dp))
            OutlinedTextField(
                value = commentText,
                onValueChange = { commentText = it },
                modifier = Modifier.fillMaxWidth().height(100.dp),
                placeholder = { Text("分享您的体验...") },
                colors = TextFieldDefaults.outlinedTextFieldColors(
                    backgroundColor = DDGGray,
                    textColor = DDGTextPrimary,
                    cursorColor = DDGRed
                ),
                shape = RoundedCornerShape(12.dp)
            )

            Spacer(modifier = Modifier.height(16.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Card(
                    modifier = Modifier.clickable { isAnonymous = !isAnonymous },
                    backgroundColor = if (isAnonymous) DDGSuccess.copy(alpha = 0.2f) else DDGGray,
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Row(
                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(
                            Icons.Default.Security,
                            contentDescription = null,
                            tint = if (isAnonymous) DDGSuccess else DDGTextSecondary,
                            modifier = Modifier.size(18.dp)
                        )
                        Spacer(modifier = Modifier.width(6.dp))
                        Text(
                            "匿名评价",
                            style = MaterialTheme.typography.caption,
                            color = if (isAnonymous) DDGSuccess else DDGTextSecondary
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(24.dp))

            Button(
                onClick = {
                    if (securityScore == 0) {
                        errorMessage = "请至少给安全性评分"
                        return@Button
                    }
                    scope.launch {
                        isSubmitting = true
                        errorMessage = null
                        val result = submitReviewUseCase(
                            SubmitReviewRequest(
                                service_area_id = serviceAreaId,
                                driver_id = driverId,
                                security_score = securityScore,
                                environment_score = environmentScore,
                                food_score = foodScore,
                                service_score = serviceScore,
                                comment_text = commentText,
                                tags = selectedTags.toList(),
                                is_anonymous = isAnonymous,
                                waybill_id = waybillId,
                                vehicle_id = vehicleId
                            )
                        )
                        isSubmitting = false
                        result.onSuccess {
                            successMessage = "评价提交成功！"
                        }.onFailure { errorMessage = it.message }
                    }
                },
                modifier = Modifier.fillMaxWidth(),
                colors = ButtonDefaults.buttonColors(backgroundColor = DDGRed),
                enabled = !isSubmitting && securityScore > 0
            ) {
                if (isSubmitting) {
                    CircularProgressIndicator(color = Color.White, modifier = Modifier.size(20.dp))
                } else {
                    Icon(Icons.Default.Star, contentDescription = null)
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("提交评价", color = Color.White, fontWeight = FontWeight.Bold)
                }
            }

            Spacer(modifier = Modifier.height(12.dp))

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
                Spacer(modifier = Modifier.height(12.dp))
                OutlinedButton(
                    onClick = onSubmitSuccess,
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Text("完成", color = DDGSuccess)
                }
            }

            errorMessage?.let { msg ->
                Spacer(modifier = Modifier.height(8.dp))
                Text(msg, color = DDGRed, style = MaterialTheme.typography.caption)
            }
        }
    }
}

@Composable
private fun StarRating(
    score: Int,
    onScoreChange: (Int) -> Unit,
    label: String
) {
    Row(verticalAlignment = Alignment.CenterVertically) {
        (1..5).forEach { star ->
            Icon(
                imageVector = if (star <= score) Icons.Default.Star else Icons.Default.StarBorder,
                contentDescription = "$star 星",
                tint = if (star <= score) DDGWarning else DDGTextSecondary,
                modifier = Modifier
                    .size(36.dp)
                    .clickable { onScoreChange(star) }
            )
            Spacer(modifier = Modifier.width(4.dp))
        }
        Spacer(modifier = Modifier.width(12.dp))
        if (score > 0) {
            Text(
                "$score 分",
                style = MaterialTheme.typography.body1,
                fontWeight = FontWeight.Bold,
                color = when {
                    score >= 4 -> DDGSuccess
                    score >= 3 -> DDGWarning
                    else -> DDGRed
                }
            )
        }
    }
}

@Composable
private fun OutlinedButton(
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    content: @Composable () -> Unit
) {
    Card(
        modifier = modifier.clickable { onClick() },
        backgroundColor = Color.Transparent,
        shape = RoundedCornerShape(8.dp),
        elevation = 0.dp
    ) {
        Box(modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp)) {
            content()
        }
    }
}
