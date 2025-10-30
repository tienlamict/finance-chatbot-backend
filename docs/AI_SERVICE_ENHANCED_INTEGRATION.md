# Enhanced AI Service Integration

## Overview

This document describes the enhanced AI service integration that sends chat requests to an external AI service with full context support including conversation history, file attachments, and configurable deep research mode.

## Architecture

### Components

1. **Enhanced AI Client** (`composer/ai_client_enhanced.go`)
   - HTTP REST client with full context support
   - Implements the `EnhancedAIClient` interface
   - Handles request/response serialization
   - Provides error handling and retry logic

2. **AI Client Interface** (`microservice/chatbot/repository/rpc/ai_client.go`)
   - Base `AIClient` interface for backward compatibility
   - `EnhancedAIClient` interface extends with context support
   - Type definitions for `HistoryItem` and `AttachFile`

3. **Business Logic** (`microservice/chatbot/business/send_message.go`)
   - Fetches conversation history
   - Generates presigned URLs for attachments
   - Calls AI service with full context

4. **Repository Layer** (`microservice/chatbot/repository/mysql/message_store.go`)
   - `GetRecentMessagesForAI()` fetches last N conversation turns
   - Excludes error messages and returns chronological order

## API Contract

### Request Format

```json
{
  "message": "Phân tích cổ phiếu VIC",
  "session_id": "conv_123",
  "user_id": "user_abc",
  "context": {
    "history": [
      {
        "role": "user",
        "message": "Give me a summary of VIC",
        "created_at": "2025-10-30T10:00:00Z"
      },
      {
        "role": "assistant",
        "message": "VIC is ...",
        "created_at": "2025-10-30T10:00:05Z"
      }
    ],
    "history_count": 8,
    "deep_research": true,
    "attach_files": [
      {
        "url": "https://minio.example/....?X-Amz-Signature=...",
        "filename": "financials_Q3.pdf",
        "mime_type": "application/pdf",
        "sha256": "abc123...",
        "pages": 12
      }
    ]
  }
}
```

### Response Format

```json
{
  "content": "Based on the provided financial statements...",
  "model": "gpt-4",
  "tokens_input": 1500,
  "tokens_output": 800,
  "request_id": "req_xyz789"
}
```

### Error Response

```json
{
  "error": "rate_limit_exceeded",
  "message": "API rate limit exceeded, please try again later",
  "request_id": "req_xyz789"
}
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `AI_SERVICE_URL` | `http://localhost:8000` | Base URL for AI service |
| `AI_PROTOCOL` | `enhanced` | Client type: `mock`, `rest`, `enhanced` |
| `AI_HISTORY_MAX` | `5` | Maximum conversation turns to include |
| `PRESIGNED_URL_EXPIRE_SEC` | `600` | Presigned URL expiry (seconds) |
| `AI_REST_TIMEOUT` | `30s` | HTTP request timeout |
| `AI_REST_API_KEY` | _(empty)_ | Optional bearer token for authentication |

### Example Configuration

```bash
# .env or docker-compose.yml
AI_SERVICE_URL=http://ai-service:8000/chat
AI_PROTOCOL=enhanced
AI_HISTORY_MAX=8
PRESIGNED_URL_EXPIRE_SEC=900
AI_REST_TIMEOUT=45s
AI_REST_API_KEY=sk-abc123...
```

## Field Semantics

### `message` (string, required)
The user's current input/prompt to the AI.

### `session_id` (string, required)
The conversation ID from the database. This ties AI replies to the same conversation thread.

### `user_id` (string, required)
The authenticated user ID (never trust client-sent user_id; always taken from JWT/session context).

### `context.history` (array)
Contains the last N pairs/turns of chat in chronological order:
- Each item has `role` (user/assistant), `message`, and `created_at`
- Excludes error messages and tool messages
- Limited to `AI_HISTORY_MAX` most recent turns

### `context.history_count` (int)
Metadata indicating the maximum history length configured on the server (not the actual count).

### `context.deep_research` (bool)
- `true`: AI can run more expensive/longer operations (e.g., web searches, multi-step reasoning)
- `false`: Standard quick response
- Controlled by client request (`deep_research` field in `SendMessageRequest`)

### `context.attach_files` (array)
Array of file references with presigned URLs:
- **Presigned URLs (recommended)**: Short-lived GET URLs that AI service can fetch directly
- Each entry contains: `url`, `filename`, `mime_type`, `sha256`, `pages` (for PDFs)
- URLs expire after `PRESIGNED_URL_EXPIRE_SEC` seconds (default: 10 minutes)
- Only includes attachments from recent user messages in the conversation

