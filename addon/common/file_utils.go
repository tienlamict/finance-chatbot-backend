// finance-chatbot-backend/addon/common/file_utils.go
package common

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

const (
	MaxFileSize = 50 * 1024 * 1024 // 50MB
)

// ValidateFileUpload validates file size and type
func ValidateFileUpload(fileHeader *multipart.FileHeader, allowedTypes []string) error {
	// Check file size
	if fileHeader.Size > MaxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", MaxFileSize)
	}

	// Check mime type if restrictions provided
	if len(allowedTypes) > 0 {
		file, err := fileHeader.Open()
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		// Read first 512 bytes for mime type detection
		buffer := make([]byte, 512)
		_, err = file.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read file: %w", err)
		}

		mtype := mimetype.Detect(buffer)
		detected := mtype.String()

		allowed := false
		for _, t := range allowedTypes {
			if strings.HasPrefix(detected, t) {
				allowed = true
				break
			}
		}

		if !allowed {
			return fmt.Errorf("file type %s is not allowed", detected)
		}
	}

	return nil
}

// ReadFileContent reads the entire file content from multipart.FileHeader
func ReadFileContent(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// CalculateSHA256 calculates SHA256 hash of file content
func CalculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// DetectMimeType detects the MIME type of file content
func DetectMimeType(data []byte) string {
	mtype := mimetype.Detect(data)
	return mtype.String()
}

// SanitizeFilename removes dangerous characters from filename
func SanitizeFilename(filename string) string {
	// Remove path separators and other dangerous chars
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, "..", "_")
	return filename
}
