package api

import (
	"finance-chatbot/microservice/rag/business"

	"github.com/gin-gonic/gin"
)

// RAGHandler handles RAG-related HTTP requests
type RAGHandler struct {
	business business.RAGBusiness
}

// NewRAGHandler creates a new RAG handler
func NewRAGHandler(biz business.RAGBusiness) *RAGHandler {
	return &RAGHandler{
		business: biz,
	}
}

// RegisterRoutes registers all RAG routes with authentication middleware
func (h *RAGHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	rag := r.Group("/rag")
	{
		// Upload document to RAG from MinIO (requires authentication)
		rag.POST("/upload", authMiddleware, h.UploadDocument)

		// Upload document directly to RAG (multipart/form-data, requires authentication)
		rag.POST("/upload-for-rag", authMiddleware, h.UploadDocumentDirect)

		// Query RAG for document retrieval (optional auth for logging)
		rag.GET("/query", h.QueryDocuments)

		// Get document info (requires authentication)
		rag.GET("/documents/:doc_id", authMiddleware, h.GetDocument)

		// List user's documents (requires authentication)
		rag.GET("/documents", authMiddleware, h.ListUserDocuments)

		// Delete document (requires authentication)
		rag.DELETE("/documents/:doc_id", authMiddleware, h.DeleteDocument)
	}
}
