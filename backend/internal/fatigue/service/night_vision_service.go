package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
)

type NightVisionService struct {
	db *gorm.DB
}

func NewNightVisionService() *NightVisionService {
	return &NightVisionService{
		db: database.GetDB(),
	}
}

func (s *NightVisionService) GetConfig(ctx context.Context, vehicleID int64) (*model.NightVisionConfig, error) {
	var config model.NightVisionConfig
	err := s.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID).First(&config).Error
	if err == gorm.ErrRecordNotFound {
		return s.createDefaultConfig(ctx, vehicleID)
	}
	return &config, err
}

func (s *NightVisionService) createDefaultConfig(ctx context.Context, vehicleID int64) (*model.NightVisionConfig, error) {
	config := &model.NightVisionConfig{
		VehicleID:            vehicleID,
		InfraredEnabled:       true,
		InfraredAutoMode:      true,
		InfraredManualOn:      false,
		InfraredIntensity:     50,
		InfraredIntensityAuto: true,
		LowLightThresholdLux:  50,
		HighLightThresholdLux: 200,
		EnhancementEnabled:    true,
		EnhanceMode:           model.EnhanceModeAuto,
		GammaValue:            1.2,
		BrightnessBoost:       30,
		ContrastBoost:         20,
		HistogramEqualization: true,
		ClaheEnabled:          true,
		DenoiseEnabled:        true,
		DenoiseStrength:       3,
		SharpenEnabled:        false,
		SharpenStrength:       2,
		NightModeAuto:         true,
		NightStartHour:        19,
		NightEndHour:          6,
		LowLightFaceDetect:    true,
		MinFaceConfidenceNight: 0.4,
	}

	if err := s.db.WithContext(ctx).Create(config).Error; err != nil {
		return nil, err
	}
	return config, nil
}

func (s *NightVisionService) UpdateConfig(ctx context.Context, req *model.NightVisionConfigUpdateRequest) (*model.NightVisionConfig, error) {
	config, err := s.GetConfig(ctx, req.VehicleID)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.InfraredEnabled != nil {
		updates["infrared_enabled"] = *req.InfraredEnabled
	}
	if req.InfraredAutoMode != nil {
		updates["infrared_auto_mode"] = *req.InfraredAutoMode
	}
	if req.InfraredManualOn != nil {
		updates["infrared_manual_on"] = *req.InfraredManualOn
	}
	if req.InfraredIntensity != nil {
		updates["infrared_intensity"] = *req.InfraredIntensity
	}
	if req.InfraredIntensityAuto != nil {
		updates["infrared_intensity_auto"] = *req.InfraredIntensityAuto
	}
	if req.LowLightThresholdLux != nil {
		updates["low_light_threshold_lux"] = *req.LowLightThresholdLux
	}
	if req.HighLightThresholdLux != nil {
		updates["high_light_threshold_lux"] = *req.HighLightThresholdLux
	}
	if req.EnhancementEnabled != nil {
		updates["enhancement_enabled"] = *req.EnhancementEnabled
	}
	if req.EnhanceMode != nil {
		updates["enhance_mode"] = *req.EnhanceMode
	}
	if req.GammaValue != nil {
		updates["gamma_value"] = *req.GammaValue
	}
	if req.BrightnessBoost != nil {
		updates["brightness_boost"] = *req.BrightnessBoost
	}
	if req.ContrastBoost != nil {
		updates["contrast_boost"] = *req.ContrastBoost
	}
	if req.HistogramEqualization != nil {
		updates["histogram_equalization"] = *req.HistogramEqualization
	}
	if req.ClaheEnabled != nil {
		updates["clahe_enabled"] = *req.ClaheEnabled
	}
	if req.DenoiseEnabled != nil {
		updates["denoise_enabled"] = *req.DenoiseEnabled
	}
	if req.DenoiseStrength != nil {
		updates["denoise_strength"] = *req.DenoiseStrength
	}
	if req.SharpenEnabled != nil {
		updates["sharpen_enabled"] = *req.SharpenEnabled
	}
	if req.SharpenStrength != nil {
		updates["sharpen_strength"] = *req.SharpenStrength
	}
	if req.NightModeAuto != nil {
		updates["night_mode_auto"] = *req.NightModeAuto
	}
	if req.NightStartHour != nil {
		updates["night_start_hour"] = *req.NightStartHour
	}
	if req.NightEndHour != nil {
		updates["night_end_hour"] = *req.NightEndHour
	}
	if req.LowLightFaceDetect != nil {
		updates["low_light_face_detect"] = *req.LowLightFaceDetect
	}
	if req.MinFaceConfidenceNight != nil {
		updates["min_face_confidence_night"] = *req.MinFaceConfidenceNight
	}

	if len(updates) == 0 {
		return config, nil
	}

	updates["updated_at"] = time.Now()

	if err := s.db.WithContext(ctx).Model(config).Updates(updates).Error; err != nil {
		return nil, err
	}

	return s.GetConfig(ctx, req.VehicleID)
}

