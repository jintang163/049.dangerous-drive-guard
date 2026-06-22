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
	BrakeTempFL    float64  `json:"brake_temp_fl"`
	BrakeTempFR    float64  `json:"brake_temp_fr"`
	BrakeTempRL    float64  `json:"brake_temp_rl"`
	BrakeTempRR    float64  `json:"brake_temp_rr"`
	TireTempFL     float64  `json:"tire_temp_fl"`
	TireTempFR     float64  `json:"tire_temp_fr"`
	TireTempRL     float64  `json:"tire_temp_rl"`
	TireTempRR     float64  `json:"tire_temp_rr"`
	BrakePadWearFR float64  `json:"brake_pad_wear_fr"`
	BrakePadWearRL float64  `json:"brake_pad_wear_rl"`
	BrakePadWearRR float64  `json:"brake_pad_wear_rr"`
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

		vehicles.GET("/:id/tire-pressure/chart", h.GetTirePressureChart)
		vehicles.GET("/:id/tire-temp/chart", h.GetTireTempChart)
		vehicles.GET("/:id/brake-temp/chart", h.GetBrakeTempChart)
		vehicles.GET("/:id/diagnostics/export", h.ExportDiagnostics)
		vehicles.POST("/:id/faults/:alert_id/ack", h.AckFaultAlert)
		vehicles.POST("/:id/faults/:alert_id/resolve", h.ResolveFaultAlert)
		vehicles.GET("/faults/alerts", h.ListAllFaultAlerts)
		vehicles.GET("/diagnostics/stats", h.GetDiagnosticsStats)
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
		BrakeTempFL:    req.BrakeTempFL,
		BrakeTempFR:    req.BrakeTempFR,
		BrakeTempRL:    req.BrakeTempRL,
		BrakeTempRR:    req.BrakeTempRR,
		TireTempFL:     req.TireTempFL,
		TireTempFR:     req.TireTempFR,
		TireTempRL:     req.TireTempRL,
		TireTempRR:     req.TireTempRR,
		BrakePadWearFR: req.BrakePadWearFR,
		BrakePadWearRL: req.BrakePadWearRL,
		BrakePadWearRR: req.BrakePadWearRR,
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

func (h *VehicleHandler) GetTirePressureChart(c context.Context, ctx *app.RequestContext) {
	vehicleID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid vehicle id")
		return
	}

	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")
	interval := ctx.DefaultQuery("interval", "hour")

	chart, err := h.vehicleService.GetTirePressureChart(c, vehicleID, startDate, endDate, interval)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, chart)
}

func (h *VehicleHandler) GetTireTempChart(c context.Context, ctx *app.RequestContext) {
	vehicleID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid vehicle id")
		return
	}

	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")
	interval := ctx.DefaultQuery("interval", "hour")

	chart, err := h.vehicleService.GetTireTempChart(c, vehicleID, startDate, endDate, interval)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, chart)
}

func (h *VehicleHandler) GetBrakeTempChart(c context.Context, ctx *app.RequestContext) {
	vehicleID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid vehicle id")
		return
	}

	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")
	interval := ctx.DefaultQuery("interval", "hour")

	chart, err := h.vehicleService.GetBrakeTempChart(c, vehicleID, startDate, endDate, interval)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, chart)
}

func (h *VehicleHandler) ExportDiagnostics(c context.Context, ctx *app.RequestContext) {
	vehicleID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid vehicle id")
		return
	}

	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")
	format := ctx.DefaultQuery("format", "json")

	data, err := h.vehicleService.ExportDiagnostics(c, vehicleID, startDate, endDate, format)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	if format == "csv" {
		ctx.Header("Content-Type", "text/csv; charset=utf-8")
		ctx.Header("Content-Disposition", "attachment; filename=diagnostics_"+strconv.FormatInt(vehicleID, 10)+".csv")
		ctx.SetBodyString(data.(string))
		return
	}

	response.Success(ctx, data)
}

func (h *VehicleHandler) AckFaultAlert(c context.Context, ctx *app.RequestContext) {
	vehicleID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid vehicle id")
		return
	}
	alertID, err := strconv.ParseInt(ctx.Param("alert_id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid alert id")
		return
	}

	userID, _ := ctx.Get("user_id")
	remark := ctx.Query("remark")

	err = h.vehicleService.AckFaultAlert(c, vehicleID, alertID, toInt64Ctx(userID), remark)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, utils.H{"acked": true})
}

func (h *VehicleHandler) ResolveFaultAlert(c context.Context, ctx *app.RequestContext) {
	vehicleID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid vehicle id")
		return
	}
	alertID, err := strconv.ParseInt(ctx.Param("alert_id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid alert id")
		return
	}

	userID, _ := ctx.Get("user_id")
	remark := ctx.Query("remark")

	err = h.vehicleService.ResolveFaultAlert(c, vehicleID, alertID, toInt64Ctx(userID), remark)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, utils.H{"resolved": true})
}

func (h *VehicleHandler) ListAllFaultAlerts(c context.Context, ctx *app.RequestContext) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	orgID, _ := strconv.ParseInt(ctx.DefaultQuery("org_id", "0"), 10, 64)
	level, _ := strconv.Atoi(ctx.DefaultQuery("level", "0"))
	status := ctx.Query("status")
	vehicleID, _ := strconv.ParseInt(ctx.DefaultQuery("vehicle_id", "0"), 10, 64)

	list, total, err := h.vehicleService.ListAllFaultAlerts(c, orgID, vehicleID, level, status, page, pageSize)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Page(ctx, list, total, page, pageSize)
}

func (h *VehicleHandler) GetDiagnosticsStats(c context.Context, ctx *app.RequestContext) {
	orgID, _ := strconv.ParseInt(ctx.DefaultQuery("org_id", "0"), 10, 64)
	vehicleID, _ := strconv.ParseInt(ctx.DefaultQuery("vehicle_id", "0"), 10, 64)

	stats, err := h.vehicleService.GetDiagnosticsStats(c, orgID, vehicleID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, stats)
}

func toInt64Ctx(v interface{}) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	case float64:
		return int64(x)
	default:
		return 0
	}
}
