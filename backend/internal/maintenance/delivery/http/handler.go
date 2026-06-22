package http

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"

	mtSvc "github.com/dangerous-drive-guard/backend/internal/maintenance/service"
	"github.com/dangerous-drive-guard/backend/pkg/middleware"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

type MaintenanceHandler struct {
	svc *mtSvc.MaintenanceService
}

func NewMaintenanceHandler(svc *mtSvc.MaintenanceService) *MaintenanceHandler {
	return &MaintenanceHandler{svc: svc}
}

func (h *MaintenanceHandler) RegisterRoutes(r *app.RouterGroup, authMiddleware app.HandlerFunc) {
	maintenance := r.Group("/maintenance", authMiddleware)
	{
		plans := maintenance.Group("/plans")
		{
			plans.GET("", h.ListPlans)
			plans.GET("/:id", h.GetPlan)
			plans.POST("", middleware.RoleAuth("admin", "dispatcher"), h.CreatePlan)
			plans.PUT("/:id", middleware.RoleAuth("admin", "dispatcher"), h.UpdatePlan)
			plans.DELETE("/:id", middleware.RoleAuth("admin"), h.DeletePlan)
			plans.POST("/:id/status", middleware.RoleAuth("admin", "dispatcher"), h.SetPlanStatus)
			plans.POST("/:id/trigger", middleware.RoleAuth("admin", "dispatcher"), h.TriggerPlanCheck)
			plans.POST("/batch-check", middleware.RoleAuth("admin", "dispatcher"), h.BatchCheckAndGenerate)
		}

		orders := maintenance.Group("/work-orders")
		{
			orders.GET("", h.ListWorkOrders)
			orders.GET("/:id", h.GetWorkOrder)
			orders.POST("", middleware.RoleAuth("admin", "dispatcher"), h.CreateWorkOrder)
			orders.PUT("/:id", middleware.RoleAuth("admin", "dispatcher"), h.UpdateWorkOrder)
			orders.POST("/:id/assign", middleware.RoleAuth("admin", "dispatcher"), h.AssignWorkOrder)
			orders.POST("/:id/appointment", middleware.RoleAuth("admin", "dispatcher"), h.SetAppointment)
			orders.POST("/:id/checkin", middleware.RoleAuth("admin", "dispatcher"), h.Checkin)
			orders.POST("/:id/start", middleware.RoleAuth("admin", "dispatcher"), h.StartWork)
			orders.POST("/:id/complete", middleware.RoleAuth("admin", "dispatcher"), h.CompleteWork)
			orders.POST("/:id/cancel", middleware.RoleAuth("admin", "dispatcher"), h.CancelWorkOrder)
			orders.GET("/:id/logs", h.GetWorkOrderLogs)
		}

		maintenance.GET("/stats", h.GetStats)
		maintenance.GET("/upcoming", h.GetUpcomingMaintenance)
		maintenance.GET("/overdue", h.GetOverdueMaintenance)
	}
}

func (h *MaintenanceHandler) ListPlans(c context.Context, ctx *app.RequestContext) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	vehicleID, _ := strconv.ParseInt(ctx.DefaultQuery("vehicle_id", "0"), 10, 64)
	status := ctx.Query("status")
	mType := ctx.Query("type")

	list, total, err := h.svc.ListPlans(c, vehicleID, status, mType, page, pageSize)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Page(ctx, list, total, page, pageSize)
}

func (h *MaintenanceHandler) GetPlan(c context.Context, ctx *app.RequestContext) {
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

func (h *MaintenanceHandler) CreatePlan(c context.Context, ctx *app.RequestContext) {
	var req mtSvc.MaintenancePlan
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}
	userID, _ := ctx.Get("user_id")
	if uid, ok := userID.(int64); ok {
		req.CreatedBy = uid
	}
	plan, err := h.svc.CreatePlan(c, &req)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, plan)
}

func (h *MaintenanceHandler) UpdatePlan(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid plan id")
		return
	}
	var req mtSvc.MaintenancePlan
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

func (h *MaintenanceHandler) DeletePlan(c context.Context, ctx *app.RequestContext) {
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
	response.Success(ctx, utils.H{"deleted": true})
}

type SetPlanStatusReq struct {
	Status string `json:"status" binding:"required"`
}

