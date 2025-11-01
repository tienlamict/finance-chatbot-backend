# RAG + AI Integration Example

This document shows how to fully integrate RAG queries into the AI request pipeline for automatic context enrichment.

## Architecture Flow

```
User Query → RAG Query → Retrieve Snippets → Include in AI Context → AI Response
```

## Complete Integration

### Option 1: Environment Variable Based (Current)

The system currently supports basic RAG integration via environment variables:

```bash
export RAG_ENABLED=true
export RAG_QUERY_K=5
export RAG_COLLECTION_DEFAULT=rag_collection
```

The `buildRAGContext()` method in `send_message.go` is called automatically when `RAG_ENABLED=true`.

### Option 2: Full Dependency Injection (Advanced)

For production deployments, inject RAG business into chatbot business:

#### Step 1: Update Chatbot Business Interface

```go
// microservice/chatbot/business/business.go
type chatUsecase struct {
    sql     interface{ ... }
    ai      rpcrepo.AIClient
    storage StorageProvider
    rag     RAGService  // <-- Add RAG service
}

type RAGService interface {
    QueryForContext(ctx context.Context, query string, k int, collection string) ([]aiclient.RAGContextSnippet, error)
}
```

#### Step 2: Create RAG Adapter

```go
// microservice/chatbot/business/rag_adapter.go
package business

import (
    "context"
    aiclient "finance-chatbot/microservice/chatbot/repository/rpc"
    ragBiz "finance-chatbot/microservice/rag/business"
    ragEntity "finance-chatbot/microservice/rag/entity"
)

type ragAdapter struct {
    business ragBiz.RAGBusiness
}

func NewRAGAdapter(biz ragBiz.RAGBusiness) RAGService {
    return &ragAdapter{business: biz}
}

func (a *ragAdapter) QueryForContext(
    ctx context.Context,
    query string,
    k int,
    collection string,
) ([]aiclient.RAGContextSnippet, error) {
    req := &ragEntity.QueryRAGRequest{
        Query:      query,
        K:          k,
        Collection: collection,
    }
    
    results, err := a.business.QueryRAG(ctx, req, "", "")
    if err != nil {
        return nil, err
    }
    
    // Convert to AI context snippets
    snippets := make([]aiclient.RAGContextSnippet, 0)
    for _, result := range results {
        for _, snippet := range result.Snippets {
            snippets = append(snippets, aiclient.RAGContextSnippet{
                Source:     result.Source,
                Title:      result.Title,
                Page:       snippet.Page,
                Score:      snippet.Score,
                Text:       snippet.Text,
                DocumentID: result.DocumentID,
            })
        }
    }
    
    return snippets, nil
}
```

#### Step 3: Update Composer

```go
// composer/service_composer.go
func ComposeChatbotAPIService(serviceCtx sctx.ServiceContext) ChatbotService {
    dbComp := serviceCtx.MustGet(common.KeyCompMySQL).(common.GormComponent)
    gormDB := dbComp.GetDB()
    
    aiClient := chooseAIClient(serviceCtx)
    
    var storage common.StorageProvider
    if storageComp, ok := serviceCtx.Get(common.KeyCompStorage); ok {
        storage = storageComp.(common.StorageProvider)
    }
    
    // Create RAG service if enabled
    var ragService chatBusiness.RAGService
    if os.Getenv("RAG_ENABLED") == "true" {
        ragAPIService := ComposeRAGAPIService(serviceCtx)
        // Note: Need to extract business from API service
        // Better: Create separate ComposeRAGBusiness() method
        ragBiz := ComposeRAGBusiness(serviceCtx)
        ragService = chatBusiness.NewRAGAdapter(ragBiz)
    }
    
    sqlRepo := chatmysql.NewMySQLRepo(gormDB)
    biz := chatBusiness.NewChatBusinessWithRAG(sqlRepo, aiClient, storage, ragService)
    api := chatAPI.NewAPI(biz)
    return api
}

// Helper to compose just RAG business
func ComposeRAGBusiness(serviceCtx sctx.ServiceContext) ragBusiness.RAGBusiness {
    dbComp := serviceCtx.MustGet(common.KeyCompMySQL).(common.GormComponent)
    gormDB := dbComp.GetDB()
    
    var storage common.StorageProvider
    if storageComp, ok := serviceCtx.Get(common.KeyCompStorage); ok {
        storage = storageComp.(common.StorageProvider)
    }
    
    ragConfig := ragrpc.Config{
        BaseURL:          getEnvOrDefault("RAG_BASE_URL", "http://localhost:8082"),
        UploadTimeoutSec: getEnvIntOrDefault("RAG_UPLOAD_TIMEOUT_SEC", 60),
        QueryTimeoutSec:  getEnvIntOrDefault("RAG_QUERY_TIMEOUT_SEC", 15),
        UploadRetries:    getEnvIntOrDefault("RAG_UPLOAD_RETRIES", 2),
        QueryRetries:     getEnvIntOrDefault("RAG_QUERY_RETRIES", 1),
    }
    
    ragClient := ragrpc.NewRAGClient(ragConfig)
    ragRepo := ragmysql.NewRAGRepository(gormDB)
    tempDir := getEnvOrDefault("RAG_TEMP_DIR", "")
    
    return ragBusiness.NewRAGBusiness(ragRepo, ragClient, storage, tempDir)
}
```

