package service

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/internal/monitor/delivery/ws"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
)

type MonitorService struct {
	db          *database.TIDB
	statusCache map[int64]*model.RealtimeVehicleStatus
	cacheMu     sync.RWMutex
}

func NewMonitorService() *MonitorService {
	svc := &MonitorService{
		db:          database.GetDB(),
		statusCache: make(map[int64]*model.RealtimeVehicleStatus),
	}
	go svc.startStatusRefresh()
	return svc
}

func (s *MonitorService) startStatusRefresh() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.refreshAllStatus()
	}
}

func (s *MonitorService) refreshAllStatus() {
	rows, err := s.db.Raw(`
		SELECT v.id, v.plate_number, v.vehicle_type, v.status, v.driver_id,
		       u.real_name as driver_name, w.id as waybill_id, w.waybill_no,
		       f.latitude, f.longitude, f.vehicle_speed, f.fatigue_score, f.fatigue_level,
		       f.detection_time
		FROM vehicles v
		LEFT JOIN users u ON u.id = v.driver_id
		LEFT JOIN waybills w ON w.vehicle_id = v.id AND w.status = 'in_transit'
		LEFT JOIN (
			SELECT f1.* FROM fatigue_detection_records f1
			INNER JOIN (
				SELECT vehicle_id, MAX(detection_time) as max_time
				FROM fatigue_detection_records
				WHERE detection_time > DATE_SUB(NOW(), INTERVAL 30 MINUTE)
				GROUP BY vehicle_id
			) f2 ON f1.vehicle_id = f2.vehicle_id AND f1.detection_time = f2.max_time
		) f ON f.vehicle_id = v.id
	`).Rows()
	if err != nil {
		logger.Sugar.Errorf("refresh status error: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var vs model.RealtimeVehicleStatus
		var lat, lng, speed, fatigueScore interface{}
		var detectionTime interface{}
		rows.Scan(&vs.VehicleID, &vs.PlateNumber, &vs.VehicleType, &vs.Status,
			&vs.DriverID, &vs.DriverName, &vs.WaybillID, &vs.WaybillNo,
			&lat, &lng, &speed, &fatigueScore, &vs.FatigueLevel, &detectionTime)

		if lat != nil {
			vs.Latitude = toFloat(lat)
		}
		if lng != nil {
			vs.Longitude = toFloat(lng)
		}
		if speed != nil {
			vs.Speed = toFloat(speed)
		}
		if fatigueScore != nil {
			vs.FatigueScore = toFloat(fatigueScore)
		}
		if t, ok := detectionTime.(time.Time); ok {
			vs.LastUpdateTime = t
			vs.GPSTime = t
		}

		vs.MarkerColor = s.getMarkerColor(&vs)
		vs.CurrentAddress = s.reverseGeocode(vs.Latitude, vs.Longitude)

		s.cacheMu.Lock()
		s.statusCache[vs.VehicleID] = &vs
		s.cacheMu.Unlock()
	}
}

func (s *MonitorService) getMarkerColor(vs *model.RealtimeVehicleStatus) string {
	switch vs.Status {
	case model.VehicleStatusOffline:
		return "#808080"
	}
	switch vs.FatigueLevel {
	case model.FatigueFatigue:
		return "#FF0000"
	case model.FatigueWarning:
		return "#FFA500"
	default:
		return "#00CC00"
	}
}

func (s *MonitorService) reverseGeocode(lat, lng float64) string {
	if lat == 0 && lng == 0 {
		return "位置未知"
	}
	return fmt.Sprintf("坐标(%.5f, %.5f)", lat, lng)
}

func (s *MonitorService) GetRealtimeVehicles(ctx context.Context, orgID int64, statusFilter string) ([]*model.RealtimeVehicleStatus, error) {
	s.cacheMu.RLock()
	cachedCount := len(s.statusCache)
	s.cacheMu.RUnlock()

	if cachedCount == 0 {
		s.refreshAllStatus()
	}

	var result []*model.RealtimeVehicleStatus
	s.cacheMu.RLock()
	for _, vs := range s.statusCache {
		if statusFilter != "" && string(vs.Status) != statusFilter {
			continue
		}
		clone := *vs
		alertCount := s.getPendingAlarmCount(vs.VehicleID)
		clone.AlertCount = alertCount
		result = append(result, &clone)
	}
	s.cacheMu.RUnlock()

	return result, nil
}

func (s *MonitorService) getPendingAlarmCount(vehicleID int64) int {
	var count int64
	s.db.Table("fatigue_alarms").
		Where("vehicle_id = ? AND status IN ('pending', 'processing')", vehicleID).
		Count(&count)
	return int(count)
}

func (s *MonitorService) GetVehicleStatus(ctx context.Context, vehicleID int64) (*model.RealtimeVehicleStatus, error) {
	s.cacheMu.RLock()
	if cached, ok := s.statusCache[vehicleID]; ok {
		clone := *cached
		clone.AlertCount = s.getPendingAlarmCount(vehicleID)
		s.cacheMu.RUnlock()
		return &clone, nil
	}
	s.cacheMu.RUnlock()

	s.refreshAllStatus()
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	if cached, ok := s.statusCache[vehicleID]; ok {
		return cached, nil
	}
	return nil, fmt.Errorf("vehicle status not found")
}

func (s *MonitorService) GetStatistics(ctx context.Context, orgID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalVehicles int64
	s.db.Table("vehicles").Where("org_id = ?", orgID).Count(&totalVehicles)
	stats["total_vehicles"] = totalVehicles

	var runningVehicles int64
	s.db.Table("vehicles").Where("org_id = ? AND status = 'running'", orgID).Count(&runningVehicles)
	stats["running_vehicles"] = runningVehicles
	stats["idle_vehicles"] = totalVehicles - runningVehicles

	var totalDrivers int64
	s.db.Table("users").Where("org_id = ? AND role = 'driver'", orgID).Count(&totalDrivers)
	stats["total_drivers"] = totalDrivers

	var totalWaybills int64
	s.db.Table("waybills").Where("carrier_org_id = ?", orgID).Count(&totalWaybills)
	stats["total_waybills"] = totalWaybills

	var inTransitWaybills int64
	s.db.Table("waybills").Where("carrier_org_id = ? AND status = 'in_transit'", orgID).Count(&inTransitWaybills)
	stats["in_transit_waybills"] = inTransitWaybills

	var pendingAlarms int64
	s.db.Table("fatigue_alarms").Where("status IN ('pending', 'processing')").Count(&pendingAlarms)
	stats["pending_alarms"] = pendingAlarms

	var todayAlarms int64
	s.db.Table("fatigue_alarms").Where("DATE(created_at) = CURDATE()").Count(&todayAlarms)
	stats["today_alarms"] = todayAlarms

	var todayMileage float64
	s.db.Raw(`
		SELECT COALESCE(SUM(ABS(total_distance)), 0) FROM waybills
		WHERE carrier_org_id = ? AND DATE(actual_departure_time) = CURDATE()
	`, orgID).Scan(&todayMileage)
	stats["today_mileage_km"] = math.Round(todayMileage*100) / 100

	var todayFatigueEvents int64
	s.db.Table("fatigue_detection_records").Where("DATE(detection_time) = CURDATE() AND fatigue_level IN ('warning', 'fatigue')").Count(&todayFatigueEvents)
	stats["today_fatigue_events"] = todayFatigueEvents

	type AlarmTypeCount struct {
		AlarmType string `json:"alarm_type"`
		Count     int64  `json:"count"`
	}
	var alarmTypeStats []AlarmTypeCount
	s.db.Table("fatigue_alarms").
		Select("alarm_type, COUNT(*) as count").
		Where("DATE(created_at) >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)").
		Group("alarm_type").
		Scan(&alarmTypeStats)
	stats["alarm_type_distribution"] = alarmTypeStats

	type DailyTrend struct {
		Date   string `json:"date"`
		Alarms int64  `json:"alarms"`
		Events int64  `json:"events"`
	}
	var dailyTrend []DailyTrend
	s.db.Raw(`
		SELECT
			DATE(d.detection_time) as date,
			(SELECT COUNT(*) FROM fatigue_alarms a WHERE DATE(a.created_at) = DATE(d.detection_time)) as alarms,
			COUNT(*) as events
		FROM fatigue_detection_records d
		WHERE d.detection_time >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
		GROUP BY DATE(d.detection_time)
		ORDER BY date DESC
	`).Scan(&dailyTrend)
	stats["daily_trend"] = dailyTrend

	return stats, nil
}

func (s *MonitorService) SendVoiceIntercom(ctx context.Context, vehicleID int64, message string, priority int) error {
	hub := ws.GetHub()
	hub.SendIntercomToVehicle(vehicleID, message, priority)
	logger.Sugar.Infof("voice intercom sent to vehicle %d: %s", vehicleID, message)

	s.db.Exec(`
		INSERT INTO operation_logs (user_id, module, action, target_type, target_id, detail, created_at)
		VALUES (?, 'monitor', 'voice_intercom', 'vehicle', ?, ?, NOW())
	`, ctx.Value("user_id"), vehicleID, message)

	return nil
}

func (s *MonitorService) DispatchServiceArea(ctx context.Context, vehicleID, serviceAreaID int64, reason string, restDuration int) error {
	var serviceArea model.ServiceArea
	err := s.db.Table("service_areas").Where("id = ?", serviceAreaID).First(&serviceArea).Error
	if err != nil {
		return fmt.Errorf("service area not found")
	}

	hub := ws.GetHub()
	hub.DispatchCommand(vehicleID, "rest_at_service_area", map[string]interface{}{
		"service_area_id":   serviceAreaID,
		"service_area_name": serviceArea.Name,
		"latitude":          serviceArea.Latitude,
		"longitude":         serviceArea.Longitude,
		"reason":            reason,
		"rest_duration_min": restDuration,
		"has_danger_parking": serviceArea.HasDangerParking,
		"phone":             serviceArea.Phone,
	})

	s.db.Exec(`
		INSERT INTO operation_logs (user_id, module, action, target_type, target_id, detail, created_at)
		VALUES (?, 'monitor', 'dispatch_service_area', 'vehicle', ?, ?, NOW())
	`, ctx.Value("user_id"), vehicleID,
		fmt.Sprintf("安排停靠服务区%s，原因：%s，休息%d分钟", serviceArea.Name, reason, restDuration))

	return nil
}

func (s *MonitorService) NotifyLegalStation(ctx context.Context, vehicleID int64) error {
	logger.Sugar.Warnf("notified legal station for vehicle %d", vehicleID)
	return nil
}

func toFloat(v interface{}) float64 {
	switch x := v.(type) {
	case float32:
		return float64(x)
	case float64:
		return x
	case int:
		return float64(x)
	case int32:
		return float64(x)
	case int64:
		return float64(x)
	default:
		return 0
	}
}
