package http

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	replanSvc "github.com/dangerous-drive-guard/backend/internal/replan/service"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

var (
	replanService *replanSvc.ReplanService
)

func initService() {
	if replanService == nil {
		replanService = replanSvc.NewReplanService()
	}
}

func getUserID(c *app.RequestContext) int64 {
	if v, ok := c.Get("user_id"); ok {
		if id, ok := v.(int64); ok {
			return id
		}
		if s, ok := v.(string); ok {
			id, _ := strconv.ParseInt(s, 10, 64)
			return id
		}
	}
	return 0
}

// ============================================================
// 实时路况事件接口
// ============================================================

func ListTrafficEvents(ctx context.Context, c *app.RequestContext) {
	initService()
	status := model.TrafficEventStatus(c.Query("status"))
	eventType := c.Query("event_type")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := replanService.ListTrafficEvents(ctx, status, eventType, keyword, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func GetTrafficEvent(ctx context.Context, c *app.RequestContext) {
	initService()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	evt, err := replanService.GetTrafficEvent(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, evt)
}

func CreateTrafficEvent(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.TrafficEventCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	userID := getUserID(c)
	evt, err := replanService.CreateTrafficEvent(ctx, &req, userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, evt)
}

func ResolveTrafficEvent(ctx context.Context, c *app.RequestContext) {
	initService()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := replanService.ResolveTrafficEvent(ctx, id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"success": true})
}

// ============================================================
// 重规划接口
// ============================================================

func TriggerReplan(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.ReplanTriggerRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	userID := getUserID(c)
	var operatorIDPtr *int64
	if userID > 0 {
		operatorIDPtr = &userID
	}
	record, err := replanService.TriggerReplan(ctx, &req, operatorIDPtr)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, record)
}

func ConfirmReplan(ctx context.Context, c *app.RequestContext) {
	initService()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.ReplanConfirmRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	userID := getUserID(c)
	var driverIDPtr *int64
	if userID > 0 {
		driverIDPtr = &userID
	}
	record, err := replanService.ConfirmReplan(ctx, id, req.Action, req.ConfirmNote, driverIDPtr)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, record)
}

func ListReplanRecords(ctx context.Context, c *app.RequestContext) {
	initService()
	waybillID, _ := strconv.ParseInt(c.Query("waybill_id"), 10, 64)
	vehicleID, _ := strconv.ParseInt(c.Query("vehicle_id"), 10, 64)
	driverID, _ := strconv.ParseInt(c.Query("driver_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	params := &model.ReplanQueryParams{
		WaybillID:   waybillID,
		VehicleID:   vehicleID,
		DriverID:    driverID,
		TriggerType: model.ReplanTriggerType(c.Query("trigger_type")),
		Status:      model.ReplanRecordStatus(c.Query("status")),
		Keyword:     c.Query("keyword"),
		StartDate:   c.Query("start_date"),
		EndDate:     c.Query("end_date"),
		Page:        page,
		PageSize:    pageSize,
	}

	list, total, err := replanService.ListReplanRecords(ctx, params)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func GetReplanRecord(ctx context.Context, c *app.RequestContext) {
	initService()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	record, err := replanService.GetReplanRecord(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, record)
}

func GetReplanStatistics(ctx context.Context, c *app.RequestContext) {
	initService()
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	orgID, _ := strconv.ParseInt(c.Query("org_id"), 10, 64)
	stats, err := replanService.GetReplanStatistics(ctx, orgID, days)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}
