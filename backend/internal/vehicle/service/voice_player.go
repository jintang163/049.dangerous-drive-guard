package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/apache/rocketmq-clients/golang/v5"
	"gorm.io/gorm"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
	"github.com/dangerous-drive-guard/backend/pkg/mq"
)

type VoiceCommand struct {
	LogID            int64  `json:"log_id"`
	StrategyID       int64  `json:"strategy_id"`
	AudioID          int64  `json:"audio_id"`
	AudioURL         string `json:"audio_url"`
	AudioName        string `json:"audio_name"`
	AudioFormat      string `json:"audio_format"`
	VolumePercent    int    `json:"volume_percent"`
	ForceHighVolume  bool   `json:"force_high_volume"`
	PlayTimes        int    `json:"play_times"`
	PlayIntervalSec  int    `json:"play_interval_sec"`
	VehicleID        int64  `json:"vehicle_id"`
	DriverID         int64  `json:"driver_id"`
	AlarmID          int64  `json:"alarm_id"`
	StrategyType     string `json:"strategy_type"`
	AlarmLevel       int    `json:"alarm_level"`
	ContinuousMinutes int   `json:"continuous_minutes"`
}

type VoicePlayResult struct {
	LogID            int64  `json:"log_id"`
	VehicleID        int64  `json:"vehicle_id"`
	DriverID         int64  `json:"driver_id"`
	Status           string `json:"status"`
	ActualVolume     int    `json:"actual_volume"`
	PlayTimes        int    `json:"play_times"`
	ErrorMsg         string `json:"error_msg,omitempty"`
	PlayedAt         int64  `json:"played_at"`
	DurationMs       int64  `json:"duration_ms,omitempty"`
}

type VehicleVoicePlayer struct {
	db           *gorm.DB
	vehicleID    int64
	cacheDir     string
	isPlaying    bool
	playMu       sync.Mutex
	currentCmd   *exec.Cmd
	httpClient   *http.Client
	consumers    []*golang.SimpleConsumer
}

