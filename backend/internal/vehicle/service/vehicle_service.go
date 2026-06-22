package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	monitorWs "github.com/dangerous-drive-guard/backend/internal/monitor/delivery/ws"
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
	BrakeTempFL    float64
	BrakeTempFR    float64
	BrakeTempRL    float64
	BrakeTempRR    float64
	TireTempFL     float64
	TireTempFR     float64
	TireTempRL     float64
	TireTempRR     float64
	BrakePadWearFR float64
	BrakePadWearRL float64
	BrakePadWearRR float64
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
		"brake_temp_fl":    data.BrakeTempFL,
		"brake_temp_fr":    data.BrakeTempFR,
		"brake_temp_rl":    data.BrakeTempRL,
		"brake_temp_rr":    data.BrakeTempRR,
		"tire_temp_fl":     data.TireTempFL,
		"tire_temp_fr":     data.TireTempFR,
		"tire_temp_rl":     data.TireTempRL,
		"tire_temp_rr":     data.TireTempRR,
		"brake_pad_wear_fr": data.BrakePadWearFR,
		"brake_pad_wear_rl": data.BrakePadWearRL,
		"brake_pad_wear_rr": data.BrakePadWearRR,
		"fault_codes":      data.FaultCodes,
	})
	faultCodesJSON, _ := json.Marshal(data.FaultCodes)

	s.db.WithContext(ctx).Exec(`
		INSERT INTO vehicle_diagnostics
		(vehicle_id, obd_data, engine_rpm, vehicle_speed, coolant_temp, fuel_level,
		 oil_pressure, battery_voltage, tire_pressure_fl, tire_pressure_fr,
		 tire_pressure_rl, tire_pressure_rr, brake_temp_fl, brake_temp_fr,
		 brake_temp_rl, brake_temp_rr, tire_temp_fl, tire_temp_fr,
		 tire_temp_rl, tire_temp_rr, brake_pad_wear_fr, brake_pad_wear_rl,
		 brake_pad_wear_rr, fault_codes, latitude, longitude, report_time, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		vehicleID, string(obdJSON), data.EngineRPM, data.VehicleSpeed,
		data.CoolantTemp, data.FuelLevel, data.OilPressure, data.BatteryVoltage,
		data.TirePressureFL, data.TirePressureFR, data.TirePressureRL, data.TirePressureRR,
		data.BrakeTempFL, data.BrakeTempFR, data.BrakeTempRL, data.BrakeTempRR,
		data.TireTempFL, data.TireTempFR, data.TireTempRL, data.TireTempRR,
		data.BrakePadWearFR, data.BrakePadWearRL, data.BrakePadWearRR,
		string(faultCodesJSON), data.Latitude, data.Longitude, data.ReportTime,
	)

	for _, code := range data.FaultCodes {
		s.processFaultCode(ctx, vehicleID, code, data.Latitude, data.Longitude)
	}

	s.processDiagnosticAnomalies(ctx, vehicleID, data, data.Latitude, data.Longitude)

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

func (s *VehicleService) processFaultCode(ctx context.Context, vehicleID int64, code string, lat, lng float64) {
	level := 1
	desc := ""
	suggestion := ""
	system := ""
	autoCallRescue := 0
	emergencyAction := ""

	if len(code) > 0 {
		switch code[0] {
		case 'P':
			level = 2
			system = "动力总成"
			desc = "动力总成故障: " + code
		case 'B':
			level = 2
			system = "车身系统"
			desc = "车身系统故障: " + code
		case 'C':
			level = 2
			system = "底盘系统"
			desc = "底盘系统故障: " + code
		case 'U':
			level = 1
			system = "网络通信"
			desc = "网络通信故障: " + code
		}
		suggestion = "建议尽快前往维修站检查处理"
	}

	var fc struct {
		FaultLevel      int
		TitleCn         string
		Description     string
		Suggestion      string
		EmergencyAction string
		AutoCallRescue  int
		FaultSystem     string
	}
	s.db.WithContext(ctx).Raw(`
		SELECT fault_level, title_cn, description, suggestion, emergency_action,
		       auto_call_rescue, fault_system
		FROM fault_code_library WHERE fault_code = ? AND status = 1 LIMIT 1`,
		code,
	).Scan(&fc)

	if fc.FaultLevel > 0 {
		level = fc.FaultLevel
		if fc.TitleCn != "" {
			desc = fc.TitleCn
		}
		if fc.Description != "" {
			desc = desc + " | " + fc.Description
		}
		if fc.Suggestion != "" {
			suggestion = fc.Suggestion
		}
		if fc.EmergencyAction != "" {
			emergencyAction = fc.EmergencyAction
		}
		autoCallRescue = fc.AutoCallRescue
		if fc.FaultSystem != "" {
			system = fc.FaultSystem
		}
	}

	var alertID int64
	now := time.Now()

	var existingAlertID int64
	s.db.WithContext(ctx).Raw(`
		SELECT id FROM vehicle_fault_alerts
		WHERE vehicle_id = ? AND fault_code = ? AND status IN (0, 1)
		LIMIT 1`, vehicleID, code,
	).Scan(&existingAlertID)

	if existingAlertID > 0 {
		alertID = existingAlertID
		s.db.WithContext(ctx).Exec(`
			UPDATE vehicle_fault_alerts SET
				last_report_time = ?,
				report_count = report_count + 1,
				status = IF(status = 2, 2, 0)
			WHERE id = ?`, now, alertID,
		)
	} else {
		s.db.WithContext(ctx).Exec(`
			INSERT INTO vehicle_fault_alerts
			(vehicle_id, fault_code, fault_level, fault_system, fault_desc,
			 fault_suggestion, emergency_action, latitude, longitude,
			 first_report_time, last_report_time, report_count, status, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, 0, NOW())`,
			vehicleID, code, level, system, desc, suggestion, emergencyAction,
			lat, lng, now, now,
		)
		s.db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&alertID)
	}

	s.recordAlertLog(ctx, alertID, "report", fmt.Sprintf("T-BOX上报故障码，经纬度:(%.6f,%.6f)", lat, lng), nil)

	alertPayload := map[string]interface{}{
		"id":               alertID,
		"vehicle_id":       vehicleID,
		"fault_code":       code,
		"fault_level":      level,
		"fault_system":     system,
		"fault_desc":       desc,
		"fault_suggestion": suggestion,
		"emergency_action": emergencyAction,
		"latitude":         lat,
		"longitude":        lng,
		"report_time":      now,
		"status":           0,
	}

	var plateNumber string
	s.db.WithContext(ctx).Raw("SELECT plate_number FROM vehicles WHERE id = ?", vehicleID).Scan(&plateNumber)
	alertPayload["plate_number"] = plateNumber

	wsMsg := &monitorWs.WSMessage{
		Type:      "fault_alert",
		Timestamp: now.Unix(),
		Data:      alertPayload,
	}
	hub := monitorWs.GetHub()
	hub.BroadcastToMonitor(wsMsg, "admin", "dispatcher")

	if autoCallRescue == 1 || level >= 4 {
		s.autoCallRescueCenter(ctx, vehicleID, code, desc, lat, lng, alertID, level)
	}
}

func (s *VehicleService) recordAlertLog(ctx context.Context, alertID int64, actionType, detail string, operatorID *int64) {
	var operatorName string
	if operatorID != nil && *operatorID > 0 {
		s.db.WithContext(ctx).Raw("SELECT real_name FROM users WHERE id = ?", *operatorID).Scan(&operatorName)
	}
	var opID interface{}
	if operatorID != nil {
		opID = *operatorID
	}
	s.db.WithContext(ctx).Exec(`
		INSERT INTO vehicle_fault_alert_logs
		(alert_id, action_type, action_detail, operator_id, operator_name, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())`,
		alertID, actionType, detail, opID, operatorName,
	)
}

func (s *VehicleService) autoCallRescueCenter(ctx context.Context, vehicleID int64, faultCode, faultDesc string, lat, lng float64, alertID int64, level int) {
	requestNo := fmt.Sprintf("RESCUE-%d-%s", vehicleID, time.Now().Format("20060102150405"))

	var driverID, orgID int64
	var driverName, plateNumber, driverPhone string
	var waybillID, waybillNo interface{}

	s.db.WithContext(ctx).Raw(`
		SELECT v.id, v.plate_number, v.driver_id, v.org_id, u.real_name, u.phone,
		       w.id, w.waybill_no
		FROM vehicles v
		LEFT JOIN users u ON u.id = v.driver_id
		LEFT JOIN waybills w ON w.vehicle_id = v.id AND w.status = 'in_transit'
		WHERE v.id = ?`, vehicleID,
	).Scan(&vehicleID, &plateNumber, &driverID, &orgID, &driverName, &driverPhone, &waybillID, &waybillNo)

	levelStr := "danger"
	if level >= 4 {
		levelStr = "critical"
	}

	s.db.WithContext(ctx).Exec(`
		INSERT INTO rescue_requests
		(request_no, waybill_id, vehicle_id, driver_id, sos_type, sos_level,
		 latitude, longitude, address, description, caller_name, caller_phone,
		 status, dispatcher_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, NOW())`,
		requestNo, waybillID, vehicleID, driverID,
		fmt.Sprintf("fault_code:%s", faultCode), levelStr,
		lat, lng,
		fmt.Sprintf("坐标(%.6f,%.6f)", lat, lng),
		fmt.Sprintf("【车辆%s 自动触发救援】故障码:%s, 描述:%s", plateNumber, faultCode, faultDesc),
		driverName, driverPhone,
		"auto_dispatched",
	)

	var rescueID int64
	s.db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&rescueID)

	s.recordAlertLog(ctx, alertID, "rescue",
		fmt.Sprintf("自动呼叫救援中心，救援请求号:%s，救援ID:%d，级别:%s", requestNo, rescueID, levelStr),
		nil,
	)

	sosPayload := map[string]interface{}{
		"rescue_id":    rescueID,
		"request_no":   requestNo,
		"vehicle_id":   vehicleID,
		"plate_number": plateNumber,
		"driver_id":    driverID,
		"driver_name":  driverName,
		"driver_phone": driverPhone,
		"waybill_id":   waybillID,
		"waybill_no":   waybillNo,
		"sos_type":     fmt.Sprintf("故障码自动触发:%s", faultCode),
		"sos_level":    levelStr,
		"latitude":     lat,
		"longitude":    lng,
		"description":  faultDesc,
		"alert_id":     alertID,
		"created_at":   time.Now(),
	}

	sosWS := &monitorWs.WSMessage{
		Type:      monitorWs.MsgSOSAlert,
		Timestamp: time.Now().Unix(),
		Data:      sosPayload,
	}
	hub := monitorWs.GetHub()
	hub.BroadcastToMonitor(sosWS, "admin", "dispatcher")

	logger.Sugar.Warnf("【自动救援触发】车辆%d(%s) 故障码:%s，级别:%d，救援号:%s，坐标:(%.6f,%.6f)",
		vehicleID, plateNumber, faultCode, level, requestNo, lat, lng)
}

func (s *VehicleService) processDiagnosticAnomalies(ctx context.Context, vehicleID int64, data *DiagnosticData, lat, lng float64) {
	const (
		StandardPressureMin = 7.0
		StandardPressureMax = 10.0
		WarningPressureMin  = 6.0
		WarningPressureMax  = 11.0
		TireTempWarning     = 75.0
		TireTempCritical    = 90.0
		BrakeTempWarning    = 250.0
		BrakeTempCritical   = 350.0
		BrakePadWearWarn    = 80.0
		BrakePadWearCrit    = 95.0
	)

	tirePressures := map[string]float64{
		"FL": data.TirePressureFL,
		"FR": data.TirePressureFR,
		"RL": data.TirePressureRL,
		"RR": data.TirePressureRR,
	}
	tireTemps := map[string]float64{
		"FL": data.TireTempFL,
		"FR": data.TireTempFR,
		"RL": data.TireTempRL,
		"RR": data.TireTempRR,
	}
	brakeTemps := map[string]float64{
		"FL": data.BrakeTempFL,
		"FR": data.BrakeTempFR,
		"RL": data.BrakeTempRL,
		"RR": data.BrakeTempRR,
	}
	brakePadWears := map[string]float64{
		"FR": data.BrakePadWearFR,
		"RL": data.BrakePadWearRL,
		"RR": data.BrakePadWearRR,
	}

	tirePosName := map[string]string{
		"FL": "左前轮", "FR": "右前轮", "RL": "左后轮", "RR": "右后轮",
	}

	for pos, pressure := range tirePressures {
		if pressure <= 0 {
			continue
		}
		faultCode := ""
		level := 0
		desc := ""
		suggestion := ""

		if pressure < WarningPressureMin {
			faultCode = fmt.Sprintf("TPMS-LOW-%s", pos)
			level = 3
			desc = fmt.Sprintf("%s胎压严重过低: %.2f bar", tirePosName[pos], pressure)
			suggestion = "【危险】立即降低车速，停车检查轮胎，充气或更换备胎！严禁继续高速行驶！"
		} else if pressure < StandardPressureMin {
			faultCode = fmt.Sprintf("TPMS-LOW-%s", pos)
			level = 2
			desc = fmt.Sprintf("%s胎压过低: %.2f bar", tirePosName[pos], pressure)
			suggestion = "尽快到就近地点检查充气，注意行驶中是否跑偏"
		} else if pressure > WarningPressureMax {
			faultCode = fmt.Sprintf("TPMS-HIGH-%s", pos)
			level = 2
			desc = fmt.Sprintf("%s胎压过高: %.2f bar", tirePosName[pos], pressure)
			suggestion = "适当放气至标准范围，高温时轮胎易过热爆胎，注意冷却"
		} else if pressure > StandardPressureMax {
			faultCode = fmt.Sprintf("TPMS-HIGH-%s", pos)
			level = 1
			desc = fmt.Sprintf("%s胎压偏高: %.2f bar", tirePosName[pos], pressure)
			suggestion = "注意检查，必要时放气调整"
		}

		if level > 0 {
			s.triggerAnomalyAlert(ctx, vehicleID, faultCode, level, "tire",
				desc, suggestion, lat, lng, data.ReportTime)
		}
	}

	for pos, temp := range tireTemps {
		if temp <= 0 {
			continue
		}
		faultCode := ""
		level := 0
		desc := ""
		suggestion := ""
		emergency := ""

		if temp >= TireTempCritical {
			faultCode = "TPMS-TEMP-HIGH"
			level = 4
			desc = fmt.Sprintf("%s轮胎温度严重过高: %.1f℃，存在爆胎风险！", tirePosName[pos], temp)
			suggestion = "【极度危险！】立即停车！人员撤离安全区域！等待轮胎完全冷却后检查，必要时更换轮胎！"
			emergency = "1.立即松油门打开双闪 2.逐步靠边停车 3.人员撤离到安全区域 4.拨打救援电话"
		} else if temp >= TireTempWarning {
			faultCode = "TPMS-TEMP-HIGH"
			level = 3
			desc = fmt.Sprintf("%s轮胎温度过高: %.1f℃", tirePosName[pos], temp)
			suggestion = "【警告】立即降低车速，进入服务区停车冷却轮胎，检查胎压！"
			emergency = "1.降低车速 2.避免急刹 3.尽快停靠安全地带冷却"
		}

		if level > 0 {
			s.triggerAnomalyAlert(ctx, vehicleID, faultCode, level, "tire",
				desc, suggestion, lat, lng, data.ReportTime)
			if level >= 3 {
				s.processFaultCode(ctx, vehicleID, faultCode, lat, lng)
			}
			if emergency != "" {
				s.checkEmergencyAndRescue(ctx, vehicleID, faultCode, desc, level, lat, lng, emergency)
			}
		}
	}

	for pos, temp := range brakeTemps {
		if temp <= 0 {
			continue
		}
		faultCode := ""
		level := 0
		desc := ""
		suggestion := ""
		emergency := ""

		if temp >= BrakeTempCritical {
			faultCode = "BRAKE-TEMP-HIGH"
			level = 4
			desc = fmt.Sprintf("%s刹车片温度严重过高: %.1f℃，制动热衰退风险极高！", tirePosName[pos], temp)
			suggestion = "【极度危险！立即减速！】停止使用刹车！使用发动机制动！立即停靠安全地带！"
			emergency = "1.立即松油门开双闪 2.换低速档利用发动机制动 3.间歇性点刹而非长踩 4.尽快靠边停车冷却制动系统 5.人员撤离安全区呼救"
		} else if temp >= BrakeTempWarning {
			faultCode = "BRAKE-TEMP-HIGH"
			level = 3
			desc = fmt.Sprintf("%s刹车片温度过高: %.1f℃，存在制动衰退风险", tirePosName[pos], temp)
			suggestion = "【警告】减少制动！使用发动机制动！进入服务区冷却制动系统！"
			emergency = "1.松油门减档 2.避免长距离刹车 3.尽快停车冷却"
		}

		if level > 0 {
			s.triggerAnomalyAlert(ctx, vehicleID, faultCode, level, "brake",
				desc, suggestion, lat, lng, data.ReportTime)
			if level >= 3 {
				s.processFaultCode(ctx, vehicleID, faultCode, lat, lng)
			}
			if emergency != "" {
				s.checkEmergencyAndRescue(ctx, vehicleID, faultCode, desc, level, lat, lng, emergency)
			}
		}
	}

	for pos, wear := range brakePadWears {
		if wear <= 0 {
			continue
		}
		posName := tirePosName[pos]
		faultCode := ""
		level := 0
		desc := ""
		suggestion := ""

		if wear >= BrakePadWearCrit {
			faultCode = fmt.Sprintf("BRAKE-PAD-WEAR-%s", pos)
			level = 4
			desc = fmt.Sprintf("%s刹车片磨损严重: %.1f%%，已到更换极限！制动失效风险！", posName, wear)
			suggestion = "【极度危险！】严禁继续行驶！立即停车更换刹车片！否则可能刹车失灵！"
		} else if wear >= BrakePadWearWarn {
			faultCode = fmt.Sprintf("BRAKE-PAD-WEAR-%s", pos)
			level = 2
			desc = fmt.Sprintf("%s刹车片磨损较大: %.1f%%，建议尽快更换", posName, wear)
			suggestion = "尽快安排保养更换刹车片，注意检查制动盘状态"
		}

		if level > 0 {
			s.triggerAnomalyAlert(ctx, vehicleID, faultCode, level, "brake",
				desc, suggestion, lat, lng, data.ReportTime)
			if level >= 3 {
				s.processFaultCode(ctx, vehicleID, faultCode, lat, lng)
			}
		}
	}
}

func (s *VehicleService) triggerAnomalyAlert(ctx context.Context, vehicleID int64, faultCode string, level int, system, desc, suggestion string, lat, lng float64, reportTime time.Time) {
	var existingCount int64
	s.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM vehicle_fault_alerts
		WHERE vehicle_id = ? AND fault_code = ? AND status IN (0, 1)
		  AND last_report_time > ?`,
		vehicleID, faultCode, reportTime.Add(-10*time.Minute),
	).Scan(&existingCount)

	if existingCount > 0 {
		s.db.WithContext(ctx).Exec(`
			UPDATE vehicle_fault_alerts SET
				last_report_time = NOW(),
				report_count = report_count + 1
			WHERE vehicle_id = ? AND fault_code = ? AND status IN (0, 1)
			ORDER BY id DESC LIMIT 1`,
			vehicleID, faultCode,
		)
		return
	}

	var alertID int64
	s.db.WithContext(ctx).Exec(`
		INSERT INTO vehicle_fault_alerts
		(vehicle_id, fault_code, fault_level, fault_system, fault_desc,
		 fault_suggestion, latitude, longitude,
		 first_report_time, last_report_time, report_count, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, 0, NOW())`,
		vehicleID, faultCode, level, system, desc, suggestion,
		lat, lng, reportTime, reportTime,
	)
	s.db.WithContext(ctx).Raw("SELECT LAST_INSERT_ID()").Scan(&alertID)

	s.recordAlertLog(ctx, alertID, "report",
		fmt.Sprintf("数据异常检测触发: %s，经纬度:(%.6f,%.6f)", desc, lat, lng),
		nil,
	)

	var plateNumber string
	s.db.WithContext(ctx).Raw("SELECT plate_number FROM vehicles WHERE id = ?", vehicleID).Scan(&plateNumber)

	alertPayload := map[string]interface{}{
		"id":               alertID,
		"vehicle_id":       vehicleID,
		"plate_number":     plateNumber,
		"fault_code":       faultCode,
		"fault_level":      level,
		"fault_system":     system,
		"fault_desc":       desc,
		"fault_suggestion": suggestion,
		"latitude":         lat,
		"longitude":        lng,
		"report_time":      reportTime,
		"status":           0,
	}

	wsMsg := &monitorWs.WSMessage{
		Type:      "fault_alert",
		Timestamp: time.Now().Unix(),
		Data:      alertPayload,
	}
	hub := monitorWs.GetHub()
	hub.BroadcastToMonitor(wsMsg, "admin", "dispatcher")
}

