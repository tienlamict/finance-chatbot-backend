// internal/infrastructure/http/handlers/chat_handler.go
package api

import (
	"net/http"

	"finance-chatbot/microservice/chatbot/business"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	UC *business.ChatBusiness
}

// POST /api/chat/messages
// body: { "conversation_id": "optional-uuid", "text": "câu hỏi..." }
func (h *ChatHandler) SendMessage(c *gin.Context) {
	var req struct {
		ConversationID string `json:"conversation_id"`
		Text           string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Lấy userID từ JWT (middleware đã set)
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	out, err := h.UC.SendMessage(c.Request.Context(), business.SendMessageInput{
		UserID:         userID,
		ConversationID: req.ConversationID,
		Text:           req.Text,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversation_id": out.ConversationID,
		"reply":           out.AssistantText,
		"model":           out.Model,
	})
}
