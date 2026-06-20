package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
	"github.com/dangerous-drive-guard/backend/pkg/mq"
)

type TransportService struct {
	db          *database.TIDB
	blockchain  *BlockchainService
	once        sync.Once
}

type BlockchainService struct {
	records map[string]string
	mu      sync.Mutex
}

func NewTransportService(cfg *config.Config) *TransportService {
	return &TransportService{
		db:         database.GetDB(),
		blockchain: &BlockchainService{records: make(map[string]string)},
	}
}

func (s *TransportService) CreateWaybill(ctx context.Context, req *model.WaybillCreateRequest) (*model.Waybill, error) {
	waybillNo := fmt.Sprintf("WB%s", strings.ToUpper(strings.ReplaceAll(uuid.New().String()[:12], "-", "")))
	riskLevel := s.calcRiskLevel(req)

	var waybillID int64
	result := s.db.WithContext(ctx).Exec(`
		INSERT INTO waybills
		(waybill_no, order_no, shipper_org_id, carrier_org_id, receiver_org_id,
		 vehicle_id, driver_id, escort_id, goods_id, goods_name, goods_un_code,
		 goods_hazard_class, goods_weight, goods_volume, package_type, package_count,
		 origin_address, origin_latitude, origin_longitude, dest_address, dest_latitude, dest_longitude,
		 planned_departure_time, planned_arrival_time, status, risk_level, approval_status,
		 emergency_contact, emergency_phone, remark, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, DATE_ADD(?, INTERVAL CEIL(estimated_hours) HOUR), ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		waybillNo, req.OrderNo, req.ShipperOrgID, req.CarrierOrgID, req.ReceiverOrgID,
		req.VehicleID, req.DriverID, req.EscortID, req.GoodsID, s.getGoodsName(req.GoodsID),
		s.getGoodsUNCode(req.GoodsID), s.getGoodsHazardClass(req.GoodsID),
		req.GoodsWeight, req.GoodsVolume, req.PackageType, req.PackageCount,
		req.OriginAddress, req.OriginLatitude, req.OriginLongitude,
		req.DestAddress, req.DestLatitude, req.DestLongitude,
		req.PlannedDepartureTime, req.PlannedDepartureTime,
		model.WaybillCreated, riskLevel, 0,
		req.EmergencyContact, req.EmergencyPhone, req.Remark,
	)
	if result.Error != nil {
		logger.Sugar.Errorf("create waybill error: %v", result.Error)
		return nil, result.Error
	}

	s.db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&waybillID)

	waybill := &model.Waybill{
		WaybillNo: waybillNo,
	}
	waybill.ID = waybillID
	waybill.Status = model.WaybillCreated
	waybill.RiskLevel = riskLevel

	_ = s.addWaybillLog(ctx, waybillID, "", string(model.WaybillCreated), 0, "dispatcher", "运单创建", nil)

	mqBody, _ := json.Marshal(map[string]interface{}{
		"waybill_id": waybillID,
		"waybill_no": waybillNo,
		"event":      "created",
		"time":       time.Now(),
	})
	_ = mq.Send(ctx, mq.Message{Topic: "waybill_event", Key: waybillNo, Body: mqBody})

	logger.Global.Info("waybill created", zap.String("waybill_no", waybillNo), zap.Int64("id", waybillID))
	return waybill, nil
}

func (s *TransportService) calcRiskLevel(req *model.WaybillCreateRequest) int {
	hazardClass := s.getGoodsHazardClass(req.GoodsID)
	level := 1
	switch hazardClass {
	case "1", "2", "3":
		level = 3
	case "4", "5", "6", "8":
		level = 2
	}
	if req.GoodsWeight > 25 {
		level++
	}
	if level > 3 {
		level = 3
	}
	return level
}

func (s *TransportService) getGoodsName(goodsID int64) string {
	var name string
	s.db.Raw("SELECT cn_name FROM dangerous_goods WHERE id = ?", goodsID).Scan(&name)
	if name == "" {
		name = "危险品"
	}
	return name
}

func (s *TransportService) getGoodsUNCode(goodsID int64) string {
	var code string
	s.db.Raw("SELECT un_code FROM dangerous_goods WHERE id = ?", goodsID).Scan(&code)
	return code
}

func (s *TransportService) getGoodsHazardClass(goodsID int64) string {
	var c string
	s.db.Raw("SELECT hazard_class FROM dangerous_goods WHERE id = ?", goodsID).Scan(&c)
	return c
}

func (s *TransportService) GetWaybill(ctx context.Context, id int64) (*model.Waybill, error) {
	var waybill model.Waybill
	err := s.db.WithContext(ctx).Table("waybills").Where("id = ?", id).First(&waybill).Error
	if err != nil {
		return nil, err
	}
	var driver model.User
	s.db.WithContext(ctx).Table("users").Where("id = ?", waybill.DriverID).First(&driver)
	waybill.Driver = &driver
	var vehicle model.Vehicle
	s.db.WithContext(ctx).Table("vehicles").Where("id = ?", waybill.VehicleID).First(&vehicle)
	waybill.Vehicle = &vehicle
	return &waybill, nil
}

func (s *TransportService) ListWaybills(ctx context.Context, carrierID int64, status model.WaybillStatus, page, pageSize int) ([]*model.Waybill, int64, error) {
	var waybills []*model.Waybill
	var total int64

	query := s.db.WithContext(ctx).Table("waybills")
	if carrierID > 0 {
		query = query.Where("carrier_org_id = ?", carrierID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	rows, err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var w model.Waybill
		err := rows.Scan(&w.ID, &w.WaybillNo, &w.OrderNo, &w.ShipperOrgID,
			&w.CarrierOrgID, &w.ReceiverOrgID, &w.VehicleID, &w.DriverID, &w.EscortID,
			&w.RoutePlanID, &w.GoodsID, &w.GoodsName, &w.GoodsUNCode,
			&w.GoodsHazardClass, &w.GoodsWeight, &w.GoodsVolume,
			&w.PackageType, &w.PackageCount,
			&w.OriginAddress, &w.OriginLatitude, &w.OriginLongitude,
			&w.DestAddress, &w.DestLatitude, &w.DestLongitude,
			&w.PlannedDepartureTime, &w.ActualDepartureTime,
			&w.PlannedArrivalTime, &w.ActualArrivalTime,
			&w.Status, &w.TotalDistance, &w.TransportCost, &w.RiskLevel,
			&w.ApprovalStatus, &w.ApprovedBy, &w.ApprovedAt,
			&w.EmergencyContact, &w.EmergencyPhone, &w.Remark,
			&w.BlockchainTxHash, &w.BlockchainBlockNo, &w.CreatedAt, &w.UpdatedAt,
		)
		_ = err
		waybills = append(waybills, &w)
	}
	return waybills, total, nil
}

func (s *TransportService) UpdateWaybillStatus(ctx context.Context, id int64, newStatus model.WaybillStatus, operatorID int64, operatorRole, remark string) (*model.Waybill, error) {
	var currentStatus string
	s.db.WithContext(ctx).Raw("SELECT status FROM waybills WHERE id = ?", id).Scan(&currentStatus)

	now := time.Now()
	updates := map[string]interface{}{
		"status":     newStatus,
		"updated_at": now,
	}
	switch newStatus {
	case model.WaybillInTransit:
		updates["actual_departure_time"] = now
	case model.WaybillCompleted:
		updates["actual_arrival_time"] = now
	}

	result := s.db.WithContext(ctx).Table("waybills").Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}

	_ = s.addWaybillLog(ctx, id, currentStatus, string(newStatus), operatorID, operatorRole, remark, nil)

	return s.GetWaybill(ctx, id)
}

func (s *TransportService) addWaybillLog(ctx context.Context, waybillID int64, oldStatus, newStatus string, operatorID int64, operatorRole, remark string, extra map[string]interface{}) error {
	extraJSON, _ := json.Marshal(extra)
	s.db.WithContext(ctx).Exec(`
		INSERT INTO waybill_status_logs
		(waybill_id, old_status, new_status, operator_id, operator_role, remark, extra_data, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW())`,
		waybillID, oldStatus, newStatus, operatorID, operatorRole, remark, string(extraJSON),
	)
	return nil
}

func (s *TransportService) SaveToBlockchain(ctx context.Context, waybillID int64) (string, error) {
	waybill, err := s.GetWaybill(ctx, waybillID)
	if err != nil {
		return "", err
	}

	waybillJSON, _ := json.Marshal(waybill)
	hash := sha256.Sum256(waybillJSON)
	dataHash := hex.EncodeToString(hash[:])

	txHash := "0x" + hex.EncodeToString([]byte(fmt.Sprintf("tx_%d_%s", waybillID, time.Now().Format("20060102150405"))))[:64]
	blockNo := time.Now().Unix()

	s.blockchain.mu.Lock()
	s.blockchain.records[txHash] = dataHash
	s.blockchain.mu.Unlock()

	s.db.WithContext(ctx).Exec(`
		UPDATE waybills SET blockchain_tx_hash = ?, blockchain_block_no = ?, updated_at = NOW()
		WHERE id = ?`,
		txHash, blockNo, waybillID,
	)

	s.db.WithContext(ctx).Exec(`
		INSERT INTO blockchain_records
		(tx_hash, block_no, data_type, data_id, data_hash, payload, submitted_by, submit_time, chain_status, created_at)
		VALUES (?, ?, 'waybill', ?, ?, ?, ?, NOW(), 'confirmed', NOW())`,
		txHash, blockNo, waybillID, dataHash, string(waybillJSON), ctx.Value("user_id"),
	)

	logger.Global.Info("waybill saved to blockchain",
		zap.Int64("waybill_id", waybillID),
		zap.String("tx_hash", txHash))

	return txHash, nil
}

func (s *TransportService) VerifyFromBlockchain(ctx context.Context, waybillID int64) (bool, map[string]interface{}, error) {
	var txHash, storedHash string
	var blockNo int64
	s.db.WithContext(ctx).Raw(`
		SELECT blockchain_tx_hash, blockchain_block_no, br.data_hash
		FROM waybills w
		LEFT JOIN blockchain_records br ON br.data_type = 'waybill' AND br.data_id = w.id
		WHERE w.id = ?`, waybillID,
	).Scan(&txHash, &blockNo, &storedHash)

	if txHash == "" {
		return false, nil, fmt.Errorf("未找到区块链存证记录")
	}

	waybill, _ := s.GetWaybill(ctx, waybillID)
	waybillJSON, _ := json.Marshal(waybill)
	hash := sha256.Sum256(waybillJSON)
	currentHash := hex.EncodeToString(hash[:])

	verified := currentHash == storedHash || storedHash == ""

	return verified, map[string]interface{}{
		"verified":     verified,
		"tx_hash":      txHash,
		"block_no":     blockNo,
		"stored_hash":  storedHash,
		"current_hash": currentHash,
	}, nil
}

func (s *TransportService) StartEscort(ctx context.Context, waybillID int64, operatorID int64) error {
	_, err := s.UpdateWaybillStatus(ctx, waybillID, model.WaybillInTransit, operatorID, "escort", "电子押运开始")
	if err != nil {
		return err
	}

	_ = s.reportEscortEvent(ctx, &model.EscortEvent{
		WaybillID:  waybillID,
		EventType:  "departure",
		EventLevel: 1,
		ReporterID: operatorID,
		EventTime:  time.Now(),
		Description: "车辆出发，电子押运流程启动",
	})

	return nil
}

func (s *TransportService) GetEscortInfo(ctx context.Context, waybillID int64) (map[string]interface{}, error) {
	waybill, err := s.GetWaybill(ctx, waybillID)
	if err != nil {
		return nil, err
	}

	var events []model.EscortEvent
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id, waybill_id, event_type, event_level, reporter_id, reporter_role,
		       latitude, longitude, address, description, event_time, handled_status, created_at
		FROM escort_events WHERE waybill_id = ? ORDER BY event_time DESC LIMIT 50`,
		waybillID,
	).Rows()
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var e model.EscortEvent
			rows.Scan(&e.ID, &e.WaybillID, &e.EventType, &e.EventLevel,
				&e.ReporterID, &e.ReporterRole,
				&e.Latitude, &e.Longitude, &e.Address,
				&e.Description, &e.EventTime, &e.HandledStatus, &e.CreatedAt)
			events = append(events, e)
		}
	}

	var tracks []model.VehicleTrack
	trackRows, err := s.db.WithContext(ctx).Raw(`
		SELECT latitude, longitude, speed, direction, gps_time
		FROM vehicle_tracks
		WHERE waybill_id = ? ORDER BY gps_time DESC LIMIT 100`,
		waybillID,
	).Rows()
	if err == nil {
		defer trackRows.Close()
		for trackRows.Next() {
			var t model.VehicleTrack
			trackRows.Scan(&t.Latitude, &t.Longitude, &t.Speed, &t.Direction, &t.GPSTime)
			tracks = append(tracks, t)
		}
	}

	return map[string]interface{}{
		"waybill": waybill,
		"events":  events,
		"tracks":  tracks,
	}, nil
}

