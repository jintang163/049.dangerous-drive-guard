package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	voiceSvc "github.com/dangerous-drive-guard/backend/internal/voice/service"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

var voiceService *voiceSvc.VoiceInterventionService

func initService() {
	if voiceService == nil {
		voiceService = voiceSvc.NewVoiceInterventionService()
	}
}

type VoiceInterventionHandler struct{}

func NewVoiceInterventionHandler() *VoiceInterventionHandler {
	initService()
	return &VoiceInterventionHandler{}
}

func (h *VoiceInterventionHandler) RegisterRoutes(router *app.RouterGroup, authMw app.HandlerFunc, roleMw func(...string) app.HandlerFunc) {
	r := router.Group("/voice-intervention", authMw)
	{
		audios := r.Group("/audios")
		{
			audios.GET("", h.ListAudios)
			audios.GET("/:id", h.GetAudio)
			audios.POST("", roleMw("admin"), h.CreateAudio)
			audios.PUT("/:id", roleMw("admin"), h.UpdateAudio)
			audios.DELETE("/:id", roleMw("admin"), h.DeleteAudio)
			audios.POST("/upload", roleMw("admin"), h.UploadAudio)
			audios.POST("/:id/set-default", roleMw("admin"), h.SetDefaultAudio)
		}
		strategies := r.Group("/strategies", roleMw("admin", "dispatcher"))
		{
			strategies.GET("", h.ListStrategies)
			strategies.GET("/:id", h.GetStrategy)
			strategies.POST("", roleMw("admin"), h.CreateStrategy)
			strategies.PUT("/:id", roleMw("admin"), h.UpdateStrategy)
			strategies.DELETE("/:id", roleMw("admin"), h.DeleteStrategy)
		}
		logs := r.Group("/logs", roleMw("admin", "dispatcher"))
		{
			logs.GET("", h.ListLogs)
			logs.PUT("/:id/status", h.UpdateLogStatus)
			logs.POST("/:id/ack", roleMw("driver", "admin"), h.DriverAckLog)
		}
		r.POST("/test-play", roleMw("admin", "dispatcher"), h.TestPlay)
		r.POST("/match", h.MatchStrategy)
		r.GET("/statistics", roleMw("admin", "dispatcher"), h.GetStatistics)
	}
}

// ============================================================
// 音频库
// ============================================================

func (h *VoiceInterventionHandler) ListAudios(ctx context.Context, c *app.RequestContext) {
	initService()
	driverID, _ := strconv.ParseInt(string(c.Query("driver_id")), 10, 64)
	orgID, _ := strconv.ParseInt(string(c.Query("org_id")), 10, 64)
	category := model.AudioCategory(c.Query("category"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	audios, total, err := voiceService.ListAudios(ctx, driverID, orgID, category, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, audios, total, page, pageSize)
}

func (h *VoiceInterventionHandler) GetAudio(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid audio id")
		return
	}
	audio, err := voiceService.GetAudio(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, audio)
}

func (h *VoiceInterventionHandler) CreateAudio(ctx context.Context, c *app.RequestContext) {
	initService()
	var audio model.VoiceInterventionAudio
	if err := c.BindAndValidate(&audio); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	audio.CreatedBy = toInt64(userID)
	err := voiceService.CreateAudio(ctx, &audio)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, audio)
}

func (h *VoiceInterventionHandler) UpdateAudio(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid audio id")
		return
	}
	var req struct {
		Name        string              `json:"name"`
		Category    model.AudioCategory `json:"category"`
		AudioURL    string              `json:"audio_url"`
		DurationSec int                 `json:"duration_sec"`
		FileSize    int64               `json:"file_size"`
		Volume      int                 `json:"volume"`
		Description string              `json:"description"`
		Tags        json.RawMessage     `json:"tags"`
		IsEnabled   *bool               `json:"is_enabled"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.AudioURL != "" {
		updates["audio_url"] = req.AudioURL
	}
	if req.DurationSec > 0 {
		updates["duration_sec"] = req.DurationSec
	}
	if req.FileSize > 0 {
		updates["file_size"] = req.FileSize
	}
	if req.Volume > 0 {
		updates["volume"] = req.Volume
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if len(req.Tags) > 0 {
		updates["tags"] = req.Tags
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}
	updates["updated_at"] = time.Now()
	err = voiceService.UpdateAudio(ctx, id, updates)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"id": id, "updated": true})
}

func (h *VoiceInterventionHandler) DeleteAudio(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid audio id")
		return
	}
	err = voiceService.DeleteAudio(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"id": id, "deleted": true})
}

func (h *VoiceInterventionHandler) SetDefaultAudio(ctx context.Context, c *app.RequestContext) {
	initService()
	audioID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid audio id")
		return
	}
	var req struct {
		DriverID int64               `json:"driver_id"`
		Category model.AudioCategory `json:"category" binding:"required"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	err = voiceService.SetDefaultAudio(ctx, req.DriverID, audioID, req.Category)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"audio_id": audioID, "is_default": true})
}

