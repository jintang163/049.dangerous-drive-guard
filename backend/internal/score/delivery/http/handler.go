package http

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	scoreSvc "github.com/dangerous-drive-guard/backend/internal/score/service"
	"github.com/dangerous-drive-guard/backend/pkg/middleware"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

type ScoreHandler struct {
	scoreService *scoreSvc.ScoreService
}

func NewScoreHandler(svc *scoreSvc.ScoreService) *ScoreHandler {
	return &ScoreHandler{scoreService: svc}
}

func (h *ScoreHandler) RegisterRoutes(r *app.RouterGroup, authMiddleware app.HandlerFunc) {
	scores := r.Group("/scores", authMiddleware)
	{
		scores.GET("/overview", h.GetOverview)
		scores.GET("/leaderboard", middleware.RoleAuth("admin"), h.GetLeaderboard)
		scores.GET("/drivers/:id", h.GetDriverScore)
		scores.GET("/drivers/:id/bonuses", h.GetDriverBonuses)
		scores.POST("/drivers/:id/check-bonus", h.CheckBonus)
		scores.GET("/monthly-reports", middleware.RoleAuth("admin", "dispatcher"), h.ListMonthlyReports)
		scores.GET("/monthly-reports/:id", middleware.RoleAuth("admin", "dispatcher"), h.GetMonthlyReport)
		scores.POST("/monthly-reports/:id/send", middleware.RoleAuth("admin"), h.SendMonthlyReport)
		scores.POST("/monthly-reports/batch-send", middleware.RoleAuth("admin"), h.BatchSendMonthlyReports)
		scores.GET("/drivers/:id/monthly-report", h.GetDriverMonthlyReport)
		scores.GET("/retraining-tasks", middleware.RoleAuth("admin", "dispatcher"), h.ListRetrainingTasks)
		scores.PUT("/retraining-tasks/:id", middleware.RoleAuth("admin"), h.UpdateRetrainingTask)
	}
}

func (h *ScoreHandler) GetOverview(c context.Context, ctx *app.RequestContext) {
	overview, err := h.scoreService.GetOverview(c)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, overview)
}

func (h *ScoreHandler) GetLeaderboard(c context.Context, ctx *app.RequestContext) {
	top, _ := strconv.Atoi(ctx.DefaultQuery("top", "20"))
	orderBy := ctx.DefaultQuery("order_by", "total_score")

	items, err := h.scoreService.GetLeaderboard(c, top, orderBy)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, items)
}

func (h *ScoreHandler) GetDriverScore(c context.Context, ctx *app.RequestContext) {
	driverID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid driver id")
		return
	}

	days, _ := strconv.Atoi(ctx.DefaultQuery("days", "30"))

	latest, history, err := h.scoreService.GetDriverScore(c, driverID, days)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"latest":  latest,
		"history": history,
	})
}

func (h *ScoreHandler) GetDriverBonuses(c context.Context, ctx *app.RequestContext) {
	driverID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid driver id")
		return
	}

	bonuses, err := h.scoreService.GetDriverBonuses(c, driverID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, bonuses)
}

func (h *ScoreHandler) CheckBonus(c context.Context, ctx *app.RequestContext) {
	driverID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid driver id")
		return
	}

	bonus, err := h.scoreService.CheckAndAwardNoViolationBonus(c, driverID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	if bonus == nil {
		response.Success(ctx, map[string]interface{}{
			"awarded": false,
			"message": "未达到连续30天无违规标准",
		})
		return
	}

	response.Success(ctx, map[string]interface{}{
		"awarded": true,
		"bonus":   bonus,
	})
}

func (h *ScoreHandler) GetMonthlyReport(c context.Context, ctx *app.RequestContext) {
	reportID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid report id")
		return
	}

	_ = reportID
	response.Success(ctx, map[string]interface{}{
		"id": reportID,
	})
}

func (h *ScoreHandler) GetDriverMonthlyReport(c context.Context, ctx *app.RequestContext) {
	driverID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid driver id")
		return
	}

	month := ctx.DefaultQuery("month", "")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	report, err := h.scoreService.GetMonthlyReport(c, driverID, month)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, report)
}

func (h *ScoreHandler) ListMonthlyReports(c context.Context, ctx *app.RequestContext) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	month := ctx.Query("month")
	driverID := ctx.Query("driver_id")

	reports, total, err := h.scoreService.ListMonthlyReports(c, page, pageSize, month, driverID)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Page(ctx, reports, total, page, pageSize)
}

func (h *ScoreHandler) SendMonthlyReport(c context.Context, ctx *app.RequestContext) {
	reportID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid report id")
		return
	}

	if err := h.scoreService.SendMonthlyReport(c, reportID); err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"sent": true,
	})
}

func (h *ScoreHandler) BatchSendMonthlyReports(c context.Context, ctx *app.RequestContext) {
	month := ctx.DefaultQuery("month", "")
	if month == "" {
		month = time.Now().AddDate(0, -1, 0).Format("2006-01")
	}

	sent, failed, err := h.scoreService.BatchSendMonthlyReports(c, month)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"month":  month,
		"sent":   sent,
		"failed": failed,
	})
}

func (h *ScoreHandler) ListRetrainingTasks(c context.Context, ctx *app.RequestContext) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	status := ctx.Query("status")

	tasks, total, err := h.scoreService.ListRetrainingTasks(c, page, pageSize, status)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Page(ctx, tasks, total, page, pageSize)
}

func (h *ScoreHandler) UpdateRetrainingTask(c context.Context, ctx *app.RequestContext) {
	taskID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid task id")
		return
	}

	var req struct {
		Status      string  `json:"status"`
		ResultNote  string  `json:"result_note"`
		ResultScore float64 `json:"result_score"`
	}
	if err := ctx.Bind(&req); err != nil {
		response.BadRequest(ctx, "invalid request body")
		return
	}

	if err := h.scoreService.UpdateRetrainingTask(c, taskID, req.Status, req.ResultNote, req.ResultScore); err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"updated": true,
	})
}
