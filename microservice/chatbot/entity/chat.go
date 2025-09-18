package entity

import (
	"time"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Conversation struct {
	ID        string    `gorm:"column:id;primaryKey;type:char(36)"`
	UserID    string    `gorm:"column:user_id;type:char(36);not null"`
	Title     *string   `gorm:"column:title"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (Conversation) TableName() string { return "conversations" }

type Message struct {
	ID             string    `gorm:"column:id;primaryKey;type:char(36)"`
	ConversationID string    `gorm:"column:conversation_id;type:char(36);not null"`
	ParentID       *string   `gorm:"column:parent_id;type:char(36)"`
	Role           Role      `gorm:"column:role;type:enum('user','assistant','tool');not null"`
	Content        string    `gorm:"column:content;type:mediumtext"`
	ContentType    string    `gorm:"column:content_type;type:enum('text','markdown','json');default:'text'"`
	ModelName      *string   `gorm:"column:model_name;type:varchar(128)"`
	TokensInput    int       `gorm:"column:tokens_input"`
	TokensOutput   int       `gorm:"column:tokens_output"`
	LatencyMS      int       `gorm:"column:latency_ms"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (Message) TableName() string { return "messages" }

type ChatRepo interface {
	GetOrCreateConversation(ctx Context, userID, conversationIDOrEmpty string, titleIfNew *string) (*Conversation, error)
	CreateMessage(ctx Context, msg *Message) error
}

type Context interface{ Done() <-chan struct{} } // alias cho context.Context (tránh import vòng)
