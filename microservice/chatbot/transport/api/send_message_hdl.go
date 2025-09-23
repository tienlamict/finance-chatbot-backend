// finance-chatbot-backend/microservice/chatbot/transport/api/send_message_hdl.go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"finance-chatbot/microservice/chatbot/entity"
)

func (a *API) resolveUserID(c *gin.Context, fallback string) string {
	// Ưu tiên: lấy từ middleware auth nếu có
	// Tuỳ hệ thống bạn đang set vào key nào:
	if v, ok := c.Get("user_id"); ok {
		if s, ok2 := v.(string); ok2 && s != "" {
			return s
		}
	}
	// Header tạm thời (nếu chưa có auth middleware)
	if s := c.GetHeader("X-User-Id"); s != "" {
		return s
	}
	return fallback
}

func (a *API) SendMessageHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req entity.SendMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userID := a.resolveUserID(c, req.UserID)
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id (from auth context or request body)"})
			return
		}

		res, err := a.uc.SendMessage(c.Request.Context(), req, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	}
}