func (h *VoiceInterventionHandler) UploadAudio(ctx context.Context, c *app.RequestContext) {
	initService()
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file not found: "+err.Error())
		return
	}
	if fileHeader.Size > 20*1024*1024 {
		response.BadRequest(c, "file too large (>20MB)")
		return
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".mp3" && ext != ".wav" && ext != ".m4a" && ext != ".ogg" {
		response.BadRequest(c, "unsupported audio format: "+ext)
		return
	}
	name := c.DefaultPostForm("name", fileHeader.Filename)
	category := model.AudioCategory(c.DefaultPostForm("category", "custom"))
	description := c.DefaultPostForm("description", "")
	driverID, _ := strconv.ParseInt(string(c.DefaultPostForm("driver_id", "0")), 10, 64)
	volume, _ := strconv.Atoi(string(c.DefaultPostForm("volume", "80")))

	uploadDir := "./uploads/audios"
	_ = os.MkdirAll(uploadDir, 0755)
	savedName := strconv.FormatInt(time.Now().UnixNano(), 10) + ext
	savedPath := filepath.Join(uploadDir, savedName)

	file, err := fileHeader.Open()
	if err != nil {
		response.InternalError(c, "open file error: "+err.Error())
		return
	}
	defer file.Close()
	out, err := os.Create(savedPath)
	if err != nil {
		response.InternalError(c, "save file error: "+err.Error())
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		response.InternalError(c, "copy file error: "+err.Error())
		return
	}

	audioURL := "/uploads/audios/" + savedName
	audio := &model.VoiceInterventionAudio{
		DriverID:    driverID,
		Name:        name,
		Category:    category,
		AudioURL:    audioURL,
		AudioFormat: strings.TrimPrefix(ext, "."),
		DurationSec: 0,
		FileSize:    fileHeader.Size,
		Volume:      volume,
		Description: description,
		IsDefault:   false,
		IsEnabled:   true,
	}
	userID, _ := c.Get("user_id")
	audio.CreatedBy = toInt64(userID)

	err = voiceService.CreateAudio(ctx, audio)
	if err != nil {
		_ = os.Remove(savedPath)
		response.InternalError(c, "save audio record error: "+err.Error())
		return
	}
	response.Success(c, audio)
}

// ============================================================
// 干预策略
// ============================================================

func (h *VoiceInterventionHandler) ListStrategies(ctx context.Context, c *app.RequestContext) {
	initService()
	driverID, _ := strconv.ParseInt(string(c.Query("driver_id")), 10, 64)
	orgID, _ := strconv.ParseInt(string(c.Query("org_id")), 10, 64)
	strategyType := model.InterventionStrategyType(c.Query("strategy_type"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	strategies, total, err := voiceService.ListStrategies(ctx, driverID, orgID, strategyType, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, strategies, total, page, pageSize)
}

func (h *VoiceInterventionHandler) GetStrategy(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid strategy id")
		return
	}
	strategy, err := voiceService.GetStrategy(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, strategy)
}

func (h *VoiceInterventionHandler) CreateStrategy(ctx context.Context, c *app.RequestContext) {
	initService()
	var strategy model.VoiceInterventionStrategy
	if err := c.BindAndValidate(&strategy); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	strategy.CreatedBy = toInt64(userID)
	err := voiceService.CreateStrategy(ctx, &strategy)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, strategy)
}

