// Package minio provides an S3-compatible storage implementation using the MinIO Go SDK.
package minio

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"go-backend/internal/logger"
)

var tracer = otel.Tracer("internal/infra/minio")

// Storage handles file operations with a MinIO/S3 bucket with integrated observability.
type Storage struct {
	client     *minio.Client
	bucketName string
	log        *zap.SugaredLogger
}

// NewStorage initializes a MinIO client, ensures the target bucket exists, and sets up logging.
func NewStorage(ctx context.Context, endpoint, accessKey, secretKey, bucketName string, useSSL bool, log *zap.SugaredLogger) (*Storage, error) {
	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	}
	client, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to init minio client: %w", err)
	}

	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket %s: %w", bucketName, err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
		}
	}

	return &Storage{
		client:     client,
		bucketName: bucketName,
		log:        log,
	}, nil
}

// Upload stores a file in a specific subdirectory and returns the object key.
func (s *Storage) Upload(ctx context.Context, subdir, filename string, src io.Reader) (string, error) {
	ctx, span := tracer.Start(ctx, "Storage.MinIO.Upload")
	defer span.End()

	log := logger.FromContext(ctx, s.log)
	objectName := filepath.Join(subdir, filename)
	contentType := s.detectContentType(filename)

	span.SetAttributes(
		attribute.String("s3.bucket", s.bucketName),
		attribute.String("s3.object_key", objectName),
	)

	// Use -1 for size to enable automatic streaming.
	info, err := s.client.PutObject(ctx, s.bucketName, objectName, src, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		s.handleError(ctx, err, "failed to upload to s3", objectName)
		return "", fmt.Errorf("s3 upload: %w", err)
	}

	log.Debugw("file uploaded to s3", "object_key", info.Key, "size", info.Size)
	return info.Key, nil
}

// GetPresignedURL generates a temporary public link for the object.
// This is essential for the Flutter app to access private files like audio or photos.
func (s *Storage) GetPresignedURL(ctx context.Context, objectKey string, expires time.Duration) (string, error) {
	ctx, span := tracer.Start(ctx, "Storage.MinIO.GetPresignedURL")
	defer span.End()

	reqParams := make(url.Values)

	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucketName, objectKey, expires, reqParams)
	if err != nil {
		s.handleError(ctx, err, "failed to generate presigned url", objectKey)
		return "", fmt.Errorf("s3 presigned url: %w", err)
	}

	return presignedURL.String(), nil
}

// Load retrieves an object from the bucket as a ReadCloser.
func (s *Storage) Load(ctx context.Context, path string) (io.ReadCloser, error) {
	ctx, span := tracer.Start(ctx, "Storage.MinIO.Load")
	defer span.End()

	obj, err := s.client.GetObject(ctx, s.bucketName, path, minio.GetObjectOptions{})
	if err != nil {
		s.handleError(ctx, err, "failed to load object from s3", path)
		return nil, fmt.Errorf("s3 load: %w", err)
	}
	return obj, nil
}

// Delete removes a single object from the bucket.
func (s *Storage) Delete(ctx context.Context, path string) error {
	ctx, span := tracer.Start(ctx, "Storage.MinIO.Delete")
	defer span.End()

	err := s.client.RemoveObject(ctx, s.bucketName, path, minio.RemoveObjectOptions{})
	if err != nil {
		s.handleError(ctx, err, "failed to delete object from s3", path)
		return fmt.Errorf("s3 delete: %w", err)
	}
	return nil
}

// GetURL возвращает временную ссылку на файл (работает 15 минут)
func (s *Storage) GetURL(ctx context.Context, objectPath string) (string, error) {
	if objectPath == "" {
		return "", nil
	}

	expires := time.Minute * 15
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucketName, objectPath, expires, nil)
	if err != nil {
		return "", fmt.Errorf("minio.PresignedGetObject: %w", err)
	}
	return presignedURL.String(), nil
}

// HealthCheck verifies if the MinIO server is reachable.
func (s *Storage) HealthCheck(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("minio connectivity failed: %w", err)
	}
	if !exists {
		return fmt.Errorf("s3 bucket %s is missing", s.bucketName)
	}
	return nil
}

// handleError records the technical failure in the trace span and logs it via Zap.
func (s *Storage) handleError(ctx context.Context, err error, message, objectKey string) {
	log := logger.FromContext(ctx, s.log)
	span := trace.SpanFromContext(ctx)

	span.RecordError(err)
	span.SetStatus(codes.Error, message)
	span.SetAttributes(attribute.String("s3.error_object", objectKey))

	log.Errorw(message, "object_key", objectKey, "error", err)
}

// detectContentType is a helper to set correct headers for the S3 object.
func (s *Storage) detectContentType(filename string) string {
	switch filepath.Ext(filename) {
	case ".m4a":
		return "audio/mp4"
	case ".wav":
		return "audio/wav"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}
