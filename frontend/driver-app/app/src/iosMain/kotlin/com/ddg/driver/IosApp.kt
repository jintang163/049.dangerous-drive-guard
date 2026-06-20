package com.ddg.driver

import androidx.compose.ui.window.ComposeUIViewController
import platform.UIKit.UIApplicationDelegate
import platform.UIKit.UIApplicationDelegateProtocol
import platform.UIKit.UIResponder
import platform.UIKit.UIWindow

class AppDelegate : UIResponder, UIApplicationDelegateProtocol {
    override fun application(
        application: platform.UIKit.UIApplication,
        didFinishLaunchingWithOptions: Map<Any?, *>?
    ): Boolean {
        return true
    }
}

fun MainViewController() = ComposeUIViewController { App() }
