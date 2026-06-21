package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	monitorWs "github.com/dangerous-drive-guard/backend/internal/monitor/delivery/ws"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
	"github.com/dangerous-drive-guard/backend/pkg/mq"
)

type RestrictedAreaService struct {
	db *gorm.DB
}

func NewRestrictedAreaService() *RestrictedAreaService {
	return &RestrictedAreaService{
		db: database.GetDB(),
	}
}

func (s *RestrictedAreaService) ListAreas(ctx context.Context, areaType string, isTemporary *int, approvalStatus *int, keyword string, page, pageSize int) ([]*model.RestrictedAreaExt, int64, error) {
	var list []*model.RestrictedAreaExt
	var total int64

	query := s.db.WithContext(ctx).Model(&model.RestrictedAreaExt{})

	if areaType != "" {
		query = query.Where("area_type = ?", areaType)
	}
	if isTemporary != nil {
		query = query.Where("is_temporary = ?", *isTemporary)
	}
	if approvalStatus != nil {
		query = query.Where("approval_status = ?", *approvalStatus)
	}
	if keyword != "" {
		query = query.Where("name LIKE ? OR address LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (s *RestrictedAreaService) GetArea(ctx context.Context, id int64) (*model.RestrictedAreaExt, error) {
	var area model.RestrictedAreaExt
	if err := s.db.WithContext(ctx).First(&area, id).Error; err != nil {
		return nil, err
	}
	return &area, nil
}

func (s *RestrictedAreaService) CreateArea(ctx context.Context, req *model.RestrictedAreaCreateRequest, userID int64) (*model.RestrictedAreaExt, error) {
	var timeScheduleJSON model.JSON
	if len(req.TimeSchedule) > 0 {
		data, _ := json.Marshal(req.TimeSchedule)
		timeScheduleJSON = model.JSON(data)
	}

	area := &model.RestrictedAreaExt{
		RestrictedArea: model.RestrictedArea{
			Name:                  req.Name,
			AreaType:              req.AreaType,
			Level:                 req.Level,
			Province:              req.Province,
			City:                  req.City,
			District:              req.District,
			Address:               req.Address,
			BoundaryPolygon:       req.BoundaryPolygon,
			CenterLatitude:        req.CenterLatitude,
			CenterLongitude:       req.CenterLongitude,
			Radius:                req.Radius,
			RestrictHazardClasses: req.RestrictHazardClasses,
			RestrictVehicleTypes:  req.RestrictVehicleTypes,
			HeightLimit:           req.HeightLimit,
			WeightLimit:           req.WeightLimit,
			EffectiveFrom:         req.EffectiveFrom,
			EffectiveTo:           req.EffectiveTo,
			Source:                req.Source,
			Status:                1,
		},
		ShapeType:      req.ShapeType,
		IsTemporary:    req.IsTemporary,
		TempReason:     req.TempReason,
		TimeSchedule:   timeScheduleJSON,
		TemplateID:     req.TemplateID,
		CreatedBy:      userID,
		ApprovalStatus: model.ApprovalPending,
	}

	if req.SubmitApproval {
		area.ApprovalStatus = model.ApprovalFirstLevel
	}

	if err := s.db.WithContext(ctx).Create(area).Error; err != nil {
		return nil, err
	}

	if req.SubmitApproval {
		s.recordApproval(ctx, area.ID, 1, userID, "", model.ApprovalActionSubmit, model.ApprovalPending, model.ApprovalFirstLevel)
	}

	logger.Sugar.Infof("Restricted area created: id=%d, name=%s, by=%d", area.ID, area.Name, userID)
	return area, nil
}

func (s *RestrictedAreaService) UpdateArea(ctx context.Context, id int64, req *model.RestrictedAreaUpdateRequest, userID int64) error {
	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.AreaType != "" {
		updates["area_type"] = req.AreaType
	}
	if req.ShapeType != "" {
		updates["shape_type"] = req.ShapeType
	}
	if req.Level > 0 {
		updates["level"] = req.Level
	}
	if req.Province != "" {
		updates["province"] = req.Province
	}
	if req.City != "" {
		updates["city"] = req.City
	}
	if req.District != "" {
		updates["district"] = req.District
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if len(req.BoundaryPolygon) > 0 {
		updates["boundary_polygon"] = req.BoundaryPolygon
	}
	if req.CenterLatitude != 0 {
		updates["center_latitude"] = req.CenterLatitude
	}
	if req.CenterLongitude != 0 {
		updates["center_longitude"] = req.CenterLongitude
	}
	if req.Radius != 0 {
		updates["radius"] = req.Radius
	}
	if req.RestrictHazardClasses != "" {
		updates["restrict_hazard_classes"] = req.RestrictHazardClasses
	}
	if req.RestrictVehicleTypes != "" {
		updates["restrict_vehicle_types"] = req.RestrictVehicleTypes
	}
	if req.HeightLimit != 0 {
		updates["height_limit"] = req.HeightLimit
	}
	if req.WeightLimit != 0 {
		updates["weight_limit"] = req.WeightLimit
	}
	if len(req.TimeSchedule) > 0 {
		data, _ := json.Marshal(req.TimeSchedule)
		updates["time_schedule"] = model.JSON(data)
	}
	if req.EffectiveFrom != nil {
		updates["effective_from"] = req.EffectiveFrom
	}
	if req.EffectiveTo != nil {
		updates["effective_to"] = req.EffectiveTo
	}
	updates["is_temporary"] = req.IsTemporary
	if req.TempReason != "" {
		updates["temp_reason"] = req.TempReason
	}
	updates["status"] = req.Status

	if len(updates) == 0 {
		return nil
	}

	if err := s.db.WithContext(ctx).Model(&model.RestrictedAreaExt{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}

	logger.Sugar.Infof("Restricted area updated: id=%d, by=%d", id, userID)
	return nil
}

func (s *RestrictedAreaService) DeleteArea(ctx context.Context, id int64, userID int64) error {
	if err := s.db.WithContext(ctx).Delete(&model.RestrictedAreaExt{}, id).Error; err != nil {
		return err
	}
	logger.Sugar.Infof("Restricted area deleted: id=%d, by=%d", id, userID)
	return nil
}

func (s *RestrictedAreaService) SubmitApproval(ctx context.Context, id int64, userID int64) error {
	var area model.RestrictedAreaExt
	if err := s.db.WithContext(ctx).First(&area, id).Error; err != nil {
		return err
	}

	if area.ApprovalStatus != model.ApprovalPending {
		return fmt.Errorf("当前状态不允许提交审批")
	}

	now := time.Now()
	if err := s.db.WithContext(ctx).Model(&area).Updates(map[string]interface{}{
		"approval_status": model.ApprovalFirstLevel,
		"updated_at":      now,
	}).Error; err != nil {
		return err
	}

	s.recordApproval(ctx, id, 1, userID, "", model.ApprovalActionSubmit, model.ApprovalPending, model.ApprovalFirstLevel)
	logger.Sugar.Infof("Restricted area submitted for approval: id=%d, by=%d", id, userID)
	return nil
}

func (s *RestrictedAreaService) ApproveFirstLevel(ctx context.Context, id int64, approverID int64, approverName string, note string) error {
	var area model.RestrictedAreaExt
	if err := s.db.WithContext(ctx).First(&area, id).Error; err != nil {
		return err
	}

	if area.ApprovalStatus != model.ApprovalFirstLevel {
		return fmt.Errorf("当前状态不允许一级审批")
	}

	now := time.Now()
	if err := s.db.WithContext(ctx).Model(&area).Updates(map[string]interface{}{
		"approval_status":      model.ApprovalSecondLevel,
		"first_approver_id":    approverID,
		"first_approval_at":    now,
		"first_approval_note":  note,
		"updated_at":           now,
	}).Error; err != nil {
		return err
	}

	s.recordApproval(ctx, id, 1, approverID, note, model.ApprovalActionApprove, model.ApprovalFirstLevel, model.ApprovalSecondLevel)
	logger.Sugar.Infof("Restricted area first level approved: id=%d, by=%d", id, approverID)
	return nil
}

func (s *RestrictedAreaService) ApproveSecondLevel(ctx context.Context, id int64, approverID int64, approverName string, note string) error {
	var area model.RestrictedAreaExt
	if err := s.db.WithContext(ctx).First(&area, id).Error; err != nil {
		return err
	}

	if area.ApprovalStatus != model.ApprovalSecondLevel {
		return fmt.Errorf("当前状态不允许二级审批")
	}

	now := time.Now()
	if err := s.db.WithContext(ctx).Model(&area).Updates(map[string]interface{}{
		"approval_status":       model.ApprovalApproved,
		"status":                1,
		"second_approver_id":    approverID,
		"second_approval_at":    now,
		"second_approval_note":  note,
		"updated_at":            now,
	}).Error; err != nil {
		return err
	}

	s.recordApproval(ctx, id, 2, approverID, note, model.ApprovalActionApprove, model.ApprovalSecondLevel, model.ApprovalApproved)
	s.notifyNavigationUpdate(ctx, id)
	logger.Sugar.Infof("Restricted area second level approved and activated: id=%d, by=%d", id, approverID)
	return nil
}

func (s *RestrictedAreaService) RejectApproval(ctx context.Context, id int64, level int, approverID int64, approverName string, note string) error {
	var area model.RestrictedAreaExt
	if err := s.db.WithContext(ctx).First(&area, id).Error; err != nil {
		return err
	}

	oldStatus := area.ApprovalStatus
	now := time.Now()

	updates := map[string]interface{}{
		"approval_status": model.ApprovalRejected,
		"status":          0,
		"updated_at":      now,
	}

	if level == 1 {
		updates["first_approver_id"] = approverID
		updates["first_approval_at"] = now
		updates["first_approval_note"] = note
	} else {
		updates["second_approver_id"] = approverID
		updates["second_approval_at"] = now
		updates["second_approval_note"] = note
	}

	if err := s.db.WithContext(ctx).Model(&area).Updates(updates).Error; err != nil {
		return err
	}

	s.recordApproval(ctx, id, level, approverID, note, model.ApprovalActionReject, oldStatus, model.ApprovalRejected)
	logger.Sugar.Infof("Restricted area rejected: id=%d, level=%d, by=%d", id, level, approverID)
	return nil
}

func (s *RestrictedAreaService) RevokeApproval(ctx context.Context, id int64, userID int64, note string) error {
	var area model.RestrictedAreaExt
	if err := s.db.WithContext(ctx).First(&area, id).Error; err != nil {
		return err
	}

	if area.ApprovalStatus == model.ApprovalApproved {
		s.notifyNavigationUpdate(ctx, id)
	}

	oldStatus := area.ApprovalStatus
	if err := s.db.WithContext(ctx).Model(&area).Updates(map[string]interface{}{
		"approval_status": model.ApprovalRevoked,
		"status":          0,
		"updated_at":      time.Now(),
	}).Error; err != nil {
		return err
	}

	s.recordApproval(ctx, id, 0, userID, note, model.ApprovalActionRevoke, oldStatus, model.ApprovalRevoked)
	logger.Sugar.Infof("Restricted area revoked: id=%d, by=%d", id, userID)
	return nil
}

func (s *RestrictedAreaService) GetApprovalHistory(ctx context.Context, areaID int64) ([]*model.RestrictedAreaApproval, error) {
	var list []*model.RestrictedAreaApproval
	if err := s.db.WithContext(ctx).Where("area_id = ?", areaID).Order("created_at ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *RestrictedAreaService) ListPendingApprovals(ctx context.Context, level int, page, pageSize int) ([]*model.RestrictedAreaExt, int64, error) {
	var list []*model.RestrictedAreaExt
	var total int64

	status := model.ApprovalFirstLevel
	if level == 2 {
		status = model.ApprovalSecondLevel
	}

	query := s.db.WithContext(ctx).Model(&model.RestrictedAreaExt{}).Where("approval_status = ?", status)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (s *RestrictedAreaService) recordApproval(ctx context.Context, areaID int64, level int, approverID int64, approverName string, action model.ApprovalAction, oldStatus, newStatus model.ApprovalStatus) {
	record := &model.RestrictedAreaApproval{
		AreaID:         areaID,
		ApprovalLevel:  level,
		ApproverID:     approverID,
		ApproverName:   approverName,
		ApprovalAction: action,
		ApprovalNote:   approverName,
		OldStatus:      oldStatus,
		NewStatus:      newStatus,
	}
	_ = s.db.WithContext(ctx).Create(record).Error
}

func (s *RestrictedAreaService) notifyNavigationUpdate(ctx context.Context, areaID int64) {
	var area model.RestrictedAreaExt
	if err := s.db.WithContext(ctx).First(&area, areaID).Error; err != nil {
		logger.Sugar.Errorf("notifyNavigationUpdate: area not found: id=%d, err=%v", areaID, err)
		return
	}

	event := "activated"
	if area.ApprovalStatus == model.ApprovalRevoked || area.Status == 0 {
		event = "deactivated"
	}

	version := time.Now().UnixMilli()

	hub := monitorWs.GetHub()
	hub.BroadcastRestrictedAreaUpdate(
		event,
		area.ID,
		area.Name,
		string(area.AreaType),
		area.Level,
		string(area.ShapeType),
		json.RawMessage(area.BoundaryPolygon),
		json.RawMessage(area.TimeSchedule),
		version,
	)

	syncMsg := map[string]interface{}{
		"event":       event,
		"area_id":     area.ID,
		"area_name":   area.Name,
		"area_type":   string(area.AreaType),
		"shape_type":  string(area.ShapeType),
		"level":       area.Level,
		"version":     version,
		"updated_at":  time.Now().Format(time.RFC3339),
	}
	msgBody, _ := json.Marshal(syncMsg)
	if err := mq.Send(ctx, mq.Message{
		Topic: "restricted_area_change",
		Tag:   event,
		Key:   fmt.Sprintf("area_%d", areaID),
		Body:  msgBody,
	}); err != nil {
		logger.Sugar.Warnf("notifyNavigationUpdate: mq send failed: err=%v", err)
	}

	if err := s.updateSyncVersion(ctx, areaID, version); err != nil {
		logger.Sugar.Warnf("notifyNavigationUpdate: update sync version failed: err=%v", err)
	}

	logger.Global.Info("Navigation sync dispatched",
		zap.Int64("area_id", areaID),
		zap.String("event", event),
		zap.Int64("version", version))
}

func (s *RestrictedAreaService) updateSyncVersion(ctx context.Context, areaID, version int64) error {
	return s.db.WithContext(ctx).Model(&model.RestrictedAreaExt{}).
		Where("id = ?", areaID).
		Update("updated_at", time.Now()).Error
}

func (s *RestrictedAreaService) PullActiveAreas(ctx context.Context, sinceVersion int64, hazardClass string) ([]*model.RestrictedAreaExt, int64, error) {
	var list []*model.RestrictedAreaExt
	sinceTime := time.UnixMilli(sinceVersion)

	query := s.db.WithContext(ctx).Model(&model.RestrictedAreaExt{}).
		Where("approval_status = ?", model.ApprovalApproved).
		Where("status = ?", 1)

	if sinceVersion > 0 {
		query = query.Where("updated_at > ?", sinceTime)
	}

	if hazardClass != "" {
		query = query.Where("restrict_hazard_classes = '' OR restrict_hazard_classes LIKE ?", "%"+hazardClass+"%")
	}

	if err := query.Find(&list).Error; err != nil {
		return nil, 0, err
	}

	var maxVersion int64
	for _, area := range list {
		v := area.UpdatedAt.UnixMilli()
		if v > maxVersion {
			maxVersion = v
		}
	}
	if maxVersion == 0 {
		maxVersion = time.Now().UnixMilli()
	}

	return list, maxVersion, nil
}

func (s *RestrictedAreaService) ListTemplates(ctx context.Context, category string, page, pageSize int) ([]*model.RestrictedAreaTemplate, int64, error) {
	var list []*model.RestrictedAreaTemplate
	var total int64

	query := s.db.WithContext(ctx).Model(&model.RestrictedAreaTemplate{})
	if category != "" {
		query = query.Where("template_category = ?", category)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Where("is_enabled = ?", 1).Order("is_builtin DESC, created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (s *RestrictedAreaService) GetTemplate(ctx context.Context, id int64) (*model.RestrictedAreaTemplate, error) {
	var tpl model.RestrictedAreaTemplate
	if err := s.db.WithContext(ctx).First(&tpl, id).Error; err != nil {
		return nil, err
	}
	return &tpl, nil
}

func (s *RestrictedAreaService) CreateTemplate(ctx context.Context, req *model.TemplateCreateRequest, userID int64) (*model.RestrictedAreaTemplate, error) {
	var timeRulesJSON model.JSON
	if len(req.TimeRules) > 0 {
		data, _ := json.Marshal(req.TimeRules)
		timeRulesJSON = model.JSON(data)
	}

	tpl := &model.RestrictedAreaTemplate{
		TemplateName:          req.TemplateName,
		TemplateCategory:      req.TemplateCategory,
		AreaType:              req.AreaType,
		Level:                 req.Level,
		DefaultRadius:         req.DefaultRadius,
		RestrictHazardClasses: req.RestrictHazardClasses,
		RestrictVehicleTypes:  req.RestrictVehicleTypes,
		HeightLimit:           req.HeightLimit,
		WeightLimit:           req.WeightLimit,
		TimeRules:             timeRulesJSON,
		Description:           req.Description,
		IsBuiltin:             0,
		IsEnabled:             req.IsEnabled,
		CreatedBy:             userID,
	}

	if err := s.db.WithContext(ctx).Create(tpl).Error; err != nil {
		return nil, err
	}

	logger.Sugar.Infof("Template created: id=%d, name=%s, by=%d", tpl.ID, tpl.TemplateName, userID)
	return tpl, nil
}

func (s *RestrictedAreaService) UpdateTemplate(ctx context.Context, id int64, req *model.TemplateCreateRequest, userID int64) error {
	updates := make(map[string]interface{})

	if req.TemplateName != "" {
		updates["template_name"] = req.TemplateName
	}
	if req.TemplateCategory != "" {
		updates["template_category"] = req.TemplateCategory
	}
	if req.AreaType != "" {
		updates["area_type"] = req.AreaType
	}
	if req.Level > 0 {
		updates["level"] = req.Level
	}
	if req.DefaultRadius > 0 {
		updates["default_radius"] = req.DefaultRadius
	}
	if req.RestrictHazardClasses != "" {
		updates["restrict_hazard_classes"] = req.RestrictHazardClasses
	}
	if req.RestrictVehicleTypes != "" {
		updates["restrict_vehicle_types"] = req.RestrictVehicleTypes
	}
	if req.HeightLimit > 0 {
		updates["height_limit"] = req.HeightLimit
	}
	if req.WeightLimit > 0 {
		updates["weight_limit"] = req.WeightLimit
	}
	if len(req.TimeRules) > 0 {
		data, _ := json.Marshal(req.TimeRules)
		updates["time_rules"] = model.JSON(data)
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	updates["is_enabled"] = req.IsEnabled

	if err := s.db.WithContext(ctx).Model(&model.RestrictedAreaTemplate{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}

	logger.Sugar.Infof("Template updated: id=%d, by=%d", id, userID)
	return nil
}

func (s *RestrictedAreaService) DeleteTemplate(ctx context.Context, id int64, userID int64) error {
	var tpl model.RestrictedAreaTemplate
	if err := s.db.WithContext(ctx).First(&tpl, id).Error; err != nil {
		return err
	}
	if tpl.IsBuiltin == 1 {
		return fmt.Errorf("内置模板不允许删除")
	}

	if err := s.db.WithContext(ctx).Delete(&tpl).Error; err != nil {
		return err
	}
	logger.Sugar.Infof("Template deleted: id=%d, by=%d", id, userID)
	return nil
}

func (s *RestrictedAreaService) ApplyTemplate(ctx context.Context, templateID int64, centerLat, centerLng float64, customName, address string) (*model.RestrictedAreaExt, error) {
	tpl, err := s.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, err
	}

	radius := tpl.DefaultRadius
	if radius <= 0 {
		radius = 500
	}

	polygon := generateCirclePolygon(centerLat, centerLng, radius)

	area := &model.RestrictedAreaExt{
		RestrictedArea: model.RestrictedArea{
			Name:                  customName,
			AreaType:              tpl.AreaType,
			Level:                 tpl.Level,
			Address:               address,
			BoundaryPolygon:       polygon,
			CenterLatitude:        centerLat,
			CenterLongitude:       centerLng,
			Radius:                radius,
			RestrictHazardClasses: tpl.RestrictHazardClasses,
			RestrictVehicleTypes:  tpl.RestrictVehicleTypes,
			HeightLimit:           tpl.HeightLimit,
			WeightLimit:           tpl.WeightLimit,
			Source:                "template",
			Status:                1,
		},
		ShapeType:      model.ShapeCircle,
		TimeSchedule:   tpl.TimeRules,
		TemplateID:     tpl.ID,
		ApprovalStatus: model.ApprovalPending,
	}

	return area, nil
}

func (s *RestrictedAreaService) ImportGisData(ctx context.Context, req *model.GisImportRequest, userID int64) (*model.GisImportRecord, error) {
	batchNo := fmt.Sprintf("GIS%s", strings.ToUpper(strings.ReplaceAll(uuid.New().String()[:12], "-", "")))

	record := &model.GisImportRecord{
		ImportBatchNo: batchNo,
		FileName:      req.FileName,
		SourceType:    req.SourceType,
		ImportStatus:  "processing",
		ImportedBy:    userID,
	}

	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}

	var features []map[string]interface{}
	if err := json.Unmarshal([]byte(req.Features), &features); err != nil {
		record.ImportStatus = "failed"
		s.db.WithContext(ctx).Model(record).Update("import_status", "failed")
		return nil, fmt.Errorf("解析GIS数据失败: %w", err)
	}

	record.TotalCount = len(features)
	successCount := 0
	failedCount := 0
	var failedDetails []map[string]interface{}

	for i, feat := range features {
		area, err := s.parseGisFeature(feat, batchNo, userID)
		if err != nil {
			failedCount++
			failedDetails = append(failedDetails, map[string]interface{}{
				"index": i,
				"error": err.Error(),
			})
			continue
		}

		if err := s.db.WithContext(ctx).Create(area).Error; err != nil {
			failedCount++
			failedDetails = append(failedDetails, map[string]interface{}{
				"index": i,
				"error": err.Error(),
			})
			continue
		}
		successCount++
	}

	record.SuccessCount = successCount
	record.FailedCount = failedCount
	record.ImportStatus = "completed"

	if len(failedDetails) > 0 {
		data, _ := json.Marshal(failedDetails)
		record.FailedDetails = model.JSON(data)
	}

	if err := s.db.WithContext(ctx).Save(record).Error; err != nil {
		return nil, err
	}

	logger.Sugar.Infof("GIS import completed: batch=%s, total=%d, success=%d, failed=%d", batchNo, record.TotalCount, successCount, failedCount)
	return record, nil
}

func (s *RestrictedAreaService) ImportGisFile(ctx context.Context, file *app.FormFile, sourceType string, userID int64) (*model.GisImportRecord, error) {
	if file == nil {
		return nil, fmt.Errorf("文件不能为空")
	}

	maxSize := int64(50 * 1024 * 1024)
	if file.Size > maxSize {
		return nil, fmt.Errorf("文件大小超过50MB限制")
	}

	f, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("无法读取文件: %w", err)
	}
	defer f.Close()

	fileBytes, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("读取文件内容失败: %w", err)
	}

	fileName := file.Filename
	var features []map[string]interface{}

	switch {
	case strings.HasSuffix(strings.ToLower(fileName), ".geojson"),
		strings.HasSuffix(strings.ToLower(fileName), ".json"):
		features, err = s.parseGeoJSONFile(fileBytes)
		if err != nil {
			return nil, fmt.Errorf("解析GeoJSON文件失败: %w", err)
		}
		if sourceType == "" {
			sourceType = "geojson"
		}

	case strings.HasSuffix(strings.ToLower(fileName), ".zip"):
		features, err = s.parseShapefileZip(fileBytes)
		if err != nil {
			return nil, fmt.Errorf("解析Shapefile压缩包失败: %w", err)
		}
		if sourceType == "" {
			sourceType = "shp"
		}

	default:
		if strings.EqualFold(sourceType, "shp") {
			return nil, fmt.Errorf("Shapefile需以ZIP压缩包形式上传（含.shp/.dbf/.shx文件）")
		}
		features, err = s.parseGeoJSONFile(fileBytes)
		if err != nil {
			return nil, fmt.Errorf("不支持的文件格式，请上传.geojson或.zip文件: %w", err)
		}
	}

	if len(features) == 0 {
		return nil, fmt.Errorf("未解析到有效的GIS要素数据")
	}

	req := &model.GisImportRequest{
		SourceType: sourceType,
		FileName:   fileName,
	}
	featuresJSON, _ := json.Marshal(features)
	req.Features = model.JSON(featuresJSON)

	return s.ImportGisData(ctx, req, userID)
}

func (s *RestrictedAreaService) parseGeoJSONFile(data []byte) ([]map[string]interface{}, error) {
	var geojson struct {
		Type     string                   `json:"type"`
		Features []map[string]interface{} `json:"features"`
	}
	if err := json.Unmarshal(data, &geojson); err != nil {
		return nil, fmt.Errorf("JSON格式错误: %w", err)
	}

	if geojson.Type == "FeatureCollection" && len(geojson.Features) > 0 {
		return geojson.Features, nil
	}

	if geojson.Type == "Feature" {
		var feat map[string]interface{}
		if err := json.Unmarshal(data, &feat); err != nil {
			return nil, err
		}
		return []map[string]interface{}{feat}, nil
	}

	var features []map[string]interface{}
	if err := json.Unmarshal(data, &features); err != nil {
		return nil, fmt.Errorf("无法识别的GeoJSON结构，需FeatureCollection或Feature数组")
	}
	return features, nil
}

func (s *RestrictedAreaService) parseShapefileZip(data []byte) ([]map[string]interface{}, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("ZIP解压失败: %w", err)
	}

	var geojsonFile *zip.File
	for _, f := range r.File {
		name := strings.ToLower(f.Name)
		if strings.HasSuffix(name, ".geojson") || strings.HasSuffix(name, ".json") {
			geojsonFile = f
			break
		}
	}

	if geojsonFile != nil {
		rc, err := geojsonFile.Open()
		if err != nil {
			return nil, fmt.Errorf("读取ZIP内GeoJSON文件失败: %w", err)
		}
		defer rc.Close()
		jsonBytes, err := io.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("读取ZIP内文件内容失败: %w", err)
		}
		return s.parseGeoJSONFile(jsonBytes)
	}

	var shpExists, dbfExists bool
	for _, f := range r.File {
		name := strings.ToLower(f.Name)
		if strings.HasSuffix(name, ".shp") {
			shpExists = true
		}
		if strings.HasSuffix(name, ".dbf") {
			dbfExists = true
		}
	}
	if !shpExists || !dbfExists {
		return nil, fmt.Errorf("ZIP压缩包中缺少必需的Shapefile组件（需包含.shp和.dbf文件），建议将Shapefile通过ogr2ogr转换为GeoJSON后打包上传")
	}

	return nil, fmt.Errorf("原生Shapefile(.shp)二进制解析暂不支持，请使用ogr2ogr将Shapefile转换为GeoJSON格式后上传，命令: ogr2ogr -f GeoJSON output.geojson input.shp")
}

func (s *RestrictedAreaService) parseGisFeature(feat map[string]interface{}, batchNo string, userID int64) (*model.RestrictedAreaExt, error) {
	props, _ := feat["properties"].(map[string]interface{})
	geom, _ := feat["geometry"].(map[string]interface{})

	if geom == nil {
		return nil, fmt.Errorf("缺少geometry字段")
	}

	geomType, _ := geom["type"].(string)
	if geomType != "Polygon" && geomType != "MultiPolygon" {
		return nil, fmt.Errorf("不支持的几何类型: %s，仅支持Polygon和MultiPolygon", geomType)
	}

	name := "GIS导入区域"
	if n, ok := props["name"].(string); ok && n != "" {
		name = n
	}

	areaType := model.AreaTypeMall
	if t, ok := props["area_type"].(string); ok && t != "" {
		areaType = model.RestrictedAreaType(t)
	}

	var level int = 2
	if l, ok := props["level"].(float64); ok {
		level = int(l)
	}

	geomJSON, _ := json.Marshal(geom)

	centerLat, centerLng := 0.0, 0.0
	if coords, ok := geom["coordinates"].([]interface{}); ok && len(coords) > 0 {
		if polygon, ok := coords[0].([]interface{}); ok && len(polygon) > 0 {
			if pt, ok := polygon[0].([]interface{}); ok && len(pt) >= 2 {
				centerLng, _ = pt[0].(float64)
				centerLat, _ = pt[1].(float64)
			}
		}
	}

	var restrictHazardClasses string
	if rhc, ok := props["restrict_hazard_classes"].(string); ok {
		restrictHazardClasses = rhc
	}
	var restrictVehicleTypes string
	if rvt, ok := props["restrict_vehicle_types"].(string); ok {
		restrictVehicleTypes = rvt
	}

	return &model.RestrictedAreaExt{
		RestrictedArea: model.RestrictedArea{
			Name:                  name,
			AreaType:              areaType,
			Level:                 level,
			BoundaryPolygon:       model.JSON(geomJSON),
			CenterLatitude:        centerLat,
			CenterLongitude:       centerLng,
			Source:                "official",
			Status:                0,
			RestrictHazardClasses: restrictHazardClasses,
			RestrictVehicleTypes:  restrictVehicleTypes,
		},
		ShapeType:      model.ShapePolygon,
		GisImportID:    batchNo,
		CreatedBy:      userID,
		ApprovalStatus: model.ApprovalPending,
	}, nil
}

func (s *RestrictedAreaService) ListGisImports(ctx context.Context, page, pageSize int) ([]*model.GisImportRecord, int64, error) {
	var list []*model.GisImportRecord
	var total int64

	query := s.db.WithContext(ctx).Model(&model.GisImportRecord{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func generateCirclePolygon(lat, lng, radius float64) model.JSON {
	points := make([][]float64, 0)
	sides := 36
	latRad := lat * 3.141592653589793 / 180.0
	mPerDegLat := 110540.0
	mPerDegLng := 111320.0 * math.Cos(latRad)
	for i := 0; i < sides; i++ {
		angle := float64(i) * (2 * math.Pi / float64(sides))
		dLat := (radius * math.Cos(angle)) / mPerDegLat
		dLng := (radius * math.Sin(angle)) / mPerDegLng
		points = append(points, []float64{lng + dLng, lat + dLat})
	}
	points = append(points, points[0])

	geometry := map[string]interface{}{
		"type":        "Polygon",
		"coordinates": [][][]float64{points},
	}
	data, _ := json.Marshal(geometry)
	return model.JSON(data)
}
