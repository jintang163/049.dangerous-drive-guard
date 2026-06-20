package com.ddg.driver

import androidx.compose.runtime.Composable
import com.tencent.kuikly.ability.AbilityPackage
import com.tencent.kuikly.runtime.HarmonyOSAbility

class MainAbility : HarmonyOSAbility() {

    override fun onCreate() {
        super.onCreate()
    }

    @Composable
    override fun Content() {
        App()
    }

    override fun onAbilityPackage(): AbilityPackage {
        return AbilityPackage(
            name = "MainAbility",
            label = "危险品运输护航"
        )
    }
}
