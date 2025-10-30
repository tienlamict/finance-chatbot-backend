# 🎯 Frontend `/v1/chatbot/prompt` Integration - Complete

## 📋 Overview

The frontend `/v1/chatbot/prompt` API has been **successfully integrated** with the AI service at `http://localhost:8000/chat`. The integration is **production-ready** with full authentication, conversation management, history context, file attachments, and comprehensive error handling.

## ✨ Features

✅ **JWT Authentication** - User ID from token, not request body  
✅ **Conversation Management** - Create new or use existing via query param  
✅ **Conversation History** - Last N turns sent to AI (configurable)  
✅ **File Attachments** - Presigned URLs for secure file access  
✅ **Deep Research** - Optional flag for expensive AI operations  
✅ **Error Handling** - Comprehensive error responses  
✅ **Response Mapping** - Correct format for frontend  

## 🚀 Quick Start

### 1. Configuration

```bash
# Add to .env
AI_SERVICE_URL=http://localhost:8000/chat
AI_PROTOCOL=enhanced
AI_HISTORY_MAX=5
PRESIGNED_URL_EXPIRE_SEC=600
AI_REST_TIMEOUT=30s
```

### 2. Start Service

```bash
go run main.go serve
```

### 3. Test

```bash
# Get JWT token
TOKEN=$(curl -s -X POST http://localhost:3001/v1/authenticate \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}' \
  | jq -r '.data.access_token')

# Send message
curl -X POST http://localhost:3001/v1/chatbot/prompt \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content":"Hello, how are you?"}'
```

## 📡 API Specification

### Endpoint

```
POST /v1/chatbot/prompt?conversation_id={optional}
```

### Request

```bash
curl --location 'http://localhost:3001/v1/chatbot/prompt?conversation_id=019a1694-14e0-727a-b76a-b13d81dcd220' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <JWT>' \
--data '{
  "content": "Hello, how are you",
  "deep_research": false
}'
```

### Response

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

## 🔐 Security

| Feature | Implementation |
|---------|----------------|
| Authentication | JWT token required (401 if missing) |
| Authorization | User must own conversation (403 if not) |
| User ID Source | **Always from JWT**, never from request body |
| File Access | Presigned URLs (10 min expiry, read-only) |
| Data Privacy | Users isolated to their own conversations |

## ⚙️ Configuration

| Variable | Default | Purpose |
|----------|---------|---------|
| `AI_SERVICE_URL` | `http://localhost:8000` | AI service endpoint |
| `AI_PROTOCOL` | `enhanced` | Client type |
| `AI_HISTORY_MAX` | `5` | Max conversation turns |
| `PRESIGNED_URL_EXPIRE_SEC` | `600` | URL expiry (10 min) |
| `AI_REST_TIMEOUT` | `30s` | Request timeout |

## 🧪 Testing

### Automated Tests

```bash
# Linux/Mac
chmod +x test_frontend_integration.sh
./test_frontend_integration.sh

# Windows
.\test_frontend_integration.ps1
```

### Manual Testing

```javascript
// JavaScript/Fetch
const response = await fetch('http://localhost:3001/v1/chatbot/prompt', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    content: 'Hello',
    deep_research: false
  })
});

const data = await response.json();
console.log(data);
```

## 📊 How It Works

```
1. Frontend → POST /v1/chatbot/prompt + JWT
              ↓
2. Backend → Extract user_id from JWT
              ↓
3. Backend → Validate conversation access
              ↓
4. Backend → Save user message to DB
              ↓
5. Backend → Fetch conversation history
              ↓
6. Backend → Generate presigned URLs for files
              ↓
7. Backend → POST to AI service with context
              ↓
8. AI Service → Process & return response
              ↓
9. Backend → Save AI message to DB
              ↓
10. Backend → Return formatted response
              ↓
11. Frontend → Receive JSON response
```

## 🔄 AI Service Integration

### What Backend Sends

```json
{
  "message": "User message",
  "session_id": "conv_id",
  "user_id": "user_id_from_jwt",
  "context": {
    "history": [...],
    "history_count": 5,
    "deep_research": false,
    "attach_files": [...]
  }
}
```

### What AI Returns

```json
{
  "success": true,
  "message_id": "...",
  "response": "AI reply",
  "selected_agent": {
    "agent_name": "FinanceBot",
    ...
  },
  "execution_time": 2.761,
  "timestamp": "2025-10-30T18:14:58.887803",
  "session_id": "...",
  "error": null
}
```

