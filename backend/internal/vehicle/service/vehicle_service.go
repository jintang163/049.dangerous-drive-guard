package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
	"github.com/dangerous-drive-guard/backend/pkg/mq"
)

type VehicleService struct {
	db *gorm.DB
}

type DiagnosticData struct {
	EngineRPM      int
	VehicleSpeed   float64
	CoolantTemp    float64
	FuelLevel      float64
	OilPressure    float64
	BatteryVoltage float64
	TirePressureFL float64
	TirePressureFR float64
	TirePressureRL float64
	TirePressureRR float64
	FaultCodes     []string
	Latitude       float64
	Longitude      float64
	ReportTime     time.Time
}

func NewVehicleService() *VehicleService {
	return &VehicleService{db: database.GetDB()}
}

func (s *VehicleService) CreateVehicle(ctx context.Context, v *model.Vehicle) (*model.Vehicle, error) {
	var id int64
	result := s.db.WithContext(ctx).Exec(`
		INSERT INTO vehicles
		(plate_number, vehicle_type, brand, model, color, vin, engine_no,
		 load_weight, load_volume, length, width, height, max_speed, fuel_type,
		 status, org_id, driver_id, escort_id, device_id, mileage, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		v.PlateNumber, v.VehicleType, v.Brand, v.Model, v.Color, v.VIN, v.EngineNo,
		v.LoadWeight, v.LoadVolume, v.Length, v.Width, v.Height, v.MaxSpeed, v.FuelType,
		v.Status, v.OrgID, v.DriverID, v.EscortID, v.DeviceID, v.Mileage,
	)
	if result.Error != nil {
		return nil, result.Error
	}
	s.db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&id)
	v.ID = id
	return v, nil
}

func (s *VehicleService) GetVehicle(ctx context.Context, id int64) (*model.Vehicle, error) {
	var v model.Vehicle
	err := s.db.WithContext(ctx).First(&v, id).Error
	return &v, err
}

func (s *VehicleService) ListVehicles(ctx context.Context, orgID int64, status model.VehicleStatus, keyword string, page, pageSize int) ([]*model.Vehicle, int64, error) {
	var vehicles []*model.Vehicle
	var total int64
	query := s.db.WithContext(ctx).Table("vehicles")
	if orgID > 0 {
		query = query.Where("org_id = ?", orgID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("plate_number LIKE ? OR vin LIKE ? OR brand LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	offset := (page - 1) * pageSize
	rows, err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var v model.Vehicle
		rows.Scan(&v.ID, &v.PlateNumber, &v.VehicleType, &v.Brand, &v.Model,
			&v.Color, &v.VIN, &v.EngineNo, &v.LoadWeight, &v.LoadVolume,
			&v.Length, &v.Width, &v.Height, &v.MaxSpeed, &v.FuelType,
			&v.Status, &v.OrgID, &v.DriverID, &v.EscortID, &v.DeviceID,
			&v.Mileage, &v.CreatedAt, &v.UpdatedAt)
		vehicles = append(vehicles, &v)
	}
	return vehicles, total, nil
}

func (s *VehicleService) UpdateVehicle(ctx context.Context, v *model.Vehicle) (*model.Vehicle, error) {
	result := s.db.WithContext(ctx).Model(&model.Vehicle{}).Where("id = ?", v.ID).Updates(map[string]interface{}{
		"plate_number": v.PlateNumber,
		"vehicle_type": v.VehicleType,
		"status":       v.Status,
		"driver_id":    v.DriverID,
		"updated_at":   time.Now(),
	})
	if result.Error != nil {
		return nil, result.Error
	}
	return v, nil
}

func (s *VehicleService) DeleteVehicle(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Delete(&model.Vehicle{}, id).Error
}

func (s *VehicleService) UploadDiagnostics(ctx context.Context, vehicleID int64, data *DiagnosticData) error {
	obdJSON, _ := json.Marshal(map[string]interface{}{
		"engine_rpm":       data.EngineRPM,
		"vehicle_speed":    data.VehicleSpeed,
		"coolant_temp":     data.CoolantTemp,
		"fuel_level":       data.FuelLevel,
		"oil_pressure":     data.OilPressure,
		"battery_voltage":  data.BatteryVoltage,
		"tire_pressure_fl": data.TirePressureFL,
		"tire_pressure_fr": data.TirePressureFR,
		"tire_pressure_rl": data.TirePressureRL,
		"tire_pressure_rr": data.TirePressureRR,
		"fault_codes":      data.FaultCodes,
	})
	faultCodesJSON, _ := json.Marshal(data.FaultCodes)

	s.db.WithContext(ctx).Exec(`
		INSERT INTO vehicle_diagnostics
		(vehicle_id, obd_data, engine_rpm, vehicle_speed, coolant_temp, fuel_level,
		 oil_pressure, battery_voltage, tire_pressure_fl, tire_pressure_fr,
		 tire_pressure_rl, tire_pressure_rr, fault_codes, latitude, longitude, report_time, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		vehicleID, string(obdJSON), data.EngineRPM, data.VehicleSpeed,
		data.CoolantTemp, data.FuelLevel, data.OilPressure, data.BatteryVoltage,
		data.TirePressureFL, data.TirePressureFR, data.TirePressureRL, data.TirePressureRR,
		string(faultCodesJSON), data.Latitude, data.Longitude, data.ReportTime,
	)

	for _, code := range data.FaultCodes {
		s.processFaultCode(ctx, vehicleID, code)
	}

	mqBody, _ := json.Marshal(map[string]interface{}{
		"vehicle_id":    vehicleID,
		"engine_rpm":    data.EngineRPM,
		"speed":         data.VehicleSpeed,
		"fuel_level":    data.FuelLevel,
		"fault_codes":   data.FaultCodes,
		"time":          data.ReportTime,
	})
	_ = mq.Send(ctx, mq.Message{Topic: "vehicle_status", Key: fmt.Sprintf("%d", vehicleID), Body: mqBody})

	return nil
}

func (s *VehicleService) processFaultCode(ctx context.Context, vehicleID int64, code string) {
	level := 1
	desc := ""
	suggestion := ""
	if len(code) > 0 {
		switch code[0] {
		case 'P':
			level = 2
			desc = "动力总成故障: " + code
		case 'B':
			level = 2
			desc = "车身系统故障: " + code
		case 'C':
			level = 2
			desc = "底盘系统故障: " + code
		case 'U':
			level = 1
			desc = "网络通信故障: " + code
		}
		suggestion = "建议尽快前往维修站检查处理"
		if level == 2 {
			level = 3
			suggestion = "严重故障！请立即停车并联系救援"
		}
	}
	s.db.WithContext(ctx).Exec(`
		INSERT INTO vehicle_fault_alerts
		(vehicle_id, fault_code, fault_level, fault_desc, fault_suggestion,
		 first_report_time, last_report_time, report_count, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1, 0, NOW())
		ON DUPLICATE KEY UPDATE
		last_report_time = NOW(),
		report_count = report_count + 1,
		status = IF(status = 2, 2, 0)`,
		vehicleID, code, level, desc, suggestion, time.Now(), time.Now(),
	)
}

type FaultAlert struct {
	ID               int64     `json:"id"`
	VehicleID        int64     `json:"vehicle_id"`
	FaultCode        string    `json:"fault_code"`
	FaultLevel       int       `json:"fault_level"`
	FaultDesc        string    `json:"fault_desc"`
	FaultSuggestion  string    `json:"fault_suggestion"`
	LastReportTime   time.Time `json:"last_report_time"`
	ReportCount      int       `json:"report_count"`
	Status           int       `json:"status"`
}

func (s *VehicleService) GetRecentDiagnostics(ctx context.Context, vehicleID int64, limit int) ([]map[string]interface{}, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id, vehicle_id, engine_rpm, vehicle_speed, coolant_temp, fuel_level,
		       oil_pressure, battery_voltage, tire_pressure_fl, tire_pressure_fr,
		       tire_pressure_rl, tire_pressure_rr, latitude, longitude, report_time
		FROM vehicle_diagnostics
		WHERE vehicle_id = ? ORDER BY report_time DESC LIMIT ?`,
		vehicleID, limit,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []map[string]interface{}
	for rows.Next() {
		item := make(map[string]interface{})
		var id, vid, rpm int
		var speed, coolant, fuel, oil, batt, tfl, tfr, trl, trr, lat, lng float64
		var rptTime time.Time
		rows.Scan(&id, &vid, &rpm, &speed, &coolant, &fuel, &oil, &batt,
			&tfl, &tfr, &trl, &trr, &lat, &lng, &rptTime)
		item["id"] = id
		item["vehicle_id"] = vid
		item["engine_rpm"] = rpm
		item["vehicle_speed"] = speed
		item["coolant_temp"] = coolant
		item["fuel_level"] = fuel
		item["oil_pressure"] = oil
		item["battery_voltage"] = batt
		item["tire_pressure"] = map[string]float64{"fl": tfl, "fr": tfr, "rl": trl, "rr": trr}
		item["latitude"] = lat
		item["longitude"] = lng
		item["report_time"] = rptTime
		result = append(result, item)
	}
	return result, nil
}

func (s *VehicleService) GetFaultAlerts(ctx context.Context, vehicleID int64) ([]*FaultAlert, error) {
	var alerts []*FaultAlert
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id, vehicle_id, fault_code, fault_level, fault_desc, fault_suggestion,
		       last_report_time, report_count, status
		FROM vehicle_fault_alerts
		WHERE vehicle_id = ? AND status IN (0, 1)
		ORDER BY fault_level DESC, last_report_time DESC`,
		vehicleID,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var a FaultAlert
		rows.Scan(&a.ID, &a.VehicleID, &a.FaultCode, &a.FaultLevel,
			&a.FaultDesc, &a.FaultSuggestion, &a.LastReportTime,
			&a.ReportCount, &a.Status)
		alerts = append(alerts, &a)
	}
	return alerts, nil
}

