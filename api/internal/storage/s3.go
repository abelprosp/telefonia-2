package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
<<<<<<< HEAD
	"os"
	"path/filepath"
=======
>>>>>>> 6b82d54 (fix tenant isolation and invoice processing reliability)
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/luxus-connect/telefonia/api/internal/config"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

// MaxObjectBytes is the hard ceiling for objects loaded *into memory*.
// Safe default for incidental GetObject callers: 512 MiB.
// Invoice import uses OpenObject + disk materialization up to MaxDiskObjectBytes.
const DefaultMaxObjectBytes int64 = 512 * 1024 * 1024

// DefaultMaxDiskObjectBytes is the ceiling for streamed invoice imports written to
// a temporary file (not held fully in RAM). Based on typical VPS disk headroom for
// MinIO-backed TXT invoices; override with OBJECT_STORAGE_MAX_DISK_OBJECT_BYTES.
const DefaultMaxDiskObjectBytes int64 = 2 * 1024 * 1024 * 1024 // 2 GiB

var MaxObjectBytes int64 = DefaultMaxObjectBytes
var MaxDiskObjectBytes int64 = DefaultMaxDiskObjectBytes

func ConfigureMaxObjectBytes(n int64) {
	if n > 0 {
		MaxObjectBytes = n
	}
}

func ConfigureMaxDiskObjectBytes(n int64) {
	if n > 0 {
		MaxDiskObjectBytes = n
	}
}

type Client struct {
	presigner *s3.PresignClient
	client    *s3.Client
}

<<<<<<< HEAD
type ObjectStream struct {
	Body          io.ReadCloser
	ContentLength *int64
}

type MaterializedObject struct {
	Path   string
	SHA256 string
	Size   int64
}
=======
const maxObjectBytes = 50 * 1024 * 1024
>>>>>>> 6b82d54 (fix tenant isolation and invoice processing reliability)

func NewClient(cfg config.Config) (*Client, error) {
	if cfg.ObjectStorageServiceURL == "" {
		return nil, fmt.Errorf("OBJECT_STORAGE_SERVICE_URL is required")
	}
	awsCfg := aws.Config{
		Region: "auto",
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.ObjectStorageAccessKeyID,
			cfg.ObjectStorageSecretKey,
			"",
		),
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.ObjectStorageServiceURL)
		o.UsePathStyle = true
	})
	publicEndpoint := cfg.ObjectStoragePublicURL
	if publicEndpoint == "" {
		publicEndpoint = cfg.ObjectStorageServiceURL
	}
	presignClient := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(publicEndpoint)
		o.UsePathStyle = true
	})
	return &Client{
		client:    client,
		presigner: s3.NewPresignClient(presignClient),
	}, nil
}

func (c *Client) CreatePresignedUploadURL(ctx context.Context, bucket, key string, expires time.Duration, contentType *string) (*models.PresignedURLModel, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if contentType != nil && *contentType != "" {
		input.ContentType = contentType
	}
	req, err := c.presigner.PresignPutObject(ctx, input, s3.WithPresignExpires(expires))
	if err != nil {
		return nil, err
	}
	return &models.PresignedURLModel{
		URL: req.URL, HTTPMethod: req.Method, ExpiresAtUTC: time.Now().UTC().Add(expires),
	}, nil
}

func (c *Client) CreatePresignedDownloadURL(ctx context.Context, bucket, key string, expires time.Duration) (*models.PresignedURLModel, error) {
	req, err := c.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return nil, err
	}
	return &models.PresignedURLModel{
		URL: req.URL, HTTPMethod: req.Method, ExpiresAtUTC: time.Now().UTC().Add(expires),
	}, nil
}

func (c *Client) OpenObject(ctx context.Context, bucket, key string) (*ObjectStream, error) {
	out, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	if out.ContentLength != nil && *out.ContentLength > MaxDiskObjectBytes {
		_ = out.Body.Close()
		return nil, fmt.Errorf("object exceeds maximum disk size of %d bytes", MaxDiskObjectBytes)
	}
	return &ObjectStream{Body: out.Body, ContentLength: out.ContentLength}, nil
}

// MaterializeObject streams the object to a temp file while computing SHA-256.
// Caller must remove Path when done (os.Remove).
func (c *Client) MaterializeObject(ctx context.Context, bucket, key string) (*MaterializedObject, error) {
	stream, err := c.OpenObject(ctx, bucket, key)
	if err != nil {
		return nil, err
	}
	defer stream.Body.Close()

	dir := filepath.Join(os.TempDir(), "luxus-invoice-import")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	f, err := os.CreateTemp(dir, "invoice-*.bin")
	if err != nil {
		return nil, err
	}
	path := f.Name()
	cleanup := true
	defer func() {
		_ = f.Close()
		if cleanup {
			_ = os.Remove(path)
		}
	}()

	h := sha256.New()
	limited := io.LimitReader(stream.Body, MaxDiskObjectBytes+1)
	n, err := io.Copy(io.MultiWriter(f, h), limited)
	if err != nil {
		return nil, err
	}
	if n > MaxDiskObjectBytes {
		return nil, fmt.Errorf("object exceeds maximum disk size of %d bytes", MaxDiskObjectBytes)
	}
	if err := f.Sync(); err != nil {
		return nil, err
	}
	cleanup = false
	return &MaterializedObject{
		Path:   path,
		SHA256: hex.EncodeToString(h.Sum(nil)),
		Size:   n,
	}, nil
}

func (c *Client) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	out, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
<<<<<<< HEAD
	if out.ContentLength != nil && *out.ContentLength > MaxObjectBytes {
		return nil, fmt.Errorf("object exceeds maximum size of %d bytes", MaxObjectBytes)
	}
	limited := io.LimitReader(out.Body, MaxObjectBytes+1)
	buf, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(buf)) > MaxObjectBytes {
		return nil, fmt.Errorf("object exceeds maximum size of %d bytes", MaxObjectBytes)
=======
	buf := make([]byte, 0, 1024*1024)
	reader := io.LimitReader(out.Body, maxObjectBytes+1)
	for {
		chunk := make([]byte, 32*1024)
		n, readErr := reader.Read(chunk)
		if n > 0 {
			buf = append(buf, chunk[:n]...)
		}
		if readErr != nil {
			break
		}
>>>>>>> 6b82d54 (fix tenant isolation and invoice processing reliability)
	}
	if int64(len(buf)) > maxObjectBytes {
		return nil, fmt.Errorf("object exceeds maximum size of %d bytes", maxObjectBytes)
	}
	return buf, nil
}
