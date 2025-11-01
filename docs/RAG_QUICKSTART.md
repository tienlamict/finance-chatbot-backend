# RAG Integration Quickstart

## Prerequisites

1. Running RAG service at `http://localhost:8082` (or configure `RAG_BASE_URL`)
2. **MinIO/S3 configured for file storage** (required for upload operations)
3. MySQL database with RAG tables (migration in `migrations/001_create_rag_tables.sql`)

> **Note**: MinIO storage is mandatory for RAG uploads. If not configured, upload requests will fail with "MinIO storage not configured" error.

## Setup

### 1. Run Database Migration

```sql
-- Apply migration
mysql -u root -p finance_chatbot < migrations/001_create_rag_tables.sql
```

### 2. Configure Environment

Add to your `.env` or docker-compose:

```bash
# Required
RAG_BASE_URL=http://localhost:8082

# Optional (defaults shown)
RAG_UPLOAD_TIMEOUT_SEC=60
RAG_QUERY_TIMEOUT_SEC=15
RAG_UPLOAD_RETRIES=2
RAG_COLLECTION_DEFAULT=rag_collection
RAG_TEMP_DIR=/tmp/rag_uploads

# Enable RAG context in AI
RAG_ENABLED=true
RAG_QUERY_K=5
```

### 3. Start Backend

```bash
go run main.go
```

## Quick Test

### Test 1: Upload Document to RAG

```bash
# First, upload a file normally (saves to MinIO)
curl -X POST http://localhost:3000/v1/chatbot/prompt \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "message=Hello" \
  -F "file=@./test.pdf"

# Note the storage_key from response
# Then upload to RAG
curl -X POST http://localhost:3000/v1/rag/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "doc_id": "test-doc-1",
    "collection": "rag_collection",
    "minio_key": "bucket/path/from/previous/upload",
    "filename": "test.pdf",
    "mime_type": "application/pdf",
    "sha256": "abc123"
  }'
```

### Test 2: Query RAG

```bash
curl "http://localhost:3000/v1/rag/query?query=machine%20learning&k=5&collection=rag_collection"
```

### Test 3: List Your Documents

```bash
curl http://localhost:3000/v1/rag/documents \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Test 4: AI with RAG Context

```bash
# Enable RAG
export RAG_ENABLED=true

# Send message (RAG snippets automatically included)
curl -X POST http://localhost:3000/v1/chatbot/prompt \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "What does the document say about machine learning?",
    "deep_research": false
  }'

# Check backend logs to see RAG snippets in AI request
```

## Verify Integration

### Check Logs

Look for these log entries:

```
[RAG Business] Downloading from MinIO: key=...
[RAG Business] Created temp file: /tmp/rag_uploads/...
[RAG] Uploading doc_id=..., file=..., size=... bytes
[RAG] Upload success: stored_path=./temp_uploads/...
[RAG Business] Document uploaded successfully
```

### Check Database

```sql
SELECT * FROM rag_documents;
SELECT * FROM rag_query_logs ORDER BY created_at DESC LIMIT 10;
```

### Check AI Request

Enable debug logging and look for:

```json
{
  "context": {
    "rag_documents": [
      {
        "source": "...",
        "title": "...",
        "page": 5,
        "score": 0.95,
        "text": "..."
      }
    ]
  }
}
```

## Common Issues

### Issue: Upload returns 502

**Solution:** Check RAG service is running:
```bash
curl http://localhost:8082/api/rag/query?query=test&k=1&collection=test
```

### Issue: MinIO storage not configured

**Solution:** Ensure MinIO component is properly configured in your service context:
```bash
# Check environment variables
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY_ID=minioadmin
MINIO_SECRET_ACCESS_KEY=minioadmin
MINIO_USE_SSL=false
MINIO_BUCKET_NAME=finance-chatbot
```

### Issue: Temp file permission denied

**Solution:** Ensure directory exists and is writable:
```bash
mkdir -p /tmp/rag_uploads
chmod 700 /tmp/rag_uploads
```

### Issue: Document not found

**Solution:** Check document was uploaded successfully:
```bash
SELECT * FROM rag_documents WHERE doc_id = 'your-doc-id';
```

### Issue: RAG context not in AI request

**Solution:** Verify environment:
```bash
echo $RAG_ENABLED  # Should be "true"
```

## Next Steps

1. Integrate RAG upload in your file upload workflow
2. Add frontend UI for document management
3. Configure RAG collections for different document types
4. Set up monitoring for RAG query performance
5. Implement RAG caching for frequent queries

## Example Integration

```go
// Upload file and immediately add to RAG
func (h *Handler) UploadFileToRAG(c *gin.Context) {
    // 1. Upload to MinIO (existing flow)
    storageKey, err := h.storage.UploadFile(...)
    
    // 2. Upload to RAG
    ragReq := &entity.CreateRAGDocumentRequest{
        DocID:      generateDocID(),
        Collection: "user_documents",
        MinioKey:   storageKey,
        Filename:   filename,
        MimeType:   mimeType,
        SHA256:     sha256Hash,
    }
    
    doc, err := h.ragBusiness.UploadDocumentForRAG(ctx, ragReq, userID)
    
    // 3. Return both storage key and RAG document ID
    return gin.H{
        "storage_key": storageKey,
        "rag_doc_id":  doc.DocID,
    }
}
```

## Monitoring

Key metrics to track:

- Upload success rate
- Upload latency (p50, p95, p99)
- Query latency
- Query results count distribution
- Top queried documents
- Failed uploads by error type

See `rag_query_logs` table for analytics data.