func (h *VoiceInterventionHandler) UpdateStrategy(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid strategy id")
		return
	}
	var req struct {
		Name               string                          `json:"name"`
		StrategyType       model.InterventionStrategyType  `json:"strategy_type"`
		Priority           *int                            `json:"priority"`
		IsEnabled          *bool                           `json:"is_enabled"`
		IsDefault          *bool                           `json:"is_default"`
		AlarmTrigger       *model.InterventionAlarmTrigger `json:"alarm_trigger"`
		AudioIDs           json.RawMessage                 `json:"audio_ids"`
		ForceHighVolume    *bool                           `json:"force_high_volume"`
		ForceVolumePercent *int                            `json:"force_volume_percent"`
		PlayTimes          *int                            `json:"play_times"`
		PlayIntervalSec    *int                            `json:"play_interval_sec"`
		ShuffleAudios      *bool                           `json:"shuffle_audios"`
		EmotionalMode      *bool                           `json:"emotional_mode"`
		CooldownSeconds    *int                            `json:"cooldown_seconds"`
		Description        string                          `json:"description"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.StrategyType != "" {
		updates["strategy_type"] = req.StrategyType
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}
	if req.AlarmTrigger != nil {
		b, _ := json.Marshal(req.AlarmTrigger)
		updates["alarm_trigger"] = string(b)
	}
	if len(req.AudioIDs) > 0 {
		updates["audio_ids"] = string(req.AudioIDs)
	}
	if req.ForceHighVolume != nil {
		updates["force_high_volume"] = *req.ForceHighVolume
	}
	if req.ForceVolumePercent != nil {
		updates["force_volume_percent"] = *req.ForceVolumePercent
	}
	if req.PlayTimes != nil {
		updates["play_times"] = *req.PlayTimes
	}
	if req.PlayIntervalSec != nil {
		updates["play_interval_sec"] = *req.PlayIntervalSec
	}
	if req.ShuffleAudios != nil {
		updates["shuffle_audios"] = *req.ShuffleAudios
	}
	if req.EmotionalMode != nil {
		updates["emotional_mode"] = *req.EmotionalMode
	}
	if req.CooldownSeconds != nil {
		updates["cooldown_seconds"] = *req.CooldownSeconds
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	updates["updated_at"] = time.Now()
	err = voiceService.UpdateStrategy(ctx, id, updates)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"id": id, "updated": true})
}

func (h *VoiceInterventionHandler) DeleteStrategy(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid strategy id")
		return
	}
	err = voiceService.DeleteStrategy(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"id": id, "deleted": true})
}

// ============================================================
// 干预日志
// ============================================================

func (h *VoiceInterventionHandler) ListLogs(ctx context.Context, c *app.RequestContext) {
	initService()
	vehicleID, _ := strconv.ParseInt(string(c.Query("vehicle_id")), 10, 64)
	driverID, _ := strconv.ParseInt(string(c.Query("driver_id")), 10, 64)
	alarmID, _ := strconv.ParseInt(string(c.Query("alarm_id")), 10, 64)
	status := model.InterventionPlayStatus(c.Query("status"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	var startTime, endTime time.Time
	if s := string(c.Query("start_time")); s != "" {
		startTime, _ = time.Parse(time.RFC3339, s)
	}
	if e := string(c.Query("end_time")); e != "" {
		endTime, _ = time.Parse(time.RFC3339, e)
	}
	logs, total, err := voiceService.ListLogs(ctx, vehicleID, driverID, alarmID, status, startTime, endTime, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, logs, total, page, pageSize)
}

func (h *VoiceInterventionHandler) UpdateLogStatus(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid log id")
		return
	}
	var req struct {
		Status       model.InterventionPlayStatus `json:"status" binding:"required"`
		ErrorMsg     string                       `json:"error_msg"`
		DurationMs   int64                        `json:"duration_ms"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	err = voiceService.UpdateLogPlayStatus(ctx, id, req.Status, req.ErrorMsg, req.DurationMs)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"id": id, "status": req.Status})
}

func (h *VoiceInterventionHandler) DriverAckLog(ctx context.Context, c *app.RequestContext) {
	initService()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid log id")
		return
	}
	err = voiceService.DriverAckLog(ctx, id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"id": id, "ack": true})
}

// ============================================================
// 测试 & 统计
// ============================================================

func (h *VoiceInterventionHandler) TestPlay(ctx context.Context, c *app.RequestContext) {
	initService()
	var req model.AudioTestPlayRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.Volume <= 0 || req.Volume > 100 {
		req.Volume = 80
	}
	userID, _ := c.Get("user_id")
	log, err := voiceService.TestPlayAudio(ctx, int(req.VehicleID), int(req.AudioID), req.Volume, int(toInt64(userID)))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, log)
}

func (h *VoiceInterventionHandler) MatchStrategy(ctx context.Context, c *app.RequestContext) {
	initService()
	var req struct {
		DriverID          int64   `json:"driver_id" binding:"required"`
		OrgID             int64   `json:"org_id"`
		AlarmLevel        int     `json:"alarm_level" binding:"required"`
		AlarmType         string  `json:"alarm_type" binding:"required"`
		FatigueScore      float64 `json:"fatigue_score"`
		ContinuousMinutes int     `json:"continuous_minutes"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := voiceService.MatchStrategy(ctx, req.DriverID, req.OrgID, req.AlarmLevel, req.AlarmType, req.FatigueScore, req.ContinuousMinutes)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *VoiceInterventionHandler) GetStatistics(ctx context.Context, c *app.RequestContext) {
	initService()
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	stats, err := voiceService.GetStatistics(ctx, days)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
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
	case string:
		if id, err := strconv.ParseInt(x, 10, 64); err == nil {
			return id
		}
	}
	return 0
}

// silence unused imports
var _ = hlog.DefaultLogger()
var _ = http.StatusOK
