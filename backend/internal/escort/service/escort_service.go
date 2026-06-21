package service

import (
	"context"
	"encoding/json"
	"fmt"
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

type EscortService struct {
	db  *database.TIDB
	mu  sync.RWMutex
}

func NewEscortService() *EscortService {
	return &EscortService{
		db: database.GetDB(),
	}
}

func (s *EscortService) generateNo(prefix string) string {
	return fmt.Sprintf("%s%s", prefix, strings.ToUpper(strings.ReplaceAll(uuid.New().String()[:12], "-", "")))
}

func (s *EscortService) CreateShift(ctx context.Context, req *model.EscortShiftCreateRequest, dispatcherID int64, dispatcherName string) (*model.EscortShift, error) {
	shiftNo := s.generateNo("ES")

	var escortName string
	s.db.Raw("SELECT real_name FROM users WHERE id = ?", req.EscortID).Scan(&escortName)

	vehicleIDsStr := int64SliceToString(req.VehicleIDs)
	waybillIDsStr := int64SliceToString(req.WaybillIDs)

	now := time.Now()
	maxConcurrent := req.MaxConcurrent
	if maxConcurrent == 0 {
		maxConcurrent = 5
	}
	pollingInterval := req.PollingInterval
	if pollingInterval == 0 {
		pollingInterval = 30
	}

	result := s.db.WithContext(ctx).Exec(`
		INSERT INTO escort_shifts
		(shift_no, escort_id, escort_name, dispatcher_id, dispatcher_name, vehicle_ids, waybill_ids,
		 scheduled_start, scheduled_end, status, remark, max_concurrent, polling_interval, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'scheduled', ?, ?, ?, NOW(), NOW())`,
		shiftNo, req.EscortID, escortName, dispatcherID, dispatcherName,
		vehicleIDsStr, waybillIDsStr,
		req.ScheduledStart, req.ScheduledEnd,
		req.Remark, maxConcurrent, pollingInterval,
	)
	if result.Error != nil {
		logger.Sugar.Errorf("create escort shift error: %v", result.Error)
		return nil, result.Error
	}

	var shiftID int64
	s.db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&shiftID)

	for _, vid := range req.VehicleIDs {
		var plateNumber string
		s.db.Raw("SELECT plate_number FROM vehicles WHERE id = ?", vid).Scan(&plateNumber)
		s.db.WithContext(ctx).Exec(`
			INSERT INTO escort_vehicle_assignments
			(shift_id, vehicle_id, plate_number, priority, assigned_by, assigned_at, is_active, created_at, updated_at)
			VALUES (?, ?, ?, 1, ?, ?, 1, NOW(), NOW())`,
			shiftID, vid, plateNumber, dispatcherID, now,
		)
	}

	shift := &model.EscortShift{
		ShiftNo:         shiftNo,
		EscortID:        req.EscortID,
		EscortName:      escortName,
		DispatcherID:    dispatcherID,
		DispatcherName:  dispatcherName,
		VehicleIDs:      vehicleIDsStr,
		WaybillIDs:      waybillIDsStr,
		ScheduledStart:  req.ScheduledStart,
		ScheduledEnd:    req.ScheduledEnd,
		Status:          model.EscortShiftScheduled,
		Remark:          req.Remark,
		MaxConcurrent:   maxConcurrent,
		PollingInterval: pollingInterval,
	}
	shift.ID = shiftID
	shift.CreatedAt = now
	shift.UpdatedAt = now

	logger.Global.Info("escort shift created",
		zap.String("shift_no", shiftNo),
		zap.Int64("escort_id", req.EscortID))

	return shift, nil
}

func (s *EscortService) GetShift(ctx context.Context, id int64) (*model.EscortShift, error) {
	var shift model.EscortShift
	err := s.db.WithContext(ctx).Table("escort_shifts").Where("id = ?", id).First(&shift).Error
	if err != nil {
		return nil, err
	}
	return &shift, nil
}

