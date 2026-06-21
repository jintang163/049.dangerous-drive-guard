package http

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	escortSvc "github.com/dangerous-drive-guard/backend/internal/escort/service"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

var geoFenceService *escortSvc.GeoFenceService

func initGeoFenceService() {
	if geoFenceService == nil {
		geoFenceService = escortSvc.NewGeoFenceService()
	}
}

func CheckGeoFenceDeviation(ctx context.Context, c *app.RequestContext) {
	initGeoFenceService()
	var req model.GeoFenceCheckRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := geoFenceService.CheckDeviation(ctx, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func GetGeoFenceAlerts(ctx context.Context, c *app.RequestContext) {
	initGeoFenceService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	vehicleID, _ := strconv.ParseInt(c.DefaultQuery("vehicle_id", "0"), 10, 64)
	waybillID, _ := strconv.ParseInt(c.DefaultQuery("waybill_id", "0"), 10, 64)
	escortID, _ := strconv.ParseInt(c.DefaultQuery("escort_id", "0"), 10, 64)
	status := model.GeoFenceAlertStatus(c.Query("status"))

	alerts, total, err := geoFenceService.ListAlerts(ctx, &model.GeoFenceListRequest{
		VehicleID: vehicleID,
		WaybillID: waybillID,
		EscortID:  escortID,
		Status:    status,
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, alerts, total, page, pageSize)
}

func ConfirmGeoFenceAlert(ctx context.Context, c *app.RequestContext) {
	initGeoFenceService()
	var req model.GeoFenceConfirmRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	confirmerID, _ := c.Get("user_id")
	confirmerName, _ := c.Get("real_name")
	role, _ := c.Get("role")

	alert, err := geoFenceService.ConfirmAlert(ctx, &req,
		toInt64(confirmerID), toString(confirmerName), toString(role))

	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, alert)
}

func ResolveGeoFenceAlert(ctx context.Context, c *app.RequestContext) {
	initGeoFenceService()
	var req model.GeoFenceResolveRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resolverID, _ := c.Get("user_id")
	resolverName, _ := c.Get("real_name")
	resolverNameStr := resolverName.(string)

	if err := geoFenceService.ResolveAlert(ctx, &req, toInt64(resolverID), resolverNameStr); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"resolved": true})
}

func GetGeoFenceConfirmLogs(ctx context.Context, c *app.RequestContext) {
	initGeoFenceService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	alertID, _ := strconv.ParseInt(c.DefaultQuery("alert_id", "0"), 10, 64)
	vehicleID, _ := strconv.ParseInt(c.DefaultQuery("vehicle_id", "0"), 10, 64)

	logs, total, err := geoFenceService.ListConfirmLogs(ctx, alertID, vehicleID, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, logs, total, page, pageSize)
}

func GetGeoFenceStats(ctx context.Context, c *app.RequestContext) {
	initGeoFenceService()
	orgID, _ := strconv.ParseInt(c.DefaultQuery("org_id", "1"), 10, 64)
	stats, err := geoFenceService.GetStatistics(ctx, orgID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}
