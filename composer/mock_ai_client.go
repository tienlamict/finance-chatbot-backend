package composer

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	aiclient "finance-chatbot/microservice/chatbot/repository/rpc"
)

type aiMockClient struct {
	model   string
	latency time.Duration
}

func NewAIMockClient() aiclient.AIClient {
	model := getenv("AI_MOCK_MODEL", "mock-chat-model-group-9")
	latStr := getenv("AI_MOCK_LATENCY_MS", "200")
	ms, _ := strconv.Atoi(latStr)
	if ms < 0 {
		ms = 0
	}
	return &aiMockClient{
		model:   model,
		latency: time.Duration(ms) * time.Millisecond,
	}
}

func (m *aiMockClient) Generate(ctx context.Context, userID, conversationID, prompt string) (*aiclient.AIResult, error) {
	// Cho phép giả lập lỗi nếu prompt chứa "error"
	if strings.Contains(strings.ToLower(prompt), "error") {
		return nil, errors.New("mock: simulated AI error")
	}

	// Giả lập trễ
	if m.latency > 0 {
		select {
		case <-time.After(m.latency):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Sinh đáp án giả (echo + chút "trí tuệ" :))
	resp := "(mock) Tôi đã nhận: " + prompt
	tin := len(prompt) / 4
	if tin < 1 {
		tin = 1
	}
	tout := len(resp) / 4
	if tout < 1 {
		tout = 1
	}

	return &aiclient.AIResult{
		Content:      resp,
		Model:        m.model,
		TokensInput:  tin,
		TokensOutput: tout,
		LatencyMs:    int(m.latency.Milliseconds()),
	}, nil
}