func NewVehicleVoicePlayer(vehicleID int64) *VehicleVoicePlayer {
	cacheDir := filepath.Join(os.TempDir(), "voice_cache", fmt.Sprintf("%d", vehicleID))
	_ = os.MkdirAll(cacheDir, 0755)

	return &VehicleVoicePlayer{
		db:         database.GetDB(),
		vehicleID:  vehicleID,
		cacheDir:   cacheDir,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *VehicleVoicePlayer) Start(ctx context.Context) error {
	logger.Sugar.Infof("[VoicePlayer] 启动车端语音播放服务，车辆ID: %d", p.vehicleID)

	topic1 := fmt.Sprintf("vehicle_%d_voice_command", p.vehicleID)
	topic2 := "voice_intervention_command"

	if err := p.subscribeTopic(ctx, topic1); err != nil {
		logger.Sugar.Errorf("[VoicePlayer] 订阅主题 %s 失败: %v", topic1, err)
	}
	if err := p.subscribeTopic(ctx, topic2); err != nil {
		logger.Sugar.Errorf("[VoicePlayer] 订阅主题 %s 失败: %v", topic2, err)
	}

	return nil
}

func (p *VehicleVoicePlayer) subscribeTopic(ctx context.Context, topicKey string) error {
	handler := func(ctx context.Context, msg *golang.MessageView) error {
		var cmd VoiceCommand
		if err := json.Unmarshal(msg.GetBody(), &cmd); err != nil {
			logger.Sugar.Errorf("[VoicePlayer] 解析语音指令失败: %v", err)
			return err
		}

		if cmd.VehicleID != 0 && cmd.VehicleID != p.vehicleID {
			logger.Sugar.Debugf("[VoicePlayer] 指令不属于本车辆 %d，忽略", cmd.VehicleID)
			return nil
		}

		go p.handleVoiceCommand(ctx, &cmd)
		return nil
	}

	if err := mq.StartConsumer(&config.Global.RocketMQ, topicKey, handler, 4); err != nil {
		return fmt.Errorf("start consumer for %s: %w", topicKey, err)
	}

	logger.Sugar.Infof("[VoicePlayer] 已订阅主题: %s", topicKey)
	return nil
}

func (p *VehicleVoicePlayer) handleVoiceCommand(ctx context.Context, cmd *VoiceCommand) {
	logger.Sugar.Infof("[VoicePlayer] 收到语音指令: log_id=%d, audio=%s, force_volume=%v, volume=%d%%",
		cmd.LogID, cmd.AudioName, cmd.ForceHighVolume, cmd.VolumePercent)

	result := &VoicePlayResult{
		LogID:     cmd.LogID,
		VehicleID: p.vehicleID,
		DriverID:  cmd.DriverID,
		PlayedAt:  time.Now().Unix(),
	}

	actualVolume := cmd.VolumePercent
	if actualVolume <= 0 {
		actualVolume = 70
	}
	if actualVolume > 100 {
		actualVolume = 100
	}

	if cmd.ForceHighVolume {
		actualVolume = 100
		logger.Sugar.Warnf("[VoicePlayer] 连续疲劳强制高音量模式，忽略车机音量设置，强制: %d%%", actualVolume)
		if err := p.setSystemVolume(actualVolume); err != nil {
			logger.Sugar.Errorf("[VoicePlayer] 设置系统音量失败: %v", err)
		}
	} else {
		if err := p.setSystemVolume(actualVolume); err != nil {
			logger.Sugar.Warnf("[VoicePlayer] 设置音量失败，使用默认播放: %v", err)
		}
	}
	result.ActualVolume = actualVolume

	audioPath, err := p.downloadAudio(cmd.AudioURL, cmd.AudioFormat)
	if err != nil {
		result.Status = "failed"
		result.ErrorMsg = fmt.Sprintf("下载音频失败: %v", err)
		logger.Sugar.Errorf("[VoicePlayer] %s", result.ErrorMsg)
		p.reportPlayResult(ctx, result)
		return
	}

	playTimes := cmd.PlayTimes
	if playTimes <= 0 {
		playTimes = 1
	}
	interval := cmd.PlayIntervalSec
	if interval < 0 {
		interval = 0
	}

	startTime := time.Now()
	successCount := 0

	for i := 0; i < playTimes; i++ {
		if err := p.playAudioFile(audioPath); err != nil {
			logger.Sugar.Errorf("[VoicePlayer] 第 %d 次播放失败: %v", i+1, err)
			continue
		}
		successCount++
		if i < playTimes-1 && interval > 0 {
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}

	result.PlayTimes = successCount
	result.DurationMs = time.Since(startTime).Milliseconds()

	if successCount == 0 {
		result.Status = "failed"
		result.ErrorMsg = "所有播放尝试均失败"
	} else if successCount < playTimes {
		result.Status = "partial"
		result.ErrorMsg = fmt.Sprintf("部分播放成功: %d/%d", successCount, playTimes)
	} else {
		result.Status = "completed"
	}

	logger.Sugar.Infof("[VoicePlayer] 播放完成: status=%s, times=%d/%d, duration=%dms",
		result.Status, successCount, playTimes, result.DurationMs)

	p.reportPlayResult(ctx, result)

	if cmd.ForceHighVolume {
		logger.Sugar.Infof("[VoicePlayer] 强制高音量播放完成，音量保持在 %d%% 直到司机确认", actualVolume)
	}
}

func (p *VehicleVoicePlayer) downloadAudio(audioURL, format string) (string, error) {
	if audioURL == "" {
		return "", fmt.Errorf("音频URL为空")
	}

	cacheFile := filepath.Join(p.cacheDir, fmt.Sprintf("%d_%d.%s",
		p.vehicleID, time.Now().UnixNano(), format))
	if format == "" {
		cacheFile += ".mp3"
	}

	if _, err := os.Stat(cacheFile); err == nil {
		logger.Sugar.Debugf("[VoicePlayer] 使用缓存音频: %s", cacheFile)
		return cacheFile, nil
	}

	logger.Sugar.Debugf("[VoicePlayer] 下载音频: %s -> %s", audioURL, cacheFile)

	resp, err := p.httpClient.Get(audioURL)
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP状态码异常: %d", resp.StatusCode)
	}

	out, err := os.Create(cacheFile)
	if err != nil {
		return "", fmt.Errorf("创建缓存文件失败: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", fmt.Errorf("写入缓存文件失败: %w", err)
	}

	return cacheFile, nil
}

func (p *VehicleVoicePlayer) playAudioFile(filePath string) error {
	p.playMu.Lock()
	defer p.playMu.Unlock()

	if p.isPlaying {
		logger.Sugar.Warnf("[VoicePlayer] 已有音频正在播放，跳过本次")
		return fmt.Errorf("busy")
	}
	p.isPlaying = true
	defer func() { p.isPlaying = false }()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("mpv"); err == nil {
			cmd = exec.Command("mpv", "--no-video", "--quiet", filePath)
		} else if _, err := exec.LookPath("ffplay"); err == nil {
			cmd = exec.Command("ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet", filePath)
		} else if _, err := exec.LookPath("aplay"); err == nil {
			cmd = exec.Command("aplay", "-q", filePath)
		} else {
			return fmt.Errorf("未找到可用的音频播放器 (mpv/ffplay/aplay)")
		}
	case "darwin":
		cmd = exec.Command("afplay", filePath)
	case "windows":
		cmd = exec.Command("powershell", "-Command",
			fmt.Sprintf("(New-Object Media.SoundPlayer '%s').PlaySync();", filePath))
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	p.currentCmd = cmd
	logger.Sugar.Debugf("[VoicePlayer] 执行播放命令: %v", cmd.Args)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动播放器失败: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("播放异常: %w", err)
	}

	return nil
}

func (p *VehicleVoicePlayer) setSystemVolume(percent int) error {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("amixer"); err == nil {
			cmd = exec.Command("amixer", "set", "Master", fmt.Sprintf("%d%%", percent))
		} else if _, err := exec.LookPath("pactl"); err == nil {
			cmd = exec.Command("pactl", "set-sink-volume", "@DEFAULT_SINK@", fmt.Sprintf("%d%%", percent))
		} else {
			return fmt.Errorf("未找到音量控制工具 (amixer/pactl)")
		}
	case "darwin":
		cmd = exec.Command("osascript", "-e",
			fmt.Sprintf("set volume output volume %d", percent))
	case "windows":
		cmd = exec.Command("powershell", "-Command",
			fmt.Sprintf("(New-Object -ComObject WScript.Shell).SendKeys([char]%d);", 175))
		logger.Sugar.Warnf("[VoicePlayer] Windows系统音量调节受限，设置: %d%%", percent)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	logger.Sugar.Debugf("[VoicePlayer] 设置系统音量: %d%%, cmd=%v", percent, cmd.Args)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行音量命令失败: %w", err)
	}

	return nil
}

