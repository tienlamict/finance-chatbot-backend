# AI Service Integration

## Overview

The chatbot is now integrated with the real AI service running on `http://localhost:8000/chat` instead of using mock responses.

## Configuration

### Environment Variables

Set the following environment variables to configure the AI service integration:

```bash
# AI Protocol (default: rest)
AI_PROTOCOL=rest

# AI Service Base URL (default: http://localhost:8000)
AI_REST_BASE_URL=http://localhost:8000

# Optional: API Key for authentication
AI_REST_API_KEY=your-api-key-here

# Optional: Request timeout (default: 5s)
AI_REST_TIMEOUT=5s
```

### Testing with Mock

To use the mock AI client for testing, set:

```bash
AI_PROTOCOL=mock
```

## API Integration

### Request Format

The system sends HTTP POST requests to the AI service:

```bash
POST http://localhost:8000/chat
Content-Type: application/json

{
  "message": "Phân tích cổ phiếu VIC",
  "query_type": "stock_research",
  "priority": "HIGH"
}
```

### Query Type Detection

The system automatically detects query types based on message content:

- **stock_research**: Messages containing keywords like "cổ phiếu", "chứng khoán", "VIC", "VNM", "stock"
- **analysis**: Messages containing keywords like "phân tích", "analyze"
- **general**: All other messages (default)

### Response Format

The AI service should return a JSON response:

```json
{
  "content": "AI response content here",
  "model": "model-name",
  "tokens_input": 100,
  "tokens_output": 150
}
```

## Error Handling

The integration includes robust error handling:

- **Network Errors**: Returns connection error messages
- **HTTP Errors**: Returns status code and error message
- **Parse Errors**: Returns JSON decode error messages
- **Timeout**: Respects the configured timeout setting

## Features

### 1. Automatic Query Type Detection

The system analyzes the user's message to determine the appropriate query type:

```go
queryType := "general"
if containsKeywords(prompt, []string{"cổ phiếu", "chứng khoán", "VIC", "VNM", "stock"}) {
    queryType = "stock_research"
} else if containsKeywords(prompt, []string{"phân tích", "analyze"}) {
    queryType = "analysis"
}
```

### 2. HTTP Client Configuration

The HTTP client is configured with:

- Connection pooling
- Timeout handling
- Keep-alive connections
- Proxy support

### 3. Priority Setting

All requests are sent with `"priority": "HIGH"` to ensure timely responses.

## Testing

### Test the Integration

1. **Start the AI service** on port 8000
2. **Set environment variables**:
   ```bash
   export AI_PROTOCOL=rest
   export AI_REST_BASE_URL=http://localhost:8000
   ```
3. **Start the chatbot service**
4. **Send a test message**:
   ```bash
   curl -X POST http://localhost:3001/v1/chatbot/prompt \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"content": "Phân tích cổ phiếu VIC"}'
   ```

### Expected Behavior

1. The chatbot receives the message
2. Sends request to `http://localhost:8000/chat`
3. Receives AI response
4. Stores both user message and AI response in database
5. Returns the conversation with AI response

## Fallback Options

If the AI service is unavailable:

1. Check the logs for error messages
2. Set `AI_PROTOCOL=mock` to use mock responses
3. Verify AI service is running: `curl http://localhost:8000/chat`

## Monitoring

### Key Metrics

- Request latency (stored in `messages.latency_ms`)
- Token usage (stored in `messages.tokens_input` and `messages.tokens_output`)
- Error rates
- Response times

### Logs

Monitor the application logs for:
- AI service connection errors
- Request/response details
- Timeout occurrences

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Verify AI service is running
   - Check `AI_REST_BASE_URL` setting
   - Verify port 8000 is accessible

2. **Timeout Errors**
   - Increase `AI_REST_TIMEOUT` value
   - Check AI service performance
   - Verify network connectivity

3. **Authentication Errors**
   - Verify `AI_REST_API_KEY` is correct
   - Check AI service authentication requirements

## Implementation Details

### Files Modified

1. `composer/http_client_ai.go` - Updated REST client implementation
2. `composer/service_composer.go` - Changed default to REST client
3. `composer/mock_ai_client.go` - Existing mock implementation (for testing)

### Key Changes

1. **Endpoint**: Changed from `/v1/generate` to `/chat`
2. **Request Format**: Updated to match AI service API
3. **Query Type Detection**: Added automatic query type detection
4. **Default Behavior**: Changed default from mock to REST

## Benefits

1. **Real AI Integration**: Live AI responses instead of mock data
2. **Flexible Configuration**: Environment variable-based configuration
3. **Error Handling**: Robust error handling and recovery
4. **Monitoring**: Built-in metrics and logging
5. **Testing**: Easy switch between mock and real AI service

## Next Steps

1. Deploy AI service to production
2. Configure production URL in environment variables
3. Monitor performance and adjust timeout settings
4. Set up alerting for AI service failures
5. Consider adding retry logic for transient failures
