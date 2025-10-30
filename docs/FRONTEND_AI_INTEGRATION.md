# Frontend `/v1/chatbot/prompt` API Integration with AI Service

## Overview

This document describes how the frontend `/v1/chatbot/prompt` API is integrated with the external AI service at `http://localhost:8000/chat`. The integration handles authentication, conversation management, AI context building, and response mapping.

## Endpoint

```
POST /v1/chatbot/prompt?conversation_id={optional}
```

## Request Format

### Headers
```
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

### Query Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `conversation_id` | string | No | Existing conversation ID. If omitted, creates new conversation |

### Request Body
```json
{
  "content": "User message here",
  "deep_research": false  // optional, default: false
}
```

### Example Request

```bash
curl --location 'http://localhost:3001/v1/chatbot/prompt?conversation_id=019a1694-14e0-727a-b76a-b13d81dcd220' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <JWT>' \
--data '{
  "content": "Hello, how are you",
  "deep_research": false
}'
```

## Response Format

### Success Response (200 OK)

```json
{
  "conversation_id": "019a1694-14e0-727a-b76a-b13d81dcd220",
  "user_message_id": "019a055f-47f3-713a-abb2-4de02db0ff19",
  "assistant_message_id": "fe19a683-7f27-49eb-905c-595dd4089a5a",
  "assistant_content": "Chào bạn! Rất vui được hỗ trợ bạn hôm nay.",
  "model_name": "FinanceBot",
  "tokens_input": 0,
  "tokens_output": 0,
  "latency_ms": 2761,
  "user_message_created_at": "2025-10-30T18:14:58.000Z",
  "ai_message_created_at": "2025-10-30T18:14:58.887803Z"
}
```

### Error Responses

#### 401 Unauthorized
```json
{
  "error": "user ID is required"
}
```

#### 403 Forbidden
```json
{
  "error": "conversation not found or access denied"
}
```

#### 500 Internal Server Error
```json
{
  "error": "AI service error: timeout"
}
```

#### 502 Bad Gateway
```json
{
  "error": "AI service unavailable"
}
```

## How It Works

### Flow Diagram

```
┌─────────┐       ┌──────────┐       ┌────────────┐       ┌─────────┐
│ Frontend│       │ Backend  │       │  Database  │       │AI Service│
└────┬────┘       └────┬─────┘       └─────┬──────┘       └────┬────┘
     │                 │                    │                   │
     │ POST /prompt    │                    │                   │
     │ + JWT + body    │                    │                   │
     ├────────────────>│                    │                   │
     │                 │                    │                   │
     │                 │ 1. Extract user_id │                   │
     │                 │    from JWT        │                   │
     │                 │                    │                   │
     │                 │ 2. Validate conv   │                   │
     │                 │    access or create│                   │
     │                 ├───────────────────>│                   │
     │                 │<───────────────────┤                   │
     │                 │                    │                   │
     │                 │ 3. Save user msg   │                   │
     │                 ├───────────────────>│                   │
     │                 │                    │                   │
     │                 │ 4. Fetch history   │                   │
     │                 ├───────────────────>│                   │
     │                 │<───────────────────┤                   │
     │                 │                    │                   │
     │                 │ 5. Generate        │                   │
     │                 │    presigned URLs  │                   │
     │                 │                    │                   │
     │                 │ 6. POST /chat      │                   │
     │                 │    (with history + │                   │
     │                 │     attachments)   │                   │
     │                 ├───────────────────────────────────────>│
     │                 │                    │                   │
     │                 │                    │            7. AI processes
     │                 │                    │               & fetches files
     │                 │                    │                   │
     │                 │<───────────────────────────────────────┤
     │                 │                    │                   │
     │                 │ 8. Save AI msg     │                   │
     │                 ├───────────────────>│                   │
     │                 │                    │                   │
     │                 │ 9. Return response │                   │
     │<────────────────┤                    │                   │
     │                 │                    │                   │