func (s *EscortService) ListShifts(ctx context.Context, escortID, dispatcherID int64, status model.EscortShiftStatus, page, pageSize int) ([]*model.EscortShift, int64, error) {
	var shifts []*model.EscortShift
	var total int64

	query := s.db.WithContext(ctx).Table("escort_shifts")
	if escortID > 0 {
		query = query.Where("escort_id = ?", escortID)
	}
	if dispatcherID > 0 {
		query = query.Where("dispatcher_id = ?", dispatcherID)
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
		var sh model.EscortShift
		rows.Scan(&sh.ID, &sh.ShiftNo, &sh.EscortID, &sh.EscortName,
			&sh.DispatcherID, &sh.DispatcherName, &sh.VehicleIDs, &sh.WaybillIDs,
			&sh.ScheduledStart, &sh.ScheduledEnd, &sh.ActualStart, &sh.ActualEnd,
			&sh.Status, &sh.Remark, &sh.MaxConcurrent, &sh.PollingInterval,
			&sh.CreatedAt, &sh.UpdatedAt)
		shifts = append(shifts, &sh)
	}
	return shifts, total, nil
}

func (s *EscortService) UpdateShiftStatus(ctx context.Context, id int64, status model.EscortShiftStatus, operatorID int64) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": now,
	}
	switch status {
	case model.EscortShiftActive:
		updates["actual_start"] = now
	case model.EscortShiftCompleted:
		updates["actual_end"] = now
	}
	result := s.db.WithContext(ctx).Table("escort_shifts").Where("id = ?", id).Updates(updates)
	return result.Error
}

func (s *EscortService) AssignVehicles(ctx context.Context, shiftID int64, vehicleIDs []int64, dispatcherID int64) error {
	now := time.Now()
	for _, vid := range vehicleIDs {
		var plateNumber string
		s.db.Raw("SELECT plate_number FROM vehicles WHERE id = ?", vid).Scan(&plateNumber)
		s.db.WithContext(ctx).Exec(`
			INSERT INTO escort_vehicle_assignments
			(shift_id, vehicle_id, plate_number, priority, assigned_by, assigned_at, is_active, created_at, updated_at)
			VALUES (?, ?, ?, 1, ?, ?, 1, NOW(), NOW())
			ON DUPLICATE KEY UPDATE is_active = 1, updated_at = NOW()`,
			shiftID, vid, plateNumber, dispatcherID, now,
		)
	}
	return nil
}

