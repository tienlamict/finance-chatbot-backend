package entity

import (
	"time"

	"gorm.io/datatypes"
)

// ====== GORM entity mapping bảng conversations, messages ======

type Conversation struct {
	ID        string    `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	UserID    string    `gorm:"column:user_id;type:varchar(50);not null" json:"user_id"`
	OrgID     *string   `gorm:"column:org_id;type:varchar(50)" json:"org_id,omitempty"`
	Title     *string   `gorm:"column:title;type:varchar(255)" json:"title,omitempty"`
	Status    string    `gorm:"column:status;type:enum('active','archived','deleted');default:'active'" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime(3)" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime(3)" json:"updated_at"`
}

func (Conversation) TableName() string { return "conversations" }

type Message struct {
	ID             string         `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	ConversationID string         `gorm:"column:conversation_id;type:char(36);not null" json:"conversation_id"`
	ParentID       *string        `gorm:"column:parent_id;type:char(36)" json:"parent_id,omitempty"`
	Role           string         `gorm:"column:role;type:enum('user','assistant','tool');not null" json:"role"`
	Content        *string        `gorm:"column:content;type:mediumtext" json:"content,omitempty"`
	ContentType    string         `gorm:"column:content_type;type:enum('text','markdown','json');default:'text'" json:"content_type"`
	Meta           datatypes.JSON `gorm:"column:meta" json:"meta,omitempty"`
	TokensInput    int            `gorm:"column:tokens_input" json:"tokens_input"`
	TokensOutput   int            `gorm:"column:tokens_output" json:"tokens_output"`
	LatencyMs      int            `gorm:"column:latency_ms" json:"latency_ms"`
	ModelName      *string        `gorm:"column:model_name;type:varchar(128)" json:"model_name,omitempty"`
	ErrorCode      *string        `gorm:"column:error_code;type:varchar(64)" json:"error_code,omitempty"`
	CreatedAt      time.Time      `gorm:"column:created_at;type:datetime(3)" json:"created_at"`
}

func (Message) TableName() string { return "messages" }

// ====== DTO ======

type ChatRequest struct {
	UserID         string `json:"user_id,omitempty"` // optional: nếu không có thì lấy từ requester context
	ConversationID string `json:"conversation_id,omitempty"`
	Content        string `json:"content" binding:"required"`
}

type ChatResponse struct {
	ConversationID string `json:"conversation_id"`
	Answer         string `json:"answer"`
	MessageID      string `json:"message_id"`
}

type ConversationItem struct {
	ID        string    `json:"id"`
	Title     *string   `json:"title,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MessageItem struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   *string   `json:"content,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	ModelName *string   `json:"model_name,omitempty"`
}
