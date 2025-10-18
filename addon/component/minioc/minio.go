// finance-chatbot-backend/addon/component/minioc/minio.go
package minioc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"finance-chatbot/addon/sctx"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	BucketName      string `mapstructure:"bucket_name"`
	Region          string `mapstructure:"region"`
}

type MinioComponent struct {
	id         string
	client     *minio.Client
	bucketName string
}

func NewMinioComponent(id string, cfg Config) (*MinioComponent, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Ensure bucket exists
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{Region: cfg.Region})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &MinioComponent{
		id:         id,
		client:     client,
		bucketName: cfg.BucketName,
	}, nil
}

func (m *MinioComponent) ID() string {
	return m.id
}

func (m *MinioComponent) InitFlags() {
	// Implement if needed for CLI flags
}

func (m *MinioComponent) Activate(serviceCtx sctx.ServiceContext) error {
	return nil
}

func (m *MinioComponent) Stop() error {
	return nil
}

// UploadFile uploads a file to MinIO/S3
// Returns the storage key (path in bucket)
func (m *MinioComponent) UploadFile(ctx context.Context, objectName string, data []byte, contentType string) (string, error) {
	reader := bytes.NewReader(data)
	size := int64(len(data))

	_, err := m.client.PutObject(ctx, m.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return the storage key (bucket + object name)
	storageKey := fmt.Sprintf("%s/%s", m.bucketName, objectName)
	return storageKey, nil
}

// DownloadFile downloads a file from MinIO/S3
func (m *MinioComponent) DownloadFile(ctx context.Context, objectName string) ([]byte, error) {
	obj, err := m.client.GetObject(ctx, m.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	return data, nil
}

// DeleteFile deletes a file from MinIO/S3
func (m *MinioComponent) DeleteFile(ctx context.Context, objectName string) error {
	err := m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// GetPresignedURL generates a presigned URL for temporary access
func (m *MinioComponent) GetPresignedURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, m.bucketName, objectName, expires, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

// GenerateStorageKey creates a unique storage key for a file
func GenerateStorageKey(userID, messageID, filename string) string {
	timestamp := time.Now().UnixNano()
	ext := filepath.Ext(filename)
	return fmt.Sprintf("attachments/%s/%s/%d%s", userID, messageID, timestamp, ext)
}
