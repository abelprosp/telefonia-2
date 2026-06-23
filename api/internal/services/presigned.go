package services

import (
	"context"
	"strings"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

const (
	defaultPresignedExpires = 900
	minPresignedExpires     = 60
	maxPresignedExpires     = 604800
)

type PresignedService struct {
	Storage ObjectStorage
}

type ObjectStorage interface {
	CreatePresignedUploadURL(ctx context.Context, bucket, key string, expires time.Duration, contentType *string) (*models.PresignedURLModel, error)
	CreatePresignedDownloadURL(ctx context.Context, bucket, key string, expires time.Duration) (*models.PresignedURLModel, error)
}

func (p *PresignedService) CreateUploadURL(ctx context.Context, input models.CreatePresignedUploadURLInput) (*models.PresignedURLModel, error) {
	if err := validateObjectStorage(input.BucketName, input.ObjectKey); err != nil {
		return nil, err
	}
	expires, err := resolveExpiration(input.ExpiresInSeconds)
	if err != nil {
		return nil, err
	}
	return p.Storage.CreatePresignedUploadURL(ctx, strings.TrimSpace(input.BucketName), strings.TrimSpace(input.ObjectKey), expires, input.ContentType)
}

func (p *PresignedService) CreateDownloadURL(ctx context.Context, input models.CreatePresignedDownloadURLInput) (*models.PresignedURLModel, error) {
	if err := validateObjectStorage(input.BucketName, input.ObjectKey); err != nil {
		return nil, err
	}
	expires, err := resolveExpiration(input.ExpiresInSeconds)
	if err != nil {
		return nil, err
	}
	return p.Storage.CreatePresignedDownloadURL(ctx, strings.TrimSpace(input.BucketName), strings.TrimSpace(input.ObjectKey), expires)
}

func validateObjectStorage(bucket, key string) error {
	if strings.TrimSpace(bucket) == "" {
		return httputil.ValidationError(notifications.ImportStorageBucketRequired)
	}
	if len(bucket) > 256 {
		return httputil.ValidationError(notifications.ImportStorageBucketMaxLength)
	}
	if strings.TrimSpace(key) == "" {
		return httputil.ValidationError(notifications.ImportStorageObjectKeyRequired)
	}
	if len(key) > 2048 {
		return httputil.ValidationError(notifications.ImportStorageObjectKeyMaxLength)
	}
	if strings.Contains(key, "..") || strings.Contains(key, "\x00") {
		return httputil.ValidationError(notifications.ObjectKeyInvalid)
	}
	return nil
}

func resolveExpiration(seconds *int) (time.Duration, error) {
	s := defaultPresignedExpires
	if seconds != nil {
		s = *seconds
	}
	if s < minPresignedExpires || s > maxPresignedExpires {
		return 0, httputil.ValidationError(notifications.PresignedExpiresInvalid)
	}
	return time.Duration(s) * time.Second, nil
}
