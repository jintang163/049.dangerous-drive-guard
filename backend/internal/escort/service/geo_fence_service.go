package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/internal/monitor/delivery/ws"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
)

const (
	DefaultDeviateThresholdMeters = 500
	MaxDailyDeviateBeforeReport   = 3
)

type GeoFenceService struct {
	db *database.TIDB
	mu sync.RWMutex
}

func NewGeoFenceService() *GeoFenceService {
	return &GeoFenceService{
		db: database.GetDB(),
	}
}

func (s *GeoFenceService) generateNo(prefix string) string {
	return fmt.Sprintf("%s%s", prefix, strings.ToUpper(strings.ReplaceAll(uuid.New().String()[:12], "-", "")))
}

func (s *GeoFenceService) CheckDeviation(ctx context.Context, req *model.GeoFenceCheckRequest) (*model.GeoFenceCheckResult, error) {
	threshold := req.Threshold
	if threshold <= 0 {
		threshold = DefaultDeviateThresholdMeters
	}

	result := &model.GeoFenceCheckResult{
		IsDeviated:         false,
		DistanceFromRouteM: 0,
		ThresholdMeters:    threshold,
		AlertLevel:         1,
		Status:             model.GeoFenceStatusPending,
	}

	routePath := s.getRoutePathForWaybill(ctx, req.WaybillID, req.VehicleID)

	if len(routePath) < 2 {
		result.Message = "route not available"
		return result, nil
	}

	currentPoint := model.GeoPoint{Lat: req.Latitude, Lng: req.Longitude}
	minDistance, nearestPoint := pointToPolylineDistance(currentPoint, routePath)

	result.DistanceFromRouteM = int(minDistance)
	result.NearestRoutePoint = &nearestPoint

	if int(minDistance) <= threshold {
		result.IsDeviated = false
		result.Message = "within route"
		return result, nil
	}

	result.IsDeviated = true
	if minDistance > float64(threshold*2) {
		result.AlertLevel = 3
	} else {
		result.AlertLevel = 2
	}

	if s.hasRecentPendingAlert(ctx, req.VehicleID, req.WaybillID, int(time.Minute*10)) {
		result.Message = "recent alert exists, skipped"
		return result, nil
	}

	dailyCount := s.getTodayDeviateCount(ctx, req.VehicleID, req.WaybillID)
	dailyCount++

	alertNo := s.generateNo("GFA")

	var plateNumber, driverName, escortName, waybillNo string
	var driverID, escortID int64
	var routePlanID int64

	s.db.Raw("SELECT plate_number, driver_id FROM vehicles WHERE id = ?", req.VehicleID).
		Scan(&plateNumber, &driverID)
	if req.DriverID > 0 {
		driverID = req.DriverID
	}
	s.db.Raw("SELECT real_name FROM users WHERE id = ?", driverID).Scan(&driverName)

	if req.WaybillID > 0 {
		s.db.Raw("SELECT waybill_no, escort_id, route_plan_id FROM waybills WHERE id = ?", req.WaybillID).
			Scan(&waybillNo, &escortID, &routePlanID)
		s.db.Raw("SELECT real_name FROM users WHERE id = ?", escortID).Scan(&escortName)
	}

	npBytes, _ := json.Marshal(nearestPoint)

	autoReported := dailyCount >= MaxDailyDeviateBeforeReport
	now := time.Now()
	status := model.GeoFenceStatusPending
	reportedAt := interface{}(nil)
	if autoReported {
		status = model.GeoFenceStatusEscalated
		reportedAt = now
	}

	insertResult := s.db.WithContext(ctx).Exec(`
		INSERT INTO geo_fence_alerts
		(alert_no, vehicle_id, plate_number, driver_id, driver_name, escort_id, escort_name,
		 waybill_id, waybill_no, route_plan_id,
		 latitude, longitude, address, distance_from_route_meters, threshold_meters,
		 alert_level, status, daily_deviate_count,
		 nearest_route_point,
		 reported_to_dispatch, reported_at,
		 created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		alertNo, req.VehicleID, plateNumber, driverID, driverName, escortID, escortName,
		req.WaybillID, waybillNo, routePlanID,
		req.Latitude, req.Longitude, req.Address, int(minDistance), threshold,
		result.AlertLevel, status, dailyCount,
		string(npBytes),
		autoReported, reportedAt,
	)

	if insertResult.Error != nil {
		logger.Sugar.Errorf("create geo fence alert error: %v", insertResult.Error)
		return nil, insertResult.Error
	}

	var alertID int64
	s.db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&alertID)

	result.AlertID = alertID
	result.AlertNo = alertNo
	result.DailyDeviateCount = dailyCount
	result.AutoReported = autoReported
	result.Status = status

	if autoReported {
		result.Message = fmt.Sprintf("deviated %d meters, auto reported to dispatch (daily count: %d)", int(minDistance), dailyCount)
	} else {
		result.Message = fmt.Sprintf("deviated %d meters, need escort confirm (daily count: %d)", int(minDistance), dailyCount)
	}

	alertData := map[string]interface{}{
		"id":                          alertID,
		"alert_no":                    alertNo,
		"vehicle_id":                  req.VehicleID,
		"plate_number":                plateNumber,
		"driver_id":                   driverID,
		"driver_name":                 driverName,
		"escort_id":                   escortID,
		"escort_name":                 escortName,
		"waybill_id":                  req.WaybillID,
		"waybill_no":                  waybillNo,
		"latitude":                    req.Latitude,
		"longitude":                   req.Longitude,
		"address":                     req.Address,
		"distance_from_route_meters":  int(minDistance),
		"threshold_meters":            threshold,
		"alert_level":                 result.AlertLevel,
		"status":                      string(status),
		"daily_deviate_count":         dailyCount,
		"auto_reported":               autoReported,
		"nearest_route_point":         nearestPoint,
		"type":                        "geo_fence",
		"popup":                       true,
		"timestamp":                   now.Format("2006-01-02 15:04:05"),
	}

	hub := ws.GetHub()
	hub.BroadcastEscort(ctx, escortID, alertData)
	hub.BroadcastMonitor(ctx, alertData)
	if autoReported {
		hub.BroadcastDispatch(ctx, alertData)
	}

	logger.Global.Warn("geo fence alert created",
		zap.Int64("alert_id", alertID),
		zap.String("alert_no", alertNo),
		zap.Int64("vehicle_id", req.VehicleID),
		zap.Int("distance_m", int(minDistance)),
		zap.Int("daily_count", dailyCount),
		zap.Bool("auto_reported", autoReported))

	return result, nil
}

func (s *GeoFenceService) ConfirmAlert(ctx context.Context, req *model.GeoFenceConfirmRequest, confirmerID int64, confirmerName, confirmerRole string) (*model.GeoFenceAlert, error) {
	var alert model.GeoFenceAlert
	err := s.db.WithContext(ctx).Table("geo_fence_alerts").Where("id = ?", req.AlertID).First(&alert).Error
	if err != nil {
		return nil, fmt.Errorf("alert not found")
	}

	if alert.Status != model.GeoFenceStatusPending {
		return nil, fmt.Errorf("alert status is not pending")
	}

	now := time.Now()
	alert.Status = model.GeoFenceStatusConfirmed
	alert.DeviateReason = &req.ConfirmType
	alert.ConfirmNote = req.Note
	alert.ConfirmedBy = &confirmerID
	alert.ConfirmedRole = confirmerRole
	alert.ConfirmedAt = &now

	saveResult := s.db.WithContext(ctx).Exec(`
		UPDATE geo_fence_alerts SET
		status = 'confirmed', deviate_reason = ?, confirm_note = ?,
		confirmed_by = ?, confirmed_role = ?, confirmed_at = ?, updated_at = ?
		WHERE id = ? AND status = 'pending'`,
		req.ConfirmType, req.Note, confirmerID, confirmerRole, now, now, req.AlertID,
	)
	if saveResult.Error != nil {
		return nil, saveResult.Error
	}

	var plateNumber, waybillNo string
	s.db.Raw("SELECT plate_number FROM vehicles WHERE id = ?", alert.VehicleID).Scan(&plateNumber)
	s.db.Raw("SELECT waybill_no FROM waybills WHERE id = ?", alert.WaybillID).Scan(&waybillNo)

	var lat, lng *float64
	if req.Latitude != 0 {
		lat = &req.Latitude
	}
	if req.Longitude != 0 {
		lng = &req.Longitude
	}

	s.db.WithContext(ctx).Exec(`
		INSERT INTO geo_fence_confirm_logs
		(alert_id, alert_no, vehicle_id, plate_number, waybill_id, waybill_no,
		 confirm_type, reason_detail, note,
		 confirmed_by, confirmed_name, confirmed_role,
		 latitude, longitude, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		req.AlertID, alert.AlertNo, alert.VehicleID, plateNumber, alert.WaybillID, waybillNo,
		req.ConfirmType, req.ReasonDetail, req.Note,
		confirmerID, confirmerName, confirmerRole,
		lat, lng,
	)

	confirmData := map[string]interface{}{
		"alert_id":     req.AlertID,
		"alert_no":     alert.AlertNo,
		"status":       "confirmed",
		"confirm_type": string(req.ConfirmType),
		"note":         req.Note,
		"confirmed_by": confirmerName,
		"confirmed_role": confirmerRole,
		"confirmed_at": now.Format("2006-01-02 15:04:05"),
		"type":         "geo_fence_confirm",
	}

	hub := ws.GetHub()
	hub.BroadcastMonitor(ctx, confirmData)
	hub.BroadcastDispatch(ctx, confirmData)

	logger.Global.Info("geo fence alert confirmed",
		zap.Int64("alert_id", req.AlertID),
		zap.String("confirm_type", string(req.ConfirmType)),
		zap.Int64("confirmer_id", confirmerID))

	return &alert, nil
}

func (s *GeoFenceService) ResolveAlert(ctx context.Context, req *model.GeoFenceResolveRequest, resolverID int64, resolverName string) error {
	var alert model.GeoFenceAlert
	err := s.db.WithContext(ctx).Table("geo_fence_alerts").Where("id = ?", req.AlertID).First(&alert).Error
	if err != nil {
		return fmt.Errorf("alert not found")
	}

	if alert.Status == model.GeoFenceStatusResolved {
		return fmt.Errorf("alert already resolved")
	}

	now := time.Now()

	result := s.db.WithContext(ctx).Exec(`
		UPDATE geo_fence_alerts SET
		status = 'resolved', resolved_by = ?, resolved_note = ?, resolved_at = ?, updated_at = ?
		WHERE id = ?`,
		resolverID, req.ResolvedNote, now, now, req.AlertID,
	)
	if result.Error != nil {
		return result.Error
	}

	logger.Global.Info("geo fence alert resolved",
		zap.Int64("alert_id", req.AlertID),
		zap.Int64("resolver_id", resolverID))

	return nil
}

func (s *GeoFenceService) ListAlerts(ctx context.Context, req *model.GeoFenceListRequest) ([]*model.GeoFenceAlert, int64, error) {
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var total int64
	query := s.db.WithContext(ctx).Table("geo_fence_alerts")
	if req.VehicleID > 0 {
		query = query.Where("vehicle_id = ?", req.VehicleID)
	}
	if req.WaybillID > 0 {
		query = query.Where("waybill_id = ?", req.WaybillID)
	}
	if req.EscortID > 0 {
		query = query.Where("escort_id = ?", req.EscortID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	rows, err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var alerts []*model.GeoFenceAlert
	for rows.Next() {
		var a model.GeoFenceAlert
		rows.Scan(&a.ID, &a.AlertNo, &a.VehicleID, &a.PlateNumber,
			&a.DriverID, &a.DriverName, &a.EscortID, &a.EscortName,
			&a.WaybillID, &a.WaybillNo, &a.RoutePlanID,
			&a.Latitude, &a.Longitude, &a.Address,
			&a.DistanceFromRouteM, &a.ThresholdMeters,
			&a.AlertLevel, &a.Status, &a.DeviateReason, &a.ConfirmNote,
			&a.ConfirmedBy, &a.ConfirmedRole, &a.ConfirmedAt,
			&a.ReportedToDispatch, &a.ReportedAt,
			&a.ResolvedBy, &a.ResolvedNote, &a.ResolvedAt,
			&a.DailyDeviateCount, &a.NearestRoutePoint, &a.SnapshotURL,
			&a.PopupDisplayed, &a.NotifiedEscort,
			&a.CreatedAt, &a.UpdatedAt)
		alerts = append(alerts, &a)
	}

	return alerts, total, nil
}

func (s *GeoFenceService) ListConfirmLogs(ctx context.Context, alertID, vehicleID int64, page, pageSize int) ([]*model.GeoFenceConfirmLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var total int64
	query := s.db.WithContext(ctx).Table("geo_fence_confirm_logs")
	if alertID > 0 {
		query = query.Where("alert_id = ?", alertID)
	}
	if vehicleID > 0 {
		query = query.Where("vehicle_id = ?", vehicleID)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	rows, err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*model.GeoFenceConfirmLog
	for rows.Next() {
		var l model.GeoFenceConfirmLog
		rows.Scan(&l.ID, &l.AlertID, &l.AlertNo,
			&l.VehicleID, &l.PlateNumber, &l.WaybillID, &l.WaybillNo,
			&l.ConfirmType, &l.ReasonDetail, &l.Note,
			&l.ConfirmedBy, &l.ConfirmedName, &l.ConfirmedRole,
			&l.Latitude, &l.Longitude, &l.CreatedAt)
		logs = append(logs, &l)
	}

	return logs, total, nil
}

func (s *GeoFenceService) GetStatistics(ctx context.Context, orgID int64) (*model.GeoFenceStats, error) {
	stats := &model.GeoFenceStats{}

	s.db.Table("geo_fence_alerts").Count(&stats.TotalAlerts)
	s.db.Table("geo_fence_alerts").Where("status = ?", "pending").Count(&stats.PendingAlerts)
	s.db.Table("geo_fence_alerts").Where("DATE(created_at) = CURDATE()").Count(&stats.TodayAlerts)
	s.db.Table("geo_fence_alerts").Where("reported_to_dispatch = ?", 1).Count(&stats.ReportedAlerts)
	s.db.Table("geo_fence_alerts").Where("status = ?", "resolved").Count(&stats.ResolvedAlerts)
	s.db.Table("geo_fence_confirm_logs").Count(&stats.TotalConfirmLogs)
	s.db.Table("geo_fence_alerts").Where("deviate_reason = ?", "detour").Count(&stats.DetourCount)
	s.db.Table("geo_fence_alerts").Where("deviate_reason = ?", "deviate").Count(&stats.DeviateCount)
	s.db.Table("geo_fence_alerts").Where("status = ?", "escalated").Count(&stats.AutoReportedCount)

	return stats, nil
}

func (s *GeoFenceService) getRoutePathForWaybill(ctx context.Context, waybillID, vehicleID int64) []model.GeoPoint {
	var routePlanID int64
	var routePathJSON string

	if waybillID > 0 {
		s.db.Raw("SELECT route_plan_id FROM waybills WHERE id = ?", waybillID).Scan(&routePlanID)
	}

	if routePlanID == 0 && vehicleID > 0 {
		s.db.Raw("SELECT id FROM route_plans WHERE waybill_id IN (SELECT id FROM waybills WHERE vehicle_id = ? AND status = 'in_transit') ORDER BY created_at DESC LIMIT 1", vehicleID).
			Scan(&routePlanID)
	}

	if routePlanID == 0 {
		return nil
	}

	s.db.Raw("SELECT route_geometry FROM route_plans WHERE id = ?", routePlanID).Scan(&routePathJSON)
	if routePathJSON == "" || routePathJSON == "[]" || routePathJSON == "null" {
		return s.getWaybillFallbackPath(ctx, waybillID, vehicleID)
	}

	var points []model.GeoPoint
	if err := json.Unmarshal([]byte(routePathJSON), &points); err == nil && len(points) >= 2 {
		return points
	}

	return s.getWaybillFallbackPath(ctx, waybillID, vehicleID)
}

func (s *GeoFenceService) getWaybillFallbackPath(ctx context.Context, waybillID, vehicleID int64) []model.GeoPoint {
	var originLat, originLng, destLat, destLng float64

	if waybillID > 0 {
		s.db.Raw("SELECT origin_latitude, origin_longitude, dest_latitude, dest_longitude FROM waybills WHERE id = ?", waybillID).
			Scan(&originLat, &originLng, &destLat, &destLng)
	} else if vehicleID > 0 {
		s.db.Raw(`SELECT origin_latitude, origin_longitude, dest_latitude, dest_longitude
			FROM waybills WHERE vehicle_id = ? AND status = 'in_transit' ORDER BY id DESC LIMIT 1`, vehicleID).
			Scan(&originLat, &originLng, &destLat, &destLng)
	}

	if originLat == 0 || destLat == 0 {
		return nil
	}

	return []model.GeoPoint{
		{Lat: originLat, Lng: originLng},
		{Lat: (originLat + destLat) / 2, Lng: (originLng + destLng) / 2},
		{Lat: destLat, Lng: destLng},
	}
}

func (s *GeoFenceService) hasRecentPendingAlert(ctx context.Context, vehicleID, waybillID int64, within time.Duration) bool {
	var count int64
	since := time.Now().Add(-within)

	query := s.db.WithContext(ctx).Table("geo_fence_alerts").
		Where("vehicle_id = ? AND status = ? AND created_at > ?",
			vehicleID, model.GeoFenceStatusPending, since)
	if waybillID > 0 {
		query = query.Where("waybill_id = ?", waybillID)
	}
	query.Count(&count)
	return count > 0
}

func (s *GeoFenceService) getTodayDeviateCount(ctx context.Context, vehicleID, waybillID int64) int {
	var count int64

	query := s.db.WithContext(ctx).Table("geo_fence_alerts").
		Where("vehicle_id = ? AND DATE(created_at) = CURDATE()", vehicleID)
	if waybillID > 0 {
		query = query.Where("waybill_id = ?", waybillID)
	}
	query.Count(&count)
	return int(count)
}

func pointToPolylineDistance(p model.GeoPoint, polyline []model.GeoPoint) (float64, model.GeoPoint) {
	if len(polyline) == 0 {
		return math.MaxFloat64, p
	}
	if len(polyline) == 1 {
		return p.DistanceTo(polyline[0]), polyline[0]
	}

	minDist := math.MaxFloat64
	nearest := polyline[0]

	for i := 0; i < len(polyline)-1; i++ {
		dist, proj := pointToSegmentDistance(p, polyline[i], polyline[i+1])
		if dist < minDist {
			minDist = dist
			nearest = proj
		}
	}

	return minDist, nearest
}

func pointToSegmentDistance(p, a, b model.GeoPoint) (float64, model.GeoPoint) {
	apLat := p.Lat - a.Lat
	apLng := p.Lng - a.Lng
	abLat := b.Lat - a.Lat
	abLng := b.Lng - a.Lng

	ab2 := abLat*abLat + abLng*abLng
	if ab2 < 1e-14 {
		return p.DistanceTo(a), a
	}

	t := (apLat*abLat + apLng*abLng) / ab2
	t = math.Max(0, math.Min(1, t))

	projLat := a.Lat + t*abLat
	projLng := a.Lng + t*abLng
	proj := model.GeoPoint{Lat: projLat, Lng: projLng}

	return p.DistanceTo(proj), proj
}