func (p *VehicleVoicePlayer) reportPlayResult(ctx context.Context, result *VoicePlayResult) {
	resultBody, _ := json.Marshal(result)

	topic := "voice_intervention_result"
	key := fmt.Sprintf("%d-%d", result.VehicleID, result.LogID)

	if err := mq.Send(ctx, mq.Message{
		Topic: topic,
		Key:   key,
		Body:  resultBody,
	}); err != nil {
		logger.Sugar.Errorf("[VoicePlayer] 上报播放结果失败: %v", err)
	} else {
		logger.Sugar.Debugf("[VoicePlayer] 播放结果已上报到 %s, status=%s", topic, result.Status)
	}

	updateFields := map[string]interface{}{
		"play_status":          result.Status,
		"actual_volume_percent": result.ActualVolume,
		"play_times":           result.PlayTimes,
		"total_play_duration_ms": result.DurationMs,
		"updated_at":           time.Now(),
	}
	if result.Status == "completed" {
		updateFields["completed_at"] = time.Now()
	}
	if result.ErrorMsg != "" {
		updateFields["error_msg"] = result.ErrorMsg
	}

	if err := p.db.WithContext(ctx).Model(&model.VoiceInterventionLog{}).
		Where("id = ?", result.LogID).
		Updates(updateFields).Error; err != nil {
		logger.Sugar.Errorf("[VoicePlayer] 更新干预日志失败: %v", err)
	}
}

func (p *VehicleVoicePlayer) Stop() {
	p.playMu.Lock()
	defer p.playMu.Unlock()

	if p.currentCmd != nil && p.currentCmd.Process != nil {
		_ = p.currentCmd.Process.Kill()
	}
	p.isPlaying = false
	logger.Sugar.Infof("[VoicePlayer] 车端语音播放服务已停止，车辆ID: %d", p.vehicleID)
}
