package mysql

import (
	"context"
	idgen "finance-chatbot/addon/common"
	"finance-chatbot/microservice/chatbot/entity"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type ChatRepoMySQL struct{ db *gorm.DB }

func NewChatRepoMySQL(db *gorm.DB) *ChatRepoMySQL { return &ChatRepoMySQL{db: db} }

func (r *ChatRepoMySQL) GetOrCreateConversation(ctx entity.Context, userID, conversationIDOrEmpty string, titleIfNew *string) (*entity.Conversation, error) {
	if conversationIDOrEmpty != "" {
		var c entity.Conversation
		if err := r.db.WithContext(ctx.(context.Context)).
			Where("id = ? AND user_id = ?", conversationIDOrEmpty, userID).
			First(&c).Error; err != nil {
			return nil, err
		}
		return &c, nil
	}
	c := &entity.Conversation{
		ID:     idgen.NewV7(), // UUID v7
		UserID: userID,
		Title:  titleIfNew,
	}
	if err := r.db.WithContext(ctx.(context.Context)).Create(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ChatRepoMySQL) CreateMessage(ctx entity.Context, msg *entity.Message) error {
	if msg.ID == "" || msg.ConversationID == "" {
		return errors.New("missing id or conversation_id")
	}
	return r.db.WithContext(ctx.(context.Context)).Create(msg).Error
}
