package http

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

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
		weather.GET("/warnings", h.ListWarnings)
		weather.GET("/warnings/active", h.GetActiveWarnings)
		weather.GET("/warnings/:id", h.GetWarning)
		weather.GET("/warnings/:id/affected-routes", h.GetAffectedRoutes)
		weather.POST("/warnings/:id/replan", h.ReplanAffectedRoutes)
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
		response.BadRequest(ctx, "invalid route id")
		return
	}

	affected, affectedWaybills, err := h.weatherService.CheckRouteAffected(c, routeID, warningID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"affected":         affected,
		"affected_waybills": affectedWaybills,
	})
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
		"replan_count": count,
	})
}