func (s *EscortService) GetShiftAssignments(ctx context.Context, shiftID int64) ([]map[string]interface{}, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT eva.id, eva.shift_id, eva.vehicle_id, eva.plate_number, eva.waybill_id, eva.waybill_no,
		       eva.priority, eva.is_active,
		       v.status as vehicle_status,
		       u.real_name as driver_name,
		       w.status as waybill_status, w.goods_name
		FROM escort_vehicle_assignments eva
		LEFT JOIN vehicles v ON v.id = eva.vehicle_id
		LEFT JOIN users u ON u.id = v.driver_id
		LEFT JOIN waybills w ON w.id = eva.waybill_id
		WHERE eva.shift_id = ? AND eva.is_active = 1
		ORDER BY eva.priority DESC, eva.id ASC`, shiftID,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		item := make(map[string]interface{})
		var id, shiftID, vehicleID, waybillID, priority int64
		var plateNumber, waybillNo, vehicleStatus, driverName, waybillStatus, goodsName string
		var isActive bool
		rows.Scan(&id, &shiftID, &vehicleID, &plateNumber, &waybillID, &waybillNo,
			&priority, &isActive, &vehicleStatus, &driverName, &waybillStatus, &goodsName)
		item["id"] = id
		item["shift_id"] = shiftID
		item["vehicle_id"] = vehicleID
		item["plate_number"] = plateNumber
		item["waybill_id"] = waybillID
		item["waybill_no"] = waybillNo
		item["priority"] = priority
		item["is_active"] = isActive
		item["vehicle_status"] = vehicleStatus
		item["driver_name"] = driverName
		item["waybill_status"] = waybillStatus
		item["goods_name"] = goodsName
		result = append(result, item)
	}
	return result, nil
}

func (s *EscortService) ReportSOS(ctx context.Context, req *model.EscortSOSReportRequest) (*model.EscortSOSAlert, error) {
	alertNo := s.generateNo("SOS")

	var plateNumber, driverName, waybillNo string
	var escortID int64
	s.db.Raw("SELECT plate_number FROM vehicles WHERE id = ?", req.VehicleID).Scan(&plateNumber)
	if req.DriverID > 0 {
		s.db.Raw("SELECT real_name FROM users WHERE id = ?", req.DriverID).Scan(&driverName)
	}
	if req.WaybillID > 0 {
		s.db.Raw("SELECT waybill_no, escort_id FROM waybills WHERE id = ?", req.WaybillID).Scan(&waybillNo, &escortID)
	}

	alertLevel := req.AlertLevel
	if alertLevel == 0 {
		switch req.AlertType {
		case "emergency_button", "fire", "leak", "accident":
			alertLevel = 3
		default:
			alertLevel = 2
		}
	}

	var id int64
	now := time.Now()
	result := s.db.WithContext(ctx).Exec(`
		INSERT INTO escort_sos_alerts
		(alert_no, vehicle_id, plate_number, driver_id, driver_name, waybill_id, waybill_no,
		 alert_type, alert_level, latitude, longitude, address, description, snapshot_url,
		 status, escort_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending', ?, ?, ?)`,
		alertNo, req.VehicleID, plateNumber, req.DriverID, driverName,
		req.WaybillID, waybillNo, req.AlertType, alertLevel,
		req.Latitude, req.Longitude, req.Address, req.Description, req.SnapshotURL,
		escortID, now, now,
	)
	if result.Error != nil {
		logger.Sugar.Errorf("report SOS error: %v", result.Error)
		return nil, result.Error
	}
	s.db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&id)

	alert := &model.EscortSOSAlert{
		ID:          id,
		AlertNo:     alertNo,
		VehicleID:   req.VehicleID,
		PlateNumber: plateNumber,
		DriverID:    req.DriverID,
		DriverName:  driverName,
		WaybillID:   req.WaybillID,
		WaybillNo:   waybillNo,
		AlertType:   req.AlertType,
		AlertLevel:  alertLevel,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Address:     req.Address,
		Description: req.Description,
		SnapshotURL: req.SnapshotURL,
		Status:      model.EscortSOSPending,
		EscortID:    escortID,
	}
	alert.CreatedAt = now

	s.db.WithContext(ctx).Exec(`
		INSERT INTO escort_video_records
		(record_no, vehicle_id, plate_number, waybill_id, waybill_no, record_type,
		 video_url, snapshot_url, start_time, latitude, longitude, trigger_reason, expire_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'alarm', ?, ?, ?, ?, ?, ?, DATE_ADD(NOW(), INTERVAL 90 DAY), NOW(), NOW())`,
		s.generateNo("VR"), req.VehicleID, plateNumber, req.WaybillID, waybillNo,
		fmt.Sprintf("/videos/escort/sos/%s.mp4", alertNo), req.SnapshotURL,
		now, req.Latitude, req.Longitude,
		fmt.Sprintf("SOS报警自动录制: %s", req.AlertType),
	)

	hub := ws.GetHub()
	alertData := map[string]interface{}{
		"id":           id,
		"alert_no":     alertNo,
		"vehicle_id":   req.VehicleID,
		"plate_number": plateNumber,
		"driver_name":  driverName,
		"alert_type":   req.AlertType,
		"alert_level":  alertLevel,
		"latitude":     req.Latitude,
		"longitude":    req.Longitude,
		"address":      req.Address,
		"description":  req.Description,
		"snapshot_url": req.SnapshotURL,
		"timestamp":    now.Format("2006-01-02 15:04:05"),
		"popup":        true,
	}
	alertMsg := &ws.WSMessage{
		Type:      ws.MsgSOSAlert,
		Timestamp: now.Unix(),
		Data:      alertData,
	}
	data, _ := json.Marshal(alertMsg)
	_ = data
	hub.BroadcastSOS(ctx, alert)

	recordNo := s.generateNo("VR")
	s.db.WithContext(ctx).Exec(`
		INSERT INTO escort_video_records
		(record_no, vehicle_id, plate_number, waybill_id, waybill_no, record_type,
		 video_url, snapshot_url, start_time, duration, latitude, longitude, trigger_reason, expire_at, created_at)
		VALUES (?, ?, ?, ?, ?, 'alarm', ?, ?, ?, 60, ?, ?, ?, DATE_ADD(NOW(), INTERVAL 90 DAY), NOW())`,
		recordNo, req.VehicleID, plateNumber, req.WaybillID, waybillNo,
		fmt.Sprintf("/videos/escort/%s.mp4", recordNo),
		req.SnapshotURL, now, req.Latitude, req.Longitude,
		fmt.Sprintf("SOS报警触发: %s", req.AlertType),
	)

	logger.Global.Warn("SOS alert reported",
		zap.String("alert_no", alertNo),
		zap.String("alert_type", req.AlertType),
		zap.Int("alert_level", alertLevel),
		zap.Int64("vehicle_id", req.VehicleID))

	return alert, nil
}

