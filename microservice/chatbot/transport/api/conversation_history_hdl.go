// finance-chatbot-backend/microservice/chatbot/transport/api/conversation_history_hdl.go
package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/chatbot/entity"
)

// GetConversationHistoryHandler handles GET /conversations/{conversationId}/messages
func (a *API) GetConversationHistoryHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		// Extract conversation ID from URL parameter
		conversationID := c.Param("conversationId")
		if conversationID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id is required"})
			return
		}

		// Get user ID from context (set by auth middleware)
		requester, ok := c.Get(core.KeyRequester)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing requester"})
			return
		}

		req := requester.(core.Requester)
		uid, err := core.FromBase58(req.GetSubject())
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
			return
		}
		userID := strconv.FormatUint(uint64(uid.GetLocalID()), 10)

		// Parse query parameters
		var reqParams entity.ConversationHistoryRequest
		reqParams.ConversationID = conversationID

		// Parse limit
		if limitStr := c.Query("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil {
				reqParams.Limit = limit
			}
		}
		if reqParams.Limit <= 0 {
			reqParams.Limit = 20
		}

		// Parse cursor parameters
		if before := c.Query("before"); before != "" {
			reqParams.Before = &before
		}
		if after := c.Query("after"); after != "" {
			reqParams.After = &after
		}

		// Parse time range
		if fromStr := c.Query("from"); fromStr != "" {
			// TODO: Parse ISO 8601 timestamp
			// For now, skip time parsing
		}
		if toStr := c.Query("to"); toStr != "" {
			// TODO: Parse ISO 8601 timestamp
			// For now, skip time parsing
		}

		// Parse order
		if order := c.Query("order"); order != "" {
			reqParams.Order = order
		}
		if reqParams.Order == "" {
			reqParams.Order = "asc"
		}

		// Parse search
		if search := c.Query("search"); search != "" {
			reqParams.Search = &search
		}

		// Parse include_attachments
		if includeAttachmentsStr := c.Query("include_attachments"); includeAttachmentsStr != "" {
			if includeAttachments, err := strconv.ParseBool(includeAttachmentsStr); err == nil {
				reqParams.IncludeAttachments = includeAttachments
			}
		}
		reqParams.IncludeAttachments = true // default

		// Call business logic
		res, err := a.uc.GetConversationHistory(c.Request.Context(), reqParams, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	}
}

// GetUserConversationsHandler handles GET /users/{userId}/conversations
func (a *API) GetUserConversationsHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		// Extract user ID from URL parameter
		userID := c.Param("userId")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}

		// Get requesting user ID from context (set by auth middleware)
		requester, ok := c.Get(core.KeyRequester)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing requester"})
			return
		}

		req := requester.(core.Requester)
		uid, err := core.FromBase58(req.GetSubject())
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
			return
		}
		requestingUserID := strconv.FormatUint(uint64(uid.GetLocalID()), 10)

		// Parse query parameters
		var reqParams entity.UserConversationsRequest
		reqParams.UserID = userID

		// Parse limit
		if limitStr := c.Query("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil {
				reqParams.Limit = limit
			}
		}
		if reqParams.Limit <= 0 {
			reqParams.Limit = 20
		}

		// Parse offset
		if offsetStr := c.Query("offset"); offsetStr != "" {
			if offset, err := strconv.Atoi(offsetStr); err == nil {
				reqParams.Offset = offset
			}
		}
		if reqParams.Offset < 0 {
			reqParams.Offset = 0
		}

		// Parse status
		if status := c.Query("status"); status != "" {
			reqParams.Status = &status
		}

		// Parse time range
		if fromStr := c.Query("from"); fromStr != "" {
			// TODO: Parse ISO 8601 timestamp
			// For now, skip time parsing
		}
		if toStr := c.Query("to"); toStr != "" {
			// TODO: Parse ISO 8601 timestamp
			// For now, skip time parsing
		}

		// Parse search
		if search := c.Query("search"); search != "" {
			reqParams.Search = &search
		}

		// Parse include_last_message
		if includeLastMessageStr := c.Query("include_last_message"); includeLastMessageStr != "" {
			if includeLastMessage, err := strconv.ParseBool(includeLastMessageStr); err == nil {
				reqParams.IncludeLastMessage = includeLastMessage
			}
		}
		reqParams.IncludeLastMessage = true // default

		// Call business logic
		res, err := a.uc.GetUserConversations(c.Request.Context(), reqParams, requestingUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	}
}

// GetConversationSummaryHandler handles GET /conversations/{conversationId}
func (a *API) GetConversationSummaryHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		// Extract conversation ID from URL parameter
		conversationID := c.Param("conversationId")
		if conversationID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id is required"})
			return
		}

		// Get user ID from context (set by auth middleware)
		requester, ok := c.Get(core.KeyRequester)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing requester"})
			return
		}

		req := requester.(core.Requester)
		uid, err := core.FromBase58(req.GetSubject())
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
			return
		}
		userID := strconv.FormatUint(uint64(uid.GetLocalID()), 10)

		// Call business logic
		res, err := a.uc.GetConversationSummary(c.Request.Context(), conversationID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	}
}
