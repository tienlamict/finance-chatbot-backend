package business

import (
	"context"
	idgen "finance-chatbot/addon/common"
	"finance-chatbot/microservice/chatbot/entity"
	"finance-chatbot/microservice/chatbot/gateways"
	"time"
)

type ChatBusiness struct {
	Repo     entity.ChatRepo
	AIClient gateways.AIClient
}

type SendMessageInput struct {
	UserID         string
	ConversationID string // optional
	Text           string
}

type SendMessageOutput struct {
	ConversationID string
	AssistantText  string
	Model          string
}

func (biz *ChatBusiness) SendMessage(ctx context.Context, in SendMessageInput) (*SendMessageOutput, error) {
	// 1) Lấy hoặc tạo conversation
	conv, err := biz.Repo.GetOrCreateConversation(ctx, in.UserID, in.ConversationID, nil)
	if err != nil {
		return nil, err
	}

	// 2) Lưu message của user
	userMsg := &entity.Message{
		ID:             idgen.NewV7(),
		ConversationID: conv.ID,
		Role:           entity.RoleUser,
		Content:        in.Text,
		ContentType:    "text",
	}
	if err := biz.Repo.CreateMessage(ctx, userMsg); err != nil {
		return nil, err
	}

	// 3) Gọi AI service
	start := time.Now()
	reply, model, ti, to, err := biz.AIClient.Chat(ctx, in.Text)
	latency := int(time.Since(start) / time.Millisecond)
	if err != nil {
		return nil, err
	}

	// 4) Lưu message assistant
	assistantMsg := &entity.Message{
		ID:             idgen.NewV7(),
		ConversationID: conv.ID,
		Role:           entity.RoleAssistant,
		Content:        reply,
		ContentType:    "markdown",
		ModelName:      &model,
		TokensInput:    ti,
		TokensOutput:   to,
		LatencyMS:      latency,
	}
	if err := biz.Repo.CreateMessage(ctx, assistantMsg); err != nil {
		return nil, err
	}

	// 5) Trả kết quả cho FE
	return &SendMessageOutput{
		ConversationID: conv.ID,
		AssistantText:  reply,
		Model:          model,
	}, nil
}
