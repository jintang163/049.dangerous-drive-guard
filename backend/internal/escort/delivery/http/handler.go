package http

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	escortSvc "github.com/dangerous-drive-guard/backend/internal/escort/service"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

var escortService *escortSvc.EscortService

func initService() {
	if escortService == nil {
		escortService = escortSvc.NewEscortService()
	}
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

func parseTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}

func CreateShift(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.EscortShiftCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	dispatcherID, _ := c.Get("user_id")
	dispatcherName, _ := c.Get("real_name")

	shift, err := escortService.CreateShift(ctx, &req, toInt64(dispatcherID), toString(dispatcherName))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, shift)
}

func GetShift(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid shift id")
		return
	}
	shift, err := escortService.GetShift(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, shift)
}

func ListShifts(ctx context.Context, c *app.RequestContext) {
	initService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	escortID, _ := strconv.ParseInt(c.DefaultQuery("escort_id", "0"), 10, 64)
	dispatcherID, _ := strconv.ParseInt(c.DefaultQuery("dispatcher_id", "0"), 10, 64)
	status := model.EscortShiftStatus(c.Query("status"))

	shifts, total, err := escortService.ListShifts(ctx, escortID, dispatcherID, status, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, shifts, total, page, pageSize)
}

func UpdateShiftStatus(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid shift id")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	err = escortService.UpdateShiftStatus(ctx, id, model.EscortShiftStatus(req.Status), toInt64(userID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"updated": true})
}

func AssignVehicles(ctx context.Context, c *app.RequestContext) {
	initService()
	shiftID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid shift id")
		return
	}
	var req struct {
		VehicleIDs []int64 `json:"vehicle_ids" binding:"required"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	dispatcherID, _ := c.Get("user_id")
	err = escortService.AssignVehicles(ctx, shiftID, req.VehicleIDs, toInt64(dispatcherID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"assigned": true})
}

func GetShiftAssignments(ctx context.Context, c *app.RequestContext) {
	initService()
	shiftID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid shift id")
		return
	}
	assignments, err := escortService.GetShiftAssignments(ctx, shiftID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{
		"list":  assignments,
		"total": len(assignments),
	})
}

func ReportSOS(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.EscortSOSReportRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	alert, err := escortService.ReportSOS(ctx, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, alert)
}

func GetSOSAlerts(ctx context.Context, c *app.RequestContext) {
	initService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	vehicleID, _ := strconv.ParseInt(c.DefaultQuery("vehicle_id", "0"), 10, 64)
	escortID, _ := strconv.ParseInt(c.DefaultQuery("escort_id", "0"), 10, 64)
	status := model.EscortSOSStatus(c.Query("status"))

	alerts, total, err := escortService.GetSOSAlerts(ctx, vehicleID, escortID, status, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, alerts, total, page, pageSize)
}

func HandleSOS(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.EscortSOSHandleRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	handlerID, _ := c.Get("user_id")
	handlerName, _ := c.Get("real_name")
	err := escortService.HandleSOS(ctx, &req, toInt64(handlerID), toString(handlerName))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"handled": true})
}

func ResolveSOS(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid alert id")
		return
	}
	var req struct {
		Note string `json:"note"`
	}
	c.BindAndValidate(&req)
	handlerID, _ := c.Get("user_id")
	handlerName, _ := c.Get("real_name")
	err = escortService.ResolveSOS(ctx, id, toInt64(handlerID), toString(handlerName), req.Note)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"resolved": true})
}

func GetTrackPlayback(ctx context.Context, c *app.RequestContext) {
	initService()
	waybillID, _ := strconv.ParseInt(c.DefaultQuery("waybill_id", "0"), 10, 64)
	vehicleID, _ := strconv.ParseInt(c.DefaultQuery("vehicle_id", "0"), 10, 64)

	startTime := parseTime(c.Query("start_time"))
	endTime := parseTime(c.Query("end_time"))

	if waybillID == 0 && vehicleID == 0 {
		response.BadRequest(c, "waybill_id or vehicle_id is required")
		return
	}

	tracks, err := escortService.GetTrackPlayback(ctx, waybillID, vehicleID, startTime, endTime)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{
		"list":  tracks,
		"total": len(tracks),
	})
}

func GetVideoRecords(ctx context.Context, c *app.RequestContext) {
	initService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	vehicleID, _ := strconv.ParseInt(c.DefaultQuery("vehicle_id", "0"), 10, 64)
	waybillID, _ := strconv.ParseInt(c.DefaultQuery("waybill_id", "0"), 10, 64)
	recordType := c.Query("record_type")
	startTime := parseTime(c.Query("start_time"))
	endTime := parseTime(c.Query("end_time"))

	records, total, err := escortService.GetVideoRecords(ctx, vehicleID, waybillID, recordType, startTime, endTime, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, records, total, page, pageSize)
}

func ViewVideoRecord(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid record id")
		return
	}
	err = escortService.IncrementVideoViewCount(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"viewed": true})
}

func SendIntercom(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.EscortIntercomRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	senderID, _ := c.Get("user_id")
	senderName, _ := c.Get("real_name")
	role, _ := c.Get("role")
	err := escortService.SendIntercom(ctx, &req, toInt64(senderID), toString(senderName), toString(role))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"sent": true})
}

func GetIntercomLogs(ctx context.Context, c *app.RequestContext) {
	initService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	vehicleID, _ := strconv.ParseInt(c.DefaultQuery("vehicle_id", "0"), 10, 64)

	logs, total, err := escortService.GetIntercomLogs(ctx, vehicleID, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, logs, total, page, pageSize)
}

func StartPollingSession(ctx context.Context, c *app.RequestContext) {
	initService()
	escortID, _ := c.Get("user_id")
	escortName, _ := c.Get("real_name")
	shiftID, _ := strconv.ParseInt(c.DefaultQuery("shift_id", "0"), 10, 64)

	session, err := escortService.StartPollingSession(ctx, toInt64(escortID), shiftID, toString(escortName))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, session)
}

func EndPollingSession(ctx context.Context, c *app.RequestContext) {
	initService()
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid session id")
		return
	}
	var req struct {
		PollingCount int `json:"polling_count"`
	}
	c.BindAndValidate(&req)
	err = escortService.EndPollingSession(ctx, sessionID, req.PollingCount)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"ended": true})
}

func GetEscortVehiclesForPolling(ctx context.Context, c *app.RequestContext) {
	initService()
	escortID, _ := strconv.ParseInt(c.DefaultQuery("escort_id", "0"), 10, 64)
	if escortID == 0 {
		uid, _ := c.Get("user_id")
		escortID = toInt64(uid)
	}
	vehicles, err := escortService.GetEscortVehiclesForPolling(ctx, escortID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{
		"list":  vehicles,
		"total": len(vehicles),
	})
}

func GetEscortStatistics(ctx context.Context, c *app.RequestContext) {
	initService()
	orgID, _ := strconv.ParseInt(c.DefaultQuery("org_id", "1"), 10, 64)
	stats, err := escortService.GetStatistics(ctx, orgID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}