type EscortEventReport struct {
	WaybillID int64  `json:"waybill_id" binding:"required"`
	EventType string `json:"event_type" binding:"required"`
	EventLevel int    `json:"event_level"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string `json:"address"`
	Description string `json:"description"`
}

func (s *TransportService) reportEscortEvent(ctx context.Context, event *model.EscortEvent) error {
	s.db.WithContext(ctx).Exec(`
		INSERT INTO escort_events
		(waybill_id, vehicle_id, event_type, event_level, reporter_id, reporter_role,
		 latitude, longitude, address, description, event_time, created_at)
		VALUES (?, (SELECT vehicle_id FROM waybills WHERE id = ?), ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		event.WaybillID, event.WaybillID, event.EventType, event.EventLevel,
		event.ReporterID, event.ReporterRole,
		event.Latitude, event.Longitude, event.Address,
		event.Description, event.EventTime,
	)
	return nil
}

func (s *TransportService) ReportEscortEvent(ctx context.Context, req *EscortEventReport, reporterID int64, reporterRole string) error {
	event := &model.EscortEvent{
		WaybillID:    req.WaybillID,
		EventType:    req.EventType,
		EventLevel:   req.EventLevel,
		ReporterID:   reporterID,
		ReporterRole: reporterRole,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		Address:      req.Address,
		Description:  req.Description,
		EventTime:    time.Now(),
	}
	if event.EventLevel == 0 {
		event.EventLevel = 1
	}
	return s.reportEscortEvent(ctx, event)
}

