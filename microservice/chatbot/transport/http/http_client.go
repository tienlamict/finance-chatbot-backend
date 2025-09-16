package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"time"

	"finance-chatbot/microservice/chatbot/entity"
)

type AIHTTPClient struct {
	baseURL string
	apiKey  string // nếu cần auth
	hc      *http.Client
}

func NewAIHTTPClient(baseURL, apiKey string) *AIHTTPClient {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   3 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &AIHTTPClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		hc: &http.Client{
			Timeout:   10 * time.Second, // request deadline
			Transport: tr,
		},
	}
}

func (c *AIHTTPClient) Close() error { return nil }

type chatItem struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}
type genReq struct {
	UserID  string     `json:"user_id"`
	Prompt  string     `json:"prompt"`
	History []chatItem `json:"history"`
}
type genResp struct {
	Reply string `json:"reply"`
}

func (c *AIHTTPClient) GenerateReply(userID string, history []entity.ChatMessage, prompt string) (string, error) {
	items := make([]chatItem, 0, len(history))
	// oldest-first (nếu AI cần ngữ cảnh theo thời gian)
	for i := len(history) - 1; i >= 0; i-- {
		h := history[i]
		items = append(items, chatItem{
			Role:      string(h.Role),
			Content:   h.Content,
			Timestamp: h.CreatedAt.Unix(),
		})
	}

	payload := genReq{UserID: userID, Prompt: prompt, History: items}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, c.baseURL+"/v1/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("ai service error: " + resp.Status)
	}
	var out genResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Reply, nil
}
