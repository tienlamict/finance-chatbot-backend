package api

import (
	"finance-chatbot/microservice/chatbot/business"
)

type API struct {
	uc business.ChatUsecase
}

func NewAPI(uc business.ChatUsecase) *API {
	return &API{uc: uc}
}
