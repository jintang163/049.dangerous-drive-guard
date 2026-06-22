package service

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
)

type ADASService struct {
	db              *database.TIDB
	cfg             *config.Config
	alertWindow     map[int64][]*model.ADASAlert
	alertWindowMu   sync.RWMutex
	frequencyWindow map[int64]*frequencyState
	frequencyMu     sync.RWMutex
}

type frequencyState struct {
	WindowStart        time.Time
	CloseFollowingCnt  int
	LaneDepartureCnt   int
	TotalAlertCnt      int
	DecelerateTriggered bool
	ReportedToCenter   bool
}

func NewADASService(cfg *config.Config) *ADASService {
	return &ADASService{
		db:              database.GetDB(),
		cfg:             cfg,
		alertWindow:     make(map[int64][]*model.ADASAlert),
		frequencyWindow: make(map[int64]*frequencyState),
	}
}

func (s *ADASService) ProcessRadarData(ctx context.Context, data *model.RadarData) (*model.RadarDataResponse, error) {
	adasCfg, err := s.getADASConfig(ctx, data.VehicleID)
	if err != nil {
		return nil, fmt.Errorf("get adas config error: %w", err)
	}

	var alerts []*model.ADASAlert
	now := time.Now()

	if adasCfg.EnableCloseFollowing && data.FollowingDistance > 0 {
		if alert := s.checkCloseFollowing(data, adasCfg, now); alert != nil {
			alerts = append(alerts, alert)
		}
	}

	if adasCfg.EnableLaneDeparture && (data.LaneDepartureLeft || data.LaneDepartureRight || math.Abs(data.LaneOffset) > 0) {
		if alert := s.checkLaneDeparture(data, adasCfg, now); alert != nil {
			alerts = append(alerts, alert)
		}
	}

	if adasCfg.EnableForwardCollision && data.ForwardCollisionTTC > 0 {
		if alert := s.checkForwardCollision(data, adasCfg, now); alert != nil {
			alerts = append(alerts, alert)
		}
	}

	var savedAlerts []*model.ADASAlert
	for _, alert := range alerts {
		if err := s.db.WithContext(ctx).Create(alert).Error; err != nil {
			logger.Sugar.Errorf("save adas alert error: %v, alert_no=%s", err, alert.AlertNo)
			continue
		}
		savedAlerts = append(savedAlerts, alert)
		s.recordToAlertWindow(data.VehicleID, alert)
	}

	decelerateTriggered := false
	var decelerateValue float64
	frequencyAlert := false

	if len(savedAlerts) > 0 && adasCfg.EnableAutoDecelerate {
		decelerateTriggered, decelerateValue, frequencyAlert = s.checkFrequencyAndDecelerate(
			ctx, data, adasCfg, savedAlerts, now,
		)
	}

	resp := &model.RadarDataResponse{
		AlertTriggered:    len(savedAlerts) > 0,
		Alerts:            savedAlerts,
		DecelerateTriggered: decelerateTriggered,
		FrequencyAlert:    frequencyAlert,
	}
	if data.FollowingDistance > 0 {
		resp.CurrentFollowingDistance = data.FollowingDistance
	}
	if data.LaneOffset != 0 {
		resp.CurrentLaneOffset = data.LaneOffset
	}
	if decelerateTriggered {
		resp.DecelerateValue = decelerateValue
	}

	return resp, nil
}

