package http

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	monitorSvc "github.com/dangerous-drive-guard/backend/internal/monitor/service"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

var monitorService *monitorSvc.MonitorService

func initService() {
	if monitorService == nil {
		monitorService = monitorSvc.NewMonitorService()
	}
}

func GetRealtimeVehicles(ctx context.Context, c *app.RequestContext) {
	initService()
	orgID, _ := strconv.ParseInt(c.DefaultQuery("org_id", "1"), 10, 64)
	statusFilter := c.Query("status")

	vehicles, err := monitorService.GetRealtimeVehicles(ctx, orgID, statusFilter)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{
		"list":  vehicles,
		"total": len(vehicles),
	})
}

func GetVehicleStatus(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}
	status, err := monitorService.GetVehicleStatus(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, status)
}

func GetStatistics(ctx context.Context, c *app.RequestContext) {
	initService()
	orgID, _ := strconv.ParseInt(c.DefaultQuery("org_id", "1"), 10, 64)
	stats, err := monitorService.GetStatistics(ctx, orgID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}

func SendVoiceIntercom(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.VoiceIntercomRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	err := monitorService.SendVoiceIntercom(ctx, req.VehicleID, req.Message, req.Priority)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"sent": true})
}

func DispatchServiceArea(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.DispatchServiceAreaRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.RestDuration == 0 {
		req.RestDuration = 20
	}
	err := monitorService.DispatchServiceArea(ctx, req.VehicleID, req.ServiceAreaID, req.Reason, req.RestDuration)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"dispatched": true})
}