func (s *EscortService) GetSOSAlerts(ctx context.Context, vehicleID, escortID int64, status model.EscortSOSStatus, page, pageSize int) ([]*model.EscortSOSAlert, int64, error) {
	var alerts []*model.EscortSOSAlert
	var total int64

	query := s.db.WithContext(ctx).Table("escort_sos_alerts")
	if vehicleID > 0 {
		query = query.Where("vehicle_id = ?", vehicleID)
	}
	if escortID > 0 {
		query = query.Where("escort_id = ?", escortID)
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
		var a model.EscortSOSAlert
		rows.Scan(&a.ID, &a.AlertNo, &a.VehicleID, &a.PlateNumber, &a.DriverID,
			&a.DriverName, &a.WaybillID, &a.WaybillNo, &a.AlertType, &a.AlertLevel,
			&a.Latitude, &a.Longitude, &a.Address, &a.Description, &a.SnapshotURL,
			&a.VideoClipURL, &a.Status, &a.HandledBy, &a.HandlerName, &a.HandledAt,
			&a.HandleNote, &a.HandleType, &a.Notified, &a.PopupDisplayed, &a.AckedAt,
			&a.EscortID, &a.CreatedAt, &a.UpdatedAt)
		alerts = append(alerts, &a)
	}
	return alerts, total, nil
}

func (s *EscortService) HandleSOS(ctx context.Context, req *model.EscortSOSHandleRequest, handlerID int64, handlerName string) error {
	now := time.Now()
	result := s.db.WithContext(ctx).Exec(`
		UPDATE escort_sos_alerts SET
		status = 'processing', handled_by = ?, handler_name = ?, handled_at = ?,
		handle_note = ?, handle_type = ?, popup_displayed = 1, updated_at = ?
		WHERE id = ? AND status = 'pending'`,
		handlerID, handlerName, now, req.HandleNote, req.HandleType, now, req.AlertID,
	)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		s.db.WithContext(ctx).Exec(`
			UPDATE escort_sos_alerts SET
			handled_by = ?, handler_name = ?, handle_note = ?, handle_type = ?, updated_at = ?
			WHERE id = ?`,
			handlerID, handlerName, req.HandleNote, req.HandleType, now, req.AlertID,
		)
	}
	logger.Sugar.Infof("SOS alert handled: id=%d handler=%s", req.AlertID, handlerName)
	return nil
}

func (s *EscortService) ResolveSOS(ctx context.Context, alertID int64, handlerID int64, handlerName, note string) error {
	now := time.Now()
	s.db.WithContext(ctx).Exec(`
		UPDATE escort_sos_alerts SET
		status = 'resolved', handled_by = ?, handler_name = ?, handled_at = ?,
		handle_note = ?, acked_at = ?, updated_at = ?
		WHERE id = ?`,
		handlerID, handlerName, now, note, now, now, alertID,
	)
	logger.Sugar.Infof("SOS alert resolved: id=%d", alertID)
	return nil
}

func (s *EscortService) GetTrackPlayback(ctx context.Context, waybillID, vehicleID int64, startTime, endTime *time.Time) ([]*model.VehicleTrack, error) {
	var tracks []*model.VehicleTrack

	query := s.db.WithContext(ctx).Table("vehicle_tracks")
	if waybillID > 0 {
		query = query.Where("waybill_id = ?", waybillID)
	}
	if vehicleID > 0 {
		query = query.Where("vehicle_id = ?", vehicleID)
	}
	if startTime != nil {
		query = query.Where("gps_time >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("gps_time <= ?", endTime)
	}

	rows, err := query.Order("gps_time ASC").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t model.VehicleTrack
		rows.Scan(&t.ID, &t.VehicleID, &t.WaybillID, &t.DriverID,
			&t.Latitude, &t.Longitude, &t.Altitude, &t.Speed, &t.Direction,
			&t.SatelliteCount, &t.Hdop, &t.Accuracy, &t.GPSTime, &t.CreatedAt)
		tracks = append(tracks, &t)
	}
	return tracks, nil
}

