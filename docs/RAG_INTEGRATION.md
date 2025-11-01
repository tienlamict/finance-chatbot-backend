# RAG Integration Documentation

## Overview

This document describes the complete RAG (Retrieval-Augmented Generation) integration in the finance-chatbot backend. The system allows uploading documents to an external RAG service, querying for relevant document snippets, and automatically including those snippets in AI context.

## Architecture

### Components

1. **Entity Layer** (`microservice/rag/entity/`)
   - Document models and query result structures
   - Error definitions
   - Request/response DTOs

2. **Repository Layer** (`microservice/rag/repository/`)
   - **MySQL** (`mysql/`): Persistence for document metadata and query logs
   - **RPC** (`rpc/`): HTTP client for external RAG API

3. **Business Layer** (`microservice/rag/business/`)
   - Upload orchestration (MinIO → temp → RAG upload → DB persistence)
   - Query orchestration (RAG query → mapping → logging)
   - Secure temp file management

4. **Transport Layer** (`microservice/rag/transport/api/`)
   - REST API handlers for upload, query, list, get, delete operations

## Database Schema

### `rag_documents` Table

```sql
CREATE TABLE `rag_documents` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `doc_id` varchar(100) NOT NULL UNIQUE,
  `collection` varchar(100) NOT NULL,
  `stored_path` varchar(500) NOT NULL,
  `filename` varchar(255) NOT NULL,
  `mime_type` varchar(100),
  `sha256` varchar(64),
  `total_pages` int(11),
  `title` varchar(500),
  `author` varchar(255),
  `uploaded_by` varchar(50) NOT NULL,
  `minio_key` varchar(500),
  `uploaded_at` datetime NOT NULL,
  `status` int(11) NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`),
  KEY `idx_collection` (`collection`),
  KEY `idx_uploaded_by` (`uploaded_by`)
);
```

### `rag_query_logs` Table

```sql
CREATE TABLE `rag_query_logs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `query_text` text NOT NULL,
  `user_id` varchar(50),
  `conversation_id` varchar(100),
  `collection` varchar(100),
  `k` int(11) NOT NULL,
  `returned_docs` json,
  `latency_ms` int(11),
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_conversation_id` (`conversation_id`)
);
```

## External RAG API

### 1. Upload Document

**Endpoint:** `POST http://localhost:8082/api/rag/upload-for-rag`

**Request:** `multipart/form-data`
- `doc_id`: Document identifier
- `file`: File upload stream
- `collection`: Collection name

**Response:** `"./temp_uploads/1111.pdf"` (stored path string)

### 2. Query Documents

**Endpoint:** `GET http://localhost:8082/api/rag/query?query=<q>&k=<int>&collection=<collection>`

**Response:**
```json
[
  {
    "source": "./temp_uploads/1111.pdf",
    "metadata": {
      "title": "Document Title",
      "total_pages": 35,
      "author": "Author Name"
    },
    "matches": [
      {
        "page_content": "Relevant text snippet",
        "score": 0.7689635,
        "page": 23
      }
    ]
  }
]
```

## Backend API Endpoints

### Upload Document to RAG

**POST** `/v1/rag/upload`

**Authentication:** Required

**Request Body:**
```json
{
  "doc_id": "doc-123",
  "collection": "rag_collection",
  "minio_key": "bucket/path/to/file.pdf",
  "filename": "document.pdf",
  "mime_type": "application/pdf",
  "sha256": "abc123..."
}
```

**Response:**
```json
{
  "data": {
    "doc_id": "doc-123",
    "collection": "rag_collection",
    "stored_path": "./temp_uploads/1111.pdf",
    "filename": "document.pdf",
    "uploaded_by": "user123",
    "uploaded_at": "2025-11-01T12:00:00Z"
  }
}
```

### Query RAG Documents

**GET** `/v1/rag/query?query=machine%20learning&k=5&collection=rag_collection`

**Authentication:** Optional (logged if authenticated)

**Response:**
```json
{
  "data": {
    "query": "machine learning",
    "k": 5,
    "collection": "rag_collection",
    "results": [
      {
        "document_id": "doc-123",
        "source": "./temp_uploads/1111.pdf",
        "title": "ML Introduction",
        "total_pages": 100,
        "snippets": [
          {
            "page": 5,
            "score": 0.95,
            "text": "Machine learning is..."
          }
        ]
      }
    ],
    "count": 1
  }
}
```

### Get Document Info

**GET** `/v1/rag/documents/:doc_id`

**Authentication:** Required

### List User Documents

**GET** `/v1/rag/documents?limit=20&offset=0`

**Authentication:** Required

### Delete Document

