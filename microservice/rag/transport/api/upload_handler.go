package api

import (
	"net/http"

	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/rag/entity"

	"github.com/gin-gonic/gin"
)

// UploadDocument handles document upload to RAG
// POST /api/v1/rag/upload
// Body: { "doc_id", "collection", "minio_key", "filename", "mime_type", "sha256" }
func (h *RAGHandler) UploadDocument(c *gin.Context) {
	var req entity.CreateRAGDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, core.ErrBadRequest.WithError("Invalid request").WithTrace(err))
		return
	}

	// Get user ID from context (set by auth middleware)
	requester, exists := c.Get("requester")
	if !exists {
		c.JSON(http.StatusUnauthorized, core.ErrUnauthorized)
		return
	}

	userID := requester.(core.Requester).GetSubject()

	// Upload document
	doc, err := h.business.UploadDocumentForRAG(c.Request.Context(), &req, userID)
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