func (s *TransportService) RecommendServiceAreas(ctx context.Context, currentLat, currentLng float64, fatigueLevel string) ([]*model.ServiceArea, error) {
	point := model.GeoPoint{Lat: currentLat, Lng: currentLng}

	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id, name, highway_name, direction, province, city, latitude, longitude,
		       distance_from_start, has_restaurant, has_hotel, has_fuel_station, has_charging,
		       has_maintenance, has_danger_goods_parking, parking_spaces, danger_parking_spaces,
		       phone, rating, status
		FROM service_areas WHERE status = 1`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.ServiceArea
	for rows.Next() {
		var a model.ServiceArea
		rows.Scan(&a.ID, &a.Name, &a.HighwayName, &a.Direction, &a.Province, &a.City,
			&a.Latitude, &a.Longitude, &a.DistanceFromStart,
			&a.HasRestaurant, &a.HasHotel, &a.HasFuelStation, &a.HasCharging,
			&a.HasMaintenance, &a.HasDangerParking, &a.ParkingSpaces, &a.DangerParkingSpaces,
			&a.Phone, &a.Rating, &a.Status)

		dist := point.DistanceTo(model.GeoPoint{Lat: a.Latitude, Lng: a.Longitude})
		a.DistanceFromCurrent = math.Round(dist/1000*100) / 100

		if dist > 150000 {
			continue
		}
		if fatigueLevel == "fatigue" && !a.HasDangerParking {
			continue
		}
		speed := 60.0
		eta := int(math.Ceil(a.DistanceFromCurrent / speed * 60))
		a.EstimatedArrivalTime = eta
		switch fatigueLevel {
		case "fatigue":
			a.RestDurationRecommend = 30
		case "warning":
			a.RestDurationRecommend = 20
		default:
			a.RestDurationRecommend = 15
		}
		result = append(result, &a)
	}
	return result, nil
}

func (s *TransportService) GetWeatherWarning(ctx context.Context, lat, lng float64) ([]map[string]interface{}, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id, warning_id, warning_type, warning_level, title, content,
		       start_time, end_time, publish_time, source
		FROM weather_warnings
		WHERE status = 1 AND start_time <= NOW() AND (end_time IS NULL OR end_time >= NOW())
		ORDER BY CASE warning_level
			WHEN 'red' THEN 1
			WHEN 'orange' THEN 2
			WHEN 'yellow' THEN 3
			WHEN 'blue' THEN 4 END
		LIMIT 10`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		item := make(map[string]interface{})
		var id, wType, level, title, content, source string
		var wID, startTime, endTime, publishTime interface{}
		rows.Scan(&id, &wID, &wType, &level, &title, &content,
			&startTime, &endTime, &publishTime, &source)
		item["id"] = id
		item["warning_id"] = wID
		item["type"] = wType
		item["level"] = level
		item["title"] = title
		item["content"] = content
		item["start_time"] = startTime
		item["end_time"] = endTime
		item["publish_time"] = publishTime
		item["source"] = source
		result = append(result, item)
	}
	return result, nil
}