func (s *NightVisionService) ReportInfraredStatus(ctx context.Context, vehicleID int64, status *model.InfraredLightStatus) error {
	triggerType := model.TriggerTypeAuto
	if !status.IsAutoMode {
		triggerType = model.TriggerTypeManual
	}

	action := model.InfraredActionTurnOff
	if status.LightOn {
		action = model.InfraredActionTurnOn
	}

	log := &model.InfraredLightLog{
		VehicleID:     vehicleID,
		DeviceID:      status.DeviceID,
		Action:        action,
		TriggerType:   triggerType,
		LightOn:       status.LightOn,
		LightLevelLux: &status.LightLevelLux,
		Reason:        status.Reason,
		Timestamp:     time.Now(),
	}

	return s.db.WithContext(ctx).Create(log).Error
}

func (s *NightVisionService) AddInfraredLog(ctx context.Context, log *model.InfraredLightLog) error {
	return s.db.WithContext(ctx).Create(log).Error
}

func (s *NightVisionService) ListInfraredLogs(ctx context.Context, vehicleID int64, page, pageSize int) ([]model.InfraredLightLog, int64, error) {
	var logs []model.InfraredLightLog
	var total int64

	query := s.db.WithContext(ctx).Model(&model.InfraredLightLog{})
	if vehicleID > 0 {
		query = query.Where("vehicle_id = ?", vehicleID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("timestamp DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

type nightVisionAvgResult struct {
	AvgImprovement   float64
	AvgProcTime      float64
	AvgConfBefore    float64
	AvgConfAfter     float64
	DetectRateBefore float64
	DetectRateAfter  float64
}

func (s *NightVisionService) GetStatistics(ctx context.Context, orgID int64) (*model.NightVisionStats, error) {
	stats := &model.NightVisionStats{}

	var configs []model.NightVisionConfig
	s.db.WithContext(ctx).Find(&configs)
	stats.TotalConfigs = len(configs)

	irEnabled := 0
	enhEnabled := 0
	for _, c := range configs {
		if c.InfraredEnabled {
			irEnabled++
		}
		if c.EnhancementEnabled {
			enhEnabled++
		}
	}
	stats.InfraredEnabledCount = irEnabled
	stats.EnhancementEnabledCount = enhEnabled

	todayStart := time.Now().Truncate(24 * time.Hour)
	todayEnd := todayStart.Add(24 * time.Hour)

	var todayOnCount int64
	s.db.WithContext(ctx).Model(&model.InfraredLightLog{}).
		Where("action = ? AND timestamp >= ? AND timestamp < ?",
			model.InfraredActionTurnOn, todayStart, todayEnd).
		Count(&todayOnCount)
	stats.TodayInfraredTurnOnCount = int(todayOnCount)

	var todayOffCount int64
	s.db.WithContext(ctx).Model(&model.InfraredLightLog{}).
		Where("action = ? AND timestamp >= ? AND timestamp < ?",
			model.InfraredActionTurnOff, todayStart, todayEnd).
		Count(&todayOffCount)
	stats.TodayInfraredTurnOffCount = int(todayOffCount)

	var totalEnhance int64
	s.db.WithContext(ctx).Model(&model.ImageEnhancementRecord{}).Count(&totalEnhance)
	stats.TotalEnhanceRecords = totalEnhance

	var todayEnhance int64
	s.db.WithContext(ctx).Model(&model.ImageEnhancementRecord{}).
		Where("timestamp >= ? AND timestamp < ?", todayStart, todayEnd).
		Count(&todayEnhance)
	stats.TodayEnhanceRecords = int(todayEnhance)

	var avg nightVisionAvgResult
	s.db.WithContext(ctx).Model(&model.ImageEnhancementRecord{}).
		Select(`
			AVG(quality_improvement_pct) as avg_improvement,
			AVG(processing_time_ms) as avg_proc_time,
			AVG(face_confidence_original) as avg_conf_before,
			AVG(face_confidence_enhanced) as avg_conf_after,
			AVG(CASE WHEN face_detected_original = 1 THEN 1.0 ELSE 0.0 END) as detect_rate_before,
			AVG(CASE WHEN face_detected_enhanced = 1 THEN 1.0 ELSE 0.0 END) as detect_rate_after
		`).
		Where("is_night_time = ?", true).
		Scan(&avg)

	stats.AvgQualityImprovement = avg.AvgImprovement
	stats.AvgProcessingTimeMs = int(avg.AvgProcTime)
	stats.AvgConfidenceBefore = avg.AvgConfBefore
	stats.AvgConfidenceAfter = avg.AvgConfAfter
	stats.NightFaceDetectRateBefore = avg.DetectRateBefore * 100
	stats.NightFaceDetectRateAfter = avg.DetectRateAfter * 100

	if avg.AvgConfBefore > 0 {
		stats.ConfidenceImprovement = (avg.AvgConfAfter - avg.AvgConfBefore) / avg.AvgConfBefore * 100
	}

	var autoCount int64
	s.db.WithContext(ctx).Model(&model.InfraredLightLog{}).
		Where("trigger_type = ?", model.TriggerTypeAuto).
		Where("timestamp >= ?", todayStart).
		Count(&autoCount)
	stats.AutoTriggerCount = int(autoCount)

	var manualCount int64
	s.db.WithContext(ctx).Model(&model.InfraredLightLog{}).
		Where("trigger_type = ?", model.TriggerTypeManual).
		Where("timestamp >= ?", todayStart).
		Count(&manualCount)
	stats.ManualTriggerCount = int(manualCount)

	return stats, nil
}

func (s *NightVisionService) IsNightTime(hour, startHour, endHour int) bool {
	if startHour == endHour {
		return false
	}
	if startHour < endHour {
		return hour >= startHour && hour < endHour
	}
	return hour >= startHour || hour < endHour
}

func (s *NightVisionService) ShouldTurnOnInfrared(cfg *model.NightVisionConfig, lightLevelLux int, currentHour int) bool {
	if !cfg.InfraredEnabled {
		return false
	}

	if !cfg.InfraredAutoMode {
		return cfg.InfraredManualOn
	}

	isNight := s.IsNightTime(currentHour, cfg.NightStartHour, cfg.NightEndHour)
	if cfg.NightModeAuto && isNight {
		return true
	}

	if lightLevelLux <= cfg.LowLightThresholdLux {
		return true
	}

	return false
}

func (s *NightVisionService) CalcInfraredIntensity(cfg *model.NightVisionConfig, lightLevelLux int) int {
	if !cfg.InfraredIntensityAuto {
		return cfg.InfraredIntensity
	}

	high := cfg.HighLightThresholdLux
	if high <= 0 {
		high = 200
	}
	if lightLevelLux <= 0 {
		return 80
	}
	if lightLevelLux >= high {
		return 30
	}

	ratio := 1.0 - float64(lightLevelLux)/float64(high)
	intensity := int(30.0 + ratio*50.0)
	if intensity < 20 {
		intensity = 20
	}
	if intensity > 100 {
		intensity = 100
	}
	return intensity
}

func (s *NightVisionService) DetermineEnhanceMode(cfg *model.NightVisionConfig, lightLevelLux int, isNight bool) model.EnhanceMode {
	if cfg.EnhanceMode != model.EnhanceModeAuto {
		return cfg.EnhanceMode
	}

	if isNight && lightLevelLux < 100 {
		return model.EnhanceModeNight
	}

	if lightLevelLux < 50 {
		return model.EnhanceModeLowLight
	}

	return model.EnhanceModeNight
}

func (s *NightVisionService) GetConfigByVehicleID(ctx context.Context, vehicleID int64) (*model.NightVisionConfig, error) {
	return s.GetConfig(ctx, vehicleID)
}

func (s *NightVisionService) ResetConfig(ctx context.Context, vehicleID int64) (*model.NightVisionConfig, error) {
	config, err := s.GetConfig(ctx, vehicleID)
	if err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Delete(config).Error; err != nil {
		return nil, err
	}

	return s.createDefaultConfig(ctx, vehicleID)
}

func (s *NightVisionService) ListConfigs(ctx context.Context, page, pageSize int, orgID int64) ([]model.NightVisionConfig, int64, error) {
	var configs []model.NightVisionConfig
	var total int64

	query := s.db.WithContext(ctx).Model(&model.NightVisionConfig{})
	if orgID > 0 {
		query = query.Where("org_id = ?", orgID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count configs: %w", err)
	}

	offset := (page - 1) * pageSize
	err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&configs).Error
	return configs, total, err
}
