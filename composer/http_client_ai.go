package composer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	aiclient "finance-chatbot/microservice/chatbot/repository/rpc"
)

// ====== Wire helper (chọn REST theo ENV) ======
//
// AI_PROTOCOL=rest|grpc  (mặc định: grpc để tương thích cũ)
// AI_REST_BASE_URL=http://ai:8080
// AI_REST_API_KEY=... (optional; nếu AI service yêu cầu auth bằng header)
// AI_REST_TIMEOUT=5s (optional; ví dụ "3s", "10s")

func composeAIRESTClient() aiclient.AIClient {
	baseURL := getenv("AI_REST_BASE_URL", "http://ai:8080")
	apiKey := getenv("AI_REST_API_KEY", "")
	timeoutStr := getenv("AI_REST_TIMEOUT", "5s")
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		timeout = 5 * time.Second
	}

	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return NewAIRESTClientAdapter(baseURL, apiKey, httpClient)
}

// ====== REST adapter implement AIClient ======

type aiRESTClientAdapter struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewAIRESTClientAdapter tạo adapter REST implement AIClient
func NewAIRESTClientAdapter(baseURL, apiKey string, httpClient *http.Client) aiclient.AIClient {
	return &aiRESTClientAdapter{
		baseURL: trimRightSlash(baseURL),
		apiKey:  apiKey,
		client:  httpClient,
	}
}

// Generate gọi REST: POST {baseURL}/v1/generate
// Request JSON: { "user_id", "conversation_id", "prompt" }
// Response JSON: { "content", "model", "tokens_input", "tokens_output" }
func (a *aiRESTClientAdapter) Generate(ctx context.Context, userID, conversationID, prompt string) (*aiclient.AIResult, error) {
	reqBody := struct {
		UserID         string `json:"user_id"`
		ConversationID string `json:"conversation_id"`
		Prompt         string `json:"prompt"`
	}{
		UserID:         userID,
		ConversationID: conversationID,
		Prompt:         prompt,
	}

	b, _ := json.Marshal(reqBody)
	url := a.baseURL + "/v1/generate"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if a.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
	}

	start := time.Now()
	resp, err := a.client.Do(httpReq)
	latency := int(time.Since(start).Milliseconds())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		if apiErr.Error == "" {
			apiErr.Error = resp.Status
		}
		return nil, errors.New("ai_rest_error: " + apiErr.Error)
	}

	var out struct {
		Content      string `json:"content"`
		Model        string `json:"model"`
		TokensInput  int    `json:"tokens_input"`
		TokensOutput int    `json:"tokens_output"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode ai response: %w", err)
	}

	return &aiclient.AIResult{
		Content:      out.Content,
		Model:        out.Model,
		TokensInput:  out.TokensInput,
		TokensOutput: out.TokensOutput,
		LatencyMs:    latency,
	}, nil
}

// ====== small helpers ======
func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func trimRightSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}