func (s *ADASService) checkCloseFollowing(data *model.RadarData, cfg *model.ADASConfig, now time.Time) *model.ADASAlert {
	dist := data.FollowingDistance

	var level model.ADASAlertLevel
	var threshold float64
	var msg string

	if dist <= cfg.CloseFollowingCritDist {
		level = model.ADASLevelCritical
		threshold = cfg.CloseFollowingCritDist
		msg = fmt.Sprintf("【严重】跟车过近！当前车距%.1f米，低于安全距离%.1f米，请立即减速拉开车距", dist, cfg.CloseFollowingCritDist)
	} else if dist <= cfg.CloseFollowingMinDist {
		level = model.ADASLevelWarning
		threshold = cfg.CloseFollowingMinDist
		msg = fmt.Sprintf("【警告】跟车过近，当前车距%.1f米，低于安全距离%.1f米，请注意保持车距", dist, cfg.CloseFollowingMinDist)
	} else if dist <= cfg.CloseFollowingWarnDist {
		level = model.ADASLevelInfo
		threshold = cfg.CloseFollowingWarnDist
		msg = fmt.Sprintf("【提示】前方车距%.1f米，接近安全距离阈值，请注意车距", dist)
	} else {
		return nil
	}

	ttc := 0.0
	if data.RelativeSpeed > 0 && data.VehicleSpeed > 0 {
		ttc = dist / (data.RelativeSpeed / 3.6)
	}

	suggestedAction := "保持安全车距，适当减速"
	if level == model.ADASLevelCritical {
		suggestedAction = "立即减速，拉大车距至安全范围"
	}

	return &model.ADASAlert{
		AlertNo:          generateADASAlertNo(now),
		VehicleID:        data.VehicleID,
		DriverID:         data.DriverID,
		WaybillID:        data.WaybillID,
		AlertType:        model.ADASAlertCloseFollowing,
		AlertLevel:       level,
		Status:           model.ADASStatusActive,
		TriggerValue:     dist,
		ThresholdValue:   threshold,
		FollowingDistance: dist,
		VehicleSpeed:     data.VehicleSpeed,
		TTC:              ttc,
		AlertMessage:     msg,
		Latitude:         data.Latitude,
		Longitude:        data.Longitude,
		SuggestedAction:  suggestedAction,
	}
}

func (s *ADASService) checkLaneDeparture(data *model.RadarData, cfg *model.ADASConfig, now time.Time) *model.ADASAlert {
	offset := math.Abs(data.LaneOffset)
	if !data.LaneDepartureLeft && !data.LaneDepartureRight && offset < cfg.LaneDepartureThreshold {
		return nil
	}

	departureSide := ""
	if data.LaneDepartureLeft || (data.LaneOffset < 0 && offset >= cfg.LaneDepartureThreshold) {
		departureSide = "left"
	} else if data.LaneDepartureRight || (data.LaneOffset > 0 && offset >= cfg.LaneDepartureThreshold) {
		departureSide = "right"
	}

	if departureSide == "" {
		return nil
	}

	var level model.ADASAlertLevel
	var msg string

	if offset >= cfg.LaneDepartureThreshold*2 {
		level = model.ADASLevelCritical
		sideText := "左侧"
		if departureSide == "right" {
			sideText = "右侧"
		}
		msg = fmt.Sprintf("【严重】车道偏离！车辆向%s偏移%.2f米，请立即修正方向", sideText, offset)
	} else {
		level = model.ADASLevelWarning
		sideText := "左侧"
		if departureSide == "right" {
			sideText = "右侧"
		}
		msg = fmt.Sprintf("【警告】车辆向%s偏移%.2f米，请注意保持车道", sideText, offset)
	}

	suggestedAction := "修正方向盘，保持车道居中行驶"
	if level == model.ADASLevelCritical {
		suggestedAction = "立即修正方向！注意后方来车，安全回归车道"
	}

	return &model.ADASAlert{
		AlertNo:         generateADASAlertNo(now),
		VehicleID:       data.VehicleID,
		DriverID:        data.DriverID,
		WaybillID:       data.WaybillID,
		AlertType:       model.ADASAlertLaneDeparture,
		AlertLevel:      level,
		Status:          model.ADASStatusActive,
		TriggerValue:    offset,
		ThresholdValue:  cfg.LaneDepartureThreshold,
		LaneOffset:      data.LaneOffset,
		VehicleSpeed:    data.VehicleSpeed,
		DepartureSide:   departureSide,
		AlertMessage:    msg,
		Latitude:        data.Latitude,
		Longitude:       data.Longitude,
		SuggestedAction: suggestedAction,
	}
}