type ScoreStat struct {
	AvgScore       float64 `json:"avg_score"`
	TotalScore     float64 `json:"total_score"`
	TripCount      int     `json:"trip_count"`
	FatigueDeduct  float64 `json:"fatigue_deduct"`
	SpeedDeduct    float64 `json:"speed_deduct"`
	PhoneDeduct    float64 `json:"phone_deduct"`
	SmokingDeduct  float64 `json:"smoking_deduct"`
	SeatbeltDeduct float64 `json:"seatbelt_deduct"`
	FatigueAlarmCount int `json:"fatigue_alarm_count"`
}

func (s *VehicleService) GetDriverScore(ctx context.Context, driverID int64, startDate, endDate string) ([]model.DrivingScore, *ScoreStat, error) {
	var scores []model.DrivingScore
	query := s.db.WithContext(ctx).Table("driving_scores").Where("driver_id = ?", driverID)
	if startDate != "" {
		query = query.Where("trip_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("trip_date <= ?", endDate)
	}
	rows, err := query.Order("trip_date DESC").Limit(90).Rows()
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	stat := &ScoreStat{}
	for rows.Next() {
		var sc model.DrivingScore
		rows.Scan(&sc.ID, &sc.DriverID, &sc.WaybillID, &sc.VehicleID, nil,
			&sc.TotalScore, &sc.ScoreLevel, nil, &sc.FatigueDeduction,
			&sc.OverspeedCount, &sc.OverspeedDeduction,
			&sc.SuddenBrakeCount, &sc.SuddenBrakeDeduction,
			&sc.SuddenAccelCount, &sc.SuddenAccelDeduction,
			&sc.SharpTurnCount, &sc.SharpTurnDeduction,
			&sc.LaneDeviationCount, &sc.LaneDeviationDeduction,
			&sc.PhoneUsageCount, &sc.PhoneUsageDeduction,
			&sc.SmokingCount, &sc.SmokingDeduction,
			&sc.SeatbeltViolationCount, &sc.SeatbeltViolationDeduction,
			&sc.RouteDeviationCount, &sc.RouteDeviationDeduction,
			&sc.FatigueAlarmCount, &sc.TotalDistance, &sc.DrivingDuration,
			&sc.NightDrivingDuration, nil, nil)
		scores = append(scores, sc)
		stat.TotalScore += sc.TotalScore
		stat.FatigueDeduct += sc.FatigueDeduction
		stat.SpeedDeduct += sc.OverspeedDeduction
		stat.PhoneDeduct += sc.PhoneUsageDeduction
		stat.SmokingDeduct += sc.SmokingDeduction
		stat.SeatbeltDeduct += sc.SeatbeltViolationDeduction
		stat.FatigueAlarmCount += sc.FatigueAlarmCount
		stat.TripCount++
	}
	if stat.TripCount > 0 {
		stat.AvgScore = stat.TotalScore / float64(stat.TripCount)
	}
	return scores, stat, nil
}

type RankingItem struct {
	Rank          int     `json:"rank"`
	DriverID      int64   `json:"driver_id"`
	DriverName    string  `json:"driver_name"`
	OrgID         int64   `json:"org_id"`
	AvgScore      float64 `json:"avg_score"`
	ScoreLevel    string  `json:"score_level"`
	TripCount     int     `json:"trip_count"`
	TotalDistance float64 `json:"total_distance"`
	AlarmCount    int     `json:"alarm_count"`
}

func (s *VehicleService) GetScoreRanking(ctx context.Context, orgID int64, period string, limit int) ([]*RankingItem, error) {
	dateCond := "DATE_SUB(CURDATE(), INTERVAL 7 DAY)"
	switch period {
	case "month":
		dateCond = "DATE_SUB(CURDATE(), INTERVAL 30 DAY)"
	case "quarter":
		dateCond = "DATE_SUB(CURDATE(), INTERVAL 90 DAY)"
	}
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT s.driver_id, u.real_name, u.org_id,
		       AVG(s.total_score) as avg_score,
		       CASE WHEN AVG(s.total_score) >= 95 THEN 'excellent'
		            WHEN AVG(s.total_score) >= 85 THEN 'good'
		            WHEN AVG(s.total_score) >= 70 THEN 'normal'
		            WHEN AVG(s.total_score) >= 60 THEN 'poor'
		            ELSE 'danger' END as score_level,
		       COUNT(*) as trip_count,
		       COALESCE(SUM(s.total_distance), 0) as total_distance,
		       COALESCE(SUM(s.fatigue_alarm_count), 0) as alarm_count
		FROM driving_scores s
		LEFT JOIN users u ON u.id = s.driver_id
		WHERE s.trip_date >= `+dateCond+`
		`+func() string {
		if orgID > 0 {
			return fmt.Sprintf(" AND u.org_id = %d", orgID)
		}
		return ""
	}()+`
		GROUP BY s.driver_id, u.real_name, u.org_id
		ORDER BY avg_score DESC
		LIMIT ?`, limit,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*RankingItem
	rank := 0
	for rows.Next() {
		rank++
		var r RankingItem
		r.Rank = rank
		rows.Scan(&r.DriverID, &r.DriverName, &r.OrgID,
			&r.AvgScore, &r.ScoreLevel, &r.TripCount,
			&r.TotalDistance, &r.AlarmCount)
		result = append(result, &r)
	}
	return result, nil
}

func (s *VehicleService) GetRealtimeStatus(ctx context.Context, vehicleID int64) (*model.RealtimeVehicleStatus, error) {
	var v model.Vehicle
	err := s.db.WithContext(ctx).Where("id = ?", vehicleID).First(&v).Error
	if err != nil {
		return nil, err
	}

	var driverName string
	if v.DriverID > 0 {
		var driver model.User
		s.db.WithContext(ctx).Select("real_name").Where("id = ?", v.DriverID).First(&driver)
		driverName = driver.RealName
	}

	status := &model.RealtimeVehicleStatus{
		VehicleID:     v.ID,
		PlateNumber:   v.PlateNumber,
		VehicleType:   v.VehicleType,
		Status:        v.Status,
		DriverID:      v.DriverID,
		DriverName:    driverName,
		LastUpdateTime: time.Now(),
	}

	var latestTrack model.VehicleTrack
	s.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID).Order("gps_time DESC").First(&latestTrack)
	if latestTrack.ID > 0 {
		status.Latitude = latestTrack.Latitude
		status.Longitude = latestTrack.Longitude
		status.Speed = latestTrack.Speed
		status.Direction = latestTrack.Direction
		status.GPSTime = latestTrack.GPSTime
	}

	var latestDiag struct {
		FuelLevel      float64
		EngineRPM      int
		TirePressureOK bool
	}
	s.db.WithContext(ctx).Table("vehicle_diagnostics").
		Select("fuel_level, engine_rpm").
		Where("vehicle_id = ?", vehicleID).
		Order("report_time DESC").
		Limit(1).
		Scan(&latestDiag)
	status.FuelLevel = latestDiag.FuelLevel
	if latestDiag.EngineRPM > 0 {
		status.EngineStatus = "running"
	} else {
		status.EngineStatus = "stopped"
	}
	status.TirePressureOK = true

	var alertCount int64
	s.db.WithContext(ctx).Table("fatigue_alarms").
		Where("vehicle_id = ? AND status IN (?, ?)", vehicleID, "pending", "processing").
		Count(&alertCount)
	status.AlertCount = int(alertCount)

	return status, nil
}

func (s *VehicleService) ListRealtimeStatus(ctx context.Context, orgID int64, status string, page, pageSize int) ([]*model.RealtimeVehicleStatus, int64, error) {
	var vehicles []*model.Vehicle
	var total int64

	query := s.db.WithContext(ctx).Model(&model.Vehicle{})
	if orgID > 0 {
		query = query.Where("org_id = ?", orgID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&vehicles).Error
	if err != nil {
		return nil, 0, err
	}

	var result []*model.RealtimeVehicleStatus
	for _, v := range vehicles {
		rtStatus, _ := s.GetRealtimeStatus(ctx, v.ID)
		if rtStatus != nil {
			result = append(result, rtStatus)
		}
	}

	return result, total, nil
}
