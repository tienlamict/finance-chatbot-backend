package mysql

import (
	"context"
	"finance-chatbot/microservice/chatbot/entity"
)

type ConversationStore interface {
	Create(ctx context.Context, conv *entity.Conversation) error
	FindByID(ctx context.Context, id string) (*entity.Conversation, error)
}

func (r *MySQLRepo) Create(ctx context.Context, conv *entity.Conversation) error {
	return r.db.WithContext(ctx).Create(conv).Error
}

func (r *MySQLRepo) FindByID(ctx context.Context, id string) (*entity.Conversation, error) {
	var conv entity.Conversation
	if err := r.db.WithContext(ctx).First(&conv, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &conv, nil
}
