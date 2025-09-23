// finance-chatbot-backend/microservice/chatbot/transport/api/list_messages_hdl.go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"finance-chatbot/microservice/chatbot/entity"
)

func (a *API) ListMessagesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req entity.ListMessagesRequest
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.ConversationID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id is required"})
			return
		}
		res, err := a.uc.ListMessages(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	}
}
