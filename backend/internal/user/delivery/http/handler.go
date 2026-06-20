package http

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	userSvc "github.com/dangerous-drive-guard/backend/internal/user/service"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

var userService *userSvc.UserService

func initService() {
	if userService == nil {
		userService = userSvc.NewUserService()
	}
}

func GetProfile(ctx context.Context, c *app.RequestContext) {
	initService()
	userID, _ := c.Get("user_id")
	user, err := userService.GetUser(ctx, toInt64(userID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, user)
}

func UpdateProfile(ctx context.Context, c *app.RequestContext) {
	initService()
	userID, _ := c.Get("user_id")
	var req map[string]interface{}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req["id"] = toInt64(userID)
	err := userService.UpdateUser(ctx, req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"updated": true})
}

func GetByID(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}
	user, err := userService.GetUser(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, user)
}

func List(ctx context.Context, c *app.RequestContext) {
	initService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	orgID, _ := strconv.ParseInt(c.DefaultQuery("org_id", "0"), 10, 64)
	role := c.Query("role")
	keyword := c.Query("keyword")
	users, total, err := userService.ListUsers(ctx, orgID, role, keyword, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, users, total, page, pageSize)
}

func Create(ctx context.Context, c *app.RequestContext) {
	initService()
	var req userSvc.UserCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	user, err := userService.CreateUser(ctx, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, user)
}

func Update(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}
	var req map[string]interface{}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req["id"] = id
	err = userService.UpdateUser(ctx, req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"updated": true})
}

func Delete(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}
	err = userService.DeleteUser(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"deleted": true})
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
