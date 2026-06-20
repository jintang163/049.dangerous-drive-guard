package http

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"golang.org/x/crypto/bcrypt"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	authSvc "github.com/dangerous-drive-guard/backend/internal/user/service"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/middleware"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	DeviceID string `json:"device_id"`
}

type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
	User         *model.User `json:"user"`
	Permissions  []string  `json:"permissions"`
}

func Login(ctx context.Context, c *app.RequestContext) {
	var req LoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	var user model.User
	err := database.GetDB().WithContext(ctx).Table("users").Where("username = ? AND status = 1", req.Username).First(&user).Error
	if err != nil {
		response.Unauthorized(c, "用户名或密码错误")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		response.Unauthorized(c, "用户名或密码错误")
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Username, string(user.Role), user.OrgID)
	if err != nil {
		response.InternalError(c, "生成Token失败")
		return
	}

	now := time.Now()
	database.GetDB().Exec(`
		UPDATE users SET last_login_at = ?, last_login_ip = ? WHERE id = ?`,
		now, c.ClientIP(), user.ID,
	)

	database.GetDB().Exec(`
		INSERT INTO user_tokens (user_id, token, expires_at, device_id, ip, created_at)
		VALUES (?, ?, DATE_ADD(NOW(), INTERVAL 72 HOUR), ?, ?, NOW())`,
		user.ID, token, req.DeviceID, c.ClientIP(),
	)

	user.Password = ""
	permissions := getRolePermissions(user.Role)

	response.Success(c, LoginResponse{
		AccessToken: token,
		ExpiresAt:   now.Add(72 * time.Hour),
		TokenType:   "Bearer",
		User:        &user,
		Permissions: permissions,
	})
}

func getRolePermissions(role model.UserRole) []string {
	perms := []string{"dashboard:view", "profile:read", "profile:update"}
	switch role {
	case model.RoleAdmin:
		perms = append(perms,
			"user:manage", "vehicle:manage", "waybill:manage",
			"route:plan", "monitor:view", "alarm:handle",
			"dispatch:command", "score:view", "system:config")
	case model.RoleDispatcher:
		perms = append(perms,
			"vehicle:view", "waybill:create", "waybill:update",
			"route:plan", "monitor:view", "alarm:handle",
			"dispatch:command", "score:view")
	case model.RoleDriver:
		perms = append(perms,
			"waybill:view", "route:navigate",
			"fatigue:report", "diagnostics:upload", "score:view")
	case model.RoleEscort:
		perms = append(perms,
			"waybill:view", "escort:event_report")
	}
	return perms
}

func Refresh(ctx context.Context, c *app.RequestContext) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	role, _ := c.Get("role")
	orgID, _ := c.Get("org_id")
	token, err := middleware.GenerateToken(
		authSvc.ToInt64(userID),
		authSvc.ToString(username),
		authSvc.ToString(role),
		authSvc.ToInt64(orgID),
	)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{
		"access_token": token,
		"expires_at":   time.Now().Add(72 * time.Hour),
		"token_type":   "Bearer",
	})
}

func Logout(ctx context.Context, c *app.RequestContext) {
	userID, _ := c.Get("user_id")
	database.GetDB().Exec(`DELETE FROM user_tokens WHERE user_id = ?`, userID)
	response.Success(c, map[string]interface{}{"logged_out": true})
}
