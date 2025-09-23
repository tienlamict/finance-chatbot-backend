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

// AIClient là abstraction để business không phụ thuộc trực tiếp protobuf
// Composer sẽ cung cấp implementation cụ thể (gRPC chẳng hạn).
type AIClient interface {
	Generate(ctx context.Context, userID, conversationID, prompt string) (*AIResult, error)
}
