package http

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/internal/weather/service"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

type WeatherHandler struct {
	weatherService *service.WeatherService
}

func NewWeatherHandler(svc *service.WeatherService) *WeatherHandler {
	return &WeatherHandler{weatherService: svc}
}

func (h *WeatherHandler) RegisterRoutes(r *app.RouterGroup, authMiddleware app.HandlerFunc) {
	weather := r.Group("/weather", authMiddleware)
	{
		weather.GET("/current", h.GetCurrentWeather)
		weather.GET("/route/:route_id", h.GetRouteWeather)
		weather.GET("/route/:route_id/analysis", h.AnalyzeRouteWeather)
		weather.GET("/route/waybill/:waybill_id/analysis", h.AnalyzeRouteByWaybill)
		weather.GET("/warnings", h.ListWarnings)
		weather.GET("/warnings/sync", h.SyncWarnings)
		weather.GET("/warnings/active", h.GetActiveWarnings)
		weather.GET("/warnings/:id", h.GetWarning)
		weather.GET("/warnings/:id/affected-routes", h.GetAffectedRoutes)
		weather.GET("/warnings/:id/routes", h.GetAffectedRoutes)
		weather.POST("/warnings/:id/replan", h.ReplanAffectedRoutes)

		weather.GET("/historical", h.GetHistoricalWeather)

		weather.POST("/push", h.PushWeatherWarning)
		weather.GET("/push/records", h.ListPushRecords)
		weather.GET("/push/records/:id", h.GetPushRecord)

		weather.POST("/operation/suspend", h.SuspendOperation)
		weather.POST("/operation/resume", h.ResumeOperation)
		weather.GET("/operation/suspensions", h.ListSuspensions)
		weather.GET("/operation/current-suspension", h.GetCurrentSuspension)
		weather.POST("/operation/auto-check", h.AutoCheckSuspend)

		driver := weather.Group("/driver")
		{
			driver.GET("/unread-count", h.GetDriverUnreadCount)
			driver.POST("/push/:id/read", h.MarkPushRead)
			driver.POST("/push/:id/respond", h.RespondToPush)
			driver.POST("/pre-departure", h.PreDepartureWarning)
			driver.POST("/en-route", h.EnRouteWarning)
		}
	}
}

func (h *WeatherHandler) GetCurrentWeather(c context.Context, ctx *app.RequestContext) {
	latStr := ctx.Query("lat")
	lngStr := ctx.Query("lng")

	if latStr == "" || lngStr == "" {
		response.BadRequest(ctx, "latitude and longitude are required")
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid latitude")
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid longitude")
		return
	}

	weather, err := h.weatherService.GetCurrentWeather(c, lat, lng)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, weather)
}

func (h *WeatherHandler) GetRouteWeather(c context.Context, ctx *app.RequestContext) {
	routeIDStr := ctx.Param("route_id")
	routeID, err := strconv.ParseInt(routeIDStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid route id")
		return
	}

	weatherPoints, err := h.weatherService.GetRouteWeather(c, routeID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, weatherPoints)
}

func (h *WeatherHandler) AnalyzeRouteWeather(c context.Context, ctx *app.RequestContext) {
	routeIDStr := ctx.Param("route_id")
	routeID, err := strconv.ParseInt(routeIDStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid route id")
		return
	}

	analysis, err := h.weatherService.AnalyzeRouteWeather(c, routeID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, analysis)
}

func (h *WeatherHandler) AnalyzeRouteByWaybill(c context.Context, ctx *app.RequestContext) {
	waybillIDStr := ctx.Param("waybill_id")
	waybillID, err := strconv.ParseInt(waybillIDStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid waybill id")
		return
	}

	var waybill struct {
		RoutePlanID int64 `json:"route_plan_id" gorm:"column:route_plan_id"`
	}
	if err := h.weatherService.DB().WithContext(c).Table("waybills").
		Where("id = ?", waybillID).Select("route_plan_id").First(&waybill).Error; err != nil {
		response.InternalError(ctx, "query waybill failed: "+err.Error())
		return
	}

	if waybill.RoutePlanID <= 0 {
		response.BadRequest(ctx, "waybill has no route plan")
		return
	}

	analysis, err := h.weatherService.AnalyzeRouteWeather(c, waybill.RoutePlanID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, analysis)
}

