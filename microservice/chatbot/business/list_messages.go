// finance-chatbot-backend/microservice/chatbot/business/list_messages.go
package business

import (
	"context"

	"finance-chatbot/microservice/chatbot/entity"
)

func (uc *chatUsecase) ListMessages(ctx context.Context, req entity.ListMessagesRequest) (*entity.ListMessagesResponse, error) {
	total, err := uc.sql.CountByConversation(ctx, req.ConversationID)
	if err != nil {
		return nil, err
	}
	items, err := uc.sql.ListByConversation(ctx, req.ConversationID, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}
	return &entity.ListMessagesResponse{
		ConversationID: req.ConversationID,
		Total:          total,
		Items:          items,
	}, nil
}
