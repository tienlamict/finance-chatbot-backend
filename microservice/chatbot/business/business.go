package business

import (
	"errors"
	"strings"

	"finance-chatbot/microservice/chatbot/entity"
)

type ChatUsecase struct {
	repo     entity.ChatRepository
	aiClient entity.AIClient
}

func NewChatBusiness(repo entity.ChatRepository, aiClient entity.AIClient) *ChatUsecase {
	return &ChatUsecase{repo: repo, aiClient: aiClient}
}

func (uc *ChatUsecase) SendMessage(userID, prompt string) (assistantMsg entity.ChatMessage, err error) {
	prompt = strings.TrimSpace(prompt)
	if userID == "" || prompt == "" {
		return assistantMsg, errors.New("user_id and prompt required")
	}

	// persist user message
	if err = uc.repo.Create(&entity.ChatMessage{UserID: userID, Role: entity.RoleUser, Content: prompt}); err != nil {
		return
	}

	// load short history (last 20, newest first)
	history, _ := uc.repo.ListByUser(userID, 20)

	// ask AI via gRPC
	reply, err := uc.aiClient.GenerateReply(userID, history, prompt)
	if err != nil {
		return
	}

	assistantMsg = entity.ChatMessage{UserID: userID, Role: entity.RoleAssistant, Content: reply}
	err = uc.repo.Create(&assistantMsg)
	return
}

func (uc *ChatUsecase) GetHistory(userID string, limit int) ([]entity.ChatMessage, error) {
	return uc.repo.ListByUser(userID, limit)
}