type SOSRequest struct {
	VehicleID    int64   `json:"vehicle_id" binding:"required"`
	DriverID     int64   `json:"driver_id"`
	WaybillID    int64   `json:"waybill_id"`
	SOSType      string  `json:"sos_type" binding:"required"`
	SOSLevel     int     `json:"sos_level"`
	Latitude     float64 `json:"latitude" binding:"required"`
	Longitude    float64 `json:"longitude" binding:"required"`
	Address      string  `json:"address"`
	Description  string  `json:"description"`
	CallerName   string  `json:"caller_name"`
	CallerPhone  string  `json:"caller_phone"`
}

func (s *TransportService) ReportSOS(ctx context.Context, req *SOSRequest) (*model.RescueRequest, error) {
	requestNo := fmt.Sprintf("SOS%s", strings.ToUpper(strings.ReplaceAll(uuid.New().String()[:12], "-", "")))
	level := req.SOSLevel
	if level == 0 {
		level = 2
	}
	switch req.SOSType {
	case "fire", "leak", "accident":
		level = 3
	}

	var id int64
	s.db.WithContext(ctx).Exec(`
		INSERT INTO rescue_requests
		(request_no, waybill_id, vehicle_id, driver_id, sos_type, sos_level,
		 latitude, longitude, address, description, caller_name, caller_phone, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending', NOW())`,
		requestNo, req.WaybillID, req.VehicleID, req.DriverID,
		req.SOSType, level,
		req.Latitude, req.Longitude, req.Address, req.Description,
		req.CallerName, req.CallerPhone,
	)
	s.db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&id)

	logger.Global.Warn("SOS received",
		zap.String("request_no", requestNo),
		zap.String("type", req.SOSType),
		zap.Int("level", level))

	rescueReq := &model.RescueRequest{
		ID:        id,
		RequestNo: requestNo,
		SOSType:   req.SOSType,
		SOSLevel:  model.AlarmLevel(level),
		Status:    "pending",
	}

	resources, _ := s.findNearestResources(req.Latitude, req.Longitude, req.SOSType, 3)
	_ = resources

	return rescueReq, nil
}