```

### Processing Steps

1. **Authentication** (Line 59 in `send_message_hdl.go`)
   - Extract `user_id` from JWT claims (`core.KeyRequester`)
   - Never trust `user_id` from request body
   - Return 401 if no valid JWT

2. **Conversation Management** (Lines 14-45 in `send_message.go`)
   - If `conversation_id` provided: validate user has access
   - If not provided: create new conversation with `user_id` from JWT
   - Return 403 if user doesn't have access to conversation

3. **Save User Message** (Lines 54-67 in `send_message.go`)
   - Create message record with role="user"
   - Process and upload file attachments to MinIO (if any)
   - Store attachment metadata in database

4. **Build AI Request Context** (Lines 147-272 in `send_message.go`)
   - Fetch last N conversation turns (configurable via `AI_HISTORY_MAX`)
   - Generate presigned URLs for recent attachments
   - Build context with history, deep_research flag, and file references

5. **Call AI Service** (Lines 165-270 in `ai_client_enhanced.go`)
   - POST to `AI_SERVICE_URL` with full context
   - Timeout: `AI_REST_TIMEOUT` (default: 30s)
   - Handle various error conditions

6. **Parse AI Response** (Lines 202-270 in `ai_client_enhanced.go`)
   - Parse JSON response from AI service
   - Extract: `success`, `message_id`, `response`, `selected_agent`
   - Map fields: `response` → `content`, `selected_agent.agent_name` → `model`

7. **Save AI Message** (Lines 99-115 in `send_message.go`)
   - Create message record with role="assistant"
   - Store AI response content, model name, tokens, latency
   - Record metadata in JSON field

8. **Return Mapped Response** (Lines 131-144 in `send_message.go`)
   - Map to frontend-expected format
   - Include all required fields with correct timestamps

## Security Features

### JWT-Based Authentication
- ✅ **User ID from JWT only** - Never accepts `user_id` from request body
- ✅ **Authorization checks** - Validates user has access to conversation
- ✅ **Role-based permissions** - Requires `chat.send` permission

### Presigned URLs for Files
- ✅ **No credentials exposed** - MinIO credentials never sent to AI service
- ✅ **Short-lived URLs** - Expire after 10 minutes (configurable)
- ✅ **Read-only access** - GET operations only
- ✅ **Scoped to specific files** - Cannot access other users' files

### Data Privacy
- ✅ **PII sanitization** - Can sanitize history before sending to AI
- ✅ **No credential logging** - Presigned URLs not logged
- ✅ **User isolation** - Users can only access their own conversations

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `AI_SERVICE_URL` | `http://localhost:8000` | AI service base URL |
| `AI_PROTOCOL` | `enhanced` | Client type: `mock`, `rest`, `enhanced` |
| `AI_HISTORY_MAX` | `5` | Max conversation turns to include in history |
| `PRESIGNED_URL_EXPIRE_SEC` | `600` | Presigned URL expiry (10 minutes) |
| `AI_REST_TIMEOUT` | `30s` | HTTP request timeout |
| `AI_REST_API_KEY` | _(empty)_ | Optional bearer token for AI service |

### Example `.env` Configuration

```bash
# AI Service Configuration
AI_SERVICE_URL=http://localhost:8000/chat
AI_PROTOCOL=enhanced
AI_HISTORY_MAX=8
PRESIGNED_URL_EXPIRE_SEC=600
AI_REST_TIMEOUT=30s

# Optional: AI service authentication
AI_REST_API_KEY=your-secret-key-here
```

## Testing

### Quick Test

```bash
# 1. Get JWT token
TOKEN=$(curl -s -X POST http://localhost:3001/v1/authenticate \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}' \
  | jq -r '.data.access_token')

# 2. Send message
curl -X POST "http://localhost:3001/v1/chatbot/prompt" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content":"Hello, how are you?"}'

# 3. Send follow-up (with conversation_id)
curl -X POST "http://localhost:3001/v1/chatbot/prompt?conversation_id=<CONV_ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content":"Tell me more"}'
```

### Comprehensive Test Scripts

Run the provided test scripts to verify the complete integration:

**Linux/Mac:**
```bash
chmod +x test_frontend_integration.sh
./test_frontend_integration.sh
```

**Windows (PowerShell):**
```powershell
.\test_frontend_integration.ps1
```

### Test Cases Covered

1. ✅ **JWT Authentication** - Token extraction and validation
2. ✅ **New Conversation** - Creating conversation without conversation_id
3. ✅ **Existing Conversation** - Using conversation_id query param
4. ✅ **Conversation History** - AI receives previous messages
5. ✅ **Deep Research Flag** - `deep_research` passed to AI
6. ✅ **File Attachments** - Presigned URLs generated and sent
7. ✅ **Error Handling** - Invalid conversation_id returns 403
8. ✅ **Auth Enforcement** - No token returns 401
9. ✅ **Response Mapping** - Correct field mapping to frontend format

## AI Service Integration

### Request to AI Service

The backend sends this payload to `http://localhost:8000/chat`:

```json
{
  "message": "User message here",
  "session_id": "019a1694-14e0-727a-b76a-b13d81dcd220",
  "user_id": "user_abc_from_jwt",
  "context": {
    "history": [
      {"role":"user","message":"Previous question","created_at":"2025-10-30T10:00:00Z"},
      {"role":"assistant","message":"Previous answer","created_at":"2025-10-30T10:00:02Z"}
    ],
    "history_count": 5,
    "deep_research": false,
    "attach_files": [
      {
        "url": "https://minio.example/file.pdf?X-Amz-Signature=...",
        "filename": "report.pdf",
        "mime_type": "application/pdf",
        "sha256": "abc123...",
        "pages": 10
      }
    ]
  }
}
```

