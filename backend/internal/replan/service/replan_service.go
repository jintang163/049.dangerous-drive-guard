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
	"gorm.io/gorm"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	monitorWs "github.com/dangerous-drive-guard/backend/internal/monitor/delivery/ws"
	routecore "github.com/dangerous-drive-guard/backend/internal/route/core"
	routeSvc "github.com/dangerous-drive-guard/backend/internal/route/service"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
	"github.com/dangerous-drive-guard/backend/pkg/mq"
)

// ============================================================
// 实时重规划服务
// 核心能力：
//  1. 路况事件管理（CRUD）
//  2. 定时扫描活跃路况 → 匹配在途运单 → 触发重规划
//  3. 基于当前位置生成多策略候选路线
//  4. 通过WebSocket推送重规划提示至司机端
//  5. 司机确认 / 拒绝后应用新路线
// ============================================================

type ReplanService struct {
	db          *database.TIDB
	routeSvc    *routeSvc.RouteService
	scannerLock sync.Mutex
	// 最近一次触发的运单ID时间戳，避免30分钟内重复触发
	recentTriggers map[int64]time.Time
	triggerMu      sync.Mutex
}

var (
	replanServiceInstance *ReplanService
	replanOnce            sync.Once
)

func NewReplanService() *ReplanService {
	replanOnce.Do(func() {
		replanServiceInstance = &ReplanService{
			db:             database.GetDB(),
			routeSvc:       routeSvc.NewRouteService(),
			recentTriggers: make(map[int64]time.Time),
		}
		go replanServiceInstance.startBackgroundScan()
	})
	return replanServiceInstance
}

// 后台扫描：每60秒检查活跃路况事件，匹配可能受影响的在途运单
func (s *ReplanService) startBackgroundScan() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	logger.Global.Info("Replan background scanner started")
	for range ticker.C {
		if err := s.scanAndTrigger(context.Background()); err != nil {
			logger.Sugar.Errorf("scanAndTrigger error: %v", err)
		}
	}
}

func (s *ReplanService) scanAndTrigger(ctx context.Context) error {
	s.scannerLock.Lock()
	defer s.scannerLock.Unlock()

	activeEvents, err := s.listActiveTrafficEvents(ctx)
	if err != nil || len(activeEvents) == 0 {
		return err
	}

	activeWaybills, err := s.listActiveWaybills(ctx)
	if err != nil || len(activeWaybills) == 0 {
		return err
	}

	for _, wb := range activeWaybills {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if s.wasRecentlyTriggered(wb.ID) {
			continue
		}

		affected, affectedEvent, _ := s.isWaybillAffected(ctx, wb, activeEvents)
		if affected && affectedEvent != nil {
			_, triggerErr := s.TriggerReplan(ctx, &model.ReplanTriggerRequest{
				WaybillID:       wb.ID,
				TriggerType:     model.ReplanTriggerTraffic,
				TriggerSourceID: affectedEvent.ID,
				TriggerReason:   fmt.Sprintf("前方路况：%s，预计延误%d分钟", affectedEvent.Title, affectedEvent.DurationMinutes),
			}, nil)
			if triggerErr != nil {
				logger.Sugar.Warnf("auto trigger replan failed: waybill=%d, err=%v", wb.ID, triggerErr)
			} else {
				s.markRecentTrigger(wb.ID)
			}
		}
	}
	return nil
}

func (s *ReplanService) wasRecentlyTriggered(waybillID int64) bool {
	s.triggerMu.Lock()
	defer s.triggerMu.Unlock()
	last, ok := s.recentTriggers[waybillID]
	if !ok {
		return false
	}
	if time.Since(last) < 30*time.Minute {
		return true
	}
	delete(s.recentTriggers, waybillID)
	return false
}

func (s *ReplanService) markRecentTrigger(waybillID int64) {
	s.triggerMu.Lock()
	s.recentTriggers[waybillID] = time.Now()
	s.triggerMu.Unlock()
}