func (s *VehicleService) checkEmergencyAndRescue(ctx context.Context, vehicleID int64, faultCode, faultDesc string, level int, lat, lng float64, _ string) {
	if level >= 4 {
		var alertID int64
		s.db.WithContext(ctx).Raw(`
			SELECT id FROM vehicle_fault_alerts
			WHERE vehicle_id = ? AND fault_code = ? ORDER BY id DESC LIMIT 1`,
			vehicleID, faultCode,
		).Scan(&alertID)
		if alertID > 0 {
			s.autoCallRescueCenter(ctx, vehicleID, faultCode, faultDesc, lat, lng, alertID, level)
		}
	}
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

type ChartPoint struct {
	Time  string  `json:"time"`
	FL    float64 `json:"fl"`
	FR    float64 `json:"fr"`
	RL    float64 `json:"rl"`
	RR    float64 `json:"rr"`
}

type ChartResponse struct {
	VehicleID   int64        `json:"vehicle_id"`
	PlateNumber string       `json:"plate_number"`
	StartDate   string       `json:"start_date"`
	EndDate     string       `json:"end_date"`
	Interval    string       `json:"interval"`
	Points      []ChartPoint `json:"points"`
	Summary     interface{}  `json:"summary"`
	StandardMin float64      `json:"standard_min"`
	StandardMax float64      `json:"standard_max"`
	WarningMin  float64      `json:"warning_min"`
	WarningMax  float64      `json:"warning_max"`
	TotalPoints int          `json:"total_points"`
}

func (s *VehicleService) GetTirePressureChart(ctx context.Context, vehicleID int64, startDate, endDate, interval string) (*ChartResponse, error) {
	return s.getChartData(ctx, vehicleID, startDate, endDate, interval, "pressure")
}

func (s *VehicleService) GetTireTempChart(ctx context.Context, vehicleID int64, startDate, endDate, interval string) (*ChartResponse, error) {
	return s.getChartData(ctx, vehicleID, startDate, endDate, interval, "tire_temp")
}

func (s *VehicleService) GetBrakeTempChart(ctx context.Context, vehicleID int64, startDate, endDate, interval string) (*ChartResponse, error) {
	return s.getChartData(ctx, vehicleID, startDate, endDate, interval, "brake_temp")
}

func (s *VehicleService) getChartData(ctx context.Context, vehicleID int64, startDate, endDate, interval, dataType string) (*ChartResponse, error) {
	var plateNumber string
	s.db.WithContext(ctx).Raw("SELECT plate_number FROM vehicles WHERE id = ?", vehicleID).Scan(&plateNumber)

	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}

	timeFormat := "%Y-%m-%d %H:00:00"
	switch interval {
	case "day":
		timeFormat = "%Y-%m-%d 00:00:00"
	case "minute":
		timeFormat = "%Y-%m-%d %H:%i:00"
	case "raw":
		timeFormat = ""
	}

	var flCol, frCol, rlCol, rrCol string
	var stdMin, stdMax, warnMin, warnMax float64
	switch dataType {
	case "pressure":
		flCol, frCol, rlCol, rrCol = "tire_pressure_fl", "tire_pressure_fr", "tire_pressure_rl", "tire_pressure_rr"
		stdMin, stdMax = 7.0, 10.0
		warnMin, warnMax = 6.0, 11.0
	case "tire_temp":
		flCol, frCol, rlCol, rrCol = "tire_temp_fl", "tire_temp_fr", "tire_temp_rl", "tire_temp_rr"
		stdMin, stdMax = 0, 75
		warnMin, warnMax = 0, 90
	case "brake_temp":
		flCol, frCol, rlCol, rrCol = "brake_temp_fl", "brake_temp_fr", "brake_temp_rl", "brake_temp_rr"
		stdMin, stdMax = 0, 250
		warnMin, warnMax = 0, 350
	}

	var rows *sql.Rows
	var err error

	if timeFormat == "" {
		querySQL := fmt.Sprintf(`
			SELECT report_time, %s, %s, %s, %s
			FROM vehicle_diagnostics
			WHERE vehicle_id = ?
			  AND report_time >= ?
			  AND report_time < ?
			ORDER BY report_time ASC
			LIMIT 5000`, flCol, frCol, rlCol, rrCol)
		rows, err = s.db.WithContext(ctx).Raw(querySQL, vehicleID, startDate, endDate).Rows()
	} else {
		querySQL := fmt.Sprintf(`
			SELECT DATE_FORMAT(report_time, '%s') as time_point,
			       AVG(%s) as fl, AVG(%s) as fr, AVG(%s) as rl, AVG(%s) as rr
			FROM vehicle_diagnostics
			WHERE vehicle_id = ?
			  AND report_time >= ?
			  AND report_time < ?
			GROUP BY time_point
			ORDER BY time_point ASC
			LIMIT 2000`, timeFormat, flCol, frCol, rlCol, rrCol)
		rows, err = s.db.WithContext(ctx).Raw(querySQL, vehicleID, startDate, endDate).Rows()
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []ChartPoint
	var fls, frs, rls, rrs []float64
	var maxFL, maxFR, maxRL, maxRR, minFL, minFR, minRL, minRR, sumFL, sumFR, sumRL, sumRR float64
	first := true
	for rows.Next() {
		var p ChartPoint
		var fl, fr, rl, rr float64
		rows.Scan(&p.Time, &fl, &fr, &rl, &rr)
		p.FL = round2(fl)
		p.FR = round2(fr)
		p.RL = round2(rl)
		p.RR = round2(rr)
		points = append(points, p)

		fls = append(fls, fl)
		frs = append(frs, fr)
		rls = append(rls, rl)
		rrs = append(rrs, rr)

		sumFL += fl
		sumFR += fr
		sumRL += rl
		sumRR += rr

		if first {
			maxFL, maxFR, maxRL, maxRR = fl, fr, rl, rr
			minFL, minFR, minRL, minRR = fl, fr, rl, rr
			first = false
		} else {
			maxFL = math.Max(maxFL, fl)
			maxFR = math.Max(maxFR, fr)
			maxRL = math.Max(maxRL, rl)
			maxRR = math.Max(maxRR, rr)
			minFL = math.Min(minFL, fl)
			minFR = math.Min(minFR, fr)
			minRL = math.Min(minRL, rl)
			minRR = math.Min(minRR, rr)
		}
	}

	n := float64(len(points))
	summary := map[string]interface{}{
		"avg_fl":  round2(div(sumFL, n)),
		"avg_fr":  round2(div(sumFR, n)),
		"avg_rl":  round2(div(sumRL, n)),
		"avg_rr":  round2(div(sumRR, n)),
		"max_fl":  round2(maxFL),
		"max_fr":  round2(maxFR),
		"max_rl":  round2(maxRL),
		"max_rr":  round2(maxRR),
		"min_fl":  round2(minFL),
		"min_fr":  round2(minFR),
		"min_rl":  round2(minRL),
		"min_rr":  round2(minRR),
		"warn_count_fl": countOutside(fls, warnMin, warnMax),
		"warn_count_fr": countOutside(frs, warnMin, warnMax),
		"warn_count_rl": countOutside(rls, warnMin, warnMax),
		"warn_count_rr": countOutside(rrs, warnMin, warnMax),
	}

	if points == nil {
		points = []ChartPoint{}
	}

	return &ChartResponse{
		VehicleID:   vehicleID,
		PlateNumber: plateNumber,
		StartDate:   startDate,
		EndDate:     endDate,
		Interval:    interval,
		Points:      points,
		Summary:     summary,
		StandardMin: stdMin,
		StandardMax: stdMax,
		WarningMin:  warnMin,
		WarningMax:  warnMax,
		TotalPoints: len(points),
	}, nil
}

func div(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func countOutside(values []float64, min, max float64) int {
	count := 0
	for _, v := range values {
		if v <= 0 {
			continue
		}
		if min > 0 && v < min {
			count++
		}
		if max > 0 && v > max {
			count++
		}
	}
	return count
}

func (s *VehicleService) ExportDiagnostics(ctx context.Context, vehicleID int64, startDate, endDate, format string) (interface{}, error) {
	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}

	var plateNumber string
	s.db.WithContext(ctx).Raw("SELECT plate_number FROM vehicles WHERE id = ?", vehicleID).Scan(&plateNumber)

	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT report_time, engine_rpm, vehicle_speed, coolant_temp, fuel_level,
		       oil_pressure, battery_voltage,
		       tire_pressure_fl, tire_pressure_fr, tire_pressure_rl, tire_pressure_rr,
		       brake_temp_fl, brake_temp_fr, brake_temp_rl, brake_temp_rr,
		       tire_temp_fl, tire_temp_fr, tire_temp_rl, tire_temp_rr,
		       brake_pad_wear_fr, brake_pad_wear_rl, brake_pad_wear_rr,
		       fault_codes, latitude, longitude
		FROM vehicle_diagnostics
		WHERE vehicle_id = ?
		  AND report_time >= ?
		  AND report_time < ?
		ORDER BY report_time ASC
		LIMIT 50000`,
		vehicleID, startDate, endDate,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type DiagExport struct {
		VehicleID     int64       `json:"vehicle_id"`
		PlateNumber   string      `json:"plate_number"`
		ExportRange   string      `json:"export_range"`
		ExportTime    string      `json:"export_time"`
		TotalRecords  int         `json:"total_records"`
		Data          []map[string]interface{} `json:"data"`
	}

	var dataList []map[string]interface{}
	for rows.Next() {
		var (
			rptTime time.Time
			rpm, speed, coolant, fuel, oil, batt,
			tpFL, tpFR, tpRL, tpRR,
			btFL, btFR, btRL, btRR,
			ttFL, ttFR, ttRL, ttRR,
			bwFR, bwRL, bwRR float64
			faultCodes, lat, lng interface{}
		)
		rows.Scan(&rptTime, &rpm, &speed, &coolant, &fuel, &oil, &batt,
			&tpFL, &tpFR, &tpRL, &tpRR,
			&btFL, &btFR, &btRL, &btRR,
			&ttFL, &ttFR, &ttRL, &ttRR,
			&bwFR, &bwRL, &bwRR,
			&faultCodes, &lat, &lng)
		item := map[string]interface{}{
			"report_time":      rptTime.Format("2006-01-02 15:04:05"),
			"engine_rpm":       rpm,
			"vehicle_speed":    speed,
			"coolant_temp":     coolant,
			"fuel_level":       fuel,
			"oil_pressure":     oil,
			"battery_voltage":  batt,
			"tire_pressure_fl": tpFL,
			"tire_pressure_fr": tpFR,
			"tire_pressure_rl": tpRL,
			"tire_pressure_rr": tpRR,
			"brake_temp_fl":    btFL,
			"brake_temp_fr":    btFR,
			"brake_temp_rl":    btRL,
			"brake_temp_rr":    btRR,
			"tire_temp_fl":     ttFL,
			"tire_temp_fr":     ttFR,
			"tire_temp_rl":     ttRL,
			"tire_temp_rr":     ttRR,
			"brake_pad_wear_fr": bwFR,
			"brake_pad_wear_rl": bwRL,
			"brake_pad_wear_rr": bwRR,
			"fault_codes":      faultCodes,
			"latitude":         lat,
			"longitude":        lng,
		}
		dataList = append(dataList, item)
	}

	if format == "csv" {
		csv := "report_time,engine_rpm,vehicle_speed,coolant_temp,fuel_level,oil_pressure,battery_voltage," +
			"tire_pressure_fl,tire_pressure_fr,tire_pressure_rl,tire_pressure_rr," +
			"brake_temp_fl,brake_temp_fr,brake_temp_rl,brake_temp_rr," +
			"tire_temp_fl,tire_temp_fr,tire_temp_rl,tire_temp_rr," +
			"brake_pad_wear_fr,brake_pad_wear_rl,brake_pad_wear_rr,fault_codes,latitude,longitude\n"
		for _, d := range dataList {
			csv += fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%q,%v,%v\n",
				d["report_time"], d["engine_rpm"], d["vehicle_speed"], d["coolant_temp"],
				d["fuel_level"], d["oil_pressure"], d["battery_voltage"],
				d["tire_pressure_fl"], d["tire_pressure_fr"], d["tire_pressure_rl"], d["tire_pressure_rr"],
				d["brake_temp_fl"], d["brake_temp_fr"], d["brake_temp_rl"], d["brake_temp_rr"],
				d["tire_temp_fl"], d["tire_temp_fr"], d["tire_temp_rl"], d["tire_temp_rr"],
				d["brake_pad_wear_fr"], d["brake_pad_wear_rl"], d["brake_pad_wear_rr"],
				d["fault_codes"], d["latitude"], d["longitude"],
			)
		}
		return csv, nil
	}

	return DiagExport{
		VehicleID:    vehicleID,
		PlateNumber:  plateNumber,
		ExportRange:  startDate + " ~ " + endDate,
		ExportTime:   time.Now().Format("2006-01-02 15:04:05"),
		TotalRecords: len(dataList),
		Data:         dataList,
	}, nil
}

func (s *VehicleService) AckFaultAlert(ctx context.Context, vehicleID, alertID, operatorID int64, remark string) error {
	result := s.db.WithContext(ctx).Exec(`
		UPDATE vehicle_fault_alerts SET
			status = 1,
			acked_by = ?,
			acked_at = NOW()
		WHERE id = ? AND vehicle_id = ? AND status = 0`,
		operatorID, alertID, vehicleID,
	)
	if result.Error != nil {
		return result.Error
	}

	s.recordAlertLog(ctx, alertID, "ack", remark, &operatorID)
	return nil
}

func (s *VehicleService) ResolveFaultAlert(ctx context.Context, vehicleID, alertID, operatorID int64, remark string) error {
	result := s.db.WithContext(ctx).Exec(`
		UPDATE vehicle_fault_alerts SET
			status = 2,
			resolved_by = ?,
			resolved_at = NOW(),
			resolved_note = ?
		WHERE id = ? AND vehicle_id = ? AND status IN (0, 1)`,
		operatorID, remark, alertID, vehicleID,
	)
	if result.Error != nil {
		return result.Error
	}

	s.recordAlertLog(ctx, alertID, "resolve", remark, &operatorID)
	return nil
}

type FaultAlertListItem struct {
	ID               int64     `json:"id"`
	VehicleID        int64     `json:"vehicle_id"`
	PlateNumber      string    `json:"plate_number"`
	FaultCode        string    `json:"fault_code"`
	FaultLevel       int       `json:"fault_level"`
	FaultSystem      string    `json:"fault_system"`
	FaultDesc        string    `json:"fault_desc"`
	FaultSuggestion  string    `json:"fault_suggestion"`
	EmergencyAction  string    `json:"emergency_action"`
	LastReportTime   time.Time `json:"last_report_time"`
	ReportCount      int       `json:"report_count"`
	Status           int       `json:"status"`
	Latitude         float64   `json:"latitude"`
	Longitude        float64   `json:"longitude"`
	Logs             interface{} `json:"logs,omitempty"`
}

func (s *VehicleService) ListAllFaultAlerts(ctx context.Context, orgID, vehicleID int64, level int, status string, page, pageSize int) ([]*FaultAlertListItem, int64, error) {
	var total int64
	countSQL := `
		SELECT COUNT(*) FROM vehicle_fault_alerts a
		LEFT JOIN vehicles v ON v.id = a.vehicle_id
		WHERE 1=1`
	listSQL := `
		SELECT a.id, a.vehicle_id, v.plate_number, a.fault_code, a.fault_level,
		       a.fault_system, a.fault_desc, a.fault_suggestion, a.emergency_action,
		       a.last_report_time, a.report_count, a.status,
		       COALESCE(a.latitude, 0), COALESCE(a.longitude, 0)
		FROM vehicle_fault_alerts a
		LEFT JOIN vehicles v ON v.id = a.vehicle_id
		WHERE 1=1`

	args := []interface{}{}
	if orgID > 0 {
		countSQL += " AND v.org_id = ?"
		listSQL += " AND v.org_id = ?"
		args = append(args, orgID)
	}
	if vehicleID > 0 {
		countSQL += " AND a.vehicle_id = ?"
		listSQL += " AND a.vehicle_id = ?"
		args = append(args, vehicleID)
	}
	if level > 0 {
		countSQL += " AND a.fault_level = ?"
		listSQL += " AND a.fault_level = ?"
		args = append(args, level)
	}
	if status != "" {
		countSQL += " AND a.status = ?"
		listSQL += " AND a.status = ?"
		var statusVal int
		switch status {
		case "pending":
			statusVal = 0
		case "acked":
			statusVal = 1
		case "resolved":
			statusVal = 2
		case "ignored":
			statusVal = 3
		}
		args = append(args, statusVal)
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	s.db.WithContext(ctx).Raw(countSQL, countArgs...).Scan(&total)

	listSQL += " ORDER BY a.fault_level DESC, a.last_report_time DESC LIMIT ? OFFSET ?"
	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)

	rows, err := s.db.WithContext(ctx).Raw(listSQL, args...).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*FaultAlertListItem
	for rows.Next() {
		var item FaultAlertListItem
		rows.Scan(&item.ID, &item.VehicleID, &item.PlateNumber, &item.FaultCode,
			&item.FaultLevel, &item.FaultSystem, &item.FaultDesc,
			&item.FaultSuggestion, &item.EmergencyAction,
			&item.LastReportTime, &item.ReportCount, &item.Status,
			&item.Latitude, &item.Longitude)
		list = append(list, &item)
	}
	return list, total, nil
}

type DiagStats struct {
	TotalVehicles       int64                  `json:"total_vehicles"`
	ReportingVehicles   int64                  `json:"reporting_vehicles"`
	TodayReports        int64                  `json:"today_reports"`
	TotalFaults         int64                  `json:"total_faults"`
	Level1Faults        int64                  `json:"level_1_faults"`
	Level2Faults        int64                  `json:"level_2_faults"`
	Level3Faults        int64                  `json:"level_3_faults"`
	Level4Faults        int64                  `json:"level_4_faults"`
	PendingFaults       int64                  `json:"pending_faults"`
	ResolvedFaults      int64                  `json:"resolved_faults"`
	AutoRescueCount     int64                  `json:"auto_rescue_count"`
	AnomalyDistribution []map[string]interface{} `json:"anomaly_distribution"`
	WeeklyTrend         []map[string]interface{} `json:"weekly_trend"`
	TopVehicles         []map[string]interface{} `json:"top_fault_vehicles"`
}

func (s *VehicleService) GetDiagnosticsStats(ctx context.Context, orgID, vehicleID int64) (*DiagStats, error) {
	stats := &DiagStats{}

	orgFilter := ""
	orgArgs := []interface{}{}
	if orgID > 0 {
		orgFilter = " WHERE v.org_id = ?"
		orgArgs = append(orgArgs, orgID)
	}

	s.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM vehicles` + orgFilter, orgArgs...).Scan(&stats.TotalVehicles)

	s.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT vehicle_id) FROM vehicle_diagnostics
		WHERE report_time >= DATE_SUB(NOW(), INTERVAL 1 DAY)`).Scan(&stats.ReportingVehicles)

	s.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM vehicle_diagnostics
		WHERE DATE(report_time) = CURDATE()`).Scan(&stats.TodayReports)

	faultFilter := "WHERE 1=1"
	faultArgs := []interface{}{}
	if orgID > 0 {
		faultFilter += " AND v.org_id = ?"
		faultArgs = append(faultArgs, orgID)
	}
	if vehicleID > 0 {
		faultFilter += " AND a.vehicle_id = ?"
		faultArgs = append(faultArgs, vehicleID)
	}

	s.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM vehicle_fault_alerts a
		LEFT JOIN vehicles v ON v.id = a.vehicle_id ` + faultFilter, faultArgs...).Scan(&stats.TotalFaults)

	levelArgs := make([]interface{}, len(faultArgs))
	copy(levelArgs, faultArgs)
	s.db.WithContext(ctx).Raw(`
		SELECT
			SUM(CASE WHEN a.fault_level = 1 THEN 1 ELSE 0 END),
			SUM(CASE WHEN a.fault_level = 2 THEN 1 ELSE 0 END),
			SUM(CASE WHEN a.fault_level = 3 THEN 1 ELSE 0 END),
			SUM(CASE WHEN a.fault_level = 4 THEN 1 ELSE 0 END),
			SUM(CASE WHEN a.status = 0 THEN 1 ELSE 0 END),
			SUM(CASE WHEN a.status = 2 THEN 1 ELSE 0 END)
		FROM vehicle_fault_alerts a
		LEFT JOIN vehicles v ON v.id = a.vehicle_id ` + faultFilter, levelArgs...).
		Scan(&stats.Level1Faults, &stats.Level2Faults, &stats.Level3Faults,
			&stats.Level4Faults, &stats.PendingFaults, &stats.ResolvedFaults)

	s.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM rescue_requests WHERE sos_type LIKE 'fault_code%'
		  AND DATE(created_at) >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)`).Scan(&stats.AutoRescueCount)

	distRows, _ := s.db.WithContext(ctx).Raw(`
		SELECT COALESCE(a.fault_system, 'unknown') as system,
		       COUNT(*) as count,
		       SUM(CASE WHEN a.status = 0 THEN 1 ELSE 0 END) as pending
		FROM vehicle_fault_alerts a
		LEFT JOIN vehicles v ON v.id = a.vehicle_id ` + faultFilter + `
		GROUP BY a.fault_system ORDER BY count DESC LIMIT 10`, faultArgs...).Rows()
	if distRows != nil {
		for distRows.Next() {
			var sys string
			var cnt, pending int64
			distRows.Scan(&sys, &cnt, &pending)
			stats.AnomalyDistribution = append(stats.AnomalyDistribution,
				map[string]interface{}{
					"system":  sys,
					"count":   cnt,
					"pending": pending,
				})
		}
		distRows.Close()
	}

	trendRows, _ := s.db.WithContext(ctx).Raw(`
		SELECT DATE(last_report_time) as date,
		       COUNT(*) as total,
		       SUM(CASE WHEN fault_level >= 3 THEN 1 ELSE 0 END) as critical
		FROM vehicle_fault_alerts a
		LEFT JOIN vehicles v ON v.id = a.vehicle_id ` + faultFilter + `
		  AND last_report_time >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
		GROUP BY DATE(last_report_time) ORDER BY date DESC`, faultArgs...).Rows()
	if trendRows != nil {
		for trendRows.Next() {
			var date string
			var total, critical int64
			trendRows.Scan(&date, &total, &critical)
			stats.WeeklyTrend = append(stats.WeeklyTrend,
				map[string]interface{}{
					"date":     date,
					"total":    total,
					"critical": critical,
				})
		}
		trendRows.Close()
	}

	topRows, _ := s.db.WithContext(ctx).Raw(`
		SELECT v.id, v.plate_number, COUNT(*) as fault_count,
		       SUM(CASE WHEN a.fault_level >= 3 THEN 1 ELSE 0 END) as critical_count
		FROM vehicle_fault_alerts a
		LEFT JOIN vehicles v ON v.id = a.vehicle_id
		WHERE DATE(a.last_report_time) >= DATE_SUB(CURDATE(), INTERVAL 30 DAY) ` +
		func() string {
			if orgID > 0 {
				return fmt.Sprintf(" AND v.org_id = %d", orgID)
			}
			return ""
		}() + `
		GROUP BY v.id, v.plate_number
		ORDER BY fault_count DESC LIMIT 10`).Rows()
	if topRows != nil {
		for topRows.Next() {
			var vid int64
			var plate string
			var fcnt, ccnt int64
			topRows.Scan(&vid, &plate, &fcnt, &ccnt)
			stats.TopVehicles = append(stats.TopVehicles,
				map[string]interface{}{
					"vehicle_id":     vid,
					"plate_number":   plate,
					"fault_count":    fcnt,
					"critical_count": ccnt,
				})
		}
		topRows.Close()
	}

	return stats, nil
}
