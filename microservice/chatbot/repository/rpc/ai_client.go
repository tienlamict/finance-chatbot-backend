package rpc

import (
	"context"
)

// AIResult là kết quả trả về từ AI service
type AIResult struct {
	Content      string
	Model        string
	TokensInput  int
	TokensOutput int
	LatencyMs    int
}

// HistoryItem represents a single turn in the conversation
type HistoryItem struct {
	Role      string `json:"role"`       // "user" or "assistant"
	Message   string `json:"message"`    // message content
	CreatedAt string `json:"created_at"` // ISO 8601 timestamp
}

// AttachFile represents a file attachment with presigned URL
type AttachFile struct {
	URL      string `json:"url"`                // Presigned GET URL
	Filename string `json:"filename,omitempty"` // Original filename
	MimeType string `json:"mime_type,omitempty"`
	SHA256   string `json:"sha256,omitempty"`
	Pages    *int   `json:"pages,omitempty"` // For PDF files
}

// AIClient là abstraction để business không phụ thuộc trực tiếp protobuf
// Composer sẽ cung cấp implementation cụ thể (gRPC chẳng hạn).
type AIClient interface {
	Generate(ctx context.Context, userID, conversationID, prompt string) (*AIResult, error)
}

// EnhancedAIClient extends AIClient with full context support
type EnhancedAIClient interface {
	AIClient
	GenerateWithContext(
		ctx context.Context,
		userID, conversationID, prompt string,
		history []HistoryItem,
		deepResearch bool,
		attachFiles []AttachFile,
	) (*AIResult, error)
	GetHistoryMaxTurns() int
	GetPresignedExpirySec() int
}
