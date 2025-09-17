package business

import (
	"context"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/chatbot/entity"
)

type ChatRepository interface {
	SendNewMessage(ctx context.Context, data *entity.ChatDataCreation) error
	//GetMessageByUserId(ctx context.Context, id string) (*entity.ChatMessage, error)
	//ListMessages(ctx context.Context, userID string, limit int) ([]entity.ChatMessage, error)
}

type UserRepository interface {
	GetUsersByIds(ctx context.Context, ids []int) ([]core.SimpleUser, error)
	GetUserById(ctx context.Context, id int) (*core.SimpleUser, error)
}

type business struct {
	chatRepo ChatRepository
	userRepo UserRepository
}

func NewBusiness(chatRepo ChatRepository, userRepo UserRepository) *business {
	return &business{
		chatRepo: chatRepo,
		userRepo: userRepo,
	}
}
