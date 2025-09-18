package entity

import "time"

type ChatMessage struct {
	ID        string    `json:"id" gorm:"column:id;primaryKey"`
	UserID    string    `json:"user_id" gorm:"column:user_id;index"`
	Role      string    `json:"role" gorm:"column:role;type:enum('user','assistant')"`
	Content   string    `json:"content" gorm:"column:content;type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (ChatMessage) TableName() string { return "chat_messages" }

type ChatRepository interface {
	Create(msg *ChatMessage) error
	ListByUser(userID string, limit int) ([]ChatMessage, error)
}

type AIClient interface {
	GenerateReply(userID string, history []ChatMessage, prompt string) (string, error)
	Close() error
}
