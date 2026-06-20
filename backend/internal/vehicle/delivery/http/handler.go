package http

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	vehSvc "github.com/dangerous-drive-guard/backend/internal/vehicle/service"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

type DiagnosticsUpload struct {
	VehicleID      int64    `json:"vehicle_id" binding:"required"`
	EngineRPM      int      `json:"engine_rpm"`
	VehicleSpeed   float64  `json:"vehicle_speed"`
	CoolantTemp    float64  `json:"coolant_temp"`
	FuelLevel      float64  `json:"fuel_level"`
	OilPressure    float64  `json:"oil_pressure"`
	BatteryVoltage float64  `json:"battery_voltage"`
	TirePressureFL float64  `json:"tire_pressure_fl"`
	TirePressureFR float64  `json:"tire_pressure_fr"`
	TirePressureRL float64  `json:"tire_pressure_rl"`
	TirePressureRR float64  `json:"tire_pressure_rr"`
	FaultCodes     []string `json:"fault_codes"`
	Latitude       float64  `json:"latitude"`
	Longitude      float64  `json:"longitude"`
	ReportTime     string   `json:"report_time"`
}

type VehicleHandler struct {
	vehicleService *vehSvc.VehicleService
}

func NewVehicleHandler(svc *vehSvc.VehicleService) *VehicleHandler {
	return &VehicleHandler{vehicleService: svc}
}

func (h *VehicleHandler) RegisterRoutes(r *app.RouterGroup, authMiddleware app.HandlerFunc) {
	vehicles := r.Group("/vehicles", authMiddleware)
	{
		vehicles.GET("", h.ListVehicles)
		vehicles.GET("/:id", h.GetVehicle)
		vehicles.POST("", h.CreateVehicle)
		vehicles.PUT("/:id", h.UpdateVehicle)
		vehicles.DELETE("/:id", h.DeleteVehicle)

		vehicles.GET("/:id/diagnostics", h.GetDiagnostics)
		vehicles.POST("/diagnostics/upload", h.UploadDiagnostic)
		vehicles.GET("/:id/faults", h.GetFaults)

		vehicles.GET("/:id/realtime", h.GetRealtimeStatus)
		vehicles.GET("/realtime", h.ListRealtimeStatus)

		vehicles.GET("/:driver_id/scores", h.GetDrivingScore)
		vehicles.GET("/scores/rank", h.GetScoreRank)
	}
}

func (h *VehicleHandler) ListVehicles(c context.Context, ctx *app.RequestContext) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	orgID, _ := strconv.ParseInt(ctx.DefaultQuery("org_id", "0"), 10, 64)
	status := ctx.Query("status")
	keyword := ctx.Query("keyword")

	vehicles, total, err := h.vehicleService.ListVehicles(c, orgID, model.VehicleStatus(status), keyword, page, pageSize)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Page(ctx, vehicles, total, page, pageSize)
}

func (h *VehicleHandler) GetVehicle(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid vehicle id")
		return
	}

	v, err := h.vehicleService.GetVehicle(c, id)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, v)
}

func (h *VehicleHandler) CreateVehicle(c context.Context, ctx *app.RequestContext) {
	var req model.Vehicle
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	v, err := h.vehicleService.CreateVehicle(c, &req)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, v)
}

func (h *VehicleHandler) UpdateVehicle(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid vehicle id")
		return
	}

	var req model.Vehicle
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	req.ID = id
	v, err := h.vehicleService.UpdateVehicle(c, &req)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, v)
}

func (h *VehicleHandler) DeleteVehicle(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid vehicle id")
		return
	}

	err = h.vehicleService.DeleteVehicle(c, id)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, utils.H{"deleted": true})
}

func (h *VehicleHandler) GetDiagnostics(c context.Context, ctx *app.RequestContext) {
	vehicleID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid vehicle id")
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "100"))
	data, err := h.vehicleService.GetRecentDiagnostics(c, vehicleID, limit)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, utils.H{
		"list":  data,
		"total": len(data),
	})
}

func (h *VehicleHandler) UploadDiagnostic(c context.Context, ctx *app.RequestContext) {
	var req DiagnosticsUpload
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	rptTime := time.Now()
	if req.ReportTime != "" {
		rptTime, _ = time.Parse(time.RFC3339, req.ReportTime)
	}

	err := h.vehicleService.UploadDiagnostics(c, req.VehicleID, &vehSvc.DiagnosticData{
		EngineRPM:      req.EngineRPM,
		VehicleSpeed:   req.VehicleSpeed,
		CoolantTemp:    req.CoolantTemp,
		FuelLevel:      req.FuelLevel,
		OilPressure:    req.OilPressure,
		BatteryVoltage: req.BatteryVoltage,
		TirePressureFL: req.TirePressureFL,
		TirePressureFR: req.TirePressureFR,
		TirePressureRL: req.TirePressureRL,
		TirePressureRR: req.TirePressureRR,
		FaultCodes:     req.FaultCodes,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		ReportTime:     rptTime,
	})
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, utils.H{"uploaded": true})
}

func (h *VehicleHandler) GetFaults(c context.Context, ctx *app.RequestContext) {
	vehicleID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid vehicle id")
		return
	}

	alerts, err := h.vehicleService.GetFaultAlerts(c, vehicleID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, utils.H{
		"list":  alerts,
		"total": len(alerts),
	})
}

func (h *VehicleHandler) GetRealtimeStatus(c context.Context, ctx *app.RequestContext) {
	vehicleID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid vehicle id")
		return
	}

	status, err := h.vehicleService.GetRealtimeStatus(c, vehicleID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, status)
}

func (h *VehicleHandler) ListRealtimeStatus(c context.Context, ctx *app.RequestContext) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	orgID, _ := strconv.ParseInt(ctx.DefaultQuery("org_id", "0"), 10, 64)
	status := ctx.Query("status")

	list, total, err := h.vehicleService.ListRealtimeStatus(c, orgID, status, page, pageSize)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Page(ctx, list, total, page, pageSize)
}

func (h *VehicleHandler) GetDrivingScore(c context.Context, ctx *app.RequestContext) {
	driverID, err := strconv.ParseInt(ctx.Param("driver_id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid driver id")
		return
	}

	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")

	scores, stats, err := h.vehicleService.GetDriverScore(c, driverID, startDate, endDate)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, utils.H{
		"list":       scores,
		"total":      len(scores),
		"statistics": stats,
	})
}

func (h *VehicleHandler) GetScoreRank(c context.Context, ctx *app.RequestContext) {
	orgID, _ := strconv.ParseInt(ctx.DefaultQuery("org_id", "0"), 10, 64)
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	period := ctx.DefaultQuery("period", "week")

	ranking, err := h.vehicleService.GetScoreRanking(c, orgID, period, limit)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, utils.H{
		"list":  ranking,
		"total": len(ranking),
	})
}
