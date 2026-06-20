package com.ddg.driver.ui.screen

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.Button
import androidx.compose.material.ButtonDefaults
import androidx.compose.material.CircularProgressIndicator
import androidx.compose.material.Icon
import androidx.compose.material.MaterialTheme
import androidx.compose.material.OutlinedTextField
import androidx.compose.material.Text
import androidx.compose.material.TextFieldDefaults
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.LocalShipping
import androidx.compose.material.icons.filled.Lock
import androidx.compose.material.icons.filled.Phone
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.ddg.driver.domain.usecase.LoginUseCase
import com.ddg.driver.ui.theme.DDGRed
import com.ddg.driver.ui.theme.DDGRedDark
import com.ddg.driver.ui.theme.DDGSurface
import com.ddg.driver.ui.theme.DDGTextHint
import com.ddg.driver.ui.theme.DDGTextPrimary
import kotlinx.coroutines.launch
import org.koin.compose.getKoin

@Composable
fun LoginScreen(
    onLoginSuccess: () -> Unit
) {
    val loginUseCase: LoginUseCase = getKoin().get()
    val scope = rememberCoroutineScope()

    var phone by remember { mutableStateOf("") }
    var password by remember { mutableStateOf("") }
    var isLoading by remember { mutableStateOf(false) }
    var errorMessage by remember { mutableStateOf<String?>(null) }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(DDGSurface)
            .padding(24.dp),
        contentAlignment = Alignment.Center
    ) {
        Column(
            modifier = Modifier.fillMaxWidth(),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Icon(
                imageVector = Icons.Default.LocalShipping,
                contentDescription = null,
                modifier = Modifier.size(80.dp),
                tint = DDGRed
            )

            Spacer(modifier = Modifier.height(16.dp))

            Text(
                text = "危险品运输护航",
                style = MaterialTheme.typography.h1.copy(
                    color = DDGTextPrimary,
                    fontSize = 28.sp,
                    fontWeight = FontWeight.Bold
                )
            )

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = "驾驶员安全驾驶管理系统",
                style = MaterialTheme.typography.body2
            )

            Spacer(modifier = Modifier.height(48.dp))

            OutlinedTextField(
                value = phone,
                onValueChange = { phone = it },
                modifier = Modifier.fillMaxWidth(),
                label = { Text("手机号", color = DDGTextHint) },
                leadingIcon = {
                    Icon(Icons.Default.Phone, contentDescription = null, tint = DDGRed)
                },
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Phone),
                colors = TextFieldDefaults.outlinedTextFieldColors(
                    textColor = DDGTextPrimary,
                    cursorColor = DDGRed,
                    focusedBorderColor = DDGRed,
                    unfocusedBorderColor = DDGTextHint
                ),
                shape = RoundedCornerShape(12.dp),
                singleLine = true
            )

            Spacer(modifier = Modifier.height(16.dp))

            OutlinedTextField(
                value = password,
                onValueChange = { password = it },
                modifier = Modifier.fillMaxWidth(),
                label = { Text("密码", color = DDGTextHint) },
                leadingIcon = {
                    Icon(Icons.Default.Lock, contentDescription = null, tint = DDGRed)
                },
                visualTransformation = PasswordVisualTransformation(),
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password),
                colors = TextFieldDefaults.outlinedTextFieldColors(
                    textColor = DDGTextPrimary,
                    cursorColor = DDGRed,
                    focusedBorderColor = DDGRed,
                    unfocusedBorderColor = DDGTextHint
                ),
                shape = RoundedCornerShape(12.dp),
                singleLine = true
            )

            errorMessage?.let {
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    text = it,
                    color = Color.Red,
                    style = MaterialTheme.typography.caption
                )
            }

            Spacer(modifier = Modifier.height(32.dp))

            Button(
                onClick = {
                    scope.launch {
                        isLoading = true
                        errorMessage = null
                        val result = loginUseCase(phone, password)
                        isLoading = false
                        result.onSuccess {
                            onLoginSuccess()
                        }.onFailure {
                            errorMessage = it.message ?: "登录失败"
                        }
                    }
                },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(52.dp),
                enabled = !isLoading && phone.isNotBlank() && password.isNotBlank(),
                colors = ButtonDefaults.buttonColors(
                    backgroundColor = DDGRed,
                    disabledBackgroundColor = DDGRedDark
                ),
                shape = RoundedCornerShape(12.dp)
            ) {
                if (isLoading) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(24.dp),
                        color = DDGTextPrimary,
                        strokeWidth = 2.dp
                    )
                } else {
                    Text(
                        text = "登 录",
                        style = MaterialTheme.typography.button
                    )
                }
            }
        }
    }
}