func (s *EscortService) GetVideoRecords(ctx context.Context, vehicleID, waybillID int64, recordType string, startTime, endTime *time.Time, page, pageSize int) ([]*model.EscortVideoRecord, int64, error) {
	var records []*model.EscortVideoRecord
	var total int64

	query := s.db.WithContext(ctx).Table("escort_video_records")
	if vehicleID > 0 {
		query = query.Where("vehicle_id = ?", vehicleID)
	}
	if waybillID > 0 {
		query = query.Where("waybill_id = ?", waybillID)
	}
	if recordType != "" {
		query = query.Where("record_type = ?", recordType)
	}
	if startTime != nil {
		query = query.Where("start_time >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("start_time <= ?", endTime)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	rows, err := query.Order("start_time DESC").Offset(offset).Limit(pageSize).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var r model.EscortVideoRecord
		rows.Scan(&r.ID, &r.RecordNo, &r.VehicleID, &r.PlateNumber, &r.WaybillID,
			&r.WaybillNo, &r.RecordType, &r.VideoURL, &r.SnapshotURL, &r.StartTime,
			&r.EndTime, &r.Duration, &r.Latitude, &r.Longitude, &r.TriggerReason,
			&r.AlertID, &r.ViewedCount, &r.ExpireAt, &r.CreatedAt, &r.UpdatedAt)
		records = append(records, &r)
	}
	return records, total, nil
}

func (s *EscortService) IncrementVideoViewCount(ctx context.Context, recordID int64) error {
	s.db.WithContext(ctx).Exec(`
		UPDATE escort_video_records SET viewed_count = viewed_count + 1 WHERE id = ?`, recordID,
	)
	return nil
}

func (s *EscortService) SendIntercom(ctx context.Context, req *model.EscortIntercomRequest, senderID int64, senderName, senderRole string) error {
	var plateNumber string
	s.db.Raw("SELECT plate_number FROM vehicles WHERE id = ?", req.VehicleID).Scan(&plateNumber)

	priority := req.Priority
	if priority == 0 {
		priority = 1
	}
	messageType := req.MessageType
	if messageType == "" {
		messageType = "text"
	}

	now := time.Now()
	s.db.WithContext(ctx).Exec(`
		INSERT INTO escort_intercom_logs
		(vehicle_id, plate_number, sender_id, sender_name, sender_role, message_type,
		 content, audio_url, priority, delivered, delivered_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		req.VehicleID, plateNumber, senderID, senderName, senderRole, messageType,
		req.Content, req.AudioURL, priority, now, now,
	)

	hub := ws.GetHub()
	hub.SendIntercomToVehicle(req.VehicleID, req.Content, priority)

	logger.Sugar.Infof("intercom sent to vehicle %d: %s", req.VehicleID, req.Content)
	return nil
}

func (s *EscortService) GetIntercomLogs(ctx context.Context, vehicleID int64, page, pageSize int) ([]*model.EscortIntercomLog, int64, error) {
	var logs []*model.EscortIntercomLog
	var total int64

	query := s.db.WithContext(ctx).Table("escort_intercom_logs")
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

	for rows.Next() {
		var l model.EscortIntercomLog
		rows.Scan(&l.ID, &l.VehicleID, &l.PlateNumber, &l.SenderID, &l.SenderName,
			&l.SenderRole, &l.MessageType, &l.Content, &l.AudioURL, &l.Priority,
			&l.Delivered, &l.DeliveredAt, &l.Acked, &l.AckedAt, &l.CreatedAt)
		logs = append(logs, &l)
	}
	return logs, total, nil
}

func (s *EscortService) StartPollingSession(ctx context.Context, escortID, shiftID int64, escortName string) (*model.EscortPollingSession, error) {
	sessionNo := s.generateNo("PS")
	now := time.Now()

	var vehiclesJSON string
	var vehicleIDs []int64
	if shiftID > 0 {
		rows, _ := s.db.Raw(`
			SELECT vehicle_id FROM escort_vehicle_assignments
			WHERE shift_id = ? AND is_active = 1`, shiftID,
		).Rows()
		for rows.Next() {
			var vid int64
			rows.Scan(&vid)
			vehicleIDs = append(vehicleIDs, vid)
		}
		rows.Close()
	}
	vehiclesBytes, _ := json.Marshal(vehicleIDs)
	vehiclesJSON = string(vehiclesBytes)

	s.db.WithContext(ctx).Exec(`
		INSERT INTO escort_polling_sessions
		(session_no, escort_id, escort_name, shift_id, start_time, vehicles, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 'active', NOW(), NOW())`,
		sessionNo, escortID, escortName, shiftID, now, vehiclesJSON,
	)

	var id int64
	s.db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&id)

	session := &model.EscortPollingSession{
		ID:         id,
		SessionNo:  sessionNo,
		EscortID:   escortID,
		EscortName: escortName,
		ShiftID:    shiftID,
		StartTime:  &now,
		Vehicles:   vehiclesJSON,
		Status:     "active",
	}

	logger.Sugar.Infof("polling session started: %s", sessionNo)
	return session, nil
}

func (s *EscortService) EndPollingSession(ctx context.Context, sessionID int64, pollingCount int) error {
	now := time.Now()
	s.db.WithContext(ctx).Exec(`
		UPDATE escort_polling_sessions SET
		end_time = ?, polling_count = ?, status = 'completed', updated_at = ?
		WHERE id = ?`,
		now, pollingCount, now, sessionID,
	)
	return nil
}

func (s *EscortService) GetEscortVehiclesForPolling(ctx context.Context, escortID int64) ([]*model.RealtimeVehicleStatus, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT DISTINCT eva.vehicle_id, v.plate_number, v.vehicle_type, v.status, v.driver_id,
		       u.real_name as driver_name, w.id as waybill_id, w.waybill_no,
		       f.latitude, f.longitude, f.vehicle_speed, f.fatigue_score, f.fatigue_level, f.detection_time
		FROM escort_vehicle_assignments eva
		INNER JOIN escort_shifts es ON es.id = eva.shift_id
		LEFT JOIN vehicles v ON v.id = eva.vehicle_id
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
		WHERE es.escort_id = ? AND es.status IN ('scheduled', 'active') AND eva.is_active = 1`,
		escortID,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.RealtimeVehicleStatus
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
		}
		result = append(result, &vs)
	}
	return result, nil
}

