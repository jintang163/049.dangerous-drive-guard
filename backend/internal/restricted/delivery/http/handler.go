package http

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	restrictedSvc "github.com/dangerous-drive-guard/backend/internal/restricted/service"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)
var restrictedService *restrictedSvc.RestrictedAreaService

func initService() {
	if restrictedService == nil {
		restrictedService = restrictedSvc.NewRestrictedAreaService()
	}
}

func getUserID(c *app.RequestContext) int64 {
	userID, _ := c.Get("user_id")
	if uid, ok := userID.(int64); ok {
		return uid
	}
	return 0
}

func getUsername(c *app.RequestContext) string {
	username, _ := c.Get("username")
	if name, ok := username.(string); ok {
		return name
	}
	return ""
}

func ListAreas(ctx context.Context, c *app.RequestContext) {
	initService()
	areaType := c.Query("area_type")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var isTemporary *int
	if tmp := c.Query("is_temporary"); tmp != "" {
		v, _ := strconv.Atoi(tmp)
		isTemporary = &v
	}

	var approvalStatus *int
	if as := c.Query("approval_status"); as != "" {
		v, _ := strconv.Atoi(as)
		approvalStatus = &v
	}

	list, total, err := restrictedService.ListAreas(ctx, areaType, isTemporary, approvalStatus, keyword, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func GetArea(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	area, err := restrictedService.GetArea(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, area)
}

func CreateArea(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.RestrictedAreaCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := getUserID(c)
	area, err := restrictedService.CreateArea(ctx, &req, userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, area)
}

func UpdateArea(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req model.RestrictedAreaUpdateRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := getUserID(c)
	if err := restrictedService.UpdateArea(ctx, id, &req, userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"success": true})
}

func DeleteArea(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	userID := getUserID(c)
	if err := restrictedService.DeleteArea(ctx, id, userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"success": true})
}

func SubmitApproval(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	userID := getUserID(c)
	if err := restrictedService.SubmitApproval(ctx, id, userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"success": true})
}

func ApproveFirstLevel(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req model.ApprovalRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID := getUserID(c)
	username := getUsername(c)
	if err := restrictedService.ApproveFirstLevel(ctx, id, userID, username, req.ApprovalNote); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"success": true})
}

func ApproveSecondLevel(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req model.ApprovalRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID := getUserID(c)
	username := getUsername(c)
	if err := restrictedService.ApproveSecondLevel(ctx, id, userID, username, req.ApprovalNote); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"success": true})
}

func RejectApproval(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	level, _ := strconv.Atoi(c.DefaultQuery("level", "1"))
	var req model.ApprovalRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID := getUserID(c)
	username := getUsername(c)
	if err := restrictedService.RejectApproval(ctx, id, level, userID, username, req.ApprovalNote); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"success": true})
}

func RevokeApproval(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req model.ApprovalRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID := getUserID(c)
	if err := restrictedService.RevokeApproval(ctx, id, userID, req.ApprovalNote); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"success": true})
}

func GetApprovalHistory(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	list, err := restrictedService.GetApprovalHistory(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"list": list, "total": len(list)})
}

func ListPendingApprovals(ctx context.Context, c *app.RequestContext) {
	initService()
	level, _ := strconv.Atoi(c.DefaultQuery("level", "1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := restrictedService.ListPendingApprovals(ctx, level, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func ListTemplates(ctx context.Context, c *app.RequestContext) {
	initService()
	category := c.Query("category")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	list, total, err := restrictedService.ListTemplates(ctx, category, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func GetTemplate(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	tpl, err := restrictedService.GetTemplate(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, tpl)
}

func CreateTemplate(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.TemplateCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID := getUserID(c)
	tpl, err := restrictedService.CreateTemplate(ctx, &req, userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, tpl)
}

func UpdateTemplate(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req model.TemplateCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID := getUserID(c)
	if err := restrictedService.UpdateTemplate(ctx, id, &req, userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"success": true})
}

func DeleteTemplate(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	userID := getUserID(c)
	if err := restrictedService.DeleteTemplate(ctx, id, userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"success": true})
}

func ApplyTemplate(ctx context.Context, c *app.RequestContext) {
	initService()
	templateID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid template id")
		return
	}

	type ApplyRequest struct {
		CenterLat float64 `json:"center_lat" binding:"required"`
		CenterLng float64 `json:"center_lng" binding:"required"`
		Name      string  `json:"name"`
		Address   string  `json:"address"`
	}
	var req ApplyRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	area, err := restrictedService.ApplyTemplate(ctx, templateID, req.CenterLat, req.CenterLng, req.Name, req.Address)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, area)
}

func ImportGisData(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.GisImportRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID := getUserID(c)
	record, err := restrictedService.ImportGisData(ctx, &req, userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, record)
}

func ListGisImports(ctx context.Context, c *app.RequestContext) {
	initService()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := restrictedService.ListGisImports(ctx, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func PullActiveAreas(ctx context.Context, c *app.RequestContext) {
	initService()
	var sinceVersion int64
	if sv := c.Query("since_version"); sv != "" {
		sinceVersion, _ = strconv.ParseInt(sv, 10, 64)
	}
	hazardClass := c.Query("hazard_class")

	list, latestVersion, err := restrictedService.PullActiveAreas(ctx, sinceVersion, hazardClass)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{
		"list":           list,
		"total":          len(list),
		"latest_version": latestVersion,
	})
}

func ImportGisFile(ctx context.Context, c *app.RequestContext) {
	initService()
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传GIS数据文件")
		return
	}

	sourceType := c.DefaultPostForm("source_type", "geojson")
	userID := getUserID(c)

	record, err := restrictedService.ImportGisFile(ctx, file, sourceType, userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, record)
}
