package http

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	fatigueSvc "github.com/dangerous-drive-guard/backend/internal/fatigue/service"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

var fatigueService *fatigueSvc.FatigueService

func initService() {
	if fatigueService == nil {
		fatigueService = fatigueSvc.NewFatigueService(config.Global)
	}
}

func DetectFatigue(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.FatigueDetectRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.DetectionTime.IsZero() {
		req.DetectionTime = time.Now()
	}

	resp, err := fatigueService.DetectFatigue(ctx, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp)
}

func UploadFrame(ctx context.Context, c *app.RequestContext) {
	initService()
	var req struct {
		VehicleID      int64  `json:"vehicle_id" form:"vehicle_id" binding:"required"`
		DriverID       int64  `json:"driver_id" form:"driver_id" binding:"required"`
		WaybillID      int64  `json:"waybill_id" form:"waybill_id"`
		FrameData      string `json:"frame_data" form:"frame_data"`
		Timestamp      int64  `json:"timestamp" form:"timestamp"`
		CameraPosition string `json:"camera_position" form:"camera_position"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	detectTime := time.Now()
	if req.Timestamp > 0 {
		detectTime = time.Unix(req.Timestamp/1000, 0)
	}

	detectReq := &model.FatigueDetectRequest{
		VehicleID:      req.VehicleID,
		DriverID:       req.DriverID,
		WaybillID:      req.WaybillID,
		ImageBase64:    req.FrameData,
		DetectionTime:  detectTime,
		EdgeComputed:   false,
		NetworkStatus:  "online",
		CameraPosition: model.CameraPosition(req.CameraPosition),
		EnableFusion:   false,
	}
	resp, err := fatigueService.DetectFatigue(ctx, detectReq)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp)
}

func UploadMultiCameraFrames(ctx context.Context, c *app.RequestContext) {
	initService()
	var req struct {
		VehicleID     int64                   `json:"vehicle_id" binding:"required"`
		DriverID      int64                   `json:"driver_id" binding:"required"`
		WaybillID     int64                   `json:"waybill_id"`
		Frames        []model.MultiCameraFrame `json:"frames" binding:"required"`
		Latitude      float64                 `json:"latitude"`
		Longitude     float64                 `json:"longitude"`
		VehicleSpeed  float64                 `json:"vehicle_speed"`
		Timestamp     int64                   `json:"timestamp"`
		EdgeComputed  bool                    `json:"edge_computed"`
		NetworkStatus string                  `json:"network_status"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if len(req.Frames) == 0 {
		response.BadRequest(c, "至少需要一个摄像头帧数据")
		return
	}

	detectTime := time.Now()
	if req.Timestamp > 0 {
		detectTime = time.Unix(req.Timestamp/1000, 0)
	}

	detectReq := &model.FatigueDetectRequest{
		VehicleID:     req.VehicleID,
		DriverID:      req.DriverID,
		WaybillID:     req.WaybillID,
		Frames:        req.Frames,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		VehicleSpeed:  req.VehicleSpeed,
		DetectionTime: detectTime,
		EdgeComputed:  req.EdgeComputed,
		NetworkStatus: req.NetworkStatus,
		EnableFusion:  true,
	}

	resp, err := fatigueService.DetectFatigue(ctx, detectReq)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp)
}

func GetMultiCameraHistory(ctx context.Context, c *app.RequestContext) {
	initService()
	vehicleID, err := strconv.ParseInt(c.Param("vehicle_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	cameraPosition := c.Query("camera_position")

	startTime := time.Time{}
	endTime := time.Time{}
	if s := c.Query("start_time"); s != "" {
		startTime, _ = time.Parse(time.RFC3339, s)
	}
	if e := c.Query("end_time"); e != "" {
		endTime, _ = time.Parse(time.RFC3339, e)
	}

	records, total, err := fatigueService.GetMultiCameraHistory(ctx, vehicleID, cameraPosition, startTime, endTime, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, records, total, page, pageSize)
}

func GetHistory(ctx context.Context, c *app.RequestContext) {
	initService()
	vehicleID, err := strconv.ParseInt(c.Param("vehicle_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	startTime := time.Time{}
	endTime := time.Time{}
	if s := c.Query("start_time"); s != "" {
		startTime, _ = time.Parse(time.RFC3339, s)
	}
	if e := c.Query("end_time"); e != "" {
		endTime, _ = time.Parse(time.RFC3339, e)
	}

	records, total, err := fatigueService.GetHistory(ctx, vehicleID, startTime, endTime, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, records, total, page, pageSize)
}

func ListAlarms(ctx context.Context, c *app.RequestContext) {
	initService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	vehicleID, _ := strconv.ParseInt(c.DefaultQuery("vehicle_id", "0"), 10, 64)
	status := model.AlarmStatus(c.Query("status"))
	level, _ := strconv.Atoi(c.DefaultQuery("level", "0"))

	alarms, total, err := fatigueService.ListAlarms(ctx, vehicleID, status, model.AlarmLevel(level), page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, alarms, total, page, pageSize)
}

func GetFusionAccuracyStats(ctx context.Context, c *app.RequestContext) {
	initService()
	days, _ := strconv.Atoi(c.DefaultQuery("days", "90"))
	if days <= 0 {
		days = 90
	}
	stats, err := fatigueService.GetFusionAccuracyStats(ctx, days)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}

func AckAlarm(ctx context.Context, c *app.RequestContext) {
	initService()
	alarmID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid alarm id")
		return
	}
	var req model.AlarmAckRequest
	req.AlarmID = alarmID
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	dispatcherID, _ := c.Get("user_id")

	alarm, err := fatigueService.AcknowledgeAlarm(ctx, alarmID, toInt64(dispatcherID), req.HandleType, req.HandleNote)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, alarm)
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
