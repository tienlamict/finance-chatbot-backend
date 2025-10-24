// finance-chatbot-backend/microservice/chatbot/transport/api/send_message_hdl.go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"finance-chatbot/addon/common"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/chatbot/entity"
)

func (a *API) resolveUserID(c *gin.Context, fallback string) string {
	// Extract user ID from authentication context (JWT token)
	if requester, ok := c.Get(core.KeyRequester); ok {
		if req, ok2 := requester.(core.Requester); ok2 {
			// The subject from the requester contains the user ID
			return req.GetSubject()
		}
	}
	// Fallback to header if no auth context (for development/testing)
	if s := c.GetHeader("X-User-Id"); s != "" {
		return s
	}
	// Note: fallback parameter (client-provided user_id) is ignored for security
	// User ID must come from authenticated context
	return ""
}

func (a *API) SendMessageHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		// Check content type - support both JSON and multipart/form-data
		contentType := c.ContentType()

		var req entity.SendMessageRequest
		var err error

		if contentType == "application/json" {
			// JSON request without files
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		} else {
			// Multipart form data with possible file uploads
			req, err = a.parseMultipartRequest(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}

		// Support conversation_id as query parameter (overrides body if present)
		if conversationID := c.Query("conversation_id"); conversationID != "" {
			req.ConversationID = conversationID
		}

		userID := a.resolveUserID(c, req.UserID)
		// UserID is now extracted from authentication context
		if userID == "" {
			common.WriteErrorResponse(c, core.ErrUnauthorized.WithError("user ID is required"))
			return
		}

		res, err := a.uc.SendMessage(c.Request.Context(), req, userID)
		if err != nil {
			// Handle different types of errors with appropriate HTTP status codes
			if err.Error() == "conversation not found or access denied" {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			if err.Error() == "user ID is required to create a new conversation" {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	}
}
