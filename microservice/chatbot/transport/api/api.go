package api

import (
	"github.com/gin-gonic/gin"

	"finance-chatbot/microservice/chatbot/business"
)

type API struct {
	uc business.ChatUsecase
}

func NewAPI(uc business.ChatUsecase) *API {
	return &API{uc: uc}
}

func (a *API) Register(r *gin.Engine) {
	g := r.Group("/v1/chatbot")
	{
		g.POST("/messages", a.SendMessageHandler())
		g.GET("/messages", a.ListMessagesHandler())
	}
}
