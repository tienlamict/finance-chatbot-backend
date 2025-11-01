package composer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	aiclient "finance-chatbot/microservice/chatbot/repository/rpc"
)

// ====== Configuration ======
// AI_SERVICE_URL=http://localhost:8000/chat (default: http://localhost:8000)
// AI_HISTORY_MAX=8 (default: 5 - number of conversation turns to include)
// PRESIGNED_URL_EXPIRE_SEC=600 (default: 600 - 10 minutes)
// AI_REST_TIMEOUT=30s (default: 30s for AI processing)
// AI_REST_API_KEY=... (optional auth header)

// AIRequestContext contains contextual information for the AI request
type AIRequestContext struct {
	History      []aiclient.HistoryItem       `json:"history"`       // Conversation history
	HistoryCount int                          `json:"history_count"` // Max history turns configured
	DeepResearch bool                         `json:"deep_research"` // Enable expensive operations
	AttachFiles  []aiclient.AttachFile        `json:"attach_files"`  // File attachments with presigned URLs
	RAGDocuments []aiclient.RAGContextSnippet `json:"rag_documents"` // RAG retrieved documents
}

// AIRequest is the full request payload sent to the AI service
type AIRequest struct {
	Message   string           `json:"message"`    // User's current message
	SessionID string           `json:"session_id"` // Conversation ID
	UserID    string           `json:"user_id"`    // Authenticated user ID
	Context   AIRequestContext `json:"context"`    // Context with history and files
}

// AIResponse is the response from the AI service (actual format from http://localhost:8000/chat)
type AIResponse struct {
	Success       bool           `json:"success"`
	MessageID     string         `json:"message_id"`
	Response      string         `json:"response"`
	SelectedAgent *SelectedAgent `json:"selected_agent"`
	ExecutionTime float64        `json:"execution_time"` // in seconds
	Timestamp     string         `json:"timestamp"`
	SessionID     string         `json:"session_id"`
	Error         *string        `json:"error"`
	// Optional fields for token tracking (if AI service provides them)
	TokensInput  *int `json:"tokens_input,omitempty"`
	TokensOutput *int `json:"tokens_output,omitempty"`
}

// SelectedAgent represents the AI agent that processed the request
type SelectedAgent struct {
	AgentName       string   `json:"agent_name"`
	AgentType       string   `json:"agent_type"`
	ConfidenceScore float64  `json:"confidence_score"`
	Reasoning       string   `json:"reasoning"`
	ExecutionTime   float64  `json:"execution_time"`
	ToolsUsed       []string `json:"tools_used"`
}

// AIErrorResponse for error cases
type AIErrorResponse struct {
	Success bool    `json:"success"`
	Error   string  `json:"error"`
	Message string  `json:"message,omitempty"`
	Details *string `json:"details,omitempty"`
}

// ====== Enhanced AI Client Adapter ======

type enhancedAIClient struct {
	baseURL            string
	apiKey             string
	client             *http.Client
	historyMaxTurns    int
	presignedExpirySec int
}

// NewEnhancedAIClient creates a new AI client with full spec support
func NewEnhancedAIClient() aiclient.EnhancedAIClient {
	baseURL := getenv("AI_SERVICE_URL", "http://host.docker.internal:8000")
	apiKey := getenv("AI_REST_API_KEY", "")
	timeoutStr := getenv("AI_REST_TIMEOUT", "90s")

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		timeout = 30 * time.Second
	}

	historyMax, err := strconv.Atoi(getenv("AI_HISTORY_MAX", "5"))
	if err != nil || historyMax < 0 {
		historyMax = 5
	}

	presignedExpiry, err := strconv.Atoi(getenv("PRESIGNED_URL_EXPIRE_SEC", "600"))
	if err != nil || presignedExpiry < 0 {
		presignedExpiry = 600
	}

	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   20,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	return &enhancedAIClient{
		baseURL:            trimRightSlash(baseURL),
		apiKey:             apiKey,
		client:             httpClient,
		historyMaxTurns:    historyMax,
		presignedExpirySec: presignedExpiry,
	}
}

// Generate sends a request to the AI service with full context
// Note: This method signature matches the existing AIClient interface
// For now, it doesn't include history and attachments - those will be added
// through a new method or by enhancing the interface
func (c *enhancedAIClient) Generate(ctx context.Context, userID, conversationID, prompt string) (*aiclient.AIResult, error) {
	// Build basic request without history (for backward compatibility)
	req := AIRequest{
		Message:   prompt,
		SessionID: conversationID,
		UserID:    userID,
		Context: AIRequestContext{
			History:      []aiclient.HistoryItem{},
			HistoryCount: c.historyMaxTurns,
			DeepResearch: false,
			AttachFiles:  []aiclient.AttachFile{},
			RAGDocuments: []aiclient.RAGContextSnippet{},
		},
	}

	return c.sendRequest(ctx, req)
}

