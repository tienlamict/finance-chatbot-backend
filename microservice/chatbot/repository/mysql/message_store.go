// finance-chatbot-backend/microservice/chatbot/repository/mysql/message_store.go
package mysql

import (
	"context"

	"finance-chatbot/microservice/chatbot/entity"
)

type MessageStore interface {
	CreateMessage(ctx context.Context, msg *entity.Message) error
	CountByConversation(ctx context.Context, convID string) (int64, error)
	ListByConversation(ctx context.Context, convID string, limit, offset int) ([]entity.Message, error)
}

func (r *MySQLRepo) CreateMessage(ctx context.Context, msg *entity.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *MySQLRepo) CountByConversation(ctx context.Context, convID string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.Message{}).
		Where("conversation_id = ?", convID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *MySQLRepo) ListByConversation(ctx context.Context, convID string, limit, offset int) ([]entity.Message, error) {
	var items []entity.Message
	q := r.db.WithContext(ctx).
		Where("conversation_id = ?", convID).
		Order("created_at ASC")
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