#### Step 4: Update buildRAGContext

```go
// microservice/chatbot/business/rag_integration.go
func (uc *chatUsecase) buildRAGContext(ctx context.Context, prompt string) ([]aiclient.RAGContextSnippet, error) {
    // If no RAG service injected, return empty
    if uc.rag == nil {
        return []aiclient.RAGContextSnippet{}, nil
    }
    
    // Get config
    ragK := getEnvIntOrDefault("RAG_QUERY_K", 5)
    ragCollection := getEnvOrDefault("RAG_COLLECTION_DEFAULT", "rag_collection")
    
    // Query RAG
    snippets, err := uc.rag.QueryForContext(ctx, prompt, ragK, ragCollection)
    if err != nil {
        // Log but don't fail the entire request
        fmt.Printf("[RAG] Query failed: %v\n", err)
        return []aiclient.RAGContextSnippet{}, nil
    }
    
    fmt.Printf("[RAG] Retrieved %d snippets for prompt\n", len(snippets))
    return snippets, nil
}
```

## Testing the Integration

### 1. Upload Test Document

```bash
# Upload a document with known content
curl -X POST http://localhost:3000/v1/rag/upload \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "doc_id": "ml-guide-001",
    "collection": "rag_collection",
    "minio_key": "attachments/user123/msg456/ml_guide.pdf",
    "filename": "machine_learning_guide.pdf",
    "mime_type": "application/pdf",
    "sha256": "abc123..."
  }'
```

### 2. Verify RAG Query Works

```bash
# Test direct query
curl "http://localhost:3000/v1/rag/query?query=what%20is%20machine%20learning&k=3&collection=rag_collection"

# Should return:
{
  "data": {
    "results": [
      {
        "document_id": "ml-guide-001",
        "snippets": [
          {
            "page": 5,
            "score": 0.95,
            "text": "Machine learning is a subset of AI..."
          }
        ]
      }
    ]
  }
}
```

### 3. Test AI with RAG Context

```bash
# Enable RAG
export RAG_ENABLED=true

# Send message that should trigger RAG
curl -X POST http://localhost:3000/v1/chatbot/prompt \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Explain machine learning based on my documents",
    "deep_research": false
  }'
```

### 4. Verify RAG in AI Request

Check backend logs for:

```
[RAG] Retrieved 3 snippets for prompt
[AI] Request JSON: {
  "message": "Explain machine learning based on my documents",
  "context": {
    "rag_documents": [
      {
        "source": "./temp_uploads/ml_guide.pdf",
        "title": "Machine Learning Guide",
        "page": 5,
        "score": 0.95,
        "text": "Machine learning is a subset of AI...",
        "document_id": "ml-guide-001"
      }
    ]
  }
}
```

## Advanced Scenarios

### Per-User RAG Collections

```go
func (uc *chatUsecase) buildRAGContext(ctx context.Context, prompt string, userID string) ([]aiclient.RAGContextSnippet, error) {
    // Use user-specific collection
    collection := fmt.Sprintf("user_%s_documents", userID)
    return uc.rag.QueryForContext(ctx, prompt, 5, collection)
}
```

### Dynamic K Based on Prompt Length

