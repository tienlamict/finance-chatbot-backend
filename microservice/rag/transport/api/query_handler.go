package api

import (
	"fmt"
	"net/http"

	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/rag/entity"

	"github.com/gin-gonic/gin"
)

// QueryDocuments handles RAG query requests
// GET /api/v1/rag/query?query=<q>&k=<int>&collection=<collection>
func (h *RAGHandler) QueryDocuments(c *gin.Context) {
	var req entity.QueryRAGRequest

	// Debug: Log raw query params
	fmt.Printf("[RAG Handler] Received query params: query=%s, k=%s, collection=%s\n",
		c.Query("query"), c.Query("k"), c.Query("collection"))

	if err := c.ShouldBindQuery(&req); err != nil {
		fmt.Printf("[RAG Handler] Binding error: %v\n", err)
		c.JSON(http.StatusBadRequest, core.ErrBadRequest.WithError("Invalid query parameters").WithTrace(err))
		return
	}

	fmt.Printf("[RAG Handler] Parsed request: query=%s, k=%d, collection=%s\n", req.Query, req.K, req.Collection)

	// Get user ID and conversation ID from context (optional)
	userID := ""
	conversationID := ""

	if requester, exists := c.Get("requester"); exists {
		userID = requester.(core.Requester).GetSubject()
	}

	// Conversation ID can be passed as query param or header
	if cid := c.Query("conversation_id"); cid != "" {
		conversationID = cid
	} else if cid := c.GetHeader("X-Conversation-ID"); cid != "" {
		conversationID = cid
	}

	// Query RAG
	results, err := h.business.QueryRAG(c.Request.Context(), &req, userID, conversationID)
	if err != nil {
		if errResp, ok := err.(*core.DefaultError); ok {
			c.JSON(errResp.StatusCode(), errResp)
			return
		}
		c.JSON(http.StatusInternalServerError, core.ErrInternalServerError.WithTrace(err))
		return
	}

	c.JSON(http.StatusOK, core.ResponseData(gin.H{
		"query":      req.Query,
		"k":          req.K,
		"collection": req.Collection,
		"results":    results,
		"count":      len(results),
	}))
}

// GetDocument retrieves a specific document by ID
// GET /api/v1/rag/documents/:doc_id
func (h *RAGHandler) GetDocument(c *gin.Context) {
	docID := c.Param("doc_id")
	if docID == "" {
		c.JSON(http.StatusBadRequest, core.ErrBadRequest.WithError("Document ID is required"))
		return
	}

	doc, err := h.business.GetDocument(c.Request.Context(), docID)
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

// ListUserDocuments lists documents uploaded by the authenticated user
// GET /api/v1/rag/documents?limit=20&offset=0
func (h *RAGHandler) ListUserDocuments(c *gin.Context) {
	// Get user ID from context
	requester, exists := c.Get("requester")
	if !exists {
		c.JSON(http.StatusUnauthorized, core.ErrUnauthorized)
		return
	}

	userID := requester.(core.Requester).GetSubject()

	// Parse pagination params
	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := parseIntParam(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := parseIntParam(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	docs, err := h.business.ListUserDocuments(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, core.ErrInternalServerError.WithTrace(err))
		return
	}

	c.JSON(http.StatusOK, core.ResponseData(gin.H{
		"documents": docs,
		"count":     len(docs),
		"limit":     limit,
		"offset":    offset,
	}))
}

// DeleteDocument deletes a document
// DELETE /api/v1/rag/documents/:doc_id
func (h *RAGHandler) DeleteDocument(c *gin.Context) {
	docID := c.Param("doc_id")
	if docID == "" {
		c.JSON(http.StatusBadRequest, core.ErrBadRequest.WithError("Document ID is required"))
		return
	}

	// Get user ID from context
	requester, exists := c.Get("requester")
	if !exists {
		c.JSON(http.StatusUnauthorized, core.ErrUnauthorized)
		return
	}

	userID := requester.(core.Requester).GetSubject()

	if err := h.business.DeleteDocument(c.Request.Context(), docID, userID); err != nil {
		if errResp, ok := err.(*core.DefaultError); ok {
			c.JSON(errResp.StatusCode(), errResp)
			return
		}
		c.JSON(http.StatusInternalServerError, core.ErrInternalServerError.WithTrace(err))
		return
	}

	c.JSON(http.StatusOK, core.ResponseData(gin.H{
		"message": "Document deleted successfully",
		"doc_id":  docID,
	}))
}

// Helper function
func parseIntParam(s string) (int, error) {
	var result int
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, &core.DefaultError{}
		}
		result = result*10 + int(s[i]-'0')
	}
	return result, nil
}
