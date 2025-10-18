// finance-chatbot-backend/addon/common/storage_interface.go
package common

import (
	"context"
	"time"
)

// StorageProvider defines the interface for object storage operations
type StorageProvider interface {
	UploadFile(ctx context.Context, objectName string, data []byte, contentType string) (string, error)
	DownloadFile(ctx context.Context, objectName string) ([]byte, error)
	DeleteFile(ctx context.Context, objectName string) error
	GetPresignedURL(ctx context.Context, objectName string, expires time.Duration) (string, error)
}
