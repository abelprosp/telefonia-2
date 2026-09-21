package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/luxus-connect/telefonia/api/internal/config"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

// MaxObjectBytes is the hard ceiling for objects loaded into memory (invoice import).
const MaxObjectBytes int64 = 256 * 1024 * 1024

type Client struct {
	presigner *s3.PresignClient
	client    *s3.Client
}

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

func (c *Client) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	out, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
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
	}
	return buf, nil
}