func (h *WeatherHandler) SyncWarnings(c context.Context, ctx *app.RequestContext) {
	province := ctx.Query("province")

	count, err := h.weatherService.SyncWarningsFromAPI(c, province)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"success":      true,
		"synced_count": count,
		"new_count":    count,
	})
}

func (h *WeatherHandler) ListWarnings(c context.Context, ctx *app.RequestContext) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	status := ctx.Query("status")

	result, err := h.weatherService.ListWarnings(c, page, pageSize, status)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Page(ctx, result.List, result.Total, result.Page, result.PageSize)
}

func (h *WeatherHandler) GetActiveWarnings(c context.Context, ctx *app.RequestContext) {
	province := ctx.Query("province")

	warnings, err := h.weatherService.GetActiveWarnings(c, province)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, warnings)
}

func (h *WeatherHandler) GetWarning(c context.Context, ctx *app.RequestContext) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid warning id")
		return
	}

	warning, err := h.weatherService.GetWarning(c, id)
	if err != nil {
		if err.Error() == "warning not found" {
			response.NotFound(ctx, err.Error())
			return
		}
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, warning)
}

func (h *WeatherHandler) GetAffectedRoutes(c context.Context, ctx *app.RequestContext) {
	routeIDStr := ctx.Query("route_id")
	warningIDStr := ctx.Param("id")

	warningID, err := strconv.ParseInt(warningIDStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid warning id")
		return
	}

	routeID, err := strconv.ParseInt(routeIDStr, 10, 64)
	if err != nil {
		routeID = 0
	}

	affected, affectedWaybills, err := h.weatherService.CheckRouteAffected(c, routeID, warningID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	type affectedRouteItem struct {
		WaybillID    int64  `json:"waybill_id"`
		WaybillNo    string `json:"waybill_no"`
		VehiclePlate string `json:"vehicle_plate"`
		Affected     bool   `json:"affected"`
	}

	var results []affectedRouteItem
	for _, wbID := range affectedWaybills {
		var wb struct {
			WaybillNo string `json:"waybill_no" gorm:"column:waybill_no"`
		}
		var v struct {
			PlateNumber string `json:"plate_number" gorm:"column:plate_number"`
		}
		h.weatherService.DB().WithContext(c).Table("waybills").Where("id = ?", wbID).Select("waybill_no").First(&wb)
		h.weatherService.DB().WithContext(c).Table("vehicles v").
			Joins("JOIN waybills w ON w.vehicle_id = v.id").
			Where("w.id = ?", wbID).Select("v.plate_number").First(&v)
		results = append(results, affectedRouteItem{
			WaybillID:    wbID,
			WaybillNo:    wb.WaybillNo,
			VehiclePlate: v.PlateNumber,
			Affected:     affected,
		})
	}

	response.Success(ctx, results)
}

func (h *WeatherHandler) ReplanAffectedRoutes(c context.Context, ctx *app.RequestContext) {
	warningIDStr := ctx.Param("id")
	warningID, err := strconv.ParseInt(warningIDStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid warning id")
		return
	}

	count, err := h.weatherService.ReplanAffectedRoutes(c, warningID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"success":         true,
		"replanned_count": count,
	})
}

func (h *WeatherHandler) GetHistoricalWeather(c context.Context, ctx *app.RequestContext) {
	var query model.HistoricalWeatherQuery
	if err := ctx.BindAndValidate(&query); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	results, err := h.weatherService.GetHistoricalWeather(c, &query)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, results)
}

func (h *WeatherHandler) PushWeatherWarning(c context.Context, ctx *app.RequestContext) {
	var req model.WeatherPushRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	record, err := h.weatherService.PushWeatherWarning(c, &req)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"success":      true,
		"push_id":      record.PushID,
		"sent_count":   record.SuccessCount,
		"failed_count": record.FailCount,
	})
}

func (h *WeatherHandler) ListPushRecords(c context.Context, ctx *app.RequestContext) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	waybillID, _ := strconv.ParseInt(ctx.Query("waybill_id"), 10, 64)
	driverID, _ := strconv.ParseInt(ctx.Query("driver_id"), 10, 64)
	phase := ctx.Query("phase")

	result, err := h.weatherService.ListPushRecords(c, page, pageSize, waybillID, driverID, phase)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Page(ctx, result.List, result.Total, result.Page, result.PageSize)
}

func (h *WeatherHandler) GetPushRecord(c context.Context, ctx *app.RequestContext) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid push record id")
		return
	}

	var record model.WeatherPushRecord
	if err := h.weatherService.DB().WithContext(c).Where("id = ?", id).First(&record).Error; err != nil {
		if err.Error() == "record not found" {
			response.NotFound(ctx, "push record not found")
			return
		}
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, record)
}

