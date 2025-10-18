// finance-chatbot-backend/microservice/chatbot/transport/api/file_handler.go
package api

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"

	"finance-chatbot/addon/common"
	"finance-chatbot/microservice/chatbot/entity"
)

const (
	MaxUploadFiles = 10 // Maximum number of files per request
)

// parseMultipartRequest handles multipart/form-data requests with file uploads
func (a *API) parseMultipartRequest(c *gin.Context) (entity.SendMessageRequest, error) {
	var req entity.SendMessageRequest

	// Parse form data
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32 MB max memory
		return req, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	// Extract text fields
	req.Content = c.PostForm("content")
	if req.Content == "" {
		return req, fmt.Errorf("content is required")
	}

	req.ConversationID = c.PostForm("conversation_id")
	req.UserID = c.PostForm("user_id")

	// Optional fields
	if orgID := c.PostForm("org_id"); orgID != "" {
		req.OrgID = &orgID
	}
	if title := c.PostForm("title"); title != "" {
		req.Title = &title
	}

	// Process uploaded files
	form, err := c.MultipartForm()
	if err != nil {
		// No files uploaded, which is okay
		return req, nil
	}

	files := form.File["files"]
	if len(files) > MaxUploadFiles {
		return req, fmt.Errorf("too many files: maximum %d files allowed", MaxUploadFiles)
	}

	// Process each uploaded file
	req.Files = make([]entity.FileUploadInfo, 0, len(files))

	for _, fileHeader := range files {
		// Validate file
		if err := common.ValidateFileUpload(fileHeader, nil); err != nil {
			return req, fmt.Errorf("invalid file %s: %w", fileHeader.Filename, err)
		}

		// Read file content
		fileData, err := common.ReadFileContent(fileHeader)
		if err != nil {
			return req, fmt.Errorf("failed to read file %s: %w", fileHeader.Filename, err)
		}

		// Detect MIME type
		mimeType := common.DetectMimeType(fileData)

		// Calculate SHA256 hash
		sha256Hash := common.CalculateSHA256(fileData)

		// Extract PDF pages count if applicable (optional, can be enhanced)
		var pages *int
		if mimeType == "application/pdf" {
			// TODO: Implement PDF page counting if needed
			// For now, leave as nil
		}

		fileInfo := entity.FileUploadInfo{
			FileHeader: fileHeader,
			Content:    fileData,
			MimeType:   mimeType,
			SHA256Hash: sha256Hash,
			Pages:      pages,
		}

		req.Files = append(req.Files, fileInfo)
	}

	return req, nil
}

// GetAttachmentHandler returns an attachment by ID
func (a *API) GetAttachmentHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		attachmentID := c.Param("id")
		if attachmentID == "" {
			c.JSON(400, gin.H{"error": "attachment_id is required"})
			return
		}

		// This would need to be implemented in business layer
		// For now, returning not implemented
		c.JSON(501, gin.H{"error": "not implemented yet"})
	}
}

// ListMessageAttachmentsHandler lists all attachments for a message
func (a *API) ListMessageAttachmentsHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		messageID := c.Param("message_id")
		if messageID == "" {
			c.JSON(400, gin.H{"error": "message_id is required"})
			return
		}

		// This would need to be implemented in business layer
		// For now, returning not implemented
		c.JSON(501, gin.H{"error": "not implemented yet"})
	}
}

// Helper function to parse JSON from form field
func parseJSONField(formValue string, target interface{}) error {
	if formValue == "" {
		return nil
	}
	return json.Unmarshal([]byte(formValue), target)
}