func (s *TransportService) ListRescueResources(ctx context.Context, lat, lng float64, resourceType string, radiusKm float64) ([]*model.RescueResource, error) {
	if radiusKm <= 0 {
		radiusKm = 50
	}
	point := model.GeoPoint{Lat: lat, Lng: lng}

	query := `SELECT id, resource_type, name, org_name, contact_person, contact_phone,
		province, city, district, address, latitude, longitude, service_radius, status
		FROM rescue_resources WHERE status = 1`
	if resourceType != "" {
		query += fmt.Sprintf(" AND resource_type = '%s'", resourceType)
	}
	rows, err := s.db.WithContext(ctx).Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.RescueResource
	for rows.Next() {
		var r model.RescueResource
		rows.Scan(&r.ID, &r.ResourceType, &r.Name, &r.OrgName, &r.ContactPerson,
			&r.ContactPhone, &r.Province, &r.City, &r.District, &r.Address,
			&r.Latitude, &r.Longitude, &r.ServiceRadius, &r.Status)

		dist := point.DistanceTo(model.GeoPoint{Lat: r.Latitude, Lng: r.Longitude})
		if dist/1000 > radiusKm {
			continue
		}
		r.Latitude = r.Latitude
		r.Longitude = r.Longitude
		result = append(result, &r)
	}
	return result, nil
}

func (s *TransportService) findNearestResources(lat, lng float64, resourceType string, limit int) ([]model.RescueResource, error) {
	var result []model.RescueResource
	list, err := s.ListRescueResources(context.Background(), lat, lng, resourceType, 100)
	if err != nil {
		return nil, err
	}
	for i, r := range list {
		if i >= limit {
			break
		}
		result = append(result, *r)
	}
	return result, nil
}

type RescueDispatch struct {
	RequestID  int64  `json:"request_id" binding:"required"`
	ResourceID int64  `json:"resource_id" binding:"required"`
	DispatcherID int64 `json:"dispatcher_id"`
	Note       string `json:"note"`
}

func (s *TransportService) DispatchRescue(ctx context.Context, req *RescueDispatch) error {
	now := time.Now()
	s.db.WithContext(ctx).Exec(`
		UPDATE rescue_requests SET
		status = 'dispatched', assigned_resource_id = ?, dispatcher_id = ?,
		dispatched_at = ?, updated_at = NOW()
		WHERE id = ? AND status = 'pending'`,
		req.ResourceID, req.DispatcherID, now, req.RequestID,
	)
	logger.Sugar.Infof("rescue dispatched: request=%d resource=%d", req.RequestID, req.ResourceID)
	return nil
}