func (s *EscortService) GetStatistics(ctx context.Context, orgID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalShifts int64
	s.db.Table("escort_shifts").Count(&totalShifts)
	stats["total_shifts"] = totalShifts

	var activeShifts int64
	s.db.Table("escort_shifts").Where("status = 'active'").Count(&activeShifts)
	stats["active_shifts"] = activeShifts

	var pendingSOS int64
	s.db.Table("escort_sos_alerts").Where("status = 'pending'").Count(&pendingSOS)
	stats["pending_sos"] = pendingSOS

	var todaySOS int64
	s.db.Table("escort_sos_alerts").Where("DATE(created_at) = CURDATE()").Count(&todaySOS)
	stats["today_sos"] = todaySOS

	var videoToday int64
	s.db.Table("escort_video_records").Where("DATE(created_at) = CURDATE()").Count(&videoToday)
	stats["today_video_records"] = videoToday

	var intercomToday int64
	s.db.Table("escort_intercom_logs").Where("DATE(created_at) = CURDATE()").Count(&intercomToday)
	stats["today_intercom"] = intercomToday

	type SOSLevelCount struct {
		Level int64 `json:"level"`
		Count int64 `json:"count"`
	}
	var sosLevelStats []SOSLevelCount
	s.db.Table("escort_sos_alerts").
		Select("alert_level as level, COUNT(*) as count").
		Where("DATE(created_at) >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)").
		Group("alert_level").
		Scan(&sosLevelStats)
	stats["sos_level_distribution"] = sosLevelStats

	type SOSTypeCount struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}
	var sosTypeStats []SOSTypeCount
	s.db.Table("escort_sos_alerts").
		Select("alert_type as type, COUNT(*) as count").
		Where("DATE(created_at) >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)").
		Group("alert_type").
		Scan(&sosTypeStats)
	stats["sos_type_distribution"] = sosTypeStats

	return stats, nil
}

func int64SliceToString(ids []int64) string {
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = fmt.Sprintf("%d", id)
	}
	return strings.Join(strs, ",")
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