func (s *ADASService) checkForwardCollision(data *model.RadarData, cfg *model.ADASConfig, now time.Time) *model.ADASAlert {
	ttc := data.ForwardCollisionTTC
	if ttc <= 0 {
		return nil
	}

	var level model.ADASAlertLevel
	var threshold float64
	var msg string

	if ttc <= cfg.ForwardCollisionTTCCrit {
		level = model.ADASLevelCritical
		threshold = cfg.ForwardCollisionTTCCrit
		msg = fmt.Sprintf("【严重】前碰撞预警！碰撞时间仅%.1f秒，请立即制动", ttc)
	} else if ttc <= cfg.ForwardCollisionTTCWarn {
		level = model.ADASLevelWarning
		threshold = cfg.ForwardCollisionTTCWarn
		msg = fmt.Sprintf("【警告】前碰撞风险，碰撞时间%.1f秒，请准备制动", ttc)
	} else {
		return nil
	}

	suggestedAction := "注意前方车辆，准备制动"
	if level == model.ADASLevelCritical {
		suggestedAction = "立即紧急制动！"
	}

	return &model.ADASAlert{
		AlertNo:          generateADASAlertNo(now),
		VehicleID:        data.VehicleID,
		DriverID:         data.DriverID,
		WaybillID:        data.WaybillID,
		AlertType:        model.ADASAlertForwardCollision,
		AlertLevel:       level,
		Status:           model.ADASStatusActive,
		TriggerValue:     ttc,
		ThresholdValue:   threshold,
		TTC:              ttc,
		FollowingDistance: data.FollowingDistance,
		VehicleSpeed:     data.VehicleSpeed,
		AlertMessage:     msg,
		Latitude:         data.Latitude,
		Longitude:        data.Longitude,
		SuggestedAction:  suggestedAction,
	}
}

func (s *ADASService) recordToAlertWindow(vehicleID int64, alert *model.ADASAlert) {
	s.alertWindowMu.Lock()
	defer s.alertWindowMu.Unlock()

	window := s.alertWindow[vehicleID]
	window = append(window, alert)

	cutoff := time.Now().Add(-10 * time.Minute)
	filtered := window[:0]
	for _, a := range window {
		if a.CreatedAt.After(cutoff) || a.CreatedAt.IsZero() {
			filtered = append(filtered, a)
		}
	}
	s.alertWindow[vehicleID] = filtered
}

func (s *ADASService) checkFrequencyAndDecelerate(ctx context.Context, data *model.RadarData, cfg *model.ADASConfig, alerts []*model.ADASAlert, now time.Time) (bool, float64, bool) {
	s.frequencyMu.Lock()
	defer s.frequencyMu.Unlock()

	state, ok := s.frequencyWindow[data.VehicleID]
	if !ok {
		state = &frequencyState{
			WindowStart: now,
		}
		s.frequencyWindow[data.VehicleID] = state
	}

	windowDuration := time.Duration(cfg.FrequencyWindowMinutes) * time.Minute
	if now.Sub(state.WindowStart) > windowDuration {
		if state.TotalAlertCnt >= cfg.FrequencyAlertThreshold {
			s.persistFrequencyTracker(ctx, data, state, cfg)
		}
		state = &frequencyState{
			WindowStart: now,
		}
		s.frequencyWindow[data.VehicleID] = state
	}

	for _, a := range alerts {
		state.TotalAlertCnt++
		switch a.AlertType {
		case model.ADASAlertCloseFollowing:
			state.CloseFollowingCnt++
		case model.ADASAlertLaneDeparture:
			state.LaneDepartureCnt++
		}
	}

	if state.TotalAlertCnt < cfg.FrequencyAlertThreshold {
		return false, 0, false
	}

	if state.DecelerateTriggered {
		return true, cfg.AutoDecelerateSpeed, true
	}

	state.DecelerateTriggered = true
	decelerateValue := cfg.AutoDecelerateSpeed

	decelAlert := &model.ADASAlert{
		AlertNo:            generateADASAlertNo(now),
		VehicleID:          data.VehicleID,
		DriverID:           data.DriverID,
		WaybillID:          data.WaybillID,
		AlertType:          model.ADASAlertAutoDecelerate,
		AlertLevel:         model.ADASLevelCritical,
		Status:             model.ADASStatusActive,
		TriggerValue:       float64(state.TotalAlertCnt),
		ThresholdValue:     float64(cfg.FrequencyAlertThreshold),
		VehicleSpeed:       data.VehicleSpeed,
		DecelerateTriggered: true,
		DecelerateValue:    decelerateValue,
		AlertMessage:       fmt.Sprintf("【自动降速】%d分钟内预警%d次（跟车过近%d次/车道偏离%d次），自动降速至%.0fkm/h并上报调度中心", cfg.FrequencyWindowMinutes, state.TotalAlertCnt, state.CloseFollowingCnt, state.LaneDepartureCnt, decelerateValue),
		SuggestedAction:    fmt.Sprintf("系统已自动降速至%.0fkm/h，请注意保持安全驾驶", decelerateValue),
		Latitude:           data.Latitude,
		Longitude:          data.Longitude,
	}
	if err := s.db.WithContext(ctx).Create(decelAlert).Error; err != nil {
		logger.Sugar.Errorf("save auto-decelerate alert error: %v", err)
	}

	state.ReportedToCenter = true

	s.reportToCenter(ctx, data, state, cfg, decelerateValue)

	logger.Sugar.Warnf("adas auto decelerate triggered: vehicle_id=%d, driver_id=%d, alerts=%d in %dmin, decelerate_to=%.0fkm/h",
		data.VehicleID, data.DriverID, state.TotalAlertCnt, cfg.FrequencyWindowMinutes, decelerateValue)

	return true, decelerateValue, true
}

