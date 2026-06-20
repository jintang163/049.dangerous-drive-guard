package http

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	vehSvc "github.com/dangerous-drive-guard/backend/internal/vehicle/service"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

var vehicleService *vehSvc.VehicleService

func initService() {
	if vehicleService == nil {
		vehicleService = vehSvc.NewVehicleService()
	}
}

func CreateVehicle(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.Vehicle
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	v, err := vehicleService.CreateVehicle(ctx, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, v)
}

func GetVehicle(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}
	v, err := vehicleService.GetVehicle(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, v)
}

func ListVehicles(ctx context.Context, c *app.RequestContext) {
	initService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	orgID, _ := strconv.ParseInt(c.DefaultQuery("org_id", "0"), 10, 64)
	status := c.Query("status")
	keyword := c.Query("keyword")
	vehicles, total, err := vehicleService.ListVehicles(ctx, orgID, model.VehicleStatus(status), keyword, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, vehicles, total, page, pageSize)
}

func UpdateVehicle(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}
	var req model.Vehicle
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.ID = id
	v, err := vehicleService.UpdateVehicle(ctx, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, v)
}

func DeleteVehicle(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}
	err = vehicleService.DeleteVehicle(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"deleted": true})
}

type DiagnosticsUpload struct {
	VehicleID     int64    `json:"vehicle_id" binding:"required"`
	EngineRPM     int      `json:"engine_rpm"`
	VehicleSpeed  float64  `json:"vehicle_speed"`
	CoolantTemp   float64  `json:"coolant_temp"`
	FuelLevel     float64  `json:"fuel_level"`
	OilPressure   float64  `json:"oil_pressure"`
	BatteryVoltage float64 `json:"battery_voltage"`
	TirePressureFL float64 `json:"tire_pressure_fl"`
	TirePressureFR float64 `json:"tire_pressure_fr"`
	TirePressureRL float64 `json:"tire_pressure_rl"`
	TirePressureRR float64 `json:"tire_pressure_rr"`
	FaultCodes    []string `json:"fault_codes"`
	Latitude      float64  `json:"latitude"`
	Longitude     float64  `json:"longitude"`
	ReportTime    string   `json:"report_time"`
}

func UploadDiagnostics(ctx context.Context, c *app.RequestContext) {
	initService()
	var req DiagnosticsUpload
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	rptTime := time.Now()
	if req.ReportTime != "" {
		rptTime, _ = time.Parse(time.RFC3339, req.ReportTime)
	}
	err := vehicleService.UploadDiagnostics(ctx, req.VehicleID, &vehSvc.DiagnosticData{
		EngineRPM:     req.EngineRPM,
		VehicleSpeed:  req.VehicleSpeed,
		CoolantTemp:   req.CoolantTemp,
		FuelLevel:     req.FuelLevel,
		OilPressure:   req.OilPressure,
		BatteryVoltage: req.BatteryVoltage,
		TirePressureFL: req.TirePressureFL,
		TirePressureFR: req.TirePressureFR,
		TirePressureRL: req.TirePressureRL,
		TirePressureRR: req.TirePressureRR,
		FaultCodes:    req.FaultCodes,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		ReportTime:    rptTime,
	})
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"uploaded": true})
}

func GetRecentDiagnostics(ctx context.Context, c *app.RequestContext) {
	initService()
	vehicleID, err := strconv.ParseInt(c.Param("vehicle_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	data, err := vehicleService.GetRecentDiagnostics(ctx, vehicleID, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{
		"list":  data,
		"total": len(data),
	})
}

func GetFaultAlerts(ctx context.Context, c *app.RequestContext) {
	initService()
	vehicleID, err := strconv.ParseInt(c.Param("vehicle_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}
	alerts, err := vehicleService.GetFaultAlerts(ctx, vehicleID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{
		"list":  alerts,
		"total": len(alerts),
	})
}

func GetDriverScore(ctx context.Context, c *app.RequestContext) {
	initService()
	driverID, err := strconv.ParseInt(c.Param("driver_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid driver id")
		return
	}
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	scores, stats, err := vehicleService.GetDriverScore(ctx, driverID, startDate, endDate)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{
		"list":        scores,
		"total":       len(scores),
		"statistics":  stats,
	})
}

func GetScoreRanking(ctx context.Context, c *app.RequestContext) {
	initService()
	orgID, _ := strconv.ParseInt(c.DefaultQuery("org_id", "0"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	period := c.DefaultQuery("period", "week")
	ranking, err := vehicleService.GetScoreRanking(ctx, orgID, period, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{
		"list":  ranking,
		"total": len(ranking),
	})
}
