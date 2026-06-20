# Assets 目录说明

此目录用于存放AI模型文件和TTS语音资源。

## 需要放置的文件：

### 1. MediaPipe FaceLandmarker 模型
- **文件名**: `face_landmarker_v2_with_blendshapes.task`
- **来源**: 从 MediaPipe 官方下载
- **下载地址**: https://developers.google.com/mediapipe/solutions/vision/face_landmarker
- **用途**: 用于检测面部468个特征点

### 2. 自定义疲劳分类模型
- **文件名**: `fatigue_classifier.tflite`
- **格式**: TensorFlow Lite 模型
- **输入**: 特征向量 [EAR, MAR, PERCLOS, pitch, yaw, gaze_angle, ...]
- **输出**: 疲劳等级分类 (0-3: 正常/轻度/中度/重度)
- **用途**: 可选，用于辅助综合评分

### 3. 中文TTS语音资源（可选）
- **目录**: `phoneme_zh/`
- **格式**: PCM 音频文件
- **用途**: 如果需要使用离线TTS而非系统TTS，可在此放置中文音素资源文件

---

## 注意事项：

1. 模型文件较大（通常 5-20MB），请确保使用 Git LFS 管理或通过其他方式分发
2. 如果使用自定义的 `fatigue_classifier.tflite`，请确保输入输出维度与代码中的特征维度匹配
3. MediaPipe 模型支持 GPU 加速，已在代码中优先启用（失败时回退到 CPU）