func (s *ReplanService) listActiveTrafficEvents(ctx context.Context) ([]*model.TrafficEvent, error) {
	var events []*model.TrafficEvent
	err := s.db.WithContext(ctx).
		Where("status = ? AND (actual_end_at IS NULL OR actual_end_at > ?)",
			model.TrafficEventActive, time.Now()).
		Find(&events).Error
	return events, err
}

func (s *ReplanService) listActiveWaybills(ctx context.Context) ([]*model.Waybill, error) {
	var waybills []*model.Waybill
	err := s.db.WithContext(ctx).
		Where("status = ? AND route_plan_id IS NOT NULL AND route_plan_id > 0",
			model.WaybillInTransit).
		Find(&waybills).Error
	return waybills, err
}

// 判断运单是否受路况事件影响：计算原路径是否经过事件影响区域
func (s *ReplanService) isWaybillAffected(ctx context.Context, wb *model.Waybill, events []*model.TrafficEvent) (bool, *model.TrafficEvent, error) {
	var plan model.RoutePlan
	if err := s.db.WithContext(ctx).First(&plan, wb.RoutePlanID).Error; err != nil {
		return false, nil, err
	}

	var currentLat, currentLng float64
	currentLat, currentLng = s.getVehicleLocation(ctx, wb.VehicleID, wb.OriginLatitude, wb.OriginLongitude)

	for _, evt := range events {
		if evt.EventLevel < 2 && evt.CongestionLevel < 3 {
			continue
		}
		if s.routePassesNearPoint(plan, currentLat, currentLng, evt.CenterLat, evt.CenterLng, 5000) {
			return true, evt, nil
		}
	}
	return false, nil, nil
}

func (s *ReplanService) getVehicleLocation(ctx context.Context, vehicleID int64, fallbackLat, fallbackLng float64) (float64, float64) {
	var lat, lng interface{}
	s.db.Raw(`
		SELECT latitude, longitude FROM fatigue_detection_records
		WHERE vehicle_id = ? AND detection_time > DATE_SUB(NOW(), INTERVAL 5 MINUTE)
		ORDER BY detection_time DESC LIMIT 1
	`, vehicleID).Scan([]interface{}{&lat, &lng})
	if lat != nil {
		return toFloat64(lat), toFloat64(lng)
	}
	return fallbackLat, fallbackLng
}

// 判断路线剩余部分是否经过某点radius米范围内
func (s *ReplanService) routePassesNearPoint(plan model.RoutePlan, curLat, curLng, targetLat, targetLng, radiusMeters float64) bool {
	currentPoint := model.GeoPoint{Lat: curLat, Lng: curLng}
	targetPoint := model.GeoPoint{Lat: targetLat, Lng: targetLng}
	hasPassedCurrent := false
	for _, pt := range plan.RoutePath {
		if !hasPassedCurrent {
			if pt.DistanceTo(currentPoint) < 1000 {
				hasPassedCurrent = true
			}
			continue
		}
		if pt.DistanceTo(targetPoint) <= radiusMeters {
			return true
		}
	}
	return false
}

// ============================================================
// 路况事件CRUD
// ============================================================

