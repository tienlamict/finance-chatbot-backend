package business

import (
	"context"
	"fmt"
	"os"
	"strconv"

	aiclient "finance-chatbot/microservice/chatbot/repository/rpc"
)

// RAGQueryService defines interface for querying RAG documents
// This allows us to optionally inject RAG without hard dependency
type RAGQueryService interface {
	QueryForContext(ctx context.Context, query string, k int, collection string) ([]aiclient.RAGContextSnippet, error)
}

// buildRAGContext queries RAG service and returns snippets for AI context
func (uc *chatUsecase) buildRAGContext(ctx context.Context, prompt string) ([]aiclient.RAGContextSnippet, error) {
	// Check if RAG is enabled via environment variable
	if os.Getenv("RAG_ENABLED") != "true" {
		return []aiclient.RAGContextSnippet{}, nil
	}

	// Get RAG query parameters from environment
	ragK := getEnvIntOrDefault("RAG_QUERY_K", 5)
	ragCollection := getEnvOrDefault("RAG_COLLECTION_DEFAULT", "rag_collection")

	// Note: In a full implementation, you would inject RAGQueryService
	// For now, we return empty to avoid hard dependency
	// TODO: Wire up RAG query service in composer
	fmt.Printf("[RAG] Context building disabled - would query: k=%d, collection=%s\n", ragK, ragCollection)

	return []aiclient.RAGContextSnippet{}, nil
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