// GenerateWithContext sends a request with full context (history + attachments + RAG)
// This is an extended method that provides the full functionality
func (c *enhancedAIClient) GenerateWithContext(
	ctx context.Context,
	userID, conversationID, prompt string,
	history []aiclient.HistoryItem,
	deepResearch bool,
	attachFiles []aiclient.AttachFile,
	ragDocs []aiclient.RAGContextSnippet,
) (*aiclient.AIResult, error) {
	// Limit history to configured max turns
	limitedHistory := history
	if len(history) > c.historyMaxTurns {
		limitedHistory = history[len(history)-c.historyMaxTurns:]
	}

	req := AIRequest{
		Message:   prompt,
		SessionID: conversationID,
		UserID:    userID,
		Context: AIRequestContext{
			History:      limitedHistory,
			HistoryCount: c.historyMaxTurns,
			DeepResearch: deepResearch,
			AttachFiles:  attachFiles,
			RAGDocuments: ragDocs,
		},
	}
	return c.sendRequest(ctx, req)
}

// sendRequest performs the actual HTTP request to the AI service
func (c *enhancedAIClient) sendRequest(ctx context.Context, req AIRequest) (*aiclient.AIResult, error) {
	// Marshal request body
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal AI request: %w", err)
	}

	// Log the actual JSON being sent (for debugging)
	fmt.Printf("AI Request JSON: %s\n", string(reqBody))

	// Log request metadata (avoid logging full presigned URLs)
	logAIRequest(req.SessionID, req.UserID, len(req.Context.History), len(req.Context.AttachFiles))

	// Build HTTP request
	url := c.baseURL + "/chat"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Send request and measure latency
	start := time.Now()
	resp, err := c.client.Do(httpReq)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		// Check if error is due to timeout
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("AI service timeout after %dms: %w", latency, err)
		}
		return nil, fmt.Errorf("AI service network error: %w", err)
	}
	defer resp.Body.Close()

	// Parse response (AI service may return 200 with success:false)
	var aiResp AIResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return nil, fmt.Errorf("failed to decode AI response: %w", err)
	}

	// Check if AI reported an error (success: false)
	if !aiResp.Success || aiResp.Error != nil {
		errMsg := "unknown error"
		if aiResp.Error != nil {
			errMsg = *aiResp.Error
		}

		// Log error details
		logAIError(req.SessionID, resp.StatusCode, errMsg, aiResp.MessageID)

		// Handle non-200 responses differently
		if resp.StatusCode != http.StatusOK {
			switch resp.StatusCode {
			case http.StatusBadRequest:
				return nil, fmt.Errorf("invalid request to AI service: %s", errMsg)
			case http.StatusUnauthorized, http.StatusForbidden:
				return nil, errors.New("AI service authentication failed")
			case http.StatusTooManyRequests:
				return nil, errors.New("AI service rate limit exceeded, please try again later")
			case http.StatusServiceUnavailable:
				return nil, errors.New("AI service is temporarily unavailable")
			case http.StatusGatewayTimeout:
				return nil, errors.New("AI service processing timeout")
			default:
				return nil, fmt.Errorf("AI service error: %s", errMsg)
			}
		}

		// Success:false with 200 status
		return nil, fmt.Errorf("AI service error: %s", errMsg)
	}

	// Validate response content
	if aiResp.Response == "" {
		return nil, errors.New("AI service returned empty response")
	}

	// Extract model name from selected agent or use default
	modelName := "unknown"
	if aiResp.SelectedAgent != nil && aiResp.SelectedAgent.AgentName != "" {
		modelName = aiResp.SelectedAgent.AgentName
	}

	// Extract tokens if provided
	tokensIn := 0
	tokensOut := 0
	if aiResp.TokensInput != nil {
		tokensIn = *aiResp.TokensInput
	}
	if aiResp.TokensOutput != nil {
		tokensOut = *aiResp.TokensOutput
	}

	// Log successful response
	logAISuccess(req.SessionID, modelName, latency, tokensIn, tokensOut)

	return &aiclient.AIResult{
		Content:      aiResp.Response, // Map "response" to "content"
		Model:        modelName,       // Use agent name as model
		TokensInput:  tokensIn,
		TokensOutput: tokensOut,
		LatencyMs:    latency,
	}, nil
}

// GetHistoryMaxTurns returns the configured maximum history turns
func (c *enhancedAIClient) GetHistoryMaxTurns() int {
	return c.historyMaxTurns
}

// GetPresignedExpirySec returns the configured presigned URL expiry in seconds
func (c *enhancedAIClient) GetPresignedExpirySec() int {
	return c.presignedExpirySec
}

// ====== Logging helpers (minimal, production should use structured logging) ======

func logAIRequest(sessionID, userID string, historyLen, filesLen int) {
	fmt.Printf("[AI] Request: session=%s user=%s history=%d files=%d\n",
		sessionID, userID, historyLen, filesLen)
}

func logAIError(sessionID string, statusCode int, errMsg, requestID string) {
	fmt.Printf("[AI] Error: session=%s status=%d error=%s request_id=%s\n",
		sessionID, statusCode, errMsg, requestID)
}

func logAISuccess(sessionID, model string, latency, tokensIn, tokensOut int) {
	fmt.Printf("[AI] Success: session=%s model=%s latency=%dms tokens_in=%d tokens_out=%d\n",
		sessionID, model, latency, tokensIn, tokensOut)
}
