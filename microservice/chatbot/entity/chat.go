// finance-chatbot-backend/microservice/chatbot/entity/chat.go
package entity

import (
	"time"

	"gorm.io/datatypes"
)

type ConversationStatus string

const (
	ConversationActive   ConversationStatus = "active"
	ConversationArchived ConversationStatus = "archived"
	ConversationDeleted  ConversationStatus = "deleted"
)

type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

type Conversation struct {
	ID        string             `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	UserID    string             `gorm:"column:user_id;type:varchar(50);index:idx_conv_user_created" json:"user_id"`
	OrgID     *string            `gorm:"column:org_id;type:varchar(50);index:idx_conv_org_created" json:"org_id,omitempty"`
	Title     *string            `gorm:"column:title;type:varchar(255)" json:"title,omitempty"`
	Status    ConversationStatus `gorm:"column:status;type:enum('active','archived','deleted');default:'active'" json:"status"`
	CreatedAt time.Time          `gorm:"column:created_at;type:datetime(3);autoCreateTime" json:"created_at"`
	UpdatedAt time.Time          `gorm:"column:updated_at;type:datetime(3);autoUpdateTime" json:"updated_at"`
}

func (Conversation) TableName() string { return "conversations" }

type Message struct {
	ID             string         `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	ConversationID string         `gorm:"column:conversation_id;type:char(36);index:idx_msg_conv_created" json:"conversation_id"`
	ParentID       *string        `gorm:"column:parent_id;type:char(36)" json:"parent_id,omitempty"`
	Role           MessageRole    `gorm:"column:role;type:enum('user','assistant','tool');index:idx_msg_role_created" json:"role"`
	Content        *string        `gorm:"column:content;type:mediumtext" json:"content,omitempty"`
	ContentType    string         `gorm:"column:content_type;type:enum('text','markdown','json');default:'text'" json:"content_type"`
	Meta           datatypes.JSON `gorm:"column:meta;type:json" json:"meta,omitempty"`
	TokensInput    int            `gorm:"column:tokens_input" json:"tokens_input"`
	TokensOutput   int            `gorm:"column:tokens_output" json:"tokens_output"`
	LatencyMs      int            `gorm:"column:latency_ms" json:"latency_ms"`
	ModelName      *string        `gorm:"column:model_name;type:varchar(128)" json:"model_name,omitempty"`
	ErrorCode      *string        `gorm:"column:error_code;type:varchar(64)" json:"error_code,omitempty"`
	CreatedAt      time.Time      `gorm:"column:created_at;type:datetime(3);autoCreateTime" json:"created_at"`
}

func (Message) TableName() string { return "messages" }

// ==== DTOs dùng cho transport ====

type SendMessageRequest struct {
	ConversationID string `json:"conversation_id"` // nếu rỗng -> tạo mới
	Content        string `json:"content" binding:"required"`
	// UserID is ignored in favor of the authenticated user ID from the JWT token.
	// This field is kept for backward compatibility but will be overridden by the auth context.
	UserID string `json:"user_id"`
	// Optional: OrgID, Title lần đầu tạo conv
	OrgID *string `json:"org_id,omitempty"`
	Title *string `json:"title,omitempty"`
	// Files uploaded with the message (populated in transport layer)
	Files []FileUploadInfo `json:"-"`
}

type SendMessageResponse struct {
	ConversationID       string          `json:"conversation_id"`
	UserMessageID        string          `json:"user_message_id"`
	AssistantMessageID   string          `json:"assistant_message_id"`
	AssistantContent     string          `json:"assistant_content"`
	ModelName            string          `json:"model_name,omitempty"`
	TokensInput          int             `json:"tokens_input"`
	TokensOutput         int             `json:"tokens_output"`
	LatencyMs            int             `json:"latency_ms"`
	UserMessageCreatedAt time.Time       `json:"user_message_created_at"`
	AIMessageCreatedAt   time.Time       `json:"ai_message_created_at"`
	Attachments          []AttachmentDTO `json:"attachments,omitempty"`
}

type ListMessagesRequest struct {
	ConversationID string `form:"conversation_id" json:"conversation_id" binding:"required"`
	Limit          int    `form:"limit,default=50" json:"limit"`
	Offset         int    `form:"offset,default=0" json:"offset"`
}

type ListMessagesResponse struct {
	ConversationID string    `json:"conversation_id"`
	Total          int64     `json:"total"`
	Items          []Message `json:"items"`
}

// Enhanced conversation history request with advanced filtering
type ConversationHistoryRequest struct {
	ConversationID     string     `form:"conversation_id" json:"conversation_id" binding:"required"`
	Limit              int        `form:"limit,default=20" json:"limit"`
	Before             *string    `form:"before" json:"before,omitempty"` // message ID or timestamp
	After              *string    `form:"after" json:"after,omitempty"`   // message ID or timestamp
	From               *time.Time `form:"from" json:"from,omitempty"`     // ISO 8601 timestamp
	To                 *time.Time `form:"to" json:"to,omitempty"`         // ISO 8601 timestamp
	Order              string     `form:"order,default=asc" json:"order"` // asc | desc
	Search             *string    `form:"search" json:"search,omitempty"` // keyword search
	IncludeAttachments bool       `form:"include_attachments,default=true" json:"include_attachments"`
}

// Enhanced conversation history response with pagination info
type ConversationHistoryResponse struct {
	ConversationID string                   `json:"conversation_id"`
	Total          int64                    `json:"total"`
	Items          []MessageWithAttachments `json:"items"`
	HasMore        bool                     `json:"has_more"`
	NextCursor     *string                  `json:"next_cursor,omitempty"`
	PreviousCursor *string                  `json:"previous_cursor,omitempty"`
}

// Message with attachments for response
type MessageWithAttachments struct {
	Message
	Attachments []AttachmentDTO `json:"attachments,omitempty"`
}

// User conversations list request
type UserConversationsRequest struct {
	UserID             string     `form:"user_id" json:"user_id" binding:"required"`
	Limit              int        `form:"limit,default=20" json:"limit"`
	Offset             int        `form:"offset,default=0" json:"offset"`
	Status             *string    `form:"status" json:"status,omitempty"` // active, archived, deleted
	From               *time.Time `form:"from" json:"from,omitempty"`
	To                 *time.Time `form:"to" json:"to,omitempty"`
	Search             *string    `form:"search" json:"search,omitempty"`
	IncludeLastMessage bool       `form:"include_last_message,default=true" json:"include_last_message"`
}

// User conversations list response
type UserConversationsResponse struct {
	UserID     string                `json:"user_id"`
	Total      int64                 `json:"total"`
	Items      []ConversationSummary `json:"items"`
	HasMore    bool                  `json:"has_more"`
	NextOffset int                   `json:"next_offset"`
}

// Conversation summary with last message preview
type ConversationSummary struct {
	Conversation
	LastMessage   *MessagePreview `json:"last_message,omitempty"`
	UnreadCount   int             `json:"unread_count"`
	TotalMessages int64           `json:"total_messages"`
}

// Message preview for conversation list
type MessagePreview struct {
	ID        string      `json:"id"`
	Role      MessageRole `json:"role"`
	Content   *string     `json:"content,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
}
