package api

import (
	"net/http"
	"strconv"

	"finance-chatbot/microservice/chatbot/business"

	"github.com/gin-gonic/gin"
)

type ChatAPI struct{ uc *business.ChatUsecase }

func NewAPI(uc *business.ChatUsecase) *ChatAPI { return &ChatAPI{uc: uc} }

func (a *ChatAPI) Register(r gin.IRouter) {
	r.POST("/v1/send-message", a.postChat)
	r.GET("/v1/get-history", a.getHistory)
}

func (a *ChatAPI) postChat(c *gin.Context) {
	var req struct{ UserID, Message string }
	if err := c.BindJSON(&req); err != nil || req.UserID == "" || req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and message are required"})
		return
	}
	msg, err := a.uc.SendMessage(req.UserID, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reply": msg.Content, "id": msg.ID})
}

func (a *ChatAPI) getHistory(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	items, err := a.uc.GetHistory(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
