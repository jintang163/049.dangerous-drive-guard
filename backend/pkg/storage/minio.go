package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type UploadResult struct {
	ObjectName  string `json:"object_name"`
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

type MinIOStorage struct {
	client *minio.Client
	cfg    *config.MinIOConfig
}

var globalMinIO *MinIOStorage

func InitMinIO(cfg *config.StorageConfig) (*MinIOStorage, error) {
	client, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client error: %w", err)
	}

	storage := &MinIOStorage{
		client: client,
		cfg:    &cfg.MinIO,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := storage.EnsureBuckets(ctx); err != nil {
		return nil, fmt.Errorf("ensure buckets error: %w", err)
	}

	globalMinIO = storage
	logger.Sugar.Infof("MinIO connected: %s, buckets: %v", cfg.MinIO.Endpoint, cfg.MinIO.Buckets)
	return storage, nil
}

func GetMinIO() *MinIOStorage {
	return globalMinIO
}

func (s *MinIOStorage) EnsureBuckets(ctx context.Context) error {
	for _, bucket := range s.cfg.Buckets {
		exists, err := s.client.BucketExists(ctx, bucket)
		if err != nil {
			return fmt.Errorf("check bucket %s error: %w", bucket, err)
		}
		if !exists {
			if err := s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
				return fmt.Errorf("create bucket %s error: %w", bucket, err)
			}
			logger.Sugar.Infof("MinIO bucket created: %s", bucket)
		}
	}
	return nil
}

func (s *MinIOStorage) UploadFile(ctx context.Context, bucket, objectName string, reader io.Reader, size int64, contentType string) (*UploadResult, error) {
	uploadInfo, err := s.client.PutObject(ctx, bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("upload file error: %w", err)
	}

	url, err := s.GetFileURL(ctx, bucket, objectName, 24*time.Hour)
	if err != nil {
		logger.Sugar.Warnf("get presigned url failed: %v", err)
		url = ""
	}

	return &UploadResult{
		ObjectName:  uploadInfo.Key,
		URL:         url,
		ContentType: contentType,
		Size:        size,
	}, nil
}

func (s *MinIOStorage) UploadFatigueVideo(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (*UploadResult, error) {
	bucket := s.cfg.Buckets["video"]
	return s.UploadFile(ctx, bucket, objectName, reader, size, contentType)
}

func (s *MinIOStorage) UploadFatigueImage(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (*UploadResult, error) {
	bucket := s.cfg.Buckets["image"]
	return s.UploadFile(ctx, bucket, objectName, reader, size, contentType)
}

func (s *MinIOStorage) GetFileURL(ctx context.Context, bucket, objectName string, expires time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := s.client.PresignedGetObject(ctx, bucket, objectName, expires, reqParams)
	if err != nil {
		return "", fmt.Errorf("get presigned url error: %w", err)
	}
	return presignedURL.String(), nil
}

func (s *MinIOStorage) GetVideoPlayURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	if expires == 0 {
		expires = 24 * time.Hour
	}
	bucket := s.cfg.Buckets["video"]
	return s.GetFileURL(ctx, bucket, objectName, expires)
}

func (s *MinIOStorage) GetImageURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	if expires == 0 {
		expires = time.Hour
	}
	bucket := s.cfg.Buckets["image"]
	return s.GetFileURL(ctx, bucket, objectName, expires)
}

func (s *MinIOStorage) DeleteVideo(ctx context.Context, objectName string) error {
	bucket := s.cfg.Buckets["video"]
	return s.DeleteFile(ctx, bucket, objectName)
}

func (s *MinIOStorage) DeleteImage(ctx context.Context, objectName string) error {
	bucket := s.cfg.Buckets["image"]
	return s.DeleteFile(ctx, bucket, objectName)
}

func (s *MinIOStorage) DeleteFile(ctx context.Context, bucket, objectName string) error {
	return s.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
}

func (s *MinIOStorage) ListFiles(ctx context.Context, bucket, prefix string) ([]minio.ObjectInfo, error) {
	var files []minio.ObjectInfo
	objectCh := s.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})
	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}
		files = append(files, object)
	}
	return files, nil
}

func (s *MinIOStorage) GetObject(ctx context.Context, bucket, objectName string) (*minio.Object, error) {
	return s.client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
}

func (s *MinIOStorage) GetVideoBucket() string {
	return s.cfg.Buckets["video"]
}

func (s *MinIOStorage) GetImageBucket() string {
	return s.cfg.Buckets["image"]
}
