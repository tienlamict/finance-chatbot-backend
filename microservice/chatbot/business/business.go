package business

import (
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
	// 1. Trim và validate input
	// 2. Persist user message vào DB qua repo
	if err = uc.repo.Create(&entity.ChatMessage{UserID: userID, Role: entity.RoleUser, Content: prompt}); err != nil {
		return
	}

	// 3. Lấy short history (last 20)
	history, _ := uc.repo.ListByUser(userID, 20)

	// 4. Gọi AIClient.GenerateReply (gRPC call sang AI service)
	reply, err := uc.aiClient.GenerateReply(userID, history, prompt)

	// 5. Persist assistant reply vào DB
	assistantMsg = entity.ChatMessage{UserID: userID, Role: entity.RoleAssistant, Content: reply}
	err = uc.repo.Create(&assistantMsg)
	return
}

func (uc *ChatUsecase) GetHistory(userID string, limit int) ([]entity.ChatMessage, error) {
	return uc.repo.ListByUser(userID, limit)
}