func (s *ADASService) persistFrequencyTracker(ctx context.Context, data *model.RadarData, state *frequencyState, cfg *model.ADASConfig) {
	tracker := &model.ADASFrequencyTracker{
		VehicleID:           data.VehicleID,
		DriverID:            data.DriverID,
		WindowStart:         state.WindowStart,
		WindowEnd:           state.WindowStart.Add(time.Duration(cfg.FrequencyWindowMinutes) * time.Minute),
		CloseFollowingCount: state.CloseFollowingCnt,
		LaneDepartureCount:  state.LaneDepartureCnt,
		TotalAlertCount:     state.TotalAlertCnt,
		DecelerateTriggered: state.DecelerateTriggered,
		ReportedToCenter:    state.ReportedToCenter,
	}
	if state.DecelerateTriggered {
		tracker.DecelerateValue = cfg.AutoDecelerateSpeed
	}
	if err := s.db.WithContext(ctx).Create(tracker).Error; err != nil {
		logger.Sugar.Errorf("persist frequency tracker error: %v", err)
	}
}

func (s *ADASService) reportToCenter(ctx context.Context, data *model.RadarData, state *frequencyState, cfg *model.ADASConfig, decelerateValue float64) {
	var vehicle model.Vehicle
	s.db.WithContext(ctx).Where("id = ?", data.VehicleID).First(&vehicle)

	var driver model.User
	s.db.WithContext(ctx).Where("id = ?", data.DriverID).First(&driver)

	logger.Sugar.Infof("adas report to center: vehicle=%s, driver=%s, close_following=%d, lane_departure=%d, total=%d, decelerate_to=%.0f",
		vehicle.PlateNumber, driver.RealName, state.CloseFollowingCnt, state.LaneDepartureCnt, state.TotalAlertCnt, decelerateValue)
}

func (s *ADASService) getADASConfig(ctx context.Context, vehicleID int64) (*model.ADASConfig, error) {
	var cfg model.ADASConfig
	err := s.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID).First(&cfg).Error
	if err == nil {
		return &cfg, nil
	}

	defaultCfg := &model.ADASConfig{
		VehicleID:               vehicleID,
		CloseFollowingMinDist:   5.0,
		CloseFollowingWarnDist:  10.0,
		CloseFollowingCritDist:  3.0,
		LaneDepartureThreshold:  0.30,
		ForwardCollisionTTCWarn: 3.0,
		ForwardCollisionTTCCrit: 1.5,
		FrequencyWindowMinutes:  5,
		FrequencyAlertThreshold: 6,
		AutoDecelerateSpeed:     20.0,
		EnableCloseFollowing:    true,
		EnableLaneDeparture:     true,
		EnableForwardCollision:  true,
		EnableAutoDecelerate:    true,
	}
	return defaultCfg, nil
}

