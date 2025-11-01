package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/rag/entity"

	"github.com/gin-gonic/gin"
)

// UploadDocumentDirect handles direct file upload to RAG (multipart/form-data)
// POST /api/rag/upload-for-rag
// Form fields: doc_id, file, collection
func (h *RAGHandler) UploadDocumentDirect(c *gin.Context) {
	// Get user ID from context
	requester, exists := c.Get("requester")
	if !exists {
		c.JSON(http.StatusUnauthorized, core.ErrUnauthorized)
		return
	}
	userID := requester.(core.Requester).GetSubject()

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(100 << 20); err != nil { // 100 MB max
		c.JSON(http.StatusBadRequest, core.ErrBadRequest.WithError("Failed to parse multipart form").WithTrace(err))
		return
	}

	// Extract form fields
	docID := c.PostForm("doc_id")
	collection := c.PostForm("collection")

	if docID == "" {
		c.JSON(http.StatusBadRequest, core.ErrBadRequest.WithError("doc_id is required"))
		return
	}
	if collection == "" {
		collection = entity.DefaultCollection
	}

	// Get uploaded file
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, core.ErrBadRequest.WithError("file is required").WithTrace(err))
		return
	}
	defer file.Close()

	// Create temp file to store uploaded content
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("rag_upload_%s_%s", docID, fileHeader.Filename))

	// Write file to temp location
	dst, err := os.Create(tempFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, core.ErrInternalServerError.WithError("Failed to create temp file").WithTrace(err))
		return
	}
	defer dst.Close()
	defer os.Remove(tempFile) // Clean up temp file

	// Copy uploaded content to temp file
	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, core.ErrInternalServerError.WithError("Failed to save temp file").WithTrace(err))
		return
	}

	// Close temp file before uploading
	dst.Close()

	// Upload to RAG service
	doc, err := h.business.UploadDocumentDirect(c.Request.Context(), tempFile, docID, collection, userID)
	if err != nil {
		if errResp, ok := err.(*core.DefaultError); ok {
			c.JSON(errResp.StatusCode(), errResp)
			return
		}
		c.JSON(http.StatusInternalServerError, core.ErrInternalServerError.WithTrace(err))
		return
	}

	c.JSON(http.StatusOK, core.ResponseData(doc))
}
