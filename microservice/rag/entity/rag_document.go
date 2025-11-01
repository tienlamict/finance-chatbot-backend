package entity

import (
	"finance-chatbot/addon/core"
	"time"
)

// RAGDocument represents a document uploaded to the RAG system
type RAGDocument struct {
	core.SQLModel `json:",inline"`
	DocID         string    `json:"doc_id" gorm:"column:doc_id;type:varchar(100);uniqueIndex;not null"`
	Collection    string    `json:"collection" gorm:"column:collection;type:varchar(100);index;not null"`
	StoredPath    string    `json:"stored_path" gorm:"column:stored_path;type:varchar(500);not null"` // Path returned by RAG API
	Filename      string    `json:"filename" gorm:"column:filename;type:varchar(255);not null"`
	MimeType      string    `json:"mime_type" gorm:"column:mime_type;type:varchar(100)"`
	SHA256        string    `json:"sha256" gorm:"column:sha256;type:varchar(64);index"`
	TotalPages    *int      `json:"total_pages" gorm:"column:total_pages"`
	Title         string    `json:"title" gorm:"column:title;type:varchar(500)"`
	Author        string    `json:"author" gorm:"column:author;type:varchar(255)"`
	UploadedBy    string    `json:"uploaded_by" gorm:"column:uploaded_by;type:varchar(50);index;not null"` // User UID
	MinioKey      string    `json:"minio_key" gorm:"column:minio_key;type:varchar(500)"`                   // Optional source
	UploadedAt    time.Time `json:"uploaded_at" gorm:"column:uploaded_at;autoCreateTime"`
}

func (RAGDocument) TableName() string {
	return "rag_documents"
}

// RAGQueryLog represents a query made to the RAG system for analytics
type RAGQueryLog struct {
	core.SQLModel  `json:",inline"`
	QueryText      string    `json:"query_text" gorm:"column:query_text;type:text;not null"`
	UserID         string    `json:"user_id" gorm:"column:user_id;type:varchar(50);index"`
	ConversationID string    `json:"conversation_id" gorm:"column:conversation_id;type:varchar(100);index"`
	Collection     string    `json:"collection" gorm:"column:collection;type:varchar(100);index"`
	K              int       `json:"k" gorm:"column:k;not null"`
	ReturnedDocs   string    `json:"returned_docs" gorm:"column:returned_docs;type:json"` // JSON array of doc_ids
	LatencyMs      int       `json:"latency_ms" gorm:"column:latency_ms"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (RAGQueryLog) TableName() string {
	return "rag_query_logs"
}

// CreateRAGDocumentRequest represents the request to upload a document from MinIO
type CreateRAGDocumentRequest struct {
	DocID      string `json:"doc_id" binding:"required"`
	Collection string `json:"collection" binding:"required"`
	MinioKey   string `json:"minio_key" binding:"required"` // The file storage key
	Filename   string `json:"filename" binding:"required"`
	MimeType   string `json:"mime_type"`
	SHA256     string `json:"sha256"`
}

// DirectUploadRAGDocumentRequest represents a direct file upload to RAG (multipart)
type DirectUploadRAGDocumentRequest struct {
	DocID      string `form:"doc_id" binding:"required"`
	Collection string `form:"collection" binding:"required"`
	File       []byte `form:"-"` // File content will be read from multipart
}

// QueryRAGRequest represents a query request
type QueryRAGRequest struct {
	Query      string `json:"query" form:"query" binding:"required"`
	K          int    `json:"k" form:"k" binding:"required,min=1,max=20"`
	Collection string `json:"collection" form:"collection" binding:"required"`
}