func (s *ADASService) GetAlerts(ctx context.Context, query *model.ADASAlertQuery) (*model.ADASAlertPage, error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	q := s.db.WithContext(ctx).Model(&model.ADASAlert{})
	if query.VehicleID > 0 {
		q = q.Where("vehicle_id = ?", query.VehicleID)
	}
	if query.DriverID > 0 {
		q = q.Where("driver_id = ?", query.DriverID)
	}
	if query.WaybillID > 0 {
		q = q.Where("waybill_id = ?", query.WaybillID)
	}
	if query.AlertType != "" {
		q = q.Where("alert_type = ?", query.AlertType)
	}
	if query.AlertLevel != "" {
		q = q.Where("alert_level = ?", query.AlertLevel)
	}
	if query.Status != "" {
		q = q.Where("status = ?", query.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count adas alerts error: %w", err)
	}

	var alerts []*model.ADASAlert
	offset := (page - 1) * pageSize
	if err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&alerts).Error; err != nil {
		return nil, fmt.Errorf("query adas alerts error: %w", err)
	}

	return &model.ADASAlertPage{
		List:     alerts,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *ADASService) GetAlert(ctx context.Context, alertID int64) (*model.ADASAlert, error) {
	var alert model.ADASAlert
	if err := s.db.WithContext(ctx).Where("id = ?", alertID).First(&alert).Error; err != nil {
		return nil, fmt.Errorf("get adas alert error: %w", err)
	}
	return &alert, nil
}

func (s *ADASService) AckAlert(ctx context.Context, req *model.ADASAlertAckRequest, operatorID int64) (*model.ADASAlert, error) {
	var alert model.ADASAlert
	if err := s.db.WithContext(ctx).Where("id = ?", req.AlertID).First(&alert).Error; err != nil {
		return nil, fmt.Errorf("alert not found")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"driver_acknowledged": true,
		"acknowledged_at":     &now,
	}

	switch req.AckType {
	case "resolve":
		updates["status"] = model.ADASStatusResolved
		updates["resolved_at"] = &now
		updates["resolution_note"] = req.Note
	case "escalate":
		updates["status"] = model.ADASStatusEscalated
		updates["reported_to_center"] = true
	default:
		updates["status"] = model.ADASStatusResolved
		updates["resolved_at"] = &now
	}

	if err := s.db.WithContext(ctx).Model(&alert).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update alert error: %w", err)
	}

	alert.DriverAcknowledged = true
	alert.AcknowledgedAt = &now
	alert.Status = model.ADASAlertStatus(updateStatus(req.AckType))

	return &alert, nil
}

func (s *ADASService) GetADASConfig(ctx context.Context, vehicleID int64) (*model.ADASConfig, error) {
	return s.getADASConfig(ctx, vehicleID)
}

