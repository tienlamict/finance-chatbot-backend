// finance-chatbot-backend/microservice/chatbot/business/send_message.go
package business

import (
	"context"
	"time"

	idgen "finance-chatbot/addon/common"
	"finance-chatbot/microservice/chatbot/entity"
)

func (uc *chatUsecase) ensureConversation(ctx context.Context, req entity.SendMessageRequest, userID string) (string, error) {
	if req.ConversationID != "" {
		// Đảm bảo conv tồn tại
		if _, err := uc.sql.FindByID(ctx, req.ConversationID); err == nil {
			return req.ConversationID, nil
		}
		// Nếu không tìm thấy thì fallback tạo mới
	}
	conv := &entity.Conversation{
		ID:        idgen.NewV7(),
		UserID:    userID,
		OrgID:     req.OrgID,
		Title:     req.Title,
		Status:    entity.ConversationActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.sql.Create(ctx, conv); err != nil {
		return "", err
	}
	return conv.ID, nil
}

func (uc *chatUsecase) SendMessage(ctx context.Context, req entity.SendMessageRequest, userID string) (*entity.SendMessageResponse, error) {
	convID, err := uc.ensureConversation(ctx, req, userID)
	if err != nil {
		return nil, err
	}

	// 1) Lưu message của user
	now := time.Now()
	content := req.Content
	userMsg := &entity.Message{
		ID:             idgen.NewV7(),
		ConversationID: convID,
		Role:           entity.RoleUser,
		Content:        &content,
		ContentType:    "text",
		CreatedAt:      now,
	}
	if err := uc.sql.CreateMessage(ctx, userMsg); err != nil {
		return nil, err
	}

	// 2) Gọi AI service
	start := time.Now()
	aiRes, err := uc.ai.Generate(ctx, userID, convID, req.Content)
	latency := int(time.Since(start).Milliseconds())
	if err != nil {
		// Lưu 1 message assistant báo lỗi (optional)
		errCode := "ai_generate_error"
		msg := &entity.Message{
			ID:             idgen.NewV7(),
			ConversationID: convID,
			Role:           entity.RoleAssistant,
			Content:        nil,
			ContentType:    "text",
			ErrorCode:      &errCode,
			LatencyMs:      latency,
			CreatedAt:      time.Now(),
		}
		_ = uc.sql.CreateMessage(ctx, msg)
		return nil, err
	}

	// 3) Lưu message của assistant
	assistantContent := aiRes.Content
	model := aiRes.Model
	assistantMsg := &entity.Message{
		ID:             idgen.NewV7(),
		ConversationID: convID,
		Role:           entity.RoleAssistant,
		Content:        &assistantContent,
		ContentType:    "text",
		TokensInput:    aiRes.TokensInput,
		TokensOutput:   aiRes.TokensOutput,
		LatencyMs:      latency,
		ModelName:      &model,
		CreatedAt:      time.Now(),
	}
	if err := uc.sql.CreateMessage(ctx, assistantMsg); err != nil {
		return nil, err
	}

	return &entity.SendMessageResponse{
		ConversationID:       convID,
		UserMessageID:        userMsg.ID,
		AssistantMessageID:   assistantMsg.ID,
		AssistantContent:     assistantContent,
		ModelName:            model,
		TokensInput:          aiRes.TokensInput,
		TokensOutput:         aiRes.TokensOutput,
		LatencyMs:            latency,
		UserMessageCreatedAt: userMsg.CreatedAt,
		AIMessageCreatedAt:   assistantMsg.CreatedAt,
	}, nil
}
