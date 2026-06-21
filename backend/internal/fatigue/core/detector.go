package detector

import (
	"fmt"
	"math"
	"strings"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
)

type FatigueDetector struct {
	config DetectorConfig
}

type DetectorConfig struct {
	PERCLOSThreshold     float64
	EyeClosedThreshold   float64
	YawnThreshold        float64
	HeadPitchThreshold   float64
	HeadYawThreshold     float64
	GazeDeviationThreshold float64
	FatigueScoreThreshold float64
	BlinkFrequencyLow    float64
	BlinkFrequencyHigh   float64
}

func DefaultDetectorConfig() DetectorConfig {
	return DetectorConfig{
		PERCLOSThreshold:     0.35,
		EyeClosedThreshold:   0.6,
		YawnThreshold:        0.55,
		HeadPitchThreshold:   30,
		HeadYawThreshold:     35,
		GazeDeviationThreshold: 0.4,
		FatigueScoreThreshold: 70,
		BlinkFrequencyLow:    5,
		BlinkFrequencyHigh:   30,
	}
}

func NewFatigueDetector(cfg DetectorConfig) *FatigueDetector {
	return &FatigueDetector{config: cfg}
}

func calcEAR(eye []model.FaceLandmark) float64 {
	if len(eye) < 6 {
		return 0
	}
	v1 := distance(eye[1], eye[5])
	v2 := distance(eye[2], eye[4])
	h := distance(eye[0], eye[3])
	if h == 0 {
		return 0
	}
	return (v1 + v2) / (2.0 * h)
}

func calcMAR(mouth []model.FaceLandmark) float64 {
	if len(mouth) < 10 {
		return 0
	}
	v1 := distance(mouth[2], mouth[10])
	v2 := distance(mouth[3], mouth[9])
	v3 := distance(mouth[4], mouth[8])
	h := distance(mouth[0], mouth[6])
	if h == 0 {
		return 0
	}
	return (v1 + v2 + v3) / (3.0 * h)
}

