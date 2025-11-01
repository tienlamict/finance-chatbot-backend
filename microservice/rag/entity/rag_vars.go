package entity

const (
	// Default collection name
	DefaultCollection = "rag_collection"

	// Max retries for RAG operations
	DefaultUploadRetries = 2
	DefaultQueryRetries  = 1

	// Timeouts
	DefaultUploadTimeoutSec = 60
	DefaultQueryTimeoutSec  = 15

	// Limits
	MaxQueryK         = 20
	DefaultQueryK     = 5
	MaxFileSizeForRAG = 50 * 1024 * 1024 // 50MB
)