**DELETE** `/v1/rag/documents/:doc_id`

**Authentication:** Required (owner only)

## Upload Flow

1. User uploads file → backend saves to MinIO
2. Backend calls `POST /v1/rag/upload` with `minio_key`
3. Business layer:
   - Downloads file from MinIO to secure temp directory (`/tmp/rag_uploads/`)
   - Uploads temp file to RAG API via multipart form
   - Receives `stored_path` from RAG API
   - Persists document metadata to `rag_documents` table
   - Deletes temp file
4. Returns document metadata to client

## Query Flow

1. Client/AI service calls `GET /v1/rag/query`
2. Business layer:
   - Queries external RAG API with parameters
   - Maps RAG response to internal result format
   - Looks up document IDs from DB by `stored_path`
   - Logs query to `rag_query_logs` (async)
3. Returns structured results with document metadata + snippets

## AI Context Integration

When sending messages to the AI service, the chatbot automatically:

1. Queries RAG for relevant snippets based on the user's prompt (if `RAG_ENABLED=true`)
2. Includes RAG results in the AI request context:

```json
{
  "message": "Tell me about machine learning",
  "session_id": "conv-123",
  "user_id": "user-456",
  "context": {
    "history": [...],
    "attach_files": [...],
    "rag_documents": [
      {
        "source": "./temp_uploads/1111.pdf",
        "title": "ML Guide",
        "page": 5,
        "score": 0.95,
        "text": "Machine learning is...",
        "document_id": "doc-123"
      }
    ]
  }
}
```

3. AI service uses RAG snippets as grounding context for generation

## Configuration

Environment variables:

```bash
# RAG Service
RAG_BASE_URL=http://localhost:8082
RAG_UPLOAD_TIMEOUT_SEC=60
RAG_QUERY_TIMEOUT_SEC=15
RAG_UPLOAD_RETRIES=2
RAG_QUERY_RETRIES=1
RAG_COLLECTION_DEFAULT=rag_collection
RAG_TEMP_DIR=/tmp/rag_uploads

# Enable RAG context in AI requests
RAG_ENABLED=true
RAG_QUERY_K=5
```

## Error Handling

### Upload Errors

- **Network/5xx errors**: Retries with exponential backoff (max 2 retries)
- **4xx errors**: No retry, immediate failure
- **Timeout**: Returns 504 Gateway Timeout
- **MinIO download failure**: Returns 502 Bad Gateway

### Query Errors

- **Network/5xx errors**: Retries once with 500ms backoff
- **4xx errors**: No retry
- **Empty results**: Returns empty array (not an error)

### Logging

All operations log:
- Request ID, user ID, conversation ID
- Latencies and status codes
- **Never** logs presigned URLs or raw file contents

## Security

1. **Temp Files**: Created with `0600` permissions (owner read/write only)
2. **Temp Directory**: Created with `0700` permissions
3. **Cleanup**: Temp files deleted immediately after upload
4. **Access Control**: Upload/delete require authentication and ownership check
5. **No Credential Leakage**: MinIO credentials never sent to external RAG API

## Testing

Run unit tests:
```bash
cd microservice/rag/business
go test -v
```

Integration tests (requires mock RAG server):
```bash
go test -v ./test/
```

## Future Enhancements

1. **Streaming Upload**: Avoid temp files by streaming from MinIO directly to RAG
2. **Bulk Operations**: Upload/query multiple documents in one request
3. **RAG Caching**: Cache frequent queries with Redis
4. **Document Versioning**: Track document updates
5. **Semantic Search UI**: Frontend interface for document search
6. **RAG Admin Dashboard**: Analytics on query patterns and document usage

## Troubleshooting

### Upload returns 422

- Check JSON keys in request match AI service contract (lowercase/snake_case)
- Verify multipart form fields: `doc_id`, `file`, `collection`

### Query returns empty results

- Verify document was successfully uploaded to RAG service
- Check collection name matches
- Test RAG API directly with curl

### Temp files not cleaned up

- Check `RAG_TEMP_DIR` permissions
- Review logs for deletion errors
- Manually clean: `rm -rf /tmp/rag_uploads/*`

### Integration not working

- Verify `RAG_ENABLED=true`
- Check AI service logs for context payload
- Test RAG query endpoint independently

## API Reference Summary

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/v1/rag/upload` | Required | Upload document to RAG |
| GET | `/v1/rag/query` | Optional | Query RAG documents |
| GET | `/v1/rag/documents/:doc_id` | Required | Get document info |
| GET | `/v1/rag/documents` | Required | List user documents |
| DELETE | `/v1/rag/documents/:doc_id` | Required | Delete document |