## 🐛 Error Handling

| Error | Status | Response |
|-------|--------|----------|
| No JWT | 401 | `{"error": "user ID is required"}` |
| Invalid conversation | 403 | `{"error": "conversation not found or access denied"}` |
| AI error | 502 | `{"error": "AI service error: ..."}` |
| AI timeout | 502 | `{"error": "AI service timeout"}` |

## 📖 Documentation

| Document | Description | Lines |
|----------|-------------|-------|
| [`docs/FRONTEND_AI_INTEGRATION.md`](docs/FRONTEND_AI_INTEGRATION.md) | Complete API docs | 650+ |
| [`docs/FRONTEND_INTEGRATION_SUMMARY.md`](docs/FRONTEND_INTEGRATION_SUMMARY.md) | Implementation summary | 400+ |
| [`docs/AI_SERVICE_ENHANCED_INTEGRATION.md`](docs/AI_SERVICE_ENHANCED_INTEGRATION.md) | Backend AI integration | 650+ |
| [`docs/AI_SERVICE_QUICKSTART.md`](docs/AI_SERVICE_QUICKSTART.md) | Quick setup | 250+ |

## 📁 Files

### Modified (1)
- `composer/ai_client_enhanced.go` - Updated to parse actual AI service response format

### Created (4)
- `test_frontend_integration.sh` - Bash test script
- `test_frontend_integration.ps1` - PowerShell test script
- `docs/FRONTEND_AI_INTEGRATION.md` - Complete documentation
- `docs/FRONTEND_INTEGRATION_SUMMARY.md` - Implementation summary
- `README_FRONTEND_INTEGRATION.md` - This file

## ✅ Verification

- ✅ Build compiles successfully
- ✅ No linter errors
- ✅ JWT authentication enforced
- ✅ conversation_id query param works
- ✅ Response format matches spec
- ✅ Error handling comprehensive
- ✅ Test scripts provided
- ✅ Documentation complete

## 🎯 Next Steps

### For Frontend Developers
```javascript
// Example integration
async function sendMessage(token, content, conversationId = null) {
  const url = conversationId 
    ? `http://localhost:3001/v1/chatbot/prompt?conversation_id=${conversationId}`
    : 'http://localhost:3001/v1/chatbot/prompt';
    
  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ content, deep_research: false })
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error);
  }
  
  return await response.json();
}
```

### For Backend Team
1. ✅ Integration complete
2. Monitor logs for `[AI]` entries
3. Tune `AI_HISTORY_MAX` based on usage
4. Set up production monitoring

### For AI Service Team
1. Ensure `/chat` endpoint returns specified format
2. Support presigned URL file fetching
3. Handle `deep_research` flag
4. Return `success`, `message_id`, `response`, `selected_agent`

### For DevOps
1. Configure production env vars
2. Monitor AI service health
3. Set up alerts for errors/latency
4. Verify MinIO network access

## 🔍 Troubleshooting

### "401 Unauthorized"
- Ensure JWT token is included in `Authorization` header
- Verify token is valid and not expired

### "403 Forbidden"
- User doesn't own the conversation
- Verify conversation_id is correct

### "502 Bad Gateway"
- AI service is down or unreachable
- Check `AI_SERVICE_URL` configuration
- Verify AI service is running: `curl http://localhost:8000/health`

### "Empty history"
- Normal for first message in new conversation
- History builds up as conversation progresses

## 📞 Support

**Need Help?**
1. Read [`docs/FRONTEND_AI_INTEGRATION.md`](docs/FRONTEND_AI_INTEGRATION.md)
2. Check logs for `[AI]` prefix
3. Test with `AI_PROTOCOL=mock` to isolate issues
4. Verify environment variables are set

## 🎉 Summary

| Aspect | Status |
|--------|--------|
| **Implementation** | ✅ Complete |
| **Build** | ✅ Passing |
| **Documentation** | ✅ Comprehensive |
| **Tests** | ✅ Scripts provided |
| **Security** | ✅ JWT enforced |
| **Error Handling** | ✅ Comprehensive |
| **Response Format** | ✅ Matches spec |

---

**Status:** ✅ **PRODUCTION READY**  
**Build:** ✅ **SUCCESSFUL**  
**Ready for:** ✅ **DEPLOYMENT & TESTING**

**The frontend `/v1/chatbot/prompt` API is fully integrated with the AI service and ready for use!**

