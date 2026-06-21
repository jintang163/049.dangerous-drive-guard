package http

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	nightSvc "github.com/dangerous-drive-guard/backend/internal/fatigue/service"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

type NightVisionHandler struct {
	nightVisionService   *nightSvc.NightVisionService
	enhanceService    *nightSvc.ImageEnhancementService
}

func NewNightVisionHandler(
	nightSvc *nightSvc.NightVisionService,
	enhanceSvc *nightSvc.ImageEnhancementService,
) *NightVisionHandler {
	return &NightVisionHandler{
		nightVisionService: nightSvc,
		enhanceService:    enhanceSvc,
	}
}

func toInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

func toString(s interface{}) string {
	if s == nil {
		return ""
	}
	if v, ok := s.(string); ok {
		return v
	}
	return ""
}

func (h *NightVisionHandler) GetConfig(c context.Context, ctx *app.RequestContext) {
	vehicleID := toInt64(ctx.Query("vehicle_id"))
	if vehicleID <= 0 {
		response.BadRequest(ctx, "invalid vehicle_id")
		return
	}

	cfg, err := h.nightVisionService.GetConfig(c, vehicleID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, cfg)
}

func (h *NightVisionHandler) UpdateConfig(c context.Context, ctx *app.RequestContext) {
	var req model.NightVisionConfigUpdateRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, "invalid params: "+err.Error())
		return
	}

	cfg, err := h.nightVisionService.UpdateConfig(c, &req)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, cfg)
}

func (h *NightVisionHandler) ResetConfig(c context.Context, ctx *app.RequestContext) {
	vehicleID := toInt64(ctx.Param("vehicle_id"))
	if vehicleID <= 0 {
		response.BadRequest(ctx, "invalid vehicle_id")
		return
	}

	cfg, err := h.nightVisionService.ResetConfig(c, vehicleID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, cfg)
}

func (h *NightVisionHandler) ListConfigs(c context.Context, ctx *app.RequestContext) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	orgID := toInt64(ctx.Query("org_id"))

	configs, total, err := h.nightVisionService.ListConfigs(c, page, pageSize, orgID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"list":     configs,
		"total":    total,
		"page":     page,
		"page_size": pageSize,
	})
}

func (h *NightVisionHandler) ReportInfraredStatus(c context.Context, ctx *app.RequestContext) {
	vehicleID := toInt64(ctx.PostForm("vehicle_id"))
	if vehicleID <= 0 {
		response.BadRequest(ctx, "invalid vehicle_id")
		return
	}

	var status model.InfraredLightStatus
	if err := ctx.BindAndValidate(&status); err != nil {
		response.BadRequest(ctx, "invalid params: "+err.Error())
		return
	}

	err := h.nightVisionService.ReportInfraredStatus(c, vehicleID, &status)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"message": "status reported",
	})
}

func (h *NightVisionHandler) ListInfraredLogs(c context.Context, ctx *app.RequestContext) {
	vehicleID := toInt64(ctx.Query("vehicle_id"))
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := h.nightVisionService.ListInfraredLogs(c, vehicleID, page, pageSize)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"list":     logs,
		"total":    total,
		"page":     page,
		"page_size": pageSize,
	})
}

func (h *NightVisionHandler) EnhanceImage(c context.Context, ctx *app.RequestContext) {
	var req model.ImageEnhanceRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, "invalid params: "+err.Error())
		return
	}

	result, err := h.enhanceService.EnhanceImage(c, &req)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, result)
}

func (h *NightVisionHandler) ListEnhanceRecords(c context.Context, ctx *app.RequestContext) {
	vehicleID := toInt64(ctx.Query("vehicle_id"))
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	records, total, err := h.enhanceService.ListEnhanceRecords(c, vehicleID, page, pageSize)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"list":     records,
		"total":    total,
		"page":     page,
		"page_size": pageSize,
	})
}

func (h *NightVisionHandler) GetEnhanceRecord(c context.Context, ctx *app.RequestContext) {
	id := toInt64(ctx.Param("id"))
	if id <= 0 {
		response.BadRequest(ctx, "invalid id")
		return
	}

	record, err := h.enhanceService.GetEnhanceRecord(c, id)
	if err != nil {
		response.NotFound(ctx, "record not found")
		return
	}

	response.Success(ctx, record)
}

func (h *NightVisionHandler) GetStatistics(c context.Context, ctx *app.RequestContext) {
	orgID := toInt64(ctx.Query("org_id"))

	stats, err := h.nightVisionService.GetStatistics(c, orgID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, stats)
}

func (h *NightVisionHandler) UploadEnhanceRecord(c context.Context, ctx *app.RequestContext) {
	var record model.ImageEnhancementRecord
	if err := ctx.BindAndValidate(&record); err != nil {
		response.BadRequest(ctx, "invalid params: "+err.Error())
		return
	}

	err := h.enhanceService.SaveEnhanceRecord(c, &record)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"id":      record.ID,
		"message": "record saved",
	})
}