### Expected AI Service Response

```json
{
  "success": true,
  "message_id": "fe19a683-7f27-49eb-905c-595dd4089a5a",
  "response": "AI-generated response text",
  "selected_agent": {
    "agent_name": "FinanceBot",
    "agent_type": "assistant",
    "confidence_score": 0.9,
    "reasoning": "Processed user request",
    "execution_time": "~1-2 seconds",
    "tools_used": []
  },
  "execution_time": 2.761198,
  "timestamp": "2025-10-30T18:14:58.887803",
  "session_id": "019a1694-14e0-727a-b76a-b13d81dcd220",
  "error": null
}
```

### Field Mapping

| AI Response Field | Frontend Response Field | Notes |
|-------------------|------------------------|-------|
| `success` | _(validation)_ | Must be `true`, else error |
| `message_id` | `assistant_message_id` | Used as DB record ID |
| `response` | `assistant_content` | Main AI reply |
| `selected_agent.agent_name` | `model_name` | Agent/model identifier |
| `execution_time` | `latency_ms` | Converted to milliseconds |
| `timestamp` | `ai_message_created_at` | ISO 8601 format |
| `session_id` | `conversation_id` | Should match request |
| `tokens_input` | `tokens_input` | Optional, 0 if not provided |
| `tokens_output` | `tokens_output` | Optional, 0 if not provided |

## Error Handling

### AI Service Errors

| Scenario | Backend Behavior | Frontend Response |
|----------|-----------------|-------------------|
| AI returns `success: false` | Return 502 | `{"error": "AI service error: <message>"}` |
| HTTP timeout | Return 502 | `{"error": "AI service timeout"}` |
| HTTP 429 | Return 502 | `{"error": "AI service rate limit exceeded"}` |
| HTTP 500+ | Return 502 | `{"error": "AI service unavailable"}` |
| Network error | Return 502 | `{"error": "AI service network error"}` |

### Database Errors

| Scenario | Backend Behavior | Frontend Response |
|----------|-----------------|-------------------|
| Conversation not found | Return 403 | `{"error": "conversation not found or access denied"}` |
| User message save fails | Return 500 | `{"error": "failed to save message"}` |
| AI message save fails | Return 500 | `{"error": "failed to save AI response"}` |

### Authentication Errors

| Scenario | Backend Behavior | Frontend Response |
|----------|-----------------|-------------------|
| No JWT token | Return 401 | `{"error": "user ID is required"}` |
| Invalid JWT | Return 401 | `{"error": "invalid token"}` |
| No `chat.send` permission | Return 403 | `{"error": "insufficient permissions"}` |

## Troubleshooting

### "AI service timeout"
**Cause:** AI processing takes >30s  
**Fix:** Increase `AI_REST_TIMEOUT=60s`

### "Conversation not found or access denied"
**Cause:** User doesn't own the conversation or conversation doesn't exist  
**Fix:** Verify conversation_id is correct and belongs to authenticated user

### "Empty history in AI requests"
**Cause:** No previous messages in conversation  
**Solution:** This is normal for first message in conversation

### "Failed to generate presigned URL"
**Cause:** MinIO not accessible or credentials invalid  
**Fix:** Verify `MINIO_*` environment variables and MinIO service is running

## Monitoring

### Log Format

Look for these log entries:

```
[AI] Request: session=conv_123 user=user_abc history=2 files=1
[AI] Success: session=conv_123 model=FinanceBot latency=2350ms tokens_in=120 tokens_out=85
[AI] Error: session=conv_456 status=502 error=timeout request_id=req_xyz
```

### Metrics to Track

- **Request latency** (p50, p95, p99)
- **Error rate** by status code
- **Token consumption** (for cost analysis)
- **History size** distribution
- **File attachment** usage

### Health Checks

```bash
# Check backend health
curl http://localhost:3001/health

# Check AI service health
curl http://localhost:8000/health

# Check MinIO health
curl http://localhost:9000/minio/health/live
```

## Related Documentation

- [Enhanced AI Integration](./AI_SERVICE_ENHANCED_INTEGRATION.md) - Complete technical docs
- [AI Integration Quickstart](./AI_SERVICE_QUICKSTART.md) - Quick setup guide
- [File Upload Documentation](./README_FILE_UPLOAD.md) - File attachment handling
- [Conversation History API](./CONVERSATION_HISTORY_API.md) - History management

## Support

For issues or questions:
1. Check this documentation
2. Review logs for `[AI]` prefix
3. Test with `AI_PROTOCOL=mock` to isolate issues
4. Verify environment variables are set correctly

---

**Status:** ✅ **PRODUCTION READY**  
**Last Updated:** 2025-10-30  
**Build Status:** ✅ **PASSING**

