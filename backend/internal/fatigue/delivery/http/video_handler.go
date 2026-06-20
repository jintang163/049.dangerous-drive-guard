package http

import (
	"context"
	"io"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	videoSvc "github.com/dangerous-drive-guard/backend/internal/fatigue/service"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

type VideoHandler struct {
	videoService *videoSvc.VideoService
}

func NewVideoHandler(svc *videoSvc.VideoService) *VideoHandler {
	return &VideoHandler{
		videoService: svc,
	}
}

func (h *VideoHandler) RegisterRoutes(r *app.RouterGroup, authMiddleware app.HandlerFunc) {
	video := r.Group("/video", authMiddleware)
	{
		video.POST("/upload", h.UploadVideo)
		video.POST("/snapshot/upload", h.UploadSnapshot)
		video.GET("/:id/video", h.GetVideoURL)
		video.GET("/:id/snapshot", h.GetSnapshotURL)
		video.DELETE("/:id", h.DeleteVideo)
		video.GET("/:id/download", h.DownloadVideo)
	}
}

func (h *VideoHandler) UploadVideo(c context.Context, ctx *app.RequestContext) {
	vehicleID, err := strconv.ParseInt(ctx.PostForm("vehicle_id"), 10, 64)
	if err != nil || vehicleID <= 0 {
		response.BadRequest(ctx, "invalid vehicle_id")
		return
	}

	alarmID, err := strconv.ParseInt(ctx.PostForm("alarm_id"), 10, 64)
	if err != nil || alarmID <= 0 {
		response.BadRequest(ctx, "invalid alarm_id")
		return
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		response.BadRequest(ctx, "missing file")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.InternalError(ctx, "failed to open file: "+err.Error())
		return
	}
	defer file.Close()

	result, err := h.videoService.UploadAlarmVideo(c, vehicleID, alarmID, fileHeader.Filename, file, fileHeader.Size)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, result)
}

func (h *VideoHandler) UploadSnapshot(c context.Context, ctx *app.RequestContext) {
	vehicleID, err := strconv.ParseInt(ctx.PostForm("vehicle_id"), 10, 64)
	if err != nil || vehicleID <= 0 {
		response.BadRequest(ctx, "invalid vehicle_id")
		return
	}

	alarmID, err := strconv.ParseInt(ctx.PostForm("alarm_id"), 10, 64)
	if err != nil || alarmID <= 0 {
		response.BadRequest(ctx, "invalid alarm_id")
		return
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		response.BadRequest(ctx, "missing file")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.InternalError(ctx, "failed to open file: "+err.Error())
		return
	}
	defer file.Close()

	result, err := h.videoService.UploadAlarmSnapshot(c, vehicleID, alarmID, fileHeader.Filename, file, fileHeader.Size)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, result)
}

func (h *VideoHandler) GetVideoURL(c context.Context, ctx *app.RequestContext) {
	alarmID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || alarmID <= 0 {
		response.BadRequest(ctx, "invalid alarm id")
		return
	}

	url, err := h.videoService.GetAlarmVideoURL(c, alarmID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(ctx, err.Error())
			return
		}
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"url":       url,
		"alarm_id":  alarmID,
		"expire_in": 86400,
	})
}

func (h *VideoHandler) GetSnapshotURL(c context.Context, ctx *app.RequestContext) {
	alarmID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || alarmID <= 0 {
		response.BadRequest(ctx, "invalid alarm id")
		return
	}

	url, err := h.videoService.GetAlarmSnapshotURL(c, alarmID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(ctx, err.Error())
			return
		}
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"url":       url,
		"alarm_id":  alarmID,
		"expire_in": 3600,
	})
}

func (h *VideoHandler) DeleteVideo(c context.Context, ctx *app.RequestContext) {
	alarmID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || alarmID <= 0 {
		response.BadRequest(ctx, "invalid alarm id")
		return
	}

	if err := h.videoService.DeleteAlarmVideo(c, alarmID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(ctx, err.Error())
			return
		}
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"message":  "video and snapshot deleted successfully",
		"alarm_id": alarmID,
	})
}

func (h *VideoHandler) DownloadVideo(c context.Context, ctx *app.RequestContext) {
	alarmID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || alarmID <= 0 {
		response.BadRequest(ctx, "invalid alarm id")
		return
	}

	minio, objectName, err := h.videoService.GetVideoObject(c, alarmID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(ctx, err.Error())
			return
		}
		response.InternalError(ctx, err.Error())
		return
	}

	info, err := minio.StatObject(c, "video", objectName)
	if err != nil {
		response.InternalError(ctx, "failed to stat object: "+err.Error())
		return
	}

	obj, err := minio.GetObject(c, "video", objectName)
	if err != nil {
		response.InternalError(ctx, "failed to get object: "+err.Error())
		return
	}
	defer obj.Close()

	ctx.Header("Content-Disposition", "attachment; filename="+extractFilename(objectName))
	ctx.Header("Content-Type", info.ContentType)
	ctx.Header("Content-Length", strconv.FormatInt(info.Size, 10))
	ctx.Header("Accept-Ranges", "bytes")
	ctx.Header("Cache-Control", "public, max-age=3600")

	ctx.SetStatusCode(consts.StatusOK)

	_, _ = io.Copy(ctx.Response.BodyWriter(), obj)
}

func extractFilename(objectName string) string {
	parts := strings.Split(objectName, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "video.mp4"
}