func (h *WeatherHandler) SuspendOperation(c context.Context, ctx *app.RequestContext) {
	var req model.OperationSuspendRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	operatorID := int64(0)
	operatorName := "system"

	suspension, err := h.weatherService.TriggerOperationSuspend(c, &req, operatorID, operatorName)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"success":    true,
		"suspension": suspension,
	})
}

func (h *WeatherHandler) ResumeOperation(c context.Context, ctx *app.RequestContext) {
	var req model.OperationResumeRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	operatorID := int64(0)
	operatorName := "system"

	suspension, err := h.weatherService.ResumeOperation(c, &req, operatorID, operatorName)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"success":       true,
		"suspension_id": suspension.ID,
	})
}

func (h *WeatherHandler) ListSuspensions(c context.Context, ctx *app.RequestContext) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	status := ctx.Query("status")

	result, err := h.weatherService.ListSuspensions(c, page, pageSize, status)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Page(ctx, result.List, result.Total, result.Page, result.PageSize)
}

func (h *WeatherHandler) GetCurrentSuspension(c context.Context, ctx *app.RequestContext) {
	suspension, err := h.weatherService.GetCurrentSuspension(c)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, suspension)
}

func (h *WeatherHandler) AutoCheckSuspend(c context.Context, ctx *app.RequestContext) {
	triggered, suspension, err := h.weatherService.CheckAndAutoSuspend(c)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"triggered":  triggered,
		"suspension": suspension,
	})
}

func (h *WeatherHandler) GetDriverUnreadCount(c context.Context, ctx *app.RequestContext) {
	driverIDStr := ctx.Query("driver_id")
	if driverIDStr == "" {
		response.BadRequest(ctx, "driver_id is required")
		return
	}

	driverID, err := strconv.ParseInt(driverIDStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid driver_id")
		return
	}

	count, err := h.weatherService.GetDriverUnreadCount(c, driverID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"unread_count": count,
	})
}

func (h *WeatherHandler) MarkPushRead(c context.Context, ctx *app.RequestContext) {
	pushID := ctx.Param("id")
	if pushID == "" {
		response.BadRequest(ctx, "push id is required")
		return
	}

	driverIDStr := ctx.Query("driver_id")
	driverID := int64(0)
	if driverIDStr != "" {
		driverID, _ = strconv.ParseInt(driverIDStr, 10, 64)
	}

	if err := h.weatherService.MarkPushRecordRead(c, pushID, driverID); err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, nil)
}

func (h *WeatherHandler) RespondToPush(c context.Context, ctx *app.RequestContext) {
	pushID := ctx.Param("id")
	if pushID == "" {
		response.BadRequest(ctx, "push id is required")
		return
	}

	var req struct {
		DriverID int64  `json:"driver_id"`
		Response string `json:"response"`
		Note     string `json:"note"`
	}
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, "invalid request body")
		return
	}

	if req.Response == "" {
		response.BadRequest(ctx, "response is required")
		return
	}

	if err := h.weatherService.RespondToPushRecord(c, pushID, req.DriverID, req.Response, req.Note); err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, nil)
}

func (h *WeatherHandler) PreDepartureWarning(c context.Context, ctx *app.RequestContext) {
	var req struct {
		WaybillID int64 `json:"waybill_id"`
	}
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, "invalid request body")
		return
	}

	if req.WaybillID <= 0 {
		response.BadRequest(ctx, "waybill_id is required")
		return
	}

	record, err := h.weatherService.PreDepartureWarning(c, req.WaybillID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, record)
}

func (h *WeatherHandler) EnRouteWarning(c context.Context, ctx *app.RequestContext) {
	var req struct {
		WaybillID   int64   `json:"waybill_id"`
		CurrentLat  float64 `json:"current_lat"`
		CurrentLng  float64 `json:"current_lng"`
	}
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, "invalid request body")
		return
	}

	if req.WaybillID <= 0 {
		response.BadRequest(ctx, "waybill_id is required")
		return
	}
	if req.CurrentLat == 0 || req.CurrentLng == 0 {
		response.BadRequest(ctx, "current location is required")
		return
	}

	record, err := h.weatherService.EnRouteWarning(c, req.WaybillID, req.CurrentLat, req.CurrentLng)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, record)
}