func (h *MaintenanceHandler) SetPlanStatus(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid plan id")
		return
	}
	var req SetPlanStatusReq
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}
	err = h.svc.SetPlanStatus(c, id, req.Status)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, utils.H{"updated": true})
}

func (h *MaintenanceHandler) TriggerPlanCheck(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid plan id")
		return
	}
	userID, _ := ctx.Get("user_id")
	var uid int64
	if u, ok := userID.(int64); ok {
		uid = u
	}
	result, err := h.svc.CheckAndGenerateWorkOrder(c, id, uid)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

func (h *MaintenanceHandler) BatchCheckAndGenerate(c context.Context, ctx *app.RequestContext) {
	orgID, _ := strconv.ParseInt(ctx.DefaultQuery("org_id", "0"), 10, 64)
	userID, _ := ctx.Get("user_id")
	var uid int64
	if u, ok := userID.(int64); ok {
		uid = u
	}
	result, err := h.svc.BatchCheckVehicles(c, orgID, uid)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

func (h *MaintenanceHandler) ListWorkOrders(c context.Context, ctx *app.RequestContext) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	vehicleID, _ := strconv.ParseInt(ctx.DefaultQuery("vehicle_id", "0"), 10, 64)
	planID, _ := strconv.ParseInt(ctx.DefaultQuery("plan_id", "0"), 10, 64)
	status := ctx.Query("status")
	mType := ctx.Query("type")
	source := ctx.Query("source")

	list, total, err := h.svc.ListWorkOrders(c, vehicleID, planID, status, mType, source, page, pageSize)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Page(ctx, list, total, page, pageSize)
}

func (h *MaintenanceHandler) GetWorkOrder(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid work order id")
		return
	}
	order, err := h.svc.GetWorkOrder(c, id)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, order)
}

func (h *MaintenanceHandler) CreateWorkOrder(c context.Context, ctx *app.RequestContext) {
	var req mtSvc.WorkOrder
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}
	userID, _ := ctx.Get("user_id")
	if uid, ok := userID.(int64); ok {
		req.CreatedBy = uid
	}
	order, err := h.svc.CreateWorkOrder(c, &req)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, order)
}

func (h *MaintenanceHandler) UpdateWorkOrder(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid work order id")
		return
	}
	var req mtSvc.WorkOrder
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}
	req.ID = id
	order, err := h.svc.UpdateWorkOrder(c, &req)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, order)
}

type AssignReq struct {
	AssignedTo    int64  `json:"assigned_to" binding:"required"`
	DispatcherID  int64  `json:"dispatcher_id"`
	Workshop      string `json:"workshop"`
	ContactPhone  string `json:"contact_phone"`
	Remark        string `json:"remark"`
}

func (h *MaintenanceHandler) AssignWorkOrder(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid work order id")
		return
	}
	var req AssignReq
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}
	dispatcherID, _ := ctx.Get("user_id")
	var did int64
	if d, ok := dispatcherID.(int64); ok {
		did = d
	}
	if req.DispatcherID == 0 {
		req.DispatcherID = did
	}
	order, err := h.svc.AssignWorkOrder(c, id, req.AssignedTo, req.DispatcherID, req.Workshop, req.ContactPhone, req.Remark)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, order)
}

type AppointmentReq struct {
	AppointmentTime string `json:"appointment_time" binding:"required"`
	Workshop        string `json:"workshop"`
	ContactPhone    string `json:"contact_phone"`
	Remark          string `json:"remark"`
}

func (h *MaintenanceHandler) SetAppointment(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid work order id")
		return
	}
	var req AppointmentReq
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}
	order, err := h.svc.SetAppointment(c, id, req.AppointmentTime, req.Workshop, req.ContactPhone, req.Remark)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, order)
}

type CheckinReq struct {
	CheckinTime  string  `json:"checkin_time"`
	CurrentMileage float64 `json:"current_mileage_km" binding:"required"`
	Mechanic     string  `json:"mechanic"`
	Remark       string  `json:"remark"`
}

func (h *MaintenanceHandler) Checkin(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid work order id")
		return
	}
	var req CheckinReq
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}
	order, err := h.svc.CheckinWorkOrder(c, id, req.CheckinTime, req.CurrentMileage, req.Mechanic, req.Remark)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, order)
}

