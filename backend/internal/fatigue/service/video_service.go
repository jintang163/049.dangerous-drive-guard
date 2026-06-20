package service

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
	"github.com/dangerous-drive-guard/backend/pkg/storage"
)

type VideoUploadResult struct {
	URL         string `json:"url"`
	ObjectName  string `json:"object_name"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Duration    int    `json:"duration"`
}

type ImageUploadResult struct {
	URL         string `json:"url"`
	ObjectName  string `json:"object_name"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

type VideoService struct {
	db    *gorm.DB
	minio *storage.MinIOStorage
}

func NewVideoService(cfg *config.Config) *VideoService {
	return &VideoService{
		db:    database.GetDB(),
		minio: storage.GetMinIO(),
	}
}

func (s *VideoService) UploadAlarmVideo(ctx context.Context, vehicleID, alarmID int64, filename string, file io.ReadCloser, size int64) (*VideoUploadResult, error) {
	defer file.Close()

	if err := s.validateAlarm(ctx, alarmID, vehicleID); err != nil {
		return nil, err
	}

	timestamp := time.Now().Unix()
	objectName := fmt.Sprintf("fatigue/videos/%d/%d/%d_%s", vehicleID, alarmID, timestamp, sanitizeFilename(filename))
	contentType := getContentType(filename)

	uploadResult, err := s.minio.UploadFatigueVideo(ctx, objectName, file, size, contentType)
	if err != nil {
		return nil, fmt.Errorf("upload video: %w", err)
	}

	if err := s.updateAlarmVideoURL(ctx, alarmID, uploadResult.ObjectName); err != nil {
		_ = s.minio.DeleteVideo(ctx, objectName)
		return nil, err
	}

	url, err := s.minio.GetVideoPlayURL(ctx, objectName, 24*time.Hour)
	if err != nil {
		logger.Global.Warn("get video presigned url failed", zap.Error(err))
		url = uploadResult.URL
	}

	return &VideoUploadResult{
		URL:         url,
		ObjectName:  uploadResult.ObjectName,
		ContentType: uploadResult.ContentType,
		Size:        uploadResult.Size,
		Duration:    0,
	}, nil
}

func (s *VideoService) UploadAlarmSnapshot(ctx context.Context, vehicleID, alarmID int64, filename string, file io.ReadCloser, size int64) (*ImageUploadResult, error) {
	defer file.Close()

	if err := s.validateAlarm(ctx, alarmID, vehicleID); err != nil {
		return nil, err
	}

	timestamp := time.Now().Unix()
	objectName := fmt.Sprintf("fatigue/images/%d/%d/%d_%s", vehicleID, alarmID, timestamp, sanitizeFilename(filename))
	contentType := getContentType(filename)

	uploadResult, err := s.minio.UploadFatigueImage(ctx, objectName, file, size, contentType)
	if err != nil {
		return nil, fmt.Errorf("upload snapshot: %w", err)
	}

	if err := s.updateAlarmSnapshotURL(ctx, alarmID, uploadResult.ObjectName); err != nil {
		_ = s.minio.DeleteImage(ctx, objectName)
		return nil, err
	}

	url, err := s.minio.GetImageURL(ctx, objectName, time.Hour)
	if err != nil {
		logger.Global.Warn("get image presigned url failed", zap.Error(err))
		url = uploadResult.URL
	}

	return &ImageUploadResult{
		URL:         url,
		ObjectName:  uploadResult.ObjectName,
		ContentType: uploadResult.ContentType,
		Size:        uploadResult.Size,
		Width:       0,
		Height:      0,
	}, nil
}

func (s *VideoService) GetAlarmVideoURL(ctx context.Context, alarmID int64) (string, error) {
	objectName, err := s.getAlarmVideoObjectName(ctx, alarmID)
	if err != nil {
		return "", err
	}
	if objectName == "" {
		return "", fmt.Errorf("video not found for alarm %d", alarmID)
	}
	return s.minio.GetVideoPlayURL(ctx, objectName, 24*time.Hour)
}

func (s *VideoService) GetAlarmSnapshotURL(ctx context.Context, alarmID int64) (string, error) {
	objectName, err := s.getAlarmSnapshotObjectName(ctx, alarmID)
	if err != nil {
		return "", err
	}
	if objectName == "" {
		return "", fmt.Errorf("snapshot not found for alarm %d", alarmID)
	}
	return s.minio.GetImageURL(ctx, objectName, time.Hour)
}

func (s *VideoService) BatchGetVideoURLs(ctx context.Context, alarmIDs []int64) (map[int64]string, error) {
	result := make(map[int64]string)
	if len(alarmIDs) == 0 {
		return result, nil
	}

	rows, err := s.db.WithContext(ctx).Table("fatigue_alarms").
		Select("id, video_clip_url").
		Where("id IN ?", alarmIDs).
		Rows()
	if err != nil {
		return nil, fmt.Errorf("query alarms: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var objectName string
		if err := rows.Scan(&id, &objectName); err != nil {
			continue
		}
		if objectName != "" {
			url, err := s.minio.GetVideoPlayURL(ctx, objectName, 24*time.Hour)
			if err != nil {
				logger.Global.Warn("get video url failed",
					zap.Int64("alarm_id", id),
					zap.Error(err),
				)
				continue
			}
			result[id] = url
		}
	}
	return result, nil
}

func (s *VideoService) DeleteAlarmVideo(ctx context.Context, alarmID int64) error {
	var result := struct {
		VideoClipURL string `gorm:"column:video_clip_url"`
		SnapImageURL string `gorm:"column:snap_image_url"`
	}{}
	err := s.db.WithContext(ctx).Table("fatigue_alarms").
		Select("video_clip_url, snap_image_url").
		Where("id = ?", alarmID).
		Scan(&result).Error
	if err != nil {
		return fmt.Errorf("query alarm: %w", err)
	}
	videoURL := result.VideoClipURL
	snapURL := result.SnapImageURL

	if videoURL != "" {
		if err := s.minio.DeleteVideo(ctx, videoURL); err != nil {
			logger.Global.Warn("delete video failed",
				zap.Int64("alarm_id", alarmID),
				zap.String("object", videoURL),
				zap.Error(err),
			)
		}
	}

	if snapURL != "" {
		if err := s.minio.DeleteImage(ctx, snapURL); err != nil {
			logger.Global.Warn("delete snapshot failed",
				zap.Int64("alarm_id", alarmID),
				zap.String("object", snapURL),
				zap.Error(err),
			)
		}
	}

	result := s.db.WithContext(ctx).Exec(`
		UPDATE fatigue_alarms SET video_clip_url = '', snap_image_url = '', updated_at = NOW()
		WHERE id = ?`, alarmID,
	)
	if result.Error != nil {
		return fmt.Errorf("update alarm: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("alarm %d not found", alarmID)
	}
	return nil
}

func (s *VideoService) GetVideoObject(ctx context.Context, alarmID int64) (*storage.MinIOStorage, string, error) {
	objectName, err := s.getAlarmVideoObjectName(ctx, alarmID)
	if err != nil {
		return nil, "", err
	}
	if objectName == "" {
		return nil, "", fmt.Errorf("video not found for alarm %d", alarmID)
	}
	return s.minio, objectName, nil
}

func (s *VideoService) validateAlarm(ctx context.Context, alarmID, vehicleID int64) error {
	var count int64
	result := s.db.WithContext(ctx).Table("fatigue_alarms").
		Where("id = ? AND vehicle_id = ?", alarmID, vehicleID).
		Count(&count)
	if result.Error != nil {
		return fmt.Errorf("query alarm: %w", result.Error)
	}
	if count == 0 {
		return fmt.Errorf("alarm %d not found or vehicle mismatch", alarmID)
	}
	return nil
}

func (s *VideoService) updateAlarmVideoURL(ctx context.Context, alarmID int64, objectName string) error {
	result := s.db.WithContext(ctx).Exec(`
		UPDATE fatigue_alarms SET video_clip_url = ?, updated_at = NOW() WHERE id = ?`,
		objectName, alarmID,
	)
	if result.Error != nil {
		return fmt.Errorf("update alarm video url: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("alarm %d not found", alarmID)
	}
	return nil
}

func (s *VideoService) updateAlarmSnapshotURL(ctx context.Context, alarmID int64, objectName string) error {
	result := s.db.WithContext(ctx).Exec(`
		UPDATE fatigue_alarms SET snap_image_url = ?, updated_at = NOW() WHERE id = ?`,
		objectName, alarmID,
	)
	if result.Error != nil {
		return fmt.Errorf("update alarm snapshot url: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("alarm %d not found", alarmID)
	}
	return nil
}

func (s *VideoService) getAlarmVideoObjectName(ctx context.Context, alarmID int64) (string, error) {
	var objectName string
	result := s.db.WithContext(ctx).Table("fatigue_alarms").
		Select("video_clip_url").
		Where("id = ?", alarmID).
		Scan(&objectName)
	if result.Error != nil {
		return "", fmt.Errorf("query alarm: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return "", fmt.Errorf("alarm %d not found", alarmID)
	}
	return objectName, nil
}

func (s *VideoService) getAlarmSnapshotObjectName(ctx context.Context, alarmID int64) (string, error) {
	var objectName string
	result := s.db.WithContext(ctx).Table("fatigue_alarms").
		Select("snap_image_url").
		Where("id = ?", alarmID).
		Scan(&objectName)
	if result.Error != nil {
		return "", fmt.Errorf("query alarm: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return "", fmt.Errorf("alarm %d not found", alarmID)
	}
	return objectName, nil
}

func sanitizeFilename(filename string) string {
	filename = filepath.Base(filename)
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "..", "")
	return filename
}

func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".avi":
		return "video/x-msvideo"
	case ".mov":
		return "video/quicktime"
	case ".mkv":
		return "video/x-matroska"
	case ".webm":
		return "video/webm"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
