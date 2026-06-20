package http

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"

	"github.com/dangerous-drive-guard/backend/internal/auth/service"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	DeviceID string `json:"device_id"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: svc}
}

func (h *AuthHandler) RegisterRoutes(r *app.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/logout", h.Logout)
		auth.GET("/me", h.GetCurrentUser)
		auth.PUT("/password", h.ChangePassword)
	}
}

func (h *AuthHandler) Login(c context.Context, ctx *app.RequestContext) {
	var req LoginRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	resp, err := h.authService.Login(c, req.Username, req.Password, req.DeviceID, ctx.ClientIP())
	if err != nil {
		response.Unauthorized(ctx, "用户名或密码错误")
		return
	}

	response.Success(ctx, resp)
}

func (h *AuthHandler) RefreshToken(c context.Context, ctx *app.RequestContext) {
	userID, _ := ctx.Get("user_id")
	username, _ := ctx.Get("username")
	role, _ := ctx.Get("role")
	orgID, _ := ctx.Get("org_id")

	token, expiresAt, err := h.authService.RefreshToken(
		c,
		service.ToInt64(userID),
		service.ToString(username),
		service.ToString(role),
		service.ToInt64(orgID),
	)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, utils.H{
		"access_token": token,
		"expires_at":   expiresAt,
		"token_type":   "Bearer",
	})
}

func (h *AuthHandler) Logout(c context.Context, ctx *app.RequestContext) {
	userID, _ := ctx.Get("user_id")
	if err := h.authService.Logout(c, service.ToInt64(userID)); err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, utils.H{"logged_out": true})
}

func (h *AuthHandler) GetCurrentUser(c context.Context, ctx *app.RequestContext) {
	userID, _ := ctx.Get("user_id")
	user, err := h.authService.GetUserByID(c, service.ToInt64(userID))
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, user)
}

func (h *AuthHandler) ChangePassword(c context.Context, ctx *app.RequestContext) {
	userID, _ := ctx.Get("user_id")
	var req ChangePasswordRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	if err := h.authService.ChangePassword(c, service.ToInt64(userID), req.OldPassword, req.NewPassword); err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, utils.H{"changed": true})
}
