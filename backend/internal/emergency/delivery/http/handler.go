package http

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	emSvc "github.com/dangerous-drive-guard/backend/internal/emergency/service"
	"github.com/dangerous-drive-guard/backend/pkg/middleware"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

type EmergencyHandler struct {
	svc *emSvc.EmergencyService
}

func NewEmergencyHandler(svc *emSvc.EmergencyService) *EmergencyHandler {
	return &EmergencyHandler{svc: svc}
}

func (h *EmergencyHandler) RegisterRoutes(r *app.RouterGroup, authMiddleware app.HandlerFunc) {
	emergency := r.Group("/emergency", authMiddleware)
	{
		plans := emergency.Group("/plans")
		{
			plans.GET("", h.ListPlans)
			plans.GET("/search", h.SearchByUNNumber)
			plans.GET("/:id", h.GetPlan)
			plans.POST("", middleware.RoleAuth("admin", "dispatcher"), h.CreatePlan)
			plans.PUT("/:id", middleware.RoleAuth("admin", "dispatcher"), h.UpdatePlan)
			plans.DELETE("/:id", middleware.RoleAuth("admin"), h.DeletePlan)
		}
		cards := emergency.Group("/task-cards")
		{
			cards.GET("", h.ListTaskCards)
			cards.POST("/generate", middleware.RoleAuth("admin", "dispatcher"), h.GenerateTaskCard)
			cards.GET("/:id", h.GetTaskCard)
			cards.POST("/:id/ack", h.AckTaskCard)
			cards.POST("/:id/complete", middleware.RoleAuth("admin", "dispatcher"), h.CompleteTaskCard)
			cards.POST("/:id/cancel", middleware.RoleAuth("admin", "dispatcher"), h.CancelTaskCard)
		}
		emergency.GET("/stats", h.GetStats)
	}
}

func (h *EmergencyHandler) ListPlans(c context.Context, ctx *app.RequestContext) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	unNumber := ctx.Query("un_number")
	dangerClass := ctx.Query("danger_class")
	keyword := ctx.Query("keyword")

	list, total, err := h.svc.ListPlans(c, unNumber, dangerClass, keyword, page, pageSize)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Page(ctx, list, total, page, pageSize)
}

func (h *EmergencyHandler) GetPlan(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid plan id")
		return
	}
	plan, err := h.svc.GetPlan(c, id)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, plan)
}

func (h *EmergencyHandler) CreatePlan(c context.Context, ctx *app.RequestContext) {
	var req emSvc.EmergencyPlan
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}
	plan, err := h.svc.CreatePlan(c, &req)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, plan)
}

func (h *EmergencyHandler) UpdatePlan(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid plan id")
		return
	}
	var req emSvc.EmergencyPlan
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}
	req.ID = id
	plan, err := h.svc.UpdatePlan(c, &req)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, plan)
}

func (h *EmergencyHandler) DeletePlan(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid plan id")
		return
	}
	err = h.svc.DeletePlan(c, id)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, map[string]bool{"deleted": true})
}

func (h *EmergencyHandler) SearchByUNNumber(c context.Context, ctx *app.RequestContext) {
	unNumber := ctx.Query("un_number")
	if unNumber == "" {
		response.BadRequest(ctx, "un_number is required")
		return
	}
	plans, err := h.svc.SearchByUNNumber(c, unNumber)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, plans)
}

func (h *EmergencyHandler) GenerateTaskCard(c context.Context, ctx *app.RequestContext) {
	var req emSvc.TaskCardGenerateData
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}
	operatorID := int64(ctx.GetUint64("userID"))
	card, err := h.svc.GenerateTaskCard(c, &req, operatorID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, card)
}

func (h *EmergencyHandler) ListTaskCards(c context.Context, ctx *app.RequestContext) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	vehicleID, _ := strconv.ParseInt(ctx.DefaultQuery("vehicle_id", "0"), 10, 64)
	driverID, _ := strconv.ParseInt(ctx.DefaultQuery("driver_id", "0"), 10, 64)
	status := ctx.Query("status")
	unNumber := ctx.Query("un_number")

	list, total, err := h.svc.ListTaskCards(c, vehicleID, driverID, status, unNumber, page, pageSize)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Page(ctx, list, total, page, pageSize)
}

func (h *EmergencyHandler) GetTaskCard(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid task card id")
		return
	}
	card, err := h.svc.GetTaskCard(c, id)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, card)
}

func (h *EmergencyHandler) AckTaskCard(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid task card id")
		return
	}
	card, err := h.svc.AckTaskCard(c, id)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, card)
}

func (h *EmergencyHandler) CompleteTaskCard(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid task card id")
		return
	}
	userID := int64(ctx.GetUint64("userID"))
	card, err := h.svc.CompleteTaskCard(c, id, userID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, card)
}

type CancelTaskCardReq struct {
	Remark string `json:"remark"`
}

func (h *EmergencyHandler) CancelTaskCard(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid task card id")
		return
	}
	var req CancelTaskCardReq
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}
	card, err := h.svc.CancelTaskCard(c, id, req.Remark)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, card)
}

func (h *EmergencyHandler) GetStats(c context.Context, ctx *app.RequestContext) {
	orgID, _ := strconv.ParseInt(ctx.DefaultQuery("org_id", "0"), 10, 64)
	stats, err := h.svc.GetStats(c, orgID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, stats)
}
