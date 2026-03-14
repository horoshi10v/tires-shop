package domain

import (
	"context"
	"io"
)

// StorageService defines the contract for saving and deleting files from object storage (MinIO/S3).
type StorageService interface {
	// UploadPhoto optimizes the image, saves it to S3, and returns the public URL.
	UploadPhoto(ctx context.Context, file io.Reader, originalFilename string) (string, error)

	// DeletePhoto removes the file from S3 using its public URL.
	DeletePhoto(ctx context.Context, fileURL string) error
}
