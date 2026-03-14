package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/horoshi10v/tires-shop/internal/domain"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioStorage struct {
	client     *minio.Client
	bucketName string
	publicURL  string
	logger     *slog.Logger
}

// NewMinioStorage initializes a new MinIO/S3 client and ensures the target bucket exists and is public.
func NewMinioStorage(endpoint, accessKey, secretKey, bucketName, publicURL string, useSSL bool, logger *slog.Logger) (domain.StorageService, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	if err == nil && !exists {
		// Automatically create the bucket if it doesn't exist
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			logger.Error("failed to create bucket", slog.String("bucket", bucketName))
		} else {
			// Make the bucket public for reading (allows frontend to access images directly)
			policy := fmt.Sprintf(`{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::%s/*"]}]}`, bucketName)
			client.SetBucketPolicy(ctx, bucketName, policy)
		}
	}

	return &minioStorage{
		client:     client,
		bucketName: bucketName,
		publicURL:  publicURL,
		logger:     logger,
	}, nil
}

func (s *minioStorage) UploadPhoto(ctx context.Context, file io.Reader, originalFilename string) (string, error) {
	// Log the original filename (resolves unused parameter warning and helps with debugging)
	s.logger.Debug("processing image upload", slog.String("original_name", originalFilename))

	// 1. Decode the original image
	img, err := imaging.Decode(file, imaging.AutoOrientation(true))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// 2. Resize. If the width is greater than 1080px, scale it down proportionally.
	if img.Bounds().Dx() > 1080 {
		img = imaging.Resize(img, 1080, 0, imaging.Lanczos)
	}

	// 3. Convert to JPEG (fast, reliable, works on ARM without CGO)
	var buf bytes.Buffer
	err = imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(80))
	if err != nil {
		return "", fmt.Errorf("failed to encode to jpeg: %w", err)
	}

	// 4. Generate a unique filename
	filename := fmt.Sprintf("lots/%s.jpg", uuid.New().String())

	// 5. Upload to MinIO
	_, err = s.client.PutObject(ctx, s.bucketName, filename, &buf, int64(buf.Len()), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to minio: %w", err)
	}

	// 6. Return the public URL to be saved in the database
	// Example: http://localhost:9000/tires-shop/lots/123-456.jpg
	fileURL := fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucketName, filename)

	s.logger.Debug("photo uploaded successfully", slog.String("url", fileURL))
	return fileURL, nil
}

func (s *minioStorage) DeletePhoto(ctx context.Context, fileURL string) error {
	// fileURL looks like "http://localhost:9000/tires-shop/lots/123-456.jpg"
	// We need to extract the object name within the bucket: "lots/123-456.jpg"
	prefix := fmt.Sprintf("%s/%s/", s.publicURL, s.bucketName)
	objectName := strings.TrimPrefix(fileURL, prefix)

	// Protection: if the prefix doesn't match, it's a foreign URL (or bad config)
	if objectName == fileURL {
		s.logger.Warn("file url does not match minio prefix, skipping deletion", slog.String("url", fileURL))
		// Do not return an error to avoid blocking Lot deletion from the DB,
		// but log a Warning instead.
		return nil
	}

	err := s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file from minio: %w", err)
	}

	s.logger.Debug("photo deleted successfully from storage", slog.String("object", objectName))
	return nil
}
