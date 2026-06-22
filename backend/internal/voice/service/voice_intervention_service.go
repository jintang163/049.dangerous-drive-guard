package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
	"github.com/dangerous-drive-guard/backend/pkg/mq"
)

type VoiceInterventionService struct {
	db               *gorm.DB
	lastIntervention map[int64]time.Time
	lastMu           sync.RWMutex
}

func NewVoiceInterventionService() *VoiceInterventionService {
	return &VoiceInterventionService{
		db:               database.GetDB(),
		lastIntervention: make(map[int64]time.Time),
	}
}

// ============================================================
// 音频库管理
// ============================================================

func (s *VoiceInterventionService) ListAudios(ctx context.Context, driverID, orgID int64, category model.AudioCategory, page, pageSize int) ([]*model.VoiceInterventionAudio, int64, error) {
	var audios []*model.VoiceInterventionAudio
	var total int64

	query := s.db.WithContext(ctx).Table("voice_intervention_audios")
	if driverID > 0 {
		query = query.Where("(driver_id = ? OR driver_id = 0)", driverID)
	}
	if orgID > 0 {
		query = query.Where("(org_id = ? OR org_id = 0)", orgID)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	err := query.Order("is_default DESC, category, created_at DESC").
		Offset(offset).Limit(pageSize).Find(&audios).Error
	return audios, total, err
}

func (s *VoiceInterventionService) GetAudio(ctx context.Context, id int64) (*model.VoiceInterventionAudio, error) {
	var audio model.VoiceInterventionAudio
	err := s.db.WithContext(ctx).First(&audio, id).Error
	if err != nil {
		return nil, err
	}
	return &audio, nil
}

func (s *VoiceInterventionService) CreateAudio(ctx context.Context, audio *model.VoiceInterventionAudio) error {
	return s.db.WithContext(ctx).Create(audio).Error
}

func (s *VoiceInterventionService) UpdateAudio(ctx context.Context, id int64, updates map[string]interface{}) error {
	return s.db.WithContext(ctx).Table("voice_intervention_audios").Where("id = ?", id).Updates(updates).Error
}

func (s *VoiceInterventionService) DeleteAudio(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Delete(&model.VoiceInterventionAudio{}, id).Error
}

func (s *VoiceInterventionService) SetDefaultAudio(ctx context.Context, driverID int64, audioID int64, category model.AudioCategory) error {
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	tx.Table("voice_intervention_audios").
		Where("(driver_id = ? OR driver_id = 0) AND category = ?", driverID, category).
		Update("is_default", false)
	err := tx.Table("voice_intervention_audios").Where("id = ?", audioID).Update("is_default", true).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// ============================================================
// 干预策略管理
// ============================================================

func (s *VoiceInterventionService) ListStrategies(ctx context.Context, driverID, orgID int64, strategyType model.InterventionStrategyType, page, pageSize int) ([]*model.VoiceInterventionStrategy, int64, error) {
	var strategies []*model.VoiceInterventionStrategy
	var total int64

	query := s.db.WithContext(ctx).Table("voice_intervention_strategies")
	if driverID > 0 {
		query = query.Where("(driver_id = ? OR driver_id = 0)", driverID)
	}
	if orgID > 0 {
		query = query.Where("(org_id = ? OR org_id = 0)", orgID)
	}
	if strategyType != "" {
		query = query.Where("strategy_type = ?", strategyType)
	}
	query.Count(&total)
	offset := (page - 1) * pageSize
	err := query.Order("priority, is_default DESC, created_at DESC").
		Offset(offset).Limit(pageSize).Find(&strategies).Error
	return strategies, total, err
}

func (s *VoiceInterventionService) GetStrategy(ctx context.Context, id int64) (*model.VoiceInterventionStrategy, error) {
	var strategy model.VoiceInterventionStrategy
	err := s.db.WithContext(ctx).First(&strategy, id).Error
	if err != nil {
		return nil, err
	}
	return &strategy, nil
}

func (s *VoiceInterventionService) CreateStrategy(ctx context.Context, strategy *model.VoiceInterventionStrategy) error {
	return s.db.WithContext(ctx).Create(strategy).Error
}

func (s *VoiceInterventionService) UpdateStrategy(ctx context.Context, id int64, updates map[string]interface{}) error {
	return s.db.WithContext(ctx).Table("voice_intervention_strategies").Where("id = ?", id).Updates(updates).Error
}

func (s *VoiceInterventionService) DeleteStrategy(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Delete(&model.VoiceInterventionStrategy{}, id).Error
}

// ============================================================
// 干预日志
// ============================================================

func (s *VoiceInterventionService) ListLogs(ctx context.Context, vehicleID, driverID, alarmID int64, status model.InterventionPlayStatus, startTime, endTime time.Time, page, pageSize int) ([]*model.VoiceInterventionLog, int64, error) {
	var logs []*model.VoiceInterventionLog
	var total int64

	query := s.db.WithContext(ctx).Table("voice_intervention_logs l").Select(`
		l.*, v.plate_number as vehicle_plate, u.real_name as driver_name
	`).Joins("LEFT JOIN vehicles v ON v.id = l.vehicle_id").
		Joins("LEFT JOIN users u ON u.id = l.driver_id")

	if vehicleID > 0 {
		query = query.Where("l.vehicle_id = ?", vehicleID)
	}
	if driverID > 0 {
		query = query.Where("l.driver_id = ?", driverID)
	}
	if alarmID > 0 {
		query = query.Where("l.alarm_id = ?", alarmID)
	}
	if status != "" {
		query = query.Where("l.play_status = ?", status)
	}
	if !startTime.IsZero() {
		query = query.Where("l.created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("l.created_at <= ?", endTime)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	err := query.Order("l.created_at DESC").Offset(offset).Limit(pageSize).Scan(&logs).Error
	return logs, total, err
}

func (s *VoiceInterventionService) UpdateLogPlayStatus(ctx context.Context, logID int64, status model.InterventionPlayStatus, errorMsg string, durationMs int64) error {
	updates := map[string]interface{}{
		"play_status": status,
	}
	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}
	if durationMs > 0 {
		updates["total_play_duration_ms"] = durationMs
	}
	if status == model.InterventionStatusSent {
		updates["sent_at"] = time.Now()
	}
	if status == model.InterventionStatusCompleted {
		updates["completed_at"] = time.Now()
	}
	return s.db.WithContext(ctx).Table("voice_intervention_logs").Where("id = ?", logID).Updates(updates).Error
}

func (s *VoiceInterventionService) DriverAckLog(ctx context.Context, logID int64) error {
	return s.db.WithContext(ctx).Table("voice_intervention_logs").
		Where("id = ?", logID).
		Updates(map[string]interface{}{
			"driver_ack": true,
			"ack_at":     time.Now(),
		}).Error
}

// ============================================================
// 策略匹配逻辑
// ============================================================

func (s *VoiceInterventionService) loadCandidateStrategies(ctx context.Context, driverID, orgID int64) ([]*model.VoiceInterventionStrategy, error) {
	var strategies []*model.VoiceInterventionStrategy
	err := s.db.WithContext(ctx).Table("voice_intervention_strategies").
		Where("is_enabled = 1 AND (driver_id = ? OR driver_id = 0) AND (org_id = ? OR org_id = 0)", driverID, orgID).
		Order("priority, is_default DESC").
		Find(&strategies).Error
	return strategies, err
}

func containsInt(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
func containsStr(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func (s *VoiceInterventionService) MatchStrategy(ctx context.Context, driverID, orgID int64, alarmLevel int, alarmType string, fatigueScore float64, continuousMinutes int) (*model.InterventionMatchedResult, error) {
	result := &model.InterventionMatchedResult{Matched: false}

	strategies, err := s.loadCandidateStrategies(ctx, driverID, orgID)
	if err != nil {
		return nil, err
	}
	if len(strategies) == 0 {
		result.Reason = "无可用策略"
		return result, nil
	}

	var matched *model.VoiceInterventionStrategy
	for _, st := range strategies {
		var trigger model.InterventionAlarmTrigger
		if len(st.AlarmTrigger) > 0 {
			_ = json.Unmarshal(st.AlarmTrigger, &trigger)
		}
		if len(trigger.AlarmLevels) > 0 && !containsInt(trigger.AlarmLevels, alarmLevel) {
			continue
		}
		if len(trigger.AlarmTypes) > 0 && !containsStr(trigger.AlarmTypes, alarmType) {
			continue
		}
		if trigger.MinContinuousMinutes > 0 && continuousMinutes < trigger.MinContinuousMinutes {
			continue
		}
		if trigger.MinFatigueScore > 0 && fatigueScore > trigger.MinFatigueScore {
			continue
		}
		matched = st
		break
	}
	if matched == nil {
		result.Reason = "无匹配策略"
		return result, nil
	}
	result.Matched = true
	result.Strategy = matched

	audioIDs := make([]int64, 0)
	if len(matched.AudioIDs) > 0 {
		_ = json.Unmarshal(matched.AudioIDs, &audioIDs)
	}
	var audios []*model.VoiceInterventionAudio
	if len(audioIDs) > 0 {
		s.db.WithContext(ctx).Table("voice_intervention_audios").
			Where("id IN ? AND is_enabled = 1", audioIDs).Find(&audios)
	}
	if len(audios) == 0 {
		s.db.WithContext(ctx).Table("voice_intervention_audios").
			Where("(driver_id = ? OR driver_id = 0) AND (org_id = ? OR org_id = 0) AND is_enabled = 1", driverID, orgID).
			Order("is_default DESC, category, created_at DESC").Limit(3).Find(&audios)
	}
	if len(audios) == 0 {
		result.Reason = "无可用音频"
		return result, nil
	}

	var selectedAudio *model.VoiceInterventionAudio
	if matched.EmotionalMode {
		for _, a := range audios {
			if a.Category == model.AudioCategoryFamily {
				selectedAudio = a
				break
			}
		}
	}
	if selectedAudio == nil && matched.ShuffleAudios && len(audios) > 1 {
		selectedAudio = audios[rand.Intn(len(audios))]
	}
	if selectedAudio == nil {
		selectedAudio = audios[0]
	}
	result.SelectedAudio = selectedAudio
	result.IsHighVolume = matched.ForceHighVolume
	result.VolumePercent = matched.ForceVolumePercent
	if result.VolumePercent <= 0 {
		result.VolumePercent = selectedAudio.Volume
	}

	return result, nil
}

// ============================================================
// 触发语音干预（供疲劳服务回调使用）
// ============================================================

func (s *VoiceInterventionService) TriggerIntervention(ctx context.Context, alarm *model.FatigueAlarm) (*model.VoiceInterventionLog, error) {
	s.lastMu.Lock()
	lastAt, ok := s.lastIntervention[alarm.DriverID]
	now := time.Now()
	s.lastMu.Unlock()

	matchResult, err := s.MatchStrategy(ctx, alarm.DriverID, 0,
		int(alarm.AlarmLevel), alarm.AlarmType, alarm.FatigueScore, alarm.ContinuousFatigueMinutes)
	if err != nil {
		logger.Sugar.Errorf("match strategy error: %v", err)
		return nil, err
	}
	if !matchResult.Matched || matchResult.SelectedAudio == nil {
		logger.Global.Debug("no voice intervention matched",
			zap.Int64("driver_id", alarm.DriverID),
			zap.String("reason", matchResult.Reason))
		return nil, nil
	}

	if ok && matchResult.Strategy.CooldownSeconds > 0 {
		if int(now.Sub(lastAt).Seconds()) < matchResult.Strategy.CooldownSeconds {
			logger.Global.Debug("voice intervention cooldown skipped",
				zap.Int64("driver_id", alarm.DriverID))
			return nil, nil
		}
	}

	audio := matchResult.SelectedAudio
	strategy := matchResult.Strategy
	logEntry := &model.VoiceInterventionLog{
		VehicleID:           alarm.VehicleID,
		DriverID:            alarm.DriverID,
		WaybillID:           alarm.WaybillID,
		AlarmID:             alarm.ID,
		StrategyID:          strategy.ID,
		AudioID:             audio.ID,
		AudioName:           audio.Name,
		AudioURL:            audio.AudioURL,
		AudioFormat:         audio.AudioFormat,
		Category:            audio.Category,
		StrategyType:        strategy.StrategyType,
		PlayStatus:          model.InterventionStatusPending,
		IsHighVolume:        matchResult.IsHighVolume,
		ActualVolumePercent: matchResult.VolumePercent,
		PlayTimes:           strategy.PlayTimes,
		AlarmLevel:          int(alarm.AlarmLevel),
		AlarmType:           alarm.AlarmType,
		FatigueScore:        alarm.FatigueScore,
		ContinuousMinutes:   alarm.ContinuousFatigueMinutes,
	}

	err = s.db.WithContext(ctx).Create(logEntry).Error
	if err != nil {
		return nil, err
	}

	msgID, err := s.dispatchToVehicle(ctx, logEntry, strategy)
	if err != nil {
		_ = s.UpdateLogPlayStatus(ctx, logEntry.ID, model.InterventionStatusFailed, err.Error(), 0)
		return logEntry, err
	}
	logEntry.MQMessageID = msgID
	_ = s.UpdateLogPlayStatus(ctx, logEntry.ID, model.InterventionStatusSent, "", 0)

	s.lastMu.Lock()
	s.lastIntervention[alarm.DriverID] = now
	s.lastMu.Unlock()

	s.db.WithContext(ctx).Table("voice_intervention_audios").
		Where("id = ?", audio.ID).
		Update("play_count", gorm.Expr("play_count + 1"))

	logger.Global.Info("voice intervention triggered",
		zap.Int64("log_id", logEntry.ID),
		zap.Int64("driver_id", alarm.DriverID),
		zap.String("strategy_type", string(strategy.StrategyType)),
		zap.String("audio", audio.Name),
		zap.Bool("high_volume", matchResult.IsHighVolume),
		zap.Int("volume", matchResult.VolumePercent))

	return logEntry, nil
}

func (s *VoiceInterventionService) dispatchToVehicle(ctx context.Context, log *model.VoiceInterventionLog, strategy *model.VoiceInterventionStrategy) (string, error) {
	msgID := uuid.New().String()
	body, _ := json.Marshal(map[string]interface{}{
		"log_id":              log.ID,
		"strategy_id":         strategy.ID,
		"audio_id":            log.AudioID,
		"audio_url":           log.AudioURL,
		"audio_name":          log.AudioName,
		"audio_format":        log.AudioFormat,
		"volume_percent":      log.ActualVolumePercent,
		"force_high_volume":   log.IsHighVolume,
		"play_times":          strategy.PlayTimes,
		"play_interval_sec":   strategy.PlayIntervalSec,
		"vehicle_id":          log.VehicleID,
		"driver_id":           log.DriverID,
		"alarm_id":            log.AlarmID,
		"strategy_type":       strategy.StrategyType,
		"alarm_level":         log.AlarmLevel,
		"continuous_minutes":  log.ContinuousMinutes,
	})
	err := mq.Send(ctx, mq.Message{
		Topic: fmt.Sprintf("vehicle_%d_voice_command", log.VehicleID),
		Key:   msgID,
		Body:  body,
	})
	if err != nil {
		return "", err
	}
	_ = mq.Send(ctx, mq.Message{
		Topic: "voice_intervention_command",
		Key:   msgID,
		Body:  body,
	})
	return msgID, nil
}

// ============================================================
// 测试播放
// ============================================================

func (s *VoiceInterventionService) TestPlayAudio(ctx context.Context, vehicleID, audioID, volume int, createdBy int64) (*model.VoiceInterventionLog, error) {
	audio, err := s.GetAudio(ctx, int64(audioID))
	if err != nil {
		return nil, err
	}
	logEntry := &model.VoiceInterventionLog{
		VehicleID:           int64(vehicleID),
		DriverID:            0,
		StrategyID:          0,
		AudioID:             audio.ID,
		AudioName:           audio.Name,
		AudioURL:            audio.AudioURL,
		AudioFormat:         audio.AudioFormat,
		Category:            audio.Category,
		StrategyType:        model.StrategyTypeNormal,
		PlayStatus:          model.InterventionStatusPending,
		IsHighVolume:        volume >= 90,
		ActualVolumePercent: volume,
		PlayTimes:           1,
	}
	if volume <= 0 {
		logEntry.ActualVolumePercent = audio.Volume
	}

	err = s.db.WithContext(ctx).Create(logEntry).Error
	if err != nil {
		return nil, err
	}
	strategy := &model.VoiceInterventionStrategy{PlayTimes: 1, PlayIntervalSec: 0}
	msgID, err := s.dispatchToVehicle(ctx, logEntry, strategy)
	if err != nil {
		_ = s.UpdateLogPlayStatus(ctx, logEntry.ID, model.InterventionStatusFailed, err.Error(), 0)
		return logEntry, err
	}
	logEntry.MQMessageID = msgID
	_ = s.UpdateLogPlayStatus(ctx, logEntry.ID, model.InterventionStatusSent, "", 0)
	return logEntry, nil
}

// ============================================================
// 统计
// ============================================================

type InterventionStats struct {
	TotalCount          int64  `json:"total_count"`
	CompletedCount      int64  `json:"completed_count"`
	FailedCount         int64  `json:"failed_count"`
	HighVolumeCount     int64  `json:"high_volume_count"`
	FamilyAudioCount    int64  `json:"family_audio_count"`
	ContinuousTriggered int64  `json:"continuous_triggered"`
	SevereTriggered     int64  `json:"severe_triggered"`
	TotalPlayDurationMs int64  `json:"total_play_duration_ms"`
	DriverAckRate       string `json:"driver_ack_rate"`
}

func (s *VoiceInterventionService) GetStatistics(ctx context.Context, days int) (*InterventionStats, error) {
	if days <= 0 {
		days = 30
	}
	stats := &InterventionStats{}
	row := s.db.WithContext(ctx).Table("voice_intervention_logs").
		Where("created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)", days).
		Select(`
			COUNT(*) as total,
			SUM(CASE WHEN play_status='completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN play_status='failed' THEN 1 ELSE 0 END) as failed,
			SUM(CASE WHEN is_high_volume=1 THEN 1 ELSE 0 END) as high_vol,
			SUM(CASE WHEN category='family' THEN 1 ELSE 0 END) as family_cnt,
			SUM(CASE WHEN strategy_type='continuous' THEN 1 ELSE 0 END) as continuous_cnt,
			SUM(CASE WHEN strategy_type='severe' THEN 1 ELSE 0 END) as severe_cnt,
			COALESCE(SUM(total_play_duration_ms),0) as total_dur,
			SUM(CASE WHEN driver_ack=1 THEN 1 ELSE 0 END) as ack_cnt
		`).Row()
	var total, completed, failed, highVol, familyCnt, contCnt, sevCnt, durMs, ackCnt int64
	row.Scan(&total, &completed, &failed, &highVol, &familyCnt, &contCnt, &sevCnt, &durMs, &ackCnt)
	stats.TotalCount = total
	stats.CompletedCount = completed
	stats.FailedCount = failed
	stats.HighVolumeCount = highVol
	stats.FamilyAudioCount = familyCnt
	stats.ContinuousTriggered = contCnt
	stats.SevereTriggered = sevCnt
	stats.TotalPlayDurationMs = durMs
	if total > 0 {
		stats.DriverAckRate = fmt.Sprintf("%.1f%%", float64(ackCnt)/float64(total)*100)
	} else {
		stats.DriverAckRate = "0%"
	}
	return stats, nil
}

// ============================================================
// 播放结果回传 MQ Consumer
// ============================================================

type VoicePlayResult struct {
	LogID        int64  `json:"log_id"`
	VehicleID    int64  `json:"vehicle_id"`
	DriverID     int64  `json:"driver_id"`
	Status       string `json:"status"`
	ActualVolume int    `json:"actual_volume"`
	PlayTimes    int    `json:"play_times"`
	ErrorMsg     string `json:"error_msg,omitempty"`
	PlayedAt     int64  `json:"played_at"`
	DurationMs   int64  `json:"duration_ms,omitempty"`
}

func (s *VoiceInterventionService) StartResultConsumer(ctx context.Context) error {
	handler := func(ctx context.Context, msg *golang.MessageView) error {
		var result VoicePlayResult
		if err := json.Unmarshal(msg.GetBody(), &result); err != nil {
			logger.Sugar.Errorf("[VoiceResult] 解析播放结果失败: %v", err)
			return err
		}

		logger.Sugar.Infof("[VoiceResult] 收到播放结果: log_id=%d, vehicle=%d, status=%s, volume=%d%%, times=%d",
			result.LogID, result.VehicleID, result.Status, result.ActualVolume, result.PlayTimes)

		updateFields := map[string]interface{}{
			"play_status":           result.Status,
			"actual_volume_percent": result.ActualVolume,
			"play_times":            result.PlayTimes,
			"updated_at":            time.Now(),
		}
		if result.DurationMs > 0 {
			updateFields["total_play_duration_ms"] = result.DurationMs
		}
		if result.Status == "completed" {
			updateFields["completed_at"] = time.Now()
		}
		if result.ErrorMsg != "" {
			updateFields["error_msg"] = result.ErrorMsg
		}

		if err := s.db.WithContext(ctx).Model(&model.VoiceInterventionLog{}).
			Where("id = ?", result.LogID).
			Updates(updateFields).Error; err != nil {
			logger.Sugar.Errorf("[VoiceResult] 更新干预日志失败: log_id=%d, err=%v", result.LogID, err)
			return err
		}

		logger.Sugar.Debugf("[VoiceResult] 干预日志已更新: log_id=%d, status=%s", result.LogID, result.Status)
		return nil
	}

	if err := mq.StartConsumer(&config.Global.RocketMQ, "voice_intervention_result", handler, 4); err != nil {
		return fmt.Errorf("start voice result consumer: %w", err)
	}

	logger.Sugar.Infof("[VoiceResult] 播放结果回传 Consumer 已启动，主题: voice_intervention_result")
	return nil
}
