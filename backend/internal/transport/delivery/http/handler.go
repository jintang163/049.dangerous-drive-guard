package http

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/internal/transport/service"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

var transportSvc *service.TransportService

func initService() {
	if transportSvc == nil {
		transportSvc = service.NewTransportService(config.Global)
	}
}

func CreateWaybill(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.WaybillCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	waybill, err := transportSvc.CreateWaybill(ctx, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, waybill)
}

func GetWaybill(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid waybill id")
		return
	}
	waybill, err := transportSvc.GetWaybill(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, waybill)
}

func ListWaybills(ctx context.Context, c *app.RequestContext) {
	initService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	carrierID, _ := strconv.ParseInt(c.DefaultQuery("carrier_org_id", "0"), 10, 64)
	status := model.WaybillStatus(c.Query("status"))

	waybills, total, err := transportSvc.ListWaybills(ctx, carrierID, status, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, waybills, total, page, pageSize)
}

func UpdateWaybillStatus(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid waybill id")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	waybill, err := transportSvc.UpdateWaybillStatus(ctx, id, model.WaybillStatus(req.Status), toInt64(userID), toString(role), req.Remark)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, waybill)
}

func SaveToBlockchain(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid waybill id")
		return
	}
	txHash, err := transportSvc.SaveToBlockchain(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{
		"success": true,
		"tx_hash": txHash,
	})
}

func VerifyFromBlockchain(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid waybill id")
		return
	}
	verified, info, err := transportSvc.VerifyFromBlockchain(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{
		"verified": verified,
		"info":     info,
	})
}

func StartEscort(ctx context.Context, c *app.RequestContext) {
	initService()
	var req struct {
		WaybillID int64 `json:"waybill_id" binding:"required"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	err := transportSvc.StartEscort(ctx, req.WaybillID, toInt64(userID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"started": true})
}

func GetEscortInfo(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("waybill_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid waybill id")
		return
	}
	info, err := transportSvc.GetEscortInfo(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, info)
}

func ReportEscortEvent(ctx context.Context, c *app.RequestContext) {
	initService()
	var req service.EscortEventReport
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	err := transportSvc.ReportEscortEvent(ctx, &req, toInt64(userID), toString(role))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"reported": true})
}

func RecommendServiceAreas(ctx context.Context, c *app.RequestContext) {
	initService()
	lat, _ := strconv.ParseFloat(c.DefaultQuery("lat", "0"), 64)
	lng, _ := strconv.ParseFloat(c.DefaultQuery("lng", "0"), 64)
	fatigueLevel := c.DefaultQuery("fatigue_level", "normal")
	areas, err := transportSvc.RecommendServiceAreas(ctx, lat, lng, fatigueLevel)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{
		"list":  areas,
		"total": len(areas),
	})
}

func GetWeatherWarning(ctx context.Context, c *app.RequestContext) {
	initService()
	lat, _ := strconv.ParseFloat(c.DefaultQuery("lat", "0"), 64)
	lng, _ := strconv.ParseFloat(c.DefaultQuery("lng", "0"), 64)
	warnings, err := transportSvc.GetWeatherWarning(ctx, lat, lng)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, warnings)
}

func ReportSOS(ctx context.Context, c *app.RequestContext) {
	initService()
	var req service.SOSRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	rescue, err := transportSvc.ReportSOS(ctx, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, rescue)
}

func ListRescueResources(ctx context.Context, c *app.RequestContext) {
	initService()
	lat, _ := strconv.ParseFloat(c.DefaultQuery("lat", "0"), 64)
	lng, _ := strconv.ParseFloat(c.DefaultQuery("lng", "0"), 64)
	resourceType := c.Query("type")
	radius, _ := strconv.ParseFloat(c.DefaultQuery("radius_km", "0"), 64)
	resources, err := transportSvc.ListRescueResources(ctx, lat, lng, resourceType, radius)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{
		"list":  resources,
		"total": len(resources),
	})
}

func DispatchRescue(ctx context.Context, c *app.RequestContext) {
	initService()
	var req service.RescueDispatch
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	dispatcherID, _ := c.Get("user_id")
	req.DispatcherID = toInt64(dispatcherID)
	err := transportSvc.DispatchRescue(ctx, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"dispatched": true})
}

func toInt64(v interface{}) int64 {
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

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
