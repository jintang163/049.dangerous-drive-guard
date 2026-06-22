package http

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/dangerous-drive-guard/backend/internal/adas/service"
	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

type ADASHandler struct {
	adasService *service.ADASService
}

var adasService *service.ADASService

func initADASService() *service.ADASService {
	if adasService == nil {
		adasService = service.NewADASService(config.Global)
	}
	return adasService
}

func NewADASHandler(svc *service.ADASService) *ADASHandler {
	return &ADASHandler{adasService: svc}
}

func (h *ADASHandler) RegisterRoutes(api *app.RouterGroup, authMiddleware ...app.HandlerFunc) {
	adas := api.Group("/adas", authMiddleware...)
	{
		adas.POST("/radar", h.ProcessRadarData)
		adas.GET("/alerts", h.GetAlerts)
		adas.GET("/alerts/:id", h.GetAlert)
		adas.POST("/alerts/:id/ack", h.AckAlert)
		adas.GET("/driver/:driver_id/summary", h.GetDriverSummary)
		adas.GET("/vehicle/:vehicle_id/frequency", h.GetFrequencyTrackers)

		configGroup := adas.Group("/config")
		{
			configGroup.GET("/:vehicle_id", h.GetConfig)
			configGroup.PUT("", h.UpdateConfig)
		}

		driver := adas.Group("/vehicle")
		{
			driver.GET("/:vehicle_id/active-alerts", h.GetVehicleActiveAlerts)
			driver.POST("/:vehicle_id/alert/:id/ack", h.VehicleAckAlert)
			driver.POST("/:vehicle_id/alert/:id/voice", h.SendVoiceAlert)
		}
	}
}

func (h *ADASHandler) ProcessRadarData(c context.Context, ctx *app.RequestContext) {
	svc := initADASService()
	var req model.RadarData
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	if req.VehicleID <= 0 || req.DriverID <= 0 {
		response.BadRequest(ctx, "vehicle_id and driver_id are required")
		return
	}

	resp, err := svc.ProcessRadarData(c, &req)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, resp)
}

func (h *ADASHandler) GetAlerts(c context.Context, ctx *app.RequestContext) {
	svc := initADASService()
	var query model.ADASAlertQuery
	if err := ctx.BindAndValidate(&query); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	result, err := svc.GetAlerts(c, &query)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Page(ctx, result.List, result.Total, result.Page, result.PageSize)
}

func (h *ADASHandler) GetAlert(c context.Context, ctx *app.RequestContext) {
	svc := initADASService()
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid alert id")
		return
	}

	alert, err := svc.GetAlert(c, id)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, alert)
}

func (h *ADASHandler) AckAlert(c context.Context, ctx *app.RequestContext) {
	svc := initADASService()
	var req model.ADASAlertAckRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	alert, err := svc.AckAlert(c, &req, 0)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, alert)
}

func (h *ADASHandler) GetDriverSummary(c context.Context, ctx *app.RequestContext) {
	svc := initADASService()
	driverIDStr := ctx.Param("driver_id")
	driverID, err := strconv.ParseInt(driverIDStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid driver id")
		return
	}

	summary, err := svc.GetDriverAlertSummary(c, driverID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, summary)
}

func (h *ADASHandler) GetFrequencyTrackers(c context.Context, ctx *app.RequestContext) {
	svc := initADASService()
	vehicleIDStr := ctx.Param("vehicle_id")
	vehicleID, err := strconv.ParseInt(vehicleIDStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid vehicle id")
		return
	}

	limitStr := ctx.Query("limit")
	limit := 20
	if limitStr != "" {
		if l, e := strconv.Atoi(limitStr); e == nil && l > 0 {
			limit = l
		}
	}

	trackers, err := svc.GetVehicleFrequencyTrackers(c, vehicleID, limit)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, trackers)
}

func (h *ADASHandler) GetConfig(c context.Context, ctx *app.RequestContext) {
	svc := initADASService()
	vehicleIDStr := ctx.Param("vehicle_id")
	vehicleID, err := strconv.ParseInt(vehicleIDStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid vehicle id")
		return
	}

	cfg, err := svc.GetADASConfig(c, vehicleID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, cfg)
}

func (h *ADASHandler) UpdateConfig(c context.Context, ctx *app.RequestContext) {
	svc := initADASService()
	var cfg model.ADASConfig
	if err := ctx.BindAndValidate(&cfg); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	if cfg.VehicleID <= 0 {
		response.BadRequest(ctx, "vehicle_id is required")
		return
	}

	result, err := svc.UpdateADASConfig(c, &cfg)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, result)
}

func (h *ADASHandler) GetVehicleActiveAlerts(c context.Context, ctx *app.RequestContext) {
	svc := initADASService()
	vehicleIDStr := ctx.Param("vehicle_id")
	vehicleID, err := strconv.ParseInt(vehicleIDStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid vehicle id")
		return
	}

	limitStr := ctx.Query("limit")
	limit := 10
	if limitStr != "" {
		if l, e := strconv.Atoi(limitStr); e == nil && l > 0 {
			limit = l
		}
	}

	alerts, err := svc.GetVehicleActiveAlerts(c, vehicleID, limit)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, alerts)
}

func (h *ADASHandler) VehicleAckAlert(c context.Context, ctx *app.RequestContext) {
	svc := initADASService()
	vehicleIDStr := ctx.Param("vehicle_id")
	vehicleID, err := strconv.ParseInt(vehicleIDStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid vehicle id")
		return
	}

	alertIDStr := ctx.Param("id")
	alertID, err := strconv.ParseInt(alertIDStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid alert id")
		return
	}

	var req struct {
		AckType      string `json:"ack_type"`
		Note         string `json:"note"`
		AckByDriver  bool   `json:"ack_by_driver"`
	}
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, "invalid request body")
		return
	}

	if req.AckType == "" {
		req.AckType = "resolve"
	}

	if err := svc.VehicleAckAlert(c, vehicleID, alertID, req.AckType, req.Note, req.AckByDriver); err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, nil)
}

func (h *ADASHandler) SendVoiceAlert(c context.Context, ctx *app.RequestContext) {
	svc := initADASService()
	vehicleIDStr := ctx.Param("vehicle_id")
	vehicleID, err := strconv.ParseInt(vehicleIDStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid vehicle id")
		return
	}

	alertIDStr := ctx.Param("id")
	alertID := int64(0)
	if alertIDStr != "" {
		if id, e := strconv.ParseInt(alertIDStr, 10, 64); e == nil {
			alertID = id
		}
	}

	var req struct {
		Message string `json:"message"`
	}
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, "invalid request body")
		return
	}

	if req.Message == "" {
		response.BadRequest(ctx, "message is required")
		return
	}

	if err := svc.SendVoiceAlertToVehicle(c, vehicleID, alertID, req.Message); err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, nil)
}
