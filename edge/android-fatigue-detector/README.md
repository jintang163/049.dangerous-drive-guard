# 车载Android边缘端疲劳检测项目

## 项目概述

基于 **Kotlin + TensorFlow Lite + MediaPipe** 的车载Android边缘端驾驶员疲劳检测系统。

## 技术栈

| 模块 | 技术 |
|------|------|
| 语言 | Kotlin 1.9.20 |
| 构建 | AGP 8.2.0 |
| 人脸检测 | MediaPipe FaceLandmarker 0.10.9 |
| 深度学习推理 | TensorFlow Lite 2.15.0 |
| 相机 | CameraX 1.3.1 |
| 本地存储 | Room 2.6.1 |
| 网络请求 | Retrofit2 + OkHttp |
| 任务调度 | WorkManager |
| 配置存储 | DataStore |
| 异步处理 | Kotlin Coroutines + Flow |
| 目标架构 | arm64-v8a, armeabi-v7a |

## 项目结构

```
app/src/main/java/com/ddg/edge/fatigue/
├── MainActivity.kt                      # 全屏CameraX预览 + 悬浮窗UI
├── FatigueDetectionService.kt           # 前台常驻Service（锁屏继续检测）
│
├── detector/                            # AI检测模块
│   ├── FaceLandmarkDetector.kt          # MediaPipe 468点面部特征
│   ├── FatigueMetricsCalculator.kt      # EAR/MAR/头部姿态欧拉角
│   ├── PERCLOSCalculator.kt             # 滑动窗口PERCLOS统计
│   ├── FatigueScoreAggregator.kt        # 综合评分公式
│   ├── YawnDetector.kt                  # 哈欠持续时间检测
│   └── GazeDirectionDetector.kt         # 虹膜视线方向估算
│
├── model/                               # 数据模型
│   ├── DetectionFrameResult.kt          # 每帧检测结果
│   ├── FatigueAlarm.kt                  # 报警事件（3级）
│   └── StoredAlarm.kt                   # Room离线存储实体
│
├── data/                                # 数据层
│   ├── local/
│   │   ├── AlarmDatabase.kt             # Room数据库
│   │   ├── AlarmDao.kt                  # Data Access Object
│   │   └── OfflineAlarmRepository.kt    # 断网缓存+批量上报
│   └── remote/
│       ├── PlatformApiService.kt        # Retrofit接口定义
│       ├── AlarmUploadWorker.kt         # WorkManager定时上传
│       └── AlarmUploadWorkerDelegate.kt # 批量上传逻辑
│
├── alert/                               # 报警处置模块
│   ├── VoiceAlertManager.kt             # TTS语音合成
│   ├── SeatVibrationController.kt       # USB转串口PWM座椅震动
│   ├── AlarmReportManager.kt            # WS实时上报+HTTP
│   └── LocalVideoRecorder.kt            # 滑动15秒缓存（前10+后5秒）
│
├── tracking/
│   └── GPSTracker.kt                    # GPS/北斗双模定位
│
└── utils/
    ├── CameraSizeSelector.kt            # 640x480分辨率优选
    └── FrameMetadata.kt                 # 帧封装+Bitmap工具
```

## 核心算法公式

### EAR (Eye Aspect Ratio) 眼睛纵横比
```
EAR = (||p2-p6|| + ||p3-p5||) / (2 * ||p1-p4||)
阈值: EAR < 0.25 → 判定闭眼
```
- p1, p4: 眼睛外角、内角
- p2, p3: 上眼睑两个点
- p5, p6: 下眼睑两个点
- 左右眼取平均值

### MAR (Mouth Aspect Ratio) 嘴巴纵横比（10点法）
```
MAR = (Σ ||pi-pj||_vertical) / (4 * ||p1-p6||_horizontal)
其中: (p2,p10), (p3,p9), (p4,p8), (p5,p7) 共4对垂直距离
阈值: MAR > 0.6 且持续 > 1秒 → 判定哈欠
```

### PERCLOS (Percentage of Eyelid Closure)
```
PERCLOS = (EAR < 0.25 的帧数) / (滑动窗口内总帧数)
滑动窗口: 60秒
阈值: >0.10 预警, >0.20 危险
```

### 综合疲劳评分公式
```
Score = 100
      - PERCLOS × 300
      - MAR超限扣分 × 2
      - 低头(pitch>15°)扣分 × 1.5
      - 视线偏离(>15°)扣分 × 2
      - [哈欠额外扣5]

分级:
  ≥80 → 正常
  60-79 → 轻度 (LEVEL_1)
  40-59 → 中度 (LEVEL_2)
  <40  → 重度 (LEVEL_3)
```

### 头部姿态欧拉角（旋转矩阵转欧拉角）
基于 SolvePnP 原理，使用7个基准点（鼻尖、下巴、左右外眼角、左右嘴角、额头）拟合旋转向量后转欧拉角：
```
pitch = atan2(-R[2][1], R[2][2])  # 上下点头
yaw   = atan2( R[2][0], R[0][0])  # 左右转头
roll  = atan2(-R[0][1], R[0][0])  # 歪头
```

### 视线方向估算
基于虹膜中心相对眼角位置的归一化偏移量：
```
水平偏移 = (虹膜X - 眼角中点X) × 2 / 眼宽
角度 = atan2(偏移量, 1.0) 转角度
阈值: 合成角度 > 15° → 判定视线偏离
```

## 权限列表

| 权限 | 用途 |
|------|------|
| CAMERA | 前置红外相机采集驾驶员面部 |
| RECORD_AUDIO | 预留语音交互/环境录音 |
| INTERNET | 报警上报、配置拉取 |
| ACCESS_NETWORK_STATE | 网络状态监测 |
| FOREGROUND_SERVICE | 常驻检测前台服务 |
| FOREGROUND_SERVICE_MEDIA_PLAYBACK | TTS播放 |
| FOREGROUND_SERVICE_LOCATION | GPS定位 |
| WAKE_LOCK | 防止系统休眠中断检测 |
| ACCESS_FINE_LOCATION | GPS/北斗高精度定位 |
| ACCESS_COARSE_LOCATION | 基站/WiFi辅助定位 |
| SYSTEM_ALERT_WINDOW | 悬浮窗疲劳指数显示 |
| MODIFY_AUDIO_SETTINGS | 调节警报音量 |

## 关键特性

1. **常驻检测**: START_STICKY 前台Service，锁屏不中断
2. **多级预警**: 正常/轻度/中度/重度 4级状态机 + 冷却防抖
3. **多模态报警**: TTS语音 + 座椅PWM震动 + 视频取证
4. **边缘智能**: MediaPipe GPU推理，无需云端即可判断
5. **离线缓存**: Room本地存储报警，网络恢复WorkManager批量回传
6. **视频取证**: MediaCodec滑动15秒缓冲，报警触发保存前10秒+后5秒
7. **双通道上报**: WebSocket实时 + HTTP批量，握手失败自动降级
8. **位置联动**: GPS/北斗3秒间隔上报，报警绑定经纬度

## 构建说明

```bash
cd edge/android-fatigue-detector
./gradlew assembleDebug
```

### 前置条件

1. 安装 Android Studio Hedgehog (2023.1.1) 或更高版本
2. Android SDK Platform 33
3. Android NDK (用于 TensorFlow Lite GPU delegate)
4. 将以下模型放入 `app/src/main/assets/`:
   - `face_landmarker_v2_with_blendshapes.task` (MediaPipe官网下载)
   - 可选: `fatigue_classifier.tflite` (自定义分类模型)