## Security & Privacy

### Presigned URLs
- **Short-lived**: Expire after 10 minutes (configurable)
- **Read-only**: GET access only, no write permissions
- **No credentials**: MinIO/S3 credentials are NEVER exposed in the payload
- **Scoped access**: URLs are tied to specific objects, can't access other files

### Best Practices
1. Never log full presigned URLs (contain signature)
2. Sanitize conversation history if it contains PII
3. Validate user authorization before sending requests
4. Use HTTPS for production AI service endpoints
5. Rotate `AI_REST_API_KEY` regularly

## Error Handling

### HTTP Status Codes

| Status | Handling |
|--------|----------|
| 200 | Success - parse response body |
| 400 | Bad Request - invalid payload format |
| 401/403 | Authentication failure - check API key |
| 429 | Rate limit exceeded - retry with backoff |
| 500 | Internal server error - log and return friendly error |
| 502/503 | Service unavailable - temporary failure |
| 504 | Gateway timeout - AI processing took too long |

### Timeout Handling
- Request timeout: 30s (default, configurable via `AI_REST_TIMEOUT`)
- If timeout occurs, save assistant message with `error_code = "ai_generate_error"`
- Return user-friendly error message

### Network Errors
- Log error details with `session_id` and `user_id`
- Return generic error to user: "AI service temporarily unavailable"
- Don't expose internal error details to clients

## Usage Examples

### Basic Request (No History)

```bash
curl -X POST http://localhost:8000/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${AI_REST_API_KEY}" \
  -d '{
    "message": "What is the capital of France?",
    "session_id": "conv_001",
    "user_id": "user_123",
    "context": {
      "history": [],
      "history_count": 5,
      "deep_research": false,
      "attach_files": []
    }
  }'
```

### Request with History

```bash
curl -X POST http://localhost:8000/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What about its population?",
    "session_id": "conv_001",
    "user_id": "user_123",
    "context": {
      "history": [
        {
          "role": "user",
          "message": "What is the capital of France?",
          "created_at": "2025-10-30T10:00:00Z"
        },
        {
          "role": "assistant",
          "message": "The capital of France is Paris.",
          "created_at": "2025-10-30T10:00:02Z"
        }
      ],
      "history_count": 5,
      "deep_research": false,
      "attach_files": []
    }
  }'
```

### Request with Attachments

```bash
curl -X POST http://localhost:8000/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Analyze this financial report",
    "session_id": "conv_002",
    "user_id": "user_123",
    "context": {
      "history": [],
      "history_count": 5,
      "deep_research": true,
      "attach_files": [
        {
          "url": "https://minio.example.com/bucket/file.pdf?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...&X-Amz-Signature=...",
          "filename": "Q3_2024_financials.pdf",
          "mime_type": "application/pdf",
          "sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
          "pages": 24
        }
      ]
    }
  }'
```

## Implementation Details

### How History is Built

1. Query database for recent messages in conversation
2. Fetch `AI_HISTORY_MAX * 2` messages (to account for user+assistant pairs)
3. Exclude messages with `error_code IS NOT NULL`
4. Order by `created_at DESC` (most recent first)
5. Reverse array to get chronological order (oldest first)
6. Convert to `HistoryItem` format with role, message, created_at

### How Attachments are Handled

