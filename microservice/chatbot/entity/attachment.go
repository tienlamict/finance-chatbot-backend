// finance-chatbot-backend/microservice/chatbot/entity/attachment.go
package entity

import (
	"mime/multipart"
	"time"

	"gorm.io/datatypes"
)

type MessageAttachment struct {
	ID         string         `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	MessageID  string         `gorm:"column:message_id;type:char(36);index:idx_att_msg" json:"message_id"`
	Filename   string         `gorm:"column:filename;type:varchar(255)" json:"filename"`
	StorageKey string         `gorm:"column:storage_key;type:varchar(512)" json:"storage_key"`
	MimeType   *string        `gorm:"column:mime_type;type:varchar(255)" json:"mime_type,omitempty"`
	SHA256     *string        `gorm:"column:sha256;type:char(64);index:idx_att_sha" json:"sha256,omitempty"`
	Pages      *int           `gorm:"column:pages" json:"pages,omitempty"`
	Meta       datatypes.JSON `gorm:"column:meta;type:json" json:"meta,omitempty"`
	CreatedAt  time.Time      `gorm:"column:created_at;type:datetime(3);autoCreateTime" json:"created_at"`
}

func (MessageAttachment) TableName() string { return "message_attachments" }

// FileUploadInfo holds information about uploaded file from the request
type FileUploadInfo struct {
	FileHeader *multipart.FileHeader
	Content    []byte // read from FileHeader
	MimeType   string
	SHA256Hash string
	Pages      *int // for PDF files
}

// AttachmentDTO for response
type AttachmentDTO struct {
	ID         string    `json:"id"`
	Filename   string    `json:"filename"`
	MimeType   string    `json:"mime_type,omitempty"`
	Size       int64     `json:"size,omitempty"`
	SHA256     string    `json:"sha256,omitempty"`
	Pages      *int      `json:"pages,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	StorageKey string    `json:"storage_key,omitempty"` // optional, can hide from client
}