```go
func (uc *chatUsecase) buildRAGContext(ctx context.Context, prompt string) ([]aiclient.RAGContextSnippet, error) {
    // Longer prompts might benefit from more context
    k := 3
    if len(prompt) > 200 {
        k = 7
    }
    
    return uc.rag.QueryForContext(ctx, prompt, k, "rag_collection")
}
```

### Conversation-Scoped RAG

```go
func (uc *chatUsecase) buildRAGContext(ctx context.Context, prompt string, conversationID string) ([]aiclient.RAGContextSnippet, error) {
    // Query only documents uploaded in this conversation
    collection := fmt.Sprintf("conversation_%s", conversationID)
    return uc.rag.QueryForContext(ctx, prompt, 5, collection)
}
```

### Conditional RAG

```go
func (uc *chatUsecase) buildRAGContext(ctx context.Context, prompt string) ([]aiclient.RAGContextSnippet, error) {
    // Only use RAG for specific types of queries
    if !shouldUseRAG(prompt) {
        return []aiclient.RAGContextSnippet{}, nil
    }
    
    return uc.rag.QueryForContext(ctx, prompt, 5, "rag_collection")
}

func shouldUseRAG(prompt string) bool {
    // Keywords that indicate document-based questions
    keywords := []string{"document", "file", "according to", "based on", "reference"}
    promptLower := strings.ToLower(prompt)
    
    for _, keyword := range keywords {
        if strings.Contains(promptLower, keyword) {
            return true
        }
    }
    return false
}
```

## Performance Optimization

### 1. Caching Frequent Queries

```go
var ragCache = make(map[string][]aiclient.RAGContextSnippet)
var ragCacheMutex sync.RWMutex

func (uc *chatUsecase) buildRAGContextCached(ctx context.Context, prompt string) ([]aiclient.RAGContextSnippet, error) {
    cacheKey := fmt.Sprintf("%s:%d", prompt, 5)
    
    // Check cache
    ragCacheMutex.RLock()
    if cached, ok := ragCache[cacheKey]; ok {
        ragCacheMutex.RUnlock()
        return cached, nil
    }
    ragCacheMutex.RUnlock()
    
    // Query RAG
    snippets, err := uc.rag.QueryForContext(ctx, prompt, 5, "rag_collection")
    if err != nil {
        return nil, err
    }
    
    // Cache result
    ragCacheMutex.Lock()
    ragCache[cacheKey] = snippets
    ragCacheMutex.Unlock()
    
    return snippets, nil
}
```

### 2. Parallel RAG + History Fetching

```go
func (uc *chatUsecase) callAIWithContext(ctx context.Context, userID, conversationID, prompt string, deepResearch bool) (*aiclient.AIResult, error) {
    enhancedClient := uc.ai.(aiclient.EnhancedAIClient)
    
    // Fetch history, attachments, and RAG in parallel
    var history []aiclient.HistoryItem
    var attachFiles []aiclient.AttachFile
    var ragDocs []aiclient.RAGContextSnippet
    
    var wg sync.WaitGroup
    wg.Add(3)
    
    go func() {
        defer wg.Done()
        history, _ = uc.buildConversationHistory(ctx, conversationID, enhancedClient.GetHistoryMaxTurns())
    }()
    
    go func() {
        defer wg.Done()
        attachFiles, _ = uc.buildAttachmentReferences(ctx, conversationID, enhancedClient.GetPresignedExpirySec())
    }()
    
    go func() {
        defer wg.Done()
        ragDocs, _ = uc.buildRAGContext(ctx, prompt)
    }()
    
    wg.Wait()
    
    return enhancedClient.GenerateWithContext(ctx, userID, conversationID, prompt, history, deepResearch, attachFiles, ragDocs)
}
```

## Monitoring

Track these metrics:

```go
// RAG query latency
ragQueryDuration := time.Since(start)
metrics.Histogram("rag.query.duration", ragQueryDuration)

// RAG results count
metrics.Gauge("rag.results.count", len(snippets))

// RAG cache hit rate
if cached {
    metrics.Increment("rag.cache.hit")
} else {
    metrics.Increment("rag.cache.miss")
}
```

## Summary

✅ **Current State**: Basic RAG integration with environment variable control  
🚀 **Future**: Full dependency injection with advanced features

The current implementation provides a solid foundation. For production, consider implementing:
- Dependency injection for cleaner testing
- Per-user/conversation collections
- Query caching
- Advanced filtering and ranking
- RAG analytics dashboard

