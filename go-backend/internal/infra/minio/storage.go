// Package minio provides an S3-compatible storage implementation using the MinIO Go SDK.
package minio

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
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
	client *minio.Client

	bucketName string

	// accessKey and secretKey are used to build a short-lived presigning client against publicEndpoint.
	accessKey string
	secretKey string

	// publicEndpoint is the host:port (or DNS name) clients use to fetch objects; it must match the Host in presigned requests.
	publicEndpoint string

	// presignSecure is whether presigned URLs use HTTPS (must match how clients open the link).
	presignSecure bool

	// signingRegion is the AWS SigV4 region used for signing; it avoids network calls to infer region when minting URLs.
	signingRegion string

	log *zap.SugaredLogger
}

// NewStorage initializes a MinIO client, ensures the target bucket exists, and stores settings for presigned GET URLs.
// endpoint is the address the application uses to reach MinIO (e.g. s3:9000 in Docker).
// publicEndpoint is the address embedded in presigned URLs for external clients (e.g. localhost:9000 on a dev machine).
func NewStorage(ctx context.Context, endpoint, publicEndpoint, accessKey, secretKey, bucketName string, useSSL bool, signingRegion string, log *zap.SugaredLogger) (*Storage, error) {
	publicEndpoint = strings.TrimSpace(publicEndpoint)
	if publicEndpoint == "" {
		return nil, fmt.Errorf("minio public endpoint is required")
	}
	signingRegion = strings.TrimSpace(signingRegion)
	if signingRegion == "" {
		signingRegion = "us-east-1"
	}

	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: signingRegion,
	}
	client, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to init minio client: %w", err)
	}

	// Ensure bucket exists
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
		client:         client,
		bucketName:     bucketName,
		accessKey:      accessKey,
		secretKey:      secretKey,
		publicEndpoint: publicEndpoint,
		presignSecure:  useSSL,
		signingRegion:  signingRegion,
		log:            log,
	}, nil
}

// GetURL returns a time-limited presigned GET URL for objectPath.
//
// Signing uses a separate MinIO client bound to publicEndpoint so the Authorization header matches the Host
// the mobile client contacts, avoiding SignatureDoesNotMatch when the app uses a different hostname than the server.
func (s *Storage) GetURL(ctx context.Context, objectPath string) (string, error) {
	if objectPath == "" {
		return "", nil
	}

	signer, err := minio.New(s.publicEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.accessKey, s.secretKey, ""),
		Secure: s.presignSecure,
		Region: s.signingRegion,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create URL signer: %w", err)
	}

	expires := time.Minute * 15
	presignedURL, err := signer.PresignedGetObject(ctx, s.bucketName, objectPath, expires, nil)
	if err != nil {
		return "", fmt.Errorf("minio.PresignedGetObject: %w", err)
	}

	return presignedURL.String(), nil
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