func (s *ReplanService) ListTrafficEvents(ctx context.Context, status model.TrafficEventStatus, eventType string, keyword string, page, pageSize int) ([]*model.TrafficEvent, int64, error) {
	var list []*model.TrafficEvent
	var total int64

	q := s.db.WithContext(ctx).Model(&model.TrafficEvent{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if eventType != "" {
		q = q.Where("event_type = ?", eventType)
	}
	if keyword != "" {
		q = q.Where("title LIKE ? OR description LIKE ? OR road_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Order("event_level DESC, created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func (s *ReplanService) GetTrafficEvent(ctx context.Context, id int64) (*model.TrafficEvent, error) {
	var evt model.TrafficEvent
	err := s.db.WithContext(ctx).First(&evt, id).Error
	if err != nil {
		return nil, err
	}
	return &evt, nil
}

func (s *ReplanService) CreateTrafficEvent(ctx context.Context, req *model.TrafficEventCreateRequest, userID int64) (*model.TrafficEvent, error) {
	eventNo := fmt.Sprintf("TE%s", strings.ToUpper(strings.ReplaceAll(uuid.New().String()[:12], "-", "")))

	startTime := time.Now()
	if req.StartedAt != nil {
		startTime = *req.StartedAt
	}

	evt := &model.TrafficEvent{
		EventNo:         eventNo,
		EventType:       req.EventType,
		EventLevel:      req.EventLevel,
		Title:           req.Title,
		Description:     req.Description,
		Source:          req.Source,
		RoadName:        req.RoadName,
		CenterLat:       req.CenterLat,
		CenterLng:       req.CenterLng,
		AffectedGeometry: req.AffectedGeometry,
		AffectedLengthKm: req.AffectedLengthKm,
		CongestionLevel: req.CongestionLevel,
		AvgSpeedKmh:     req.AvgSpeedKmh,
		DurationMinutes: req.DurationMinutes,
		StartedAt:       startTime,
		ExpectedEndAt:   req.ExpectedEndAt,
		Status:          model.TrafficEventActive,
	}

	if err := s.db.WithContext(ctx).Create(evt).Error; err != nil {
		return nil, err
	}

	msgBody, _ := json.Marshal(map[string]interface{}{
		"event_id":   evt.ID,
		"event_type": evt.EventType,
		"event_no":   evt.EventNo,
		"level":      evt.EventLevel,
		"title":      evt.Title,
	})
	mq.Send(ctx, mq.Message{Topic: "traffic_event", Tag: "created", Key: fmt.Sprintf("evt_%d", evt.ID), Body: msgBody})

	logger.Sugar.Infof("traffic event created: id=%d, type=%s, title=%s", evt.ID, evt.EventType, evt.Title)
	return evt, nil
}

func (s *ReplanService) ResolveTrafficEvent(ctx context.Context, id int64) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&model.TrafficEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        model.TrafficEventResolved,
			"actual_end_at": now,
			"updated_at":    now,
		}).Error
}

// ============================================================
// 重规划核心流程
// ============================================================

// TriggerReplan 手动/自动触发重规划；生成候选路线并推送提示
func (s *ReplanService) TriggerReplan(ctx context.Context, req *model.ReplanTriggerRequest, operatorID *int64) (*model.RouteReplanRecord, error) {
	var wb model.Waybill
	if err := s.db.WithContext(ctx).First(&wb, req.WaybillID).Error; err != nil {
		return nil, fmt.Errorf("运单不存在")
	}
	if wb.Status != model.WaybillInTransit {
		return nil, fmt.Errorf("仅运输中状态的运单可触发重规划")
	}

	curLat, curLng := req.CurrentLat, req.CurrentLng
	if curLat == 0 && curLng == 0 {
		curLat, curLng = s.getVehicleLocation(ctx, wb.VehicleID, wb.OriginLatitude, wb.OriginLongitude)
	}

	var originalPlan model.RoutePlan
	if wb.RoutePlanID > 0 {
		s.db.WithContext(ctx).First(&originalPlan, wb.RoutePlanID)
	}

	origRemainingDist, origRemainingDur := s.estimateRemainingRoute(originalPlan, curLat, curLng)

	replanNo := fmt.Sprintf("RP%s", strings.ToUpper(strings.ReplaceAll(uuid.New().String()[:12], "-", "")))

	record := &model.RouteReplanRecord{
		ReplanNo:                 replanNo,
		WaybillID:                wb.ID,
		WaybillNo:                wb.WaybillNo,
		VehicleID:                wb.VehicleID,
		VehiclePlate:             "",
		DriverID:                 wb.DriverID,
		DriverName:               "",
		OriginalRoutePlanID:      wb.RoutePlanID,
		TriggerType:              req.TriggerType,
		TriggerSourceID:          req.TriggerSourceID,
		TriggerReason:            req.TriggerReason,
		CurrentLat:               curLat,
		CurrentLng:               curLng,
		OriginalDistanceRemaining: origRemainingDist,
		OriginalDurationRemaining: origRemainingDur,
		Status:                   model.ReplanStatusPending,
	}
	if req.EventType != "" {
		record.EventType = req.EventType
	}
	if operatorID != nil {
		record.OperatorID = *operatorID
	}

	var vehicle model.Vehicle
	if s.db.WithContext(ctx).First(&vehicle, wb.VehicleID).Error == nil {
		record.VehiclePlate = vehicle.PlateNumber
	}
	var driver model.User
	if s.db.WithContext(ctx).First(&driver, wb.DriverID).Error == nil {
		record.DriverName = driver.RealName
	}

	var triggerEvent *model.TrafficEvent
	if req.TriggerType == model.ReplanTriggerTraffic && req.TriggerSourceID > 0 {
		var te model.TrafficEvent
		if s.db.WithContext(ctx).First(&te, req.TriggerSourceID).Error == nil {
			triggerEvent = &te
			record.EventType = string(te.EventType)
		}
	}

	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}

	candidates, bestCandidate, err := s.generateCandidateRoutes(ctx, &wb, &originalPlan, curLat, curLng, triggerEvent)
	if err != nil {
		logger.Sugar.Warnf("generate candidate routes failed: err=%v", err)
	}

	bestDuration := origRemainingDur
	bestDistance := origRemainingDist
	if bestCandidate != nil {
		bestDuration = bestCandidate.EstimatedDuration
		bestDistance = bestCandidate.TotalDistance
	}
	record.NewDurationRemaining = bestDuration
	record.NewDistanceRemaining = bestDistance
	record.DistanceDelta = bestDistance - origRemainingDist
	record.DurationDelta = bestDuration - origRemainingDur
	s.db.WithContext(ctx).Save(record)

	record.CandidateRoutes = candidates

	notifiedAt := time.Now()
	record.NotifiedAt = &notifiedAt
	s.db.WithContext(ctx).Save(record)

	s.pushReplanNotification(ctx, record, candidates, triggerEvent)

	logger.Sugar.Infof("replan triggered: no=%s, waybill=%s, trigger=%s, reason=%s",
		record.ReplanNo, record.WaybillNo, record.TriggerType, record.TriggerReason)
	return record, nil
}

// 生成多策略候选路线（安全优先、距离最短、时间最快）
func (s *ReplanService) generateCandidateRoutes(
	ctx context.Context,
	wb *model.Waybill,
	originalPlan *model.RoutePlan,
	curLat, curLng float64,
	triggerEvent *model.TrafficEvent,
) ([]*model.ReplanCandidateRoute, *model.ReplanCandidateRoute, error) {
	strategies := []model.RouteStrategy{
		model.StrategySafest,
		model.StrategyShortest,
		model.StrategyEconomic,
	}

	origin := model.Coordinate{
		Latitude:  curLat,
		Longitude: curLng,
		Address:   "当前位置",
	}
	dest := model.Coordinate{
		Latitude:  wb.DestLatitude,
		Longitude: wb.DestLongitude,
		Address:   wb.DestAddress,
	}

	// 需规避的路况事件ID列表
	var avoidEventIDs []int64
	if triggerEvent != nil {
		avoidEventIDs = append(avoidEventIDs, triggerEvent.ID)
	}

	var candidates []*model.ReplanCandidateRoute
	rankOrder := 0
	for _, strategy := range strategies {
		planReq := &model.RoutePlanRequest{
			Origin:        origin,
			Destination:   dest,
			Strategy:      strategy,
			WaybillID:     wb.ID,
			VehicleID:     wb.VehicleID,
			DriverID:      wb.DriverID,
			HazardClass:   wb.GoodsHazardClass,
		}

		plan, err := s.routeSvc.PlanRoute(ctx, planReq)
		if err != nil || plan == nil {
			logger.Sugar.Warnf("strategy %s route plan failed: %v", strategy, err)
			continue
		}

		if !s.routeAvoidsEvent(plan, triggerEvent, curLat, curLng) {
			continue
		}

		candidate := &model.ReplanCandidateRoute{
			ReplanRecordID:    0,
			RouteGeometry:     plan.RouteGeometry,
			Strategy:          string(strategy),
			TotalDistance:     plan.TotalDistance,
			EstimatedDuration: plan.EstimatedDuration,
			TollFee:           plan.TollFee,
			FuelCost:          plan.FuelCost,
			SafetyScore:       plan.SafetyScore,
			IsRecommended:     0,
		}
		candidates = append(candidates, candidate)
		if len(candidates) >= 3 {
			break
		}
	}

	if len(candidates) == 0 {
		logger.Sugar.Warnf("no valid candidate routes after filtering, fallback to best-effort plan")
		planReq := &model.RoutePlanRequest{
			Origin:      origin,
			Destination: dest,
			Strategy:    model.StrategySafest,
			WaybillID:   wb.ID,
			VehicleID:   wb.VehicleID,
			DriverID:    wb.DriverID,
			HazardClass: wb.GoodsHazardClass,
		}
		if plan, err := s.routeSvc.PlanRoute(ctx, planReq); err == nil && plan != nil {
			candidates = append(candidates, &model.ReplanCandidateRoute{
				RouteGeometry:     plan.RouteGeometry,
				Strategy:          "safest",
				TotalDistance:     plan.TotalDistance,
				EstimatedDuration: plan.EstimatedDuration,
				TollFee:           plan.TollFee,
				FuelCost:          plan.FuelCost,
				SafetyScore:       plan.SafetyScore,
				IsRecommended:     1,
			})
		}
	}

	_ = avoidEventIDs
	best := s.selectBestCandidate(candidates)
	if best != nil {
		best.IsRecommended = 1
	}

	for _, c := range candidates {
		c.ReplanRecordID = 0
	}

	// 持久化候选路线
	for i, c := range candidates {
		c.RankOrder = rankOrder
		rankOrder++
		candidates[i] = c
	}
	return candidates, best, nil
}

func (s *ReplanService) routeAvoidsEvent(plan *model.RoutePlan, evt *model.TrafficEvent, curLat, curLng float64) bool {
	if evt == nil {
		return true
	}
	return !s.routePassesNearPoint(*plan, curLat, curLng, evt.CenterLat, evt.CenterLng, 3000)
}

// 选择最优路线：安全优先，其次距离、时延权衡
func (s *ReplanService) selectBestCandidate(candidates []*model.ReplanCandidateRoute) *model.ReplanCandidateRoute {
	if len(candidates) == 0 {
		return nil
	}
	best := candidates[0]
	for _, c := range candidates {
		if c.SafetyScore > best.SafetyScore {
			best = c
		} else if c.SafetyScore == best.SafetyScore {
			if c.TotalDistance < best.TotalDistance {
				best = c
			}
		}
	}
	return best
}

func (s *ReplanService) estimateRemainingRoute(plan model.RoutePlan, curLat, curLng float64) (distanceKm float64, durationMin int) {
	currentPoint := model.GeoPoint{Lat: curLat, Lng: curLng}
	passedCurrent := false
	totalDist := 0.0
	var prevPt *model.GeoPoint
	for i := range plan.RoutePath {
		pt := plan.RoutePath[i]
		if !passedCurrent {
			if pt.DistanceTo(currentPoint) < 1000 {
				passedCurrent = true
				prevPt = &pt
			}
			continue
		}
		if prevPt != nil {
			totalDist += pt.DistanceTo(*prevPt) / 1000.0
		}
		prevPt = &pt
	}
	avgSpeed := 50.0
	if plan.ExpectedSpeed > 0 {
		avgSpeed = plan.ExpectedSpeed
	}
	durationMin = int(math.Round(totalDist / avgSpeed * 60))
	return totalDist, durationMin
}

// 通过 WebSocket 推送重规划提示（管理端监控 + 司机端导航）
func (s *ReplanService) pushReplanNotification(
	ctx context.Context,
	record *model.RouteReplanRecord,
	candidates []*model.ReplanCandidateRoute,
	evt *model.TrafficEvent,
) {
	hub := monitorWs.GetHub()
	candidateInfo := make([]map[string]interface{}, 0)
	for _, c := range candidates {
		candidateInfo = append(candidateInfo, map[string]interface{}{
			"strategy":           c.Strategy,
			"total_distance":     c.TotalDistance,
			"estimated_duration": c.EstimatedDuration,
			"safety_score":       c.SafetyScore,
			"toll_fee":           c.TollFee,
			"is_recommended":     c.IsRecommended,
		})
	}

	payload := map[string]interface{}{
		"replan_id":    record.ID,
		"replan_no":    record.ReplanNo,
		"waybill_id":   record.WaybillID,
		"waybill_no":   record.WaybillNo,
		"vehicle_id":   record.VehicleID,
		"vehicle_plate": record.VehiclePlate,
		"driver_id":    record.DriverID,
		"driver_name":  record.DriverName,
		"trigger_type": record.TriggerType,
		"trigger_reason": record.TriggerReason,
		"original_distance_remaining": record.OriginalDistanceRemaining,
		"original_duration_remaining": record.OriginalDurationRemaining,
		"new_distance_remaining":      record.NewDistanceRemaining,
		"new_duration_remaining":      record.NewDurationRemaining,
		"distance_delta":              record.DistanceDelta,
		"duration_delta":              record.DurationDelta,
		"current_lat": record.CurrentLat,
		"current_lng": record.CurrentLng,
		"candidates":  candidateInfo,
		"status":      string(record.Status),
		"created_at":  record.CreatedAt.Format(time.RFC3339),
	}
	if evt != nil {
		payload["traffic_event"] = map[string]interface{}{
			"id":        evt.ID,
			"type":      evt.EventType,
			"level":     evt.EventLevel,
			"title":     evt.Title,
			"road_name": evt.RoadName,
			"center_lat": evt.CenterLat,
			"center_lng": evt.CenterLng,
		}
	}

	hub.BroadcastReplanSuggestion(record.VehicleID, record.DriverID, payload)

	msgBody, _ := json.Marshal(payload)
	mq.Send(ctx, mq.Message{
		Topic: "route_replan",
		Tag:   "suggested",
		Key:   fmt.Sprintf("rp_%d", record.ID),
		Body:  msgBody,
	})
}

// ConfirmReplan 司机确认/拒绝重规划；确认后应用新路线并更新运单
func (s *ReplanService) ConfirmReplan(ctx context.Context, recordID int64, action, note string, driverID *int64) (*model.RouteReplanRecord, error) {
	var record model.RouteReplanRecord
	if err := s.db.WithContext(ctx).First(&record, recordID).Error; err != nil {
		return nil, fmt.Errorf("重规划记录不存在")
	}
	if record.Status != model.ReplanStatusPending {
		return nil, fmt.Errorf("重规划记录当前状态为 %s，不允许此操作", record.Status)
	}

	now := time.Now()

	switch action {
	case "confirm":
		var wb model.Waybill
		if err := s.db.WithContext(ctx).First(&wb, record.WaybillID).Error; err == nil {
			newPlanID, err := s.applyBestRoutePlan(ctx, &record, &wb, driverID)
			if err != nil {
				logger.Sugar.Warnf("apply best route plan failed: err=%v", err)
			} else if newPlanID > 0 {
				record.NewRoutePlanID = newPlanID
			}
		}
		record.Status = model.ReplanStatusConfirmed
		record.DriverConfirmAt = &now
		record.AppliedAt = &now
		s.notifyRouteApplied(ctx, &record)

	case "reject":
		record.Status = model.ReplanStatusRejected
		record.DriverConfirmAt = &now

	default:
		return nil, fmt.Errorf("不支持的操作类型: %s", action)
	}

	record.ConfirmNote = note
	if err := s.db.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// 将最优候选路线保存为正式路线规划，并更新运单引用
func (s *ReplanService) applyBestRoutePlan(ctx context.Context, record *model.RouteReplanRecord, wb *model.Waybill, driverID *int64) (int64, error) {
	var candidates []*model.ReplanCandidateRoute
	s.db.WithContext(ctx).Where("replan_record_id = ?", record.ID).
		Order("is_recommended DESC, rank_order ASC").
		Find(&candidates)

	var best *model.ReplanCandidateRoute
	for _, c := range candidates {
		if c.IsRecommended == 1 {
			best = c
			break
		}
	}
	if best == nil && len(candidates) > 0 {
		best = candidates[0]
	}
	if best == nil {
		return 0, nil
	}

	var path []model.GeoPoint
	if len(best.RoutePath) > 0 {
		json.Unmarshal(best.RoutePath, &path)
	}

	plan := &model.RoutePlan{
		PlanNo:             fmt.Sprintf("RP%s", strings.ToUpper(uuid.New().String()[:12])),
		WaybillID:          record.WaybillID,
		VehicleID:          record.VehicleID,
		DriverID:           wb.DriverID,
		Strategy:           model.RouteStrategy(best.Strategy),
		Origin:             model.Coordinate{Latitude: record.CurrentLat, Longitude: record.CurrentLng, Address: "当前位置"},
		Destination:        model.Coordinate{Latitude: wb.DestLatitude, Longitude: wb.DestLongitude, Address: wb.DestAddress},
		RouteGeometry:      best.RouteGeometry,
		RoutePath:          path,
		TotalDistance:      best.TotalDistance,
		EstimatedDuration:  best.EstimatedDuration,
		TollFee:            best.TollFee,
		FuelCost:           best.FuelCost,
		SafetyScore:        best.SafetyScore,
		Status:             "active",
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	if err := tx.Create(plan).Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := tx.Model(wb).Update("route_plan_id", plan.ID).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return plan.ID, nil
}

func (s *ReplanService) notifyRouteApplied(ctx context.Context, record *model.RouteReplanRecord) {
	hub := monitorWs.GetHub()
	hub.NotifyRouteApplied(record.VehicleID, record.DriverID, map[string]interface{}{
		"replan_id":         record.ID,
		"replan_no":         record.ReplanNo,
		"new_route_plan_id": record.NewRoutePlanID,
		"waybill_id":        record.WaybillID,
		"applied_at":        record.AppliedAt.Format(time.RFC3339),
	})

	msgBody, _ := json.Marshal(map[string]interface{}{
		"replan_id":         record.ID,
		"new_route_plan_id": record.NewRoutePlanID,
		"waybill_id":        record.WaybillID,
		"vehicle_id":        record.VehicleID,
	})
	mq.Send(ctx, mq.Message{
		Topic: "route_replan",
		Tag:   "applied",
		Key:   fmt.Sprintf("rp_%d", record.ID),
		Body:  msgBody,
	})
}

// ============================================================
// 重规划历史查询
// ============================================================

func (s *ReplanService) ListReplanRecords(ctx context.Context, params *model.ReplanQueryParams) ([]*model.RouteReplanRecord, int64, error) {
	var list []*model.RouteReplanRecord
	var total int64

	q := s.db.WithContext(ctx).Model(&model.RouteReplanRecord{})
	if params.WaybillID > 0 {
		q = q.Where("waybill_id = ?", params.WaybillID)
	}
	if params.VehicleID > 0 {
		q = q.Where("vehicle_id = ?", params.VehicleID)
	}
	if params.DriverID > 0 {
		q = q.Where("driver_id = ?", params.DriverID)
	}
	if params.TriggerType != "" {
		q = q.Where("trigger_type = ?", params.TriggerType)
	}
	if params.Status != "" {
		q = q.Where("status = ?", params.Status)
	}
	if params.Keyword != "" {
		q = q.Where("waybill_no LIKE ? OR vehicle_plate LIKE ? OR trigger_reason LIKE ? OR replan_no LIKE ?",
			"%"+params.Keyword+"%", "%"+params.Keyword+"%",
			"%"+params.Keyword+"%", "%"+params.Keyword+"%")
	}
	if params.StartDate != "" {
		q = q.Where("DATE(created_at) >= ?", params.StartDate)
	}
	if params.EndDate != "" {
		q = q.Where("DATE(created_at) <= ?", params.EndDate)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Order("created_at DESC").
		Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).
		Find(&list).Error
	return list, total, err
}

func (s *ReplanService) GetReplanRecord(ctx context.Context, id int64) (*model.RouteReplanRecord, error) {
	var record model.RouteReplanRecord
	if err := s.db.WithContext(ctx).First(&record, id).Error; err != nil {
		return nil, err
	}

	var candidates []*model.ReplanCandidateRoute
	s.db.WithContext(ctx).Where("replan_record_id = ?", id).
		Order("is_recommended DESC, rank_order ASC").Find(&candidates)
	record.CandidateRoutes = candidates

	if record.OriginalRoutePlanID > 0 {
		var orig model.RoutePlan
		if s.db.WithContext(ctx).First(&orig, record.OriginalRoutePlanID).Error == nil {
			record.OriginalRoutePlan = &orig
		}
	}
	if record.NewRoutePlanID > 0 {
		var newPlan model.RoutePlan
		if s.db.WithContext(ctx).First(&newPlan, record.NewRoutePlanID).Error == nil {
			record.NewRoutePlan = &newPlan
		}
	}
	if record.TriggerSourceID > 0 && record.TriggerType == model.ReplanTriggerTraffic {
		var evt model.TrafficEvent
		if s.db.WithContext(ctx).First(&evt, record.TriggerSourceID).Error == nil {
			record.TrafficEvents = []*model.TrafficEvent{&evt}
		}
	}

	return &record, nil
}

// 统计接口
func (s *ReplanService) GetReplanStatistics(ctx context.Context, orgID int64, days int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	baseQ := s.db.WithContext(ctx).Model(&model.RouteReplanRecord{})
	if days > 0 {
		baseQ = baseQ.Where("created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)", days)
	}

	var total int64
	baseQ.Count(&total)
	stats["total"] = total

	for status, key := range map[model.ReplanRecordStatus]string{
		model.ReplanStatusPending:     "pending",
		model.ReplanStatusConfirmed:   "confirmed",
		model.ReplanStatusRejected:    "rejected",
		model.ReplanStatusAutoApplied: "auto_applied",
	} {
		var c int64
		tmp := *baseQ
		tmp.Where("status = ?", status).Count(&c)
		stats[key] = c
	}

	triggerStats := make(map[string]int64)
	rows, _ := baseQ.Group("trigger_type").Select("trigger_type, COUNT(*) as cnt").Rows()
	defer rows.Close()
	for rows.Next() {
		var t string
		var c int64
		rows.Scan(&t, &c)
		triggerStats[t] = c
	}
	stats["trigger_distribution"] = triggerStats

	var avgDelayMin float64
	tmp := *baseQ
	tmp.Select("COALESCE(AVG(NULLIF(duration_delta, 0)), 0)").Scan(&avgDelayMin)
	stats["avg_delay_minutes"] = math.Round(avgDelayMin*100) / 100

	return stats, nil
}

// ============================================================
// 工具函数
// ============================================================

func toFloat64(v interface{}) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int64:
		return float64(x)
	default:
		return 0
	}
}

// 确保 routeSvc.NewRouteService 是幂等单例
type _routeSvcCompat = routeSvc.RouteService
type _astarCompat = routecore.AStarPlanner
type _gormCompat = gorm.DB
type _zapLogger = zap.Logger
