package detector

import (
	"math"

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
