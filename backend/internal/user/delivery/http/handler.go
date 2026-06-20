package http

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"

	userSvc "github.com/dangerous-drive-guard/backend/internal/user/service"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	RealName string `json:"real_name" binding:"required"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Role     string `json:"role" binding:"required"`
	OrgID    int64  `json:"org_id"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
}

type UserHandler struct {
	userService *userSvc.UserService
}

func NewUserHandler(svc *userSvc.UserService) *UserHandler {
	return &UserHandler{userService: svc}
}

func (h *UserHandler) RegisterRoutes(r *app.RouterGroup, authMiddleware app.HandlerFunc) {
	users := r.Group("/users", authMiddleware)
	{
		users.GET("", h.ListUsers)
		users.GET("/:id", h.GetUser)
		users.POST("", h.CreateUser)
		users.PUT("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUser)
		users.POST("/:id/reset-password", h.ResetPassword)
	}
}

func (h *UserHandler) ListUsers(c context.Context, ctx *app.RequestContext) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	orgID, _ := strconv.ParseInt(ctx.DefaultQuery("org_id", "0"), 10, 64)
	role := ctx.Query("role")
	keyword := ctx.Query("keyword")

	users, total, err := h.userService.ListUsers(c, orgID, role, keyword, page, pageSize)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Page(ctx, users, total, page, pageSize)
}

func (h *UserHandler) GetUser(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid user id")
		return
	}

	user, err := h.userService.GetUser(c, id)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, user)
}

func (h *UserHandler) CreateUser(c context.Context, ctx *app.RequestContext) {
	var req CreateUserRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	user, err := h.userService.CreateUser(c, &userSvc.UserCreateRequest{
		Username: req.Username,
		Password: req.Password,
		RealName: req.RealName,
		Phone:    req.Phone,
		Email:    req.Email,
		Role:     userSvc.ToUserRole(req.Role),
		OrgID:    req.OrgID,
	})
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, user)
}

func (h *UserHandler) UpdateUser(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid user id")
		return
	}

	var req map[string]interface{}
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	req["id"] = id
	err = h.userService.UpdateUser(c, req)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, utils.H{"updated": true})
}

func (h *UserHandler) DeleteUser(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid user id")
		return
	}

	err = h.userService.DeleteUser(c, id)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, utils.H{"deleted": true})
}

func (h *UserHandler) ResetPassword(c context.Context, ctx *app.RequestContext) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid user id")
		return
	}

	var req ResetPasswordRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	err = h.userService.ResetPassword(c, id, req.NewPassword)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, utils.H{"reset": true})
}
