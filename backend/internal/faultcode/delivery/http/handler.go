package http

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"

	fcSvc "github.com/dangerous-drive-guard/backend/internal/faultcode/service"
	"github.com/dangerous-drive-guard/backend/pkg/middleware"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

type FaultCodeHandler struct {
	svc *fcSvc.FaultCodeService
}

func NewFaultCodeHandler(svc *fcSvc.FaultCodeService) *FaultCodeHandler {
	return &FaultCodeHandler{svc: svc}
}

func (h *FaultCodeHandler) RegisterRoutes(r *app.RouterGroup, authMiddleware app.HandlerFunc) {
	faultCodes := r.Group("/fault-codes", authMiddleware)
	{
		faultCodes.GET("", h.List)
		faultCodes.GET("/stats", h.GetStats)
		faultCodes.GET("/by-code", h.GetByCode)
		faultCodes.GET("/:id", h.Get)
		faultCodes.POST("", h.Create)
		faultCodes.PUT("/:id", h.Update)
		faultCodes.DELETE("/:id", middleware.RoleAuth("admin"), h.Delete)
		faultCodes.POST("/:id/status", h.SetStatus)
		faultCodes.POST("/batch-import", h.BatchImport)
	}
}

func (h *FaultCodeHandler) List(c context.Context, ctx *app.RequestContext) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	system := ctx.Query("system")
	category := ctx.Query("category")
	level, _ := strconv.Atoi(ctx.DefaultQuery("level", "0"))
	status, _ := strconv.Atoi(ctx.DefaultQuery("status", "-1"))
	keyword := ctx.Query("keyword")

	list, total, err := h.svc.ListFaultCodes(c, system, category, level, status, keyword, page, pageSize)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Page(ctx, list, total, page, pageSize)
}

func (h *FaultCodeHandler) Get(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid fault code id")
		return
	}

	fc, err := h.svc.GetFaultCode(c, id)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, fc)
}

func (h *FaultCodeHandler) GetByCode(c context.Context, ctx *app.RequestContext) {
	code := ctx.Query("code")
	if code == "" {
		response.BadRequest(ctx, "code is required")
		return
	}

	fc, err := h.svc.GetByCode(c, code)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, fc)
}

func (h *FaultCodeHandler) GetStats(c context.Context, ctx *app.RequestContext) {
	stats, err := h.svc.GetStats(c)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, stats)
}

func (h *FaultCodeHandler) Create(c context.Context, ctx *app.RequestContext) {
	var req fcSvc.FaultCode
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	userID, _ := ctx.Get("user_id")
	if uid, ok := userID.(int64); ok {
		req.CreatedBy = uid
	}

	fc, err := h.svc.CreateFaultCode(c, &req)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, fc)
}

func (h *FaultCodeHandler) Update(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid fault code id")
		return
	}

	var req fcSvc.FaultCode
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	req.ID = id
	fc, err := h.svc.UpdateFaultCode(c, &req)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, fc)
}

func (h *FaultCodeHandler) Delete(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid fault code id")
		return
	}

	err = h.svc.DeleteFaultCode(c, id)
	if err != nil {
		if err.Error() == "builtin fault code cannot be deleted" {
			response.BadRequest(ctx, err.Error())
		} else {
			response.InternalError(ctx, err.Error())
		}
		return
	}

	response.Success(ctx, utils.H{"deleted": true})
}

type SetStatusReq struct {
	Status int `json:"status" binding:"required"`
}

func (h *FaultCodeHandler) SetStatus(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid fault code id")
		return
	}

	var req SetStatusReq
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	err = h.svc.SetStatus(c, id, req.Status)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, utils.H{"updated": true})
}

type BatchImportReq struct {
	Codes []*fcSvc.FaultCode `json:"codes" binding:"required"`
}

func (h *FaultCodeHandler) BatchImport(c context.Context, ctx *app.RequestContext) {
	var req BatchImportReq
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	userID, _ := ctx.Get("user_id")
	if uid, ok := userID.(int64); ok {
		for _, fc := range req.Codes {
			fc.CreatedBy = uid
		}
	}

	success, failed, err := h.svc.BatchImport(c, req.Codes)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, utils.H{
		"success": success,
		"failed":  failed,
		"total":   len(req.Codes),
	})
}