func distance(p1, p2 model.FaceLandmark) float64 {
	dx := float64(p1.X - p2.X)
	dy := float64(p1.Y - p2.Y)
	dz := float64(p1.Z - p2.Z)
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

func (d *FatigueDetector) Detect(landmarks model.FaceLandmarks, metrics model.FatigueMetrics, windowMetrics []model.FatigueMetrics) (float64, model.FatigueLevel, string, string) {
	score := 100.0
	var reasons []string
	var alarmType string

	perclos := metrics.PERCLOS
	if perclos <= 0 && landmarks.FaceDetected {
		leftEAR := calcEAR(landmarks.LeftEye)
		rightEAR := calcEAR(landmarks.RightEye)
		avgEAR := (leftEAR + rightEAR) / 2
		eyeClosed := 0.0
		if avgEAR < 0.2 {
			eyeClosed = 1.0
		} else if avgEAR < 0.3 {
			eyeClosed = 0.5
		}
		perclos = eyeClosed * d.config.PERCLOSThreshold * 1.5
	}

	if perclos > d.config.PERCLOSThreshold {
		penalty := (perclos - d.config.PERCLOSThreshold) * 150
		score -= math.Min(penalty, 35)
		reasons = append(reasons, "眼睛闭合时间过长")
		if alarmType == "" {
			alarmType = "fatigue_perclos"
		}
	}

	if metrics.EyeClosedRatio > d.config.EyeClosedThreshold {
		penalty := (metrics.EyeClosedRatio - d.config.EyeClosedThreshold) * 100
		score -= math.Min(penalty, 25)
		reasons = append(reasons, "双眼闭合")
		if alarmType == "" {
			alarmType = "eyes_closed"
		}
	}

	if metrics.BlinkFrequency < d.config.BlinkFrequencyLow {
		penalty := (d.config.BlinkFrequencyLow - metrics.BlinkFrequency) * 2
		score -= math.Min(penalty, 10)
		reasons = append(reasons, "眨眼频率过低(疑似瞌睡)")
	}
	if metrics.BlinkFrequency > d.config.BlinkFrequencyHigh {
		penalty := (metrics.BlinkFrequency - d.config.BlinkFrequencyHigh) * 0.5
		score -= math.Min(penalty, 8)
		reasons = append(reasons, "眨眼频率过高(眼睛不适)")
	}

	mar := metrics.MouthOpenRatio
	if mar <= 0 && landmarks.FaceDetected {
		mar = calcMAR(landmarks.Mouth)
	}
	if mar > d.config.YawnThreshold {
		penalty := (mar - d.config.YawnThreshold) * 100
		score -= math.Min(penalty, 20)
		reasons = append(reasons, "频繁打哈欠")
		if alarmType == "" {
			alarmType = "excessive_yawn"
		}
	}
	if metrics.YawnCount >= 3 {
		score -= 10
		reasons = append(reasons, "连续打哈欠")
		if alarmType == "" {
			alarmType = "excessive_yawn"
		}
	}

	if math.Abs(metrics.HeadPitch) > d.config.HeadPitchThreshold {
		penalty := (math.Abs(metrics.HeadPitch) - d.config.HeadPitchThreshold) * 1.5
		score -= math.Min(penalty, 20)
		if metrics.HeadPitch > 0 {
			reasons = append(reasons, "低头姿势")
		} else {
			reasons = append(reasons, "仰头姿势")
		}
		if alarmType == "" {
			alarmType = "abnormal_head_posture"
		}
	}

	if math.Abs(metrics.HeadYaw) > d.config.HeadYawThreshold {
		penalty := (math.Abs(metrics.HeadYaw) - d.config.HeadYawThreshold)
		score -= math.Min(penalty, 15)
		reasons = append(reasons, "左顾右盼")
		if alarmType == "" {
			alarmType = "gaze_distraction"
		}
	}

	if metrics.GazeDeviation > d.config.GazeDeviationThreshold {
		penalty := (metrics.GazeDeviation - d.config.GazeDeviationThreshold) * 80
		score -= math.Min(penalty, 20)
		reasons = append(reasons, "视线长时间偏离前方")
		if alarmType == "" {
			alarmType = "gaze_distraction"
		}
	}

	if metrics.PhoneUsageDetected {
		score -= 30
		reasons = append(reasons, "检测到使用手机")
		if alarmType == "" {
			alarmType = "phone_usage"
		}
	}

	if metrics.SmokingDetected {
		score -= 25
		reasons = append(reasons, "检测到抽烟行为")
		if alarmType == "" {
			alarmType = "smoking"
		}
	}

	if !metrics.SeatbeltOn {
		score -= 15
		reasons = append(reasons, "未系安全带")
		if alarmType == "" {
			alarmType = "no_seatbelt"
		}
	}

	continuousFatigue := 0
	if len(windowMetrics) >= 5 {
		for i := len(windowMetrics) - 5; i < len(windowMetrics); i++ {
			if windowMetrics[i].PERCLOS > d.config.PERCLOSThreshold*0.8 {
				continuousFatigue++
			}
		}
		if continuousFatigue >= 4 {
			score -= 15
			reasons = append(reasons, "连续疲劳状态")
			if alarmType == "" {
				alarmType = "continuous_fatigue"
			}
		}
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	level := model.FatigueNormal
	var msg string
	if score >= 70 {
		level = model.FatigueNormal
		msg = "状态正常，请继续保持良好驾驶习惯"
	} else if score >= d.config.FatigueScoreThreshold {
		level = model.FatigueWarning
		if len(reasons) > 0 {
			msg = "注意：" + reasons[0]
		} else {
			msg = "请注意驾驶状态，建议适当休息"
		}
	} else {
		level = model.FatigueFatigue
		if len(reasons) > 0 {
			msg = "严重警告：" + reasons[0] + "，请立即停车休息！"
		} else {
			msg = "严重疲劳！请立即停车休息！"
		}
	}

	return score, level, alarmType, msg
}

func (d *FatigueDetector) CalcPERCLOSFromFrames(frames []model.FaceLandmarks, threshold float64) float64 {
	if len(frames) == 0 {
		return 0
	}
	closedFrames := 0
	for _, fm := range frames {
		if !fm.FaceDetected {
			continue
		}
		leftEAR := calcEAR(fm.LeftEye)
		rightEAR := calcEAR(fm.RightEye)
		avgEAR := (leftEAR + rightEAR) / 2
		if avgEAR < threshold {
			closedFrames++
		}
	}
	return float64(closedFrames) / float64(len(frames))
}

func (d *FatigueDetector) DetectMultiCamera(frames []model.MultiCameraFrame, windowMetrics []model.FatigueMetrics) (*model.FusionResult, string, string) {
	cameraScores := make(map[string]float64)
	cameraConfidences := make(map[string]float64)
	cameraQualities := make(map[string]float64)
	var validCameras []string
	var bestCamera string
	bestQuality := 0.0

	for i := range frames {
		frame := &frames[i]
		pos := string(frame.Position)
		if !frame.FaceDetected {
			cameraScores[pos] = -1
			cameraConfidences[pos] = 0
			cameraQualities[pos] = 0
			continue
		}

		score, _, alarmType, _ := d.Detect(frame.Landmarks, frame.Metrics, windowMetrics)
		cameraScores[pos] = score
		cameraConfidences[pos] = frame.Confidence
		cameraQualities[pos] = frame.Quality

		if frame.Occluded || frame.Backlit {
			cameraConfidences[pos] *= 0.6
		}

		validCameras = append(validCameras, pos)
		if frame.Quality > bestQuality {
			bestQuality = frame.Quality
			bestCamera = pos
		}

		_ = alarmType
	}

	if len(validCameras) == 0 {
		return &model.FusionResult{
			FatigueScore:      50,
			FatigueLevel:      model.FatigueWarning,
			FusionMethod:      "no_face",
			FusionConfidence:  0,
			OcclusionDetected: true,
		}, "no_face_detected", "所有摄像头均未检测到人脸"
	}

	fusionScore, fusionConfidence, method := d.fuseScores(cameraScores, cameraConfidences, cameraQualities, validCameras)

	var occlusionDetected, backlitDetected bool
	for i := range frames {
		if frames[i].Occluded {
			occlusionDetected = true
		}
		if frames[i].Backlit {
			backlitDetected = true
		}
	}

	if len(validCameras) >= 2 {
		fusionConfidence = math.Min(fusionConfidence*1.15, 0.98)
	}
	if len(validCameras) >= 3 {
		fusionConfidence = math.Min(fusionConfidence*1.08, 0.98)
	}

	level := model.FatigueNormal
	if fusionScore < d.config.FatigueScoreThreshold {
		if fusionScore < 40 {
			level = model.FatigueFatigue
		} else {
			level = model.FatigueWarning
		}
	}

	primaryCamera := bestCamera
	if primaryCamera == "" && len(validCameras) > 0 {
		primaryCamera = validCameras[0]
	}

	result := &model.FusionResult{
		FatigueScore:      math.Round(fusionScore*100) / 100,
		FatigueLevel:      level,
		FusionMethod:      method,
		UsedCameras:       validCameras,
		PrimaryCamera:     primaryCamera,
		FusionConfidence:  math.Round(fusionConfidence*10000) / 10000,
		OcclusionDetected: occlusionDetected,
		BacklitDetected:   backlitDetected,
	}

	if s, ok := cameraScores["left"]; ok {
		result.LeftScore = math.Round(s*100) / 100
	}
	if s, ok := cameraScores["center"]; ok {
		result.CenterScore = math.Round(s*100) / 100
	}
	if s, ok := cameraScores["right"]; ok {
		result.RightScore = math.Round(s*100) / 100
	}

	var alarmType string
	if level == model.FatigueFatigue {
		alarmType = "multi_camera_fatigue"
	} else if level == model.FatigueWarning {
		alarmType = "multi_camera_warning"
	}

	msg := d.buildFusionMessage(result, cameraScores, occlusionDetected, backlitDetected)

	return result, alarmType, msg
}

func (d *FatigueDetector) fuseScores(scores map[string]float64, confidences, qualities map[string]float64, validCameras []string) (float64, float64, string) {
	if len(validCameras) == 0 {
		return 50, 0, "none"
	}

	if len(validCameras) == 1 {
		pos := validCameras[0]
		return scores[pos], confidences[pos], "single_camera"
	}

	var totalWeight float64
	var weightedScore float64
	var totalConfidence float64

	for _, pos := range validCameras {
		score := scores[pos]
		if score < 0 {
			continue
		}
		conf := confidences[pos]
		quality := qualities[pos]
		if quality <= 0 {
			quality = 0.5
		}

		var weight float64
		if pos == "center" {
			weight = 0.4 * conf * quality
		} else {
			weight = 0.3 * conf * quality
		}

		weightedScore += score * weight
		totalWeight += weight
		totalConfidence += conf * weight
	}

	if totalWeight == 0 {
		if s, ok := scores["center"]; ok && s >= 0 {
			return s, confidences["center"], "center_fallback"
		}
		for _, pos := range validCameras {
			if scores[pos] >= 0 {
				return scores[pos], confidences[pos], pos + "_fallback"
			}
		}
		return 50, 0, "none"
	}

	fusedScore := weightedScore / totalWeight
	fusedConfidence := totalConfidence / totalWeight
	method := "weighted_fusion_" + strings.Join(validCameras, "+")

	return fusedScore, fusedConfidence, method
}

func (d *FatigueDetector) buildFusionMessage(result *model.FusionResult, scores map[string]float64, occlusion, backlit bool) string {
	var parts []string

	if occlusion {
		parts = append(parts, "检测到遮挡，已切换多视角融合")
	}
	if backlit {
		parts = append(parts, "检测到逆光，已启用多视角补偿")
	}

	cameraScores := []string{}
	if result.LeftScore > 0 {
		cameraScores = append(cameraScores, fmt.Sprintf("左:%.0f", result.LeftScore))
	}
	if result.CenterScore > 0 {
		cameraScores = append(cameraScores, fmt.Sprintf("中:%.0f", result.CenterScore))
	}
	if result.RightScore > 0 {
		cameraScores = append(cameraScores, fmt.Sprintf("右:%.0f", result.RightScore))
	}
	if len(cameraScores) > 0 {
		parts = append(parts, "各视角评分["+strings.Join(cameraScores, " ")+"]")
	}

	parts = append(parts, fmt.Sprintf("融合置信度:%.1f%%", result.FusionConfidence*100))

	switch result.FatigueLevel {
	case model.FatigueFatigue:
		parts = append([]string{"严重疲劳！请立即停车休息！"}, parts...)
	case model.FatigueWarning:
		parts = append([]string{"请注意驾驶状态，建议适当休息"}, parts...)
	default:
		parts = append([]string{"状态正常"}, parts...)
	}

	return strings.Join(parts, "；")
}
