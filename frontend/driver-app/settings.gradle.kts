pluginManagement {
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
        maven("https://maven.pkg.jetbrains.space/public/p/compose/dev")
        maven("https://mirrors.tencent.com/nexus/repository/maven-public/")
    }
    plugins {
        kotlin("multiplatform").version("1.9.20")
        kotlin("plugin.serialization").version("1.9.20")
        id("com.android.application").version("8.1.4")
        id("com.android.library").version("8.1.4")
        id("org.jetbrains.compose").version("1.5.10")
        id("com.tencent.kuikly").version("0.5.0")
    }
}

dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        google()
        mavenCentral()
        maven("https://maven.pkg.jetbrains.space/public/p/compose/dev")
        maven("https://mirrors.tencent.com/nexus/repository/maven-public/")
    }
}

rootProject.name = "DangerousDriveGuard"
include(":app")