1. Fetch last 10 messages in conversation
2. For each user message, get its attachments
3. Extract object name from storage key (format: `bucket/path/to/file`)
4. Generate presigned GET URL with configured expiry
5. Build `AttachFile` objects with URL, filename, mime_type, sha256, pages
6. Only include successfully generated URLs (failures are logged but don't block request)

### Fallback Behavior

If enhanced client features fail:
- **History fetch failure**: Continue with empty history array
- **Attachment URL generation failure**: Skip that attachment, continue with others
- **Client doesn't support enhanced features**: Fall back to basic `Generate()` method

## Testing

### Unit Tests

```go
// Test history limiting
func TestHistoryLimiting(t *testing.T) {
    // Create 20 messages, expect only last AI_HISTORY_MAX turns
}

// Test presigned URL generation
func TestPresignedURLs(t *testing.T) {
    // Generate URL, verify it's valid and expires correctly
}

// Test error handling
func TestAIServiceTimeout(t *testing.T) {
    // Mock timeout, verify fallback message
}
```

### Integration Tests

1. **Mock AI Service**: Use `AI_PROTOCOL=mock` for testing
2. **Verify History**: Check that correct number of turns are sent
3. **Verify Attachments**: Check that presigned URLs work
4. **Error Scenarios**: Test timeout, 429, 500 responses

### Manual Testing with Mock Service

```bash
# Set mock mode
export AI_PROTOCOL=mock

# Send request with deep_research enabled
curl -X POST http://localhost:3000/api/chatbot/send \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "conversation_id": "conv_123",
    "content": "Analyze VIC stock",
    "deep_research": true
  }'
```

## Performance Considerations

### Request Size
- Each history turn: ~500 bytes (average message length)
- 10 turns: ~5 KB
- Presigned URLs: ~300 bytes each
- Total typical request: 5-15 KB

### Response Time
- Network latency: 10-50ms (local network)
- AI processing: 1-30s (depends on complexity and deep_research flag)
- Total: typically 2-10s for standard queries

### Optimization
- **Connection pooling**: HTTP client reuses connections
- **Timeout tuning**: Adjust `AI_REST_TIMEOUT` based on AI service capabilities
- **History caching**: Consider caching recent messages to reduce DB queries
- **Async processing**: For very long AI operations, consider async/webhook pattern

## Monitoring & Logging

### Metrics to Track
- Request latency (p50, p95, p99)
- Error rate by status code
- Tokens consumed (input/output)
- History size distribution
- Attachment count per request

### Log Format

```
[AI] Request: session=conv_123 user=user_abc history=4 files=1
[AI] Success: session=conv_123 model=gpt-4 latency=2350ms tokens_in=1200 tokens_out=450
[AI] Error: session=conv_456 status=429 error=rate_limit request_id=req_xyz
```

### Alerts
- **High error rate** (>5%): Check AI service health
- **High latency** (>30s): Investigate AI service performance
- **Presigned URL failures**: Check MinIO/S3 connectivity

## Troubleshooting

### Common Issues

**Issue**: "AI service timeout"
- **Cause**: AI processing takes >30s
- **Fix**: Increase `AI_REST_TIMEOUT` or optimize AI prompts

**Issue**: "Failed to generate presigned URL"
- **Cause**: MinIO/S3 credentials invalid or bucket not accessible
- **Fix**: Verify `MINIO_*` environment variables

**Issue**: "Empty history in requests"
- **Cause**: Database query failing or no messages in conversation
- **Fix**: Check logs for `GetRecentMessagesForAI` errors

**Issue**: "AI service returns 401"
- **Cause**: Invalid or missing `AI_REST_API_KEY`
- **Fix**: Set correct API key in environment

## Migration Guide

### From Basic REST Client to Enhanced Client

1. **Update Environment**:
   ```bash
   # Change from:
   AI_PROTOCOL=rest
   
   # To:
   AI_PROTOCOL=enhanced  # or remove (enhanced is default)
   ```

2. **Add New Config**:
   ```bash
   AI_HISTORY_MAX=8
   PRESIGNED_URL_EXPIRE_SEC=600
   ```

3. **Update Request DTOs** (if sending from client):
   ```json
   {
     "conversation_id": "...",
     "content": "...",
     "deep_research": true  // NEW FIELD
   }
   ```

4. **No Code Changes Required**: The enhanced client is backward compatible with the `AIClient` interface.

### Rollback Procedure

If issues occur, quickly rollback:
```bash
# Use basic REST client
AI_PROTOCOL=rest

# Or use mock for testing
AI_PROTOCOL=mock
```

## Future Enhancements

### Planned Features
1. **Retry with exponential backoff** for transient failures
2. **Circuit breaker** pattern to prevent cascading failures
3. **Request/response caching** for identical prompts
4. **Streaming responses** for real-time AI output
5. **Multi-model support** (choose model per request)
6. **Token budget enforcement** (prevent expensive queries)

### AI Service Requirements (for AI team)
1. Support GET requests to presigned URLs for file fetching
2. Validate content-length and mime-type of fetched files
3. Return structured error responses with request_id
4. Support request timeout headers (if needed)
5. Provide metrics/monitoring endpoints

## References

- [AI Service Integration Spec](./AI_SERVICE_INTEGRATION.md)
- [File Upload Documentation](./README_FILE_UPLOAD.md)
- [MinIO Configuration](../addon/component/minioc/minio.go)
- [Conversation History API](./CONVERSATION_HISTORY_API.md)

## Contact

For questions or issues:
- Backend Team: Review this documentation
- AI Team: Coordinate on request/response formats
- DevOps: Monitor AI service health and adjust configs

