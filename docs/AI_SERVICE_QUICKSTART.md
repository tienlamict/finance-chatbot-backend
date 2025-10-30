# Enhanced AI Service Integration - Quickstart Guide

## Quick Setup (5 minutes)

### 1. Configure Environment Variables

Add these to your `.env` or `docker-compose.yml`:

```bash
# AI Service Configuration
AI_SERVICE_URL=http://localhost:8000/chat
AI_PROTOCOL=enhanced
AI_HISTORY_MAX=5
PRESIGNED_URL_EXPIRE_SEC=600
AI_REST_TIMEOUT=30s
AI_REST_API_KEY=your-api-key-here  # Optional
```

### 2. Start the Service

```bash
# Using Go
go run main.go serve

# Or with Docker
docker-compose up
```

### 3. Test the Integration

#### Test 1: Simple Message (No History)

```bash
# Get JWT token first
JWT_TOKEN=$(curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}' \
  | jq -r '.data.access_token')

# Send a message
curl -X POST http://localhost:3000/api/chatbot/send \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "What is the capital of France?",
    "conversation_id": ""
  }' | jq .
```

**Expected Response:**
```json
{
  "data": {
    "conversation_id": "conv_abc123",
    "user_message_id": "msg_xyz789",
    "assistant_message_id": "msg_def456",
    "assistant_content": "The capital of France is Paris.",
    "model_name": "gpt-4",
    "tokens_input": 20,
    "tokens_output": 10,
    "latency_ms": 850
  }
}
```

#### Test 2: Follow-up Message (With History)

```bash
# Send follow-up in same conversation
curl -X POST http://localhost:3000/api/chatbot/send \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "What about its population?",
    "conversation_id": "conv_abc123"
  }' | jq .
```

**The server will automatically:**
- Fetch previous conversation history
- Send history to AI service
- AI will use context to answer

#### Test 3: Deep Research Mode

```bash
curl -X POST http://localhost:3000/api/chatbot/send \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Analyze VIC stock performance",
    "conversation_id": "",
    "deep_research": true
  }' | jq .
```

#### Test 4: Message with File Attachment

```bash
# Upload a file with the message
curl -X POST http://localhost:3000/api/chatbot/send \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -F 'content=Analyze this document' \
  -F 'conversation_id=' \
  -F 'files=@./test_document.pdf' \
  | jq .
```

**The server will automatically:**
- Upload file to MinIO
- Generate presigned URL
- Send URL to AI service
- AI can fetch and analyze the file

## Verify Integration

### Check Logs

Look for these log entries to confirm proper operation:

```
[AI] Request: session=conv_abc123 user=user_123 history=2 files=0
[AI] Success: session=conv_abc123 model=gpt-4 latency=1250ms tokens_in=150 tokens_out=80
```

### Test with Mock AI Service

For testing without an actual AI service:

```bash
# Set mock mode
AI_PROTOCOL=mock

# Restart service
go run main.go serve

# Send a test message
curl -X POST http://localhost:3000/api/chatbot/send \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"content": "test", "conversation_id": ""}'
```

The mock will return a predefined response instantly.

## Common Configurations

### Development (Local)

```bash
AI_SERVICE_URL=http://localhost:8000/chat
AI_PROTOCOL=enhanced
AI_HISTORY_MAX=3
PRESIGNED_URL_EXPIRE_SEC=300
AI_REST_TIMEOUT=60s
```

### Production

```bash
AI_SERVICE_URL=https://ai-service.example.com/chat
AI_PROTOCOL=enhanced
AI_HISTORY_MAX=8
PRESIGNED_URL_EXPIRE_SEC=600
AI_REST_TIMEOUT=30s
AI_REST_API_KEY=${AI_API_KEY_SECRET}
```

### Testing (Mock)

```bash
AI_PROTOCOL=mock
AI_HISTORY_MAX=5
```

## Troubleshooting

### Issue: "AI service timeout"

```bash
# Increase timeout
AI_REST_TIMEOUT=60s
```

### Issue: "Failed to generate presigned URL"

```bash
# Check MinIO configuration
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY_ID=minioadmin
MINIO_SECRET_ACCESS_KEY=minioadmin
MINIO_BUCKET_NAME=chatbot-files
```

Verify MinIO is running:
```bash
curl http://localhost:9000/minio/health/live
```

### Issue: "Empty history in requests"

Check database for messages:
```sql
SELECT * FROM messages WHERE conversation_id = 'conv_abc123' ORDER BY created_at DESC;
```

## Next Steps

1. **Read Full Documentation**: See [AI_SERVICE_ENHANCED_INTEGRATION.md](./AI_SERVICE_ENHANCED_INTEGRATION.md)
2. **Implement AI Service**: Coordinate with AI team on endpoint implementation
3. **Monitor Performance**: Set up metrics for latency, errors, tokens used
4. **Tune Configuration**: Adjust `AI_HISTORY_MAX` and timeouts based on usage

## Example AI Service Implementation (Python)

For AI service developers, here's a minimal FastAPI endpoint:

```python
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List, Optional

app = FastAPI()

class HistoryItem(BaseModel):
    role: str
    message: str
    created_at: str

class AttachFile(BaseModel):
    url: str
    filename: Optional[str] = None
    mime_type: Optional[str] = None
    sha256: Optional[str] = None
    pages: Optional[int] = None

class Context(BaseModel):
    history: List[HistoryItem]
    history_count: int
    deep_research: bool
    attach_files: List[AttachFile]

class ChatRequest(BaseModel):
    message: str
    session_id: str
    user_id: str
    context: Context

class ChatResponse(BaseModel):
    content: str
    model: str
    tokens_input: int
    tokens_output: int
    request_id: str

@app.post("/chat", response_model=ChatResponse)
async def chat(request: ChatRequest):
    # 1. Build conversation context from history
    conversation = ""
    for item in request.context.history:
        conversation += f"{item.role}: {item.message}\n"
    
    # 2. Fetch attached files if any
    file_contents = []
    for file in request.context.attach_files:
        # Fetch file from presigned URL
        import httpx
        async with httpx.AsyncClient() as client:
            response = await client.get(file.url)
            if response.status_code == 200:
                file_contents.append(response.content)
    
    # 3. Call your AI model
    ai_response = call_your_ai_model(
        prompt=request.message,
        history=conversation,
        files=file_contents,
        deep_research=request.context.deep_research
    )
    
    # 4. Return response
    return ChatResponse(
        content=ai_response.text,
        model="gpt-4",
        tokens_input=ai_response.tokens_in,
        tokens_output=ai_response.tokens_out,
        request_id=f"req_{request.session_id}"
    )

def call_your_ai_model(prompt, history, files, deep_research):
    # Your AI implementation here
    pass
```

## Support

For issues or questions:
- Check the [full documentation](./AI_SERVICE_ENHANCED_INTEGRATION.md)
- Review logs for error messages
- Test with `AI_PROTOCOL=mock` to isolate issues

