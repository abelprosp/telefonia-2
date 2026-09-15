package services

import (
	"context"
	"strings"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/store"
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
	bucket := strings.TrimSpace(input.BucketName)
	key, err := tenantScopedObjectKey(ctx, strings.TrimSpace(input.ObjectKey))
	if err != nil {
		return nil, err
	}
	result, err := p.Storage.CreatePresignedUploadURL(ctx, bucket, key, expires, input.ContentType)
	if err != nil {
		return nil, err
	}
	result.BucketName = bucket
	result.ObjectKey = key
	return result, nil
}

func (p *PresignedService) CreateDownloadURL(ctx context.Context, input models.CreatePresignedDownloadURLInput) (*models.PresignedURLModel, error) {
	if err := validateObjectStorage(input.BucketName, input.ObjectKey); err != nil {
		return nil, err
	}
	expires, err := resolveExpiration(input.ExpiresInSeconds)
	if err != nil {
		return nil, err
	}
	bucket := strings.TrimSpace(input.BucketName)
	key := strings.TrimSpace(input.ObjectKey)
	if err := ensureTenantObjectKey(ctx, key); err != nil {
		return nil, err
	}
	result, err := p.Storage.CreatePresignedDownloadURL(ctx, bucket, key, expires)
	if err != nil {
		return nil, err
	}
	result.BucketName = bucket
	result.ObjectKey = key
	return result, nil
}

func tenantScopedObjectKey(ctx context.Context, key string) (string, error) {
	org := auth.OrganizationFromContext(ctx)
	if org == nil || strings.TrimSpace(org.ID) == "" {
		return "", httputil.BusinessError(notifications.SharedOrganizationRequired)
	}
	orgID := strings.TrimSpace(org.ID)
	key = strings.TrimLeft(strings.TrimSpace(key), "/")
	prefix := orgID + "/"
	if strings.HasPrefix(key, prefix) {
		return key, nil
	}
	// Reject attempts to write into another tenant's prefix.
	if parts := strings.SplitN(key, "/", 2); len(parts) == 2 && looksLikeOrgID(parts[0]) && parts[0] != orgID {
		return "", httputil.ForbiddenError(notifications.N("STORAGE_TENANT_MISMATCH", "Chave de armazenamento fora da organização autenticada."))
	}
	return prefix + key, nil
}

func ensureTenantObjectKey(ctx context.Context, key string) error {
	org := auth.OrganizationFromContext(ctx)
	if org == nil || strings.TrimSpace(org.ID) == "" {
		return httputil.BusinessError(notifications.SharedOrganizationRequired)
	}
	orgID := strings.TrimSpace(org.ID)
	key = strings.TrimLeft(strings.TrimSpace(key), "/")
	if strings.HasPrefix(key, orgID+"/") {
		return nil
	}
	// Objetos legados sem prefixo de tenant: só a org Luxus pode acessar.
	first := strings.SplitN(key, "/", 2)[0]
	if orgID == store.DefaultLuxusOrgID && !looksLikeOrgID(first) {
		return nil
	}
	return httputil.ForbiddenError(notifications.N("STORAGE_TENANT_MISMATCH", "Chave de armazenamento fora da organização autenticada."))
}

func looksLikeOrgID(v string) bool {
	v = strings.TrimSpace(v)
	if len(v) < 8 {
		return false
	}
	for _, r := range v {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') || r == '-' {
			continue
		}
		return false
	}
	return strings.Contains(v, "-")
}

func validateObjectStorage(bucket, key string) error {
	if strings.TrimSpace(bucket) == "" {
		return httputil.ValidationError(notifications.N("STORAGE_BUCKET_REQUIRED", "Bucket is required."))
	}
	if strings.TrimSpace(key) == "" {
		return httputil.ValidationError(notifications.N("STORAGE_OBJECT_KEY_REQUIRED", "Object key is required."))
	}
	if strings.Contains(key, "..") {
		return httputil.ValidationError(notifications.N("STORAGE_OBJECT_KEY_INVALID", "Object key is invalid."))
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
