package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"finance-chatbot/microservice/rag/entity"
)

// RAGClient defines the interface for RAG service communication
type RAGClient interface {
	// UploadDocument uploads a file to the RAG service
	UploadDocument(ctx context.Context, docID, filePath, collection string) (storedPath string, err error)

	// QueryDocuments queries the RAG service for relevant documents
	QueryDocuments(ctx context.Context, query string, k int, collection string) ([]entity.RAGQueryDoc, error)
}

type ragClient struct {
	baseURL       string
	client        *http.Client
	uploadTimeout time.Duration
	queryTimeout  time.Duration
	uploadRetries int
	queryRetries  int
}

// Config holds RAG client configuration
type Config struct {
	BaseURL          string
	UploadTimeoutSec int
	QueryTimeoutSec  int
	UploadRetries    int
	QueryRetries     int
}

// NewRAGClient creates a new RAG HTTP client
func NewRAGClient(cfg Config) RAGClient {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://host.docker.internal:8082"
	}
	if cfg.UploadTimeoutSec <= 0 {
		cfg.UploadTimeoutSec = entity.DefaultUploadTimeoutSec
	}
	if cfg.QueryTimeoutSec <= 0 {
		cfg.QueryTimeoutSec = entity.DefaultQueryTimeoutSec
	}
	if cfg.UploadRetries < 0 {
		cfg.UploadRetries = entity.DefaultUploadRetries
	}
	if cfg.QueryRetries < 0 {
		cfg.QueryRetries = entity.DefaultQueryRetries
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          50,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	return &ragClient{
		baseURL:       trimRightSlash(cfg.BaseURL),
		client:        httpClient,
		uploadTimeout: time.Duration(cfg.UploadTimeoutSec) * time.Second,
		queryTimeout:  time.Duration(cfg.QueryTimeoutSec) * time.Second,
		uploadRetries: cfg.UploadRetries,
		queryRetries:  cfg.QueryRetries,
	}
}

// UploadDocument uploads a file to RAG service with retry logic
func (c *ragClient) UploadDocument(ctx context.Context, docID, filePath, collection string) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= c.uploadRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s...
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			fmt.Printf("[RAG] Upload retry %d/%d after %v\n", attempt, c.uploadRetries, backoff)
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
			}
		}

		storedPath, err := c.uploadDocumentOnce(ctx, docID, filePath, collection)
		if err == nil {
			return storedPath, nil
		}

		lastErr = err

		// Don't retry on client errors (4xx)
		if isClientError(err) {
			return "", err
		}

		fmt.Printf("[RAG] Upload attempt %d failed: %v\n", attempt+1, err)
	}

	return "", fmt.Errorf("upload failed after %d retries: %w", c.uploadRetries+1, lastErr)
}

// uploadDocumentOnce performs a single upload attempt
func (c *ragClient) uploadDocumentOnce(ctx context.Context, docID, filePath, collection string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add doc_id field
	if err := writer.WriteField("doc_id", docID); err != nil {
		return "", fmt.Errorf("failed to write doc_id field: %w", err)
	}

	// Add collection field
	if err := writer.WriteField("collection", collection); err != nil {
		return "", fmt.Errorf("failed to write collection field: %w", err)
	}

	// Add file field
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create HTTP request with timeout
	uploadCtx, cancel := context.WithTimeout(ctx, c.uploadTimeout)
	defer cancel()

	url := c.baseURL + "/api/rag/upload-for-rag"
	req, err := http.NewRequestWithContext(uploadCtx, http.MethodPost, url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Log request details
	fmt.Printf("[RAG] Uploading doc_id=%s, file=%s, size=%d bytes, collection=%s\n",
		docID, filepath.Base(filePath), fileInfo.Size(), collection)

	start := time.Now()
	resp, err := c.client.Do(req)
	latency := time.Since(start)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return "", entity.ErrRAGTimeout
		}
		return "", fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[RAG] Upload failed: status=%d, body=%s, latency=%v\n",
			resp.StatusCode, string(respBody), latency)

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return "", fmt.Errorf("client error %d: %s", resp.StatusCode, string(respBody))
		}
		return "", fmt.Errorf("server error %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response (should be a string path)
	storedPath := string(bytes.TrimSpace(bytes.Trim(respBody, "\"")))

	if storedPath == "" {
		return "", entity.ErrRAGInvalidResponse
	}

	fmt.Printf("[RAG] Upload success: doc_id=%s, stored_path=%s, latency=%v\n",
		docID, storedPath, latency)

	return storedPath, nil
}

// QueryDocuments queries the RAG service with retry logic
func (c *ragClient) QueryDocuments(ctx context.Context, query string, k int, collection string) ([]entity.RAGQueryDoc, error) {
	var lastErr error

	for attempt := 0; attempt <= c.queryRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(500<<uint(attempt-1)) * time.Millisecond
			fmt.Printf("[RAG] Query retry %d/%d after %v\n", attempt, c.queryRetries, backoff)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		docs, err := c.queryDocumentsOnce(ctx, query, k, collection)
		if err == nil {
			return docs, nil
		}

		lastErr = err

		// Don't retry on client errors
		if isClientError(err) {
			return nil, err
		}

		fmt.Printf("[RAG] Query attempt %d failed: %v\n", attempt+1, err)
	}

	return nil, fmt.Errorf("query failed after %d retries: %w", c.queryRetries+1, lastErr)
}

// queryDocumentsOnce performs a single query attempt
func (c *ragClient) queryDocumentsOnce(ctx context.Context, query string, k int, collection string) ([]entity.RAGQueryDoc, error) {
	// Create request with timeout
	queryCtx, cancel := context.WithTimeout(ctx, c.queryTimeout)
	defer cancel()

	// Properly URL encode query parameters
	baseURL, err := url.Parse(c.baseURL + "/api/rag/query")
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	params := url.Values{}
	params.Set("query", query)
	params.Set("k", fmt.Sprintf("%d", k))
	params.Set("collection", collection)
	baseURL.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(queryCtx, http.MethodGet, baseURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	fmt.Printf("[RAG] Querying: query=%s, k=%d, collection=%s\n", query, k, collection)
	fmt.Printf("[RAG] Full URL: %s\n", baseURL.String())

	start := time.Now()
	resp, err := c.client.Do(req)
	latency := time.Since(start)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, entity.ErrRAGTimeout
		}
		return nil, fmt.Errorf("query request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[RAG] Query failed: status=%d, body=%s, latency=%v\n",
			resp.StatusCode, string(respBody), latency)

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return nil, fmt.Errorf("client error %d: %s", resp.StatusCode, string(respBody))
		}
		return nil, fmt.Errorf("server error %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse JSON response
	var docs []entity.RAGQueryDoc
	if err := json.Unmarshal(respBody, &docs); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Printf("[RAG] Query success: found %d documents, latency=%v\n", len(docs), latency)

	return docs, nil
}

// Helper functions

func trimRightSlash(s string) string {
	if len(s) > 0 && s[len(s)-1] == '/' {
		return s[:len(s)-1]
	}
	return s
}

func isClientError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "client error") || contains(msg, "400") || contains(msg, "404")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (hasPrefix(s, substr) || hasSuffix(s, substr) || hasInfix(s, substr))))
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func hasInfix(s, infix string) bool {
	for i := 0; i <= len(s)-len(infix); i++ {
		if s[i:i+len(infix)] == infix {
			return true
		}
	}
	return false
}

// GetEnv helper
func getenv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt helper
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
