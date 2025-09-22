package gateways

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type AIClient interface {
	Chat(ctx context.Context, prompt string) (reply string, model string, tokensIn, tokensOut int, err error)
}

type HTTPAIClient struct {
	Endpoint string
	Client   *http.Client
}

func NewHTTPAIClient(endpoint string) *HTTPAIClient {
	return &HTTPAIClient{
		Endpoint: endpoint,
		Client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *HTTPAIClient) Chat(ctx context.Context, prompt string) (string, string, int, int, error) {
	// Stub: POST {"input": "..."} -> {"output":"...", "model":"...", "ti":..., "to":...}
	payload := map[string]string{"input": prompt}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", "", 0, 0, err
	}
	defer resp.Body.Close()

	var out struct {
		Output string `json:"output"`
		Model  string `json:"model"`
		TI     int    `json:"ti"`
		TO     int    `json:"to"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", "", 0, 0, err
	}
	return out.Output, out.Model, out.TI, out.TO, nil
}
