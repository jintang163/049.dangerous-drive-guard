package http

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/internal/route/service"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

var routeService *service.RouteService

func initService() {
	if routeService == nil {
		routeService = service.NewRouteService(config.Global)
	}
}

func PlanRoute(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.RoutePlanRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	plan, err := routeService.PlanRoute(ctx, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, plan)
}

func PlanMultiStrategy(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.RoutePlanRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := routeService.PlanMultiStrategy(ctx, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func GetRoute(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid route id")
		return
	}
	_ = id
	response.Success(c, map[string]interface{}{"id": id, "status": "active"})
}

func ListRoutes(ctx context.Context, c *app.RequestContext) {
	initService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	response.Page(c, []interface{}{}, 0, page, pageSize)
}

func ReplanRoute(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid route id")
		return
	}
	var req model.RoutePlanRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	plan, err := routeService.PlanRoute(ctx, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	plan.ID = id
	response.Success(c, plan)
}

func ListRestrictedAreas(ctx context.Context, c *app.RequestContext) {
	initService()
	hazardClass := c.Query("hazard_class")
	vehicleType := c.Query("vehicle_type")
	areas, err := routeService.GetRestrictedAreas(ctx, hazardClass, model.VehicleType(vehicleType))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{
		"list":  areas,
		"total": len(areas),
	})
}
