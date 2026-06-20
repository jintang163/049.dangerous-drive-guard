package com.ddg.driver.ui.components

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.Card
import androidx.compose.material.Icon
import androidx.compose.material.MaterialTheme
import androidx.compose.material.Text
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.LocalPhone
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.scale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.ddg.driver.ui.theme.DDGRed
import com.ddg.driver.ui.theme.DDGRedDark
import com.ddg.driver.ui.theme.DDGSurface
import com.ddg.driver.ui.theme.DDGTextPrimary

@Composable
fun SOSButton(
    onClick: () -> Unit
) {
    var isPressed by remember { mutableStateOf(false) }
    val scale by animateFloatAsState(
        targetValue = if (isPressed) 0.95f else 1f,
        animationSpec = tween(durationMillis = 100)
    )

    Card(
        modifier = Modifier
            .fillMaxWidth(),
        backgroundColor = DDGSurface,
        shape = CircleShape
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .padding(8.dp),
            contentAlignment = Alignment.Center
        ) {
            Column(
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Box(
                    modifier = Modifier
                        .size(120.dp)
                        .scale(scale)
                        .background(
                            color = DDGRed,
                            shape = CircleShape
                        )
                        .clickable {
                            isPressed = true
                            onClick()
                            kotlinx.coroutines.GlobalScope.launch {
                                kotlinx.coroutines.delay(100)
                                isPressed = false
                            }
                        },
                    contentAlignment = Alignment.Center
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Icon(
                            imageVector = Icons.Default.LocalPhone,
                            contentDescription = null,
                            tint = DDGTextPrimary,
                            modifier = Modifier.size(40.dp)
                        )
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(
                            text = "SOS",
                            style = MaterialTheme.typography.h3,
                            fontWeight = FontWeight.Bold,
                            color = DDGTextPrimary,
                            fontSize = 28.sp
                        )
                    }
                }
                Spacer(modifier = Modifier.height(12.dp))
                Text(
                    text = "一键紧急求助",
                    style = MaterialTheme.typography.body2,
                    color = DDGRed
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = "紧急情况下点击联系救援中心",
                    style = MaterialTheme.typography.caption,
                    color = DDGTextPrimary.copy(alpha = 0.7f)
                )
            }
        }
    }
}

private fun kotlinx.coroutines.GlobalScope.launch(block: suspend kotlinx.coroutines.GlobalScope.() -> Unit): kotlinx.coroutines.Job {
    return kotlinx.coroutines.GlobalScope.launch(block = block)
}