type StartWorkReq struct {
	Items     string `json:"items"`
	PartsUsed string `json:"parts_used"`
	Mechanic  string `json:"mechanic"`
	Remark    string `json:"remark"`
}

func (h *MaintenanceHandler) StartWork(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid work order id")
		return
	}
	var req StartWorkReq
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}
	order, err := h.svc.StartWork(c, id, req.Items, req.PartsUsed, req.Mechanic, req.Remark)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, order)
}

type CompleteReq struct {
	CheckoutTime       string  `json:"checkout_time"`
	ActualCost         float64 `json:"actual_cost"`
	QualityCheckDone   int     `json:"quality_check_done"`
	QualityCheckNote   string  `json:"quality_check_note"`
	NextMileageSuggest float64 `json:"next_mileage_km"`
	NextDateSuggest    string  `json:"next_date"`
	PartsUsed          string  `json:"parts_used"`
	Items              string  `json:"items"`
	PhotosBefore       string  `json:"photos_before"`
	PhotosAfter        string  `json:"photos_after"`
	Remark             string  `json:"remark"`
}

func (h *MaintenanceHandler) CompleteWork(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid work order id")
		return
	}
	var req CompleteReq
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}
	userID, _ := ctx.Get("user_id")
	var uid int64
	if u, ok := userID.(int64); ok {
		uid = u
	}
	order, err := h.svc.CompleteWork(c, id, uid, &mtSvc.WorkOrderCompleteData{
		CheckoutTime:       req.CheckoutTime,
		ActualCost:         req.ActualCost,
		QualityCheckDone:   req.QualityCheckDone,
		QualityCheckNote:   req.QualityCheckNote,
		NextMileageSuggest: req.NextMileageSuggest,
		NextDateSuggest:    req.NextDateSuggest,
		PartsUsed:          req.PartsUsed,
		Items:              req.Items,
		PhotosBefore:       req.PhotosBefore,
		PhotosAfter:        req.PhotosAfter,
		Remark:             req.Remark,
	})
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, order)
}

type CancelReq struct {
	CancelledReason string `json:"cancelled_reason" binding:"required"`
}

func (h *MaintenanceHandler) CancelWorkOrder(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid work order id")
		return
	}
	var req CancelReq
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}
	userID, _ := ctx.Get("user_id")
	var uid int64
	if u, ok := userID.(int64); ok {
		uid = u
	}
	order, err := h.svc.CancelWorkOrder(c, id, uid, req.CancelledReason)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, order)
}

func (h *MaintenanceHandler) GetWorkOrderLogs(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid work order id")
		return
	}
	logs, err := h.svc.GetWorkOrderLogs(c, id)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, utils.H{
		"list":  logs,
		"total": len(logs),
	})
}

func (h *MaintenanceHandler) GetStats(c context.Context, ctx *app.RequestContext) {
	orgID, _ := strconv.ParseInt(ctx.DefaultQuery("org_id", "0"), 10, 64)
	stats, err := h.svc.GetStats(c, orgID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, stats)
}

func (h *MaintenanceHandler) GetUpcomingMaintenance(c context.Context, ctx *app.RequestContext) {
	orgID, _ := strconv.ParseInt(ctx.DefaultQuery("org_id", "0"), 10, 64)
	vehicleID, _ := strconv.ParseInt(ctx.DefaultQuery("vehicle_id", "0"), 10, 64)
	days, _ := strconv.Atoi(ctx.DefaultQuery("days", "30"))
	km, _ := strconv.ParseFloat(ctx.DefaultQuery("km", "1000"), 64)
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))

	list, err := h.svc.GetUpcomingMaintenance(c, orgID, vehicleID, days, km, limit)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, utils.H{
		"list":  list,
		"total": len(list),
	})
}

func (h *MaintenanceHandler) GetOverdueMaintenance(c context.Context, ctx *app.RequestContext) {
	orgID, _ := strconv.ParseInt(ctx.DefaultQuery("org_id", "0"), 10, 64)
	vehicleID, _ := strconv.ParseInt(ctx.DefaultQuery("vehicle_id", "0"), 10, 64)
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))

	list, err := h.svc.GetOverdueMaintenance(c, orgID, vehicleID, limit)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, utils.H{
		"list":  list,
		"total": len(list),
	})
}