func (s *ADASService) UpdateADASConfig(ctx context.Context, cfg *model.ADASConfig) (*model.ADASConfig, error) {
	var existing model.ADASConfig
	err := s.db.WithContext(ctx).Where("vehicle_id = ?", cfg.VehicleID).First(&existing).Error

	if err != nil {
		if err := s.db.WithContext(ctx).Create(cfg).Error; err != nil {
			return nil, fmt.Errorf("create adas config error: %w", err)
		}
		return cfg, nil
	}

	updates := map[string]interface{}{
		"close_following_min_dist_m":    cfg.CloseFollowingMinDist,
		"close_following_warn_dist_m":   cfg.CloseFollowingWarnDist,
		"close_following_crit_dist_m":   cfg.CloseFollowingCritDist,
		"lane_departure_threshold_m":    cfg.LaneDepartureThreshold,
		"forward_collision_ttc_warn_s":  cfg.ForwardCollisionTTCWarn,
		"forward_collision_ttc_crit_s":  cfg.ForwardCollisionTTCCrit,
		"frequency_window_minutes":      cfg.FrequencyWindowMinutes,
		"frequency_alert_threshold":     cfg.FrequencyAlertThreshold,
		"auto_decelerate_speed_kmh":     cfg.AutoDecelerateSpeed,
		"enable_close_following":        cfg.EnableCloseFollowing,
		"enable_lane_departure":         cfg.EnableLaneDeparture,
		"enable_forward_collision":      cfg.EnableForwardCollision,
		"enable_auto_decelerate":        cfg.EnableAutoDecelerate,
	}

	if err := s.db.WithContext(ctx).Model(&existing).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update adas config error: %w", err)
	}

	cfg.ID = existing.ID
	return cfg, nil
}

func (s *ADASService) GetDriverAlertSummary(ctx context.Context, driverID int64) (map[string]interface{}, error) {
	if driverID <= 0 {
		return nil, fmt.Errorf("invalid driver id")
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var todayTotal int64
	s.db.WithContext(ctx).Model(&model.ADASAlert{}).
		Where("driver_id = ? AND created_at >= ?", driverID, today).
		Count(&todayTotal)

	var todayCloseFollowing int64
	s.db.WithContext(ctx).Model(&model.ADASAlert{}).
		Where("driver_id = ? AND alert_type = ? AND created_at >= ?", driverID, model.ADASAlertCloseFollowing, today).
		Count(&todayCloseFollowing)

	var todayLaneDeparture int64
	s.db.WithContext(ctx).Model(&model.ADASAlert{}).
		Where("driver_id = ? AND alert_type = ? AND created_at >= ?", driverID, model.ADASAlertLaneDeparture, today).
		Count(&todayLaneDeparture)

	var todayCritical int64
	s.db.WithContext(ctx).Model(&model.ADASAlert{}).
		Where("driver_id = ? AND alert_level = ? AND created_at >= ?", driverID, model.ADASLevelCritical, today).
		Count(&todayCritical)

	var activeAlerts int64
	s.db.WithContext(ctx).Model(&model.ADASAlert{}).
		Where("driver_id = ? AND status = ?", driverID, model.ADASStatusActive).
		Count(&activeAlerts)

	sevenDaysAgo := now.AddDate(0, 0, -7)
	var weekTotal int64
	s.db.WithContext(ctx).Model(&model.ADASAlert{}).
		Where("driver_id = ? AND created_at >= ?", driverID, sevenDaysAgo).
		Count(&weekTotal)

	return map[string]interface{}{
		"today_total":            todayTotal,
		"today_close_following":  todayCloseFollowing,
		"today_lane_departure":   todayLaneDeparture,
		"today_critical":         todayCritical,
		"active_alerts":          activeAlerts,
		"week_total":             weekTotal,
	}, nil
}

func (s *ADASService) GetVehicleFrequencyTrackers(ctx context.Context, vehicleID int64, limit int) ([]*model.ADASFrequencyTracker, error) {
	if vehicleID <= 0 {
		return nil, fmt.Errorf("invalid vehicle id")
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	var trackers []*model.ADASFrequencyTracker
	if err := s.db.WithContext(ctx).
		Where("vehicle_id = ?", vehicleID).
		Order("window_start DESC").
		Limit(limit).
		Find(&trackers).Error; err != nil {
		return nil, fmt.Errorf("query frequency trackers error: %w", err)
	}
	return trackers, nil
}

func generateADASAlertNo(now time.Time) string {
	return fmt.Sprintf("ADAS-%s-%04d", now.Format("20060102-150405"), now.UnixNano()%10000)
}

func updateStatus(ackType string) model.ADASAlertStatus {
	switch ackType {
	case "escalate":
		return model.ADASStatusEscalated
	case "resolve":
		return model.ADASStatusResolved
	default:
		return model.ADASStatusResolved
	}
}
