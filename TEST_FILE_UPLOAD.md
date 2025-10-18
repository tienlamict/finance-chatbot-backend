# Testing the File Upload Feature - Comprehensive testing guide

## Quick Start Guide

### 1. Start MinIO (Local Testing)

Using Docker:
```bash
docker run -d \
  -p 9000:9000 \
  -p 9001:9001 \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin" \
  --name minio \
  minio/minio server /data --console-address ":9001"
```

Or add to your `docker-compose.yaml`:
```yaml
services:
  minio:
    image: minio/minio:latest
    container_name: finance-chatbot-minio
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    command: server /data --console-address ":9001"
    volumes:
      - minio_data:/data
    networks:
      - app-network

volumes:
  minio_data:

networks:
  app-network:
```

Access MinIO Console: http://localhost:9001 (login: minioadmin/minioadmin)

### 2. Configure Environment Variables

Create or update `.env` file:
```env
# MinIO Configuration
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY_ID=minioadmin
MINIO_SECRET_ACCESS_KEY=minioadmin
MINIO_USE_SSL=false
MINIO_BUCKET_NAME=finance-chatbot-attachments
MINIO_REGION=us-east-1
```

### 3. Run the Application

```bash
go run main.go
```

Expected output should include:
```
MinIO endpoint not configured, file upload will be disabled
# OR if configured:
# (MinIO will be initialized)
```

### 4. Test File Upload

#### Create Test Files
```bash
# Create a test text file
echo "This is a test document for the chatbot" > test.txt

# Create a test JSON file
echo '{"data": "test"}' > test.json

# Or use any existing PDF, image, etc.
```

#### Test with curl

**Test 1: Send message with single file**
```bash
curl -X POST http://localhost:8080/v1/chatbot/promt \
  -H "X-User-Id: test-user-123" \
  -F "content=Please analyze this document" \
  -F "files=@test.txt"
```

**Test 2: Send message with multiple files**
```bash
curl -X POST http://localhost:8080/v1/chatbot/promt \
  -H "X-User-Id: test-user-123" \
  -F "content=Here are multiple documents" \
  -F "files=@test.txt" \
  -F "files=@test.json"
```

**Test 3: Send message with conversation context**
```bash
curl -X POST http://localhost:8080/v1/chatbot/promt \
  -H "X-User-Id: test-user-123" \
  -F "content=Follow-up question" \
  -F "conversation_id=YOUR_CONVERSATION_ID" \
  -F "files=@test.txt"
```

**Test 4: Send message without files (JSON)**
```bash
curl -X POST http://localhost:8080/v1/chatbot/promt \
  -H "X-User-Id: test-user-123" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Hello chatbot",
    "conversation_id": ""
  }'
```

### 5. Expected Response

Successful upload response:
```json
{
  "conversation_id": "01JCXXX...",
  "user_message_id": "01JCYYY...",
  "assistant_message_id": "01JCZZZ...",
  "assistant_content": "AI response here...",
  "model_name": "mock-model",
  "tokens_input": 100,
  "tokens_output": 50,
  "latency_ms": 150,
  "user_message_created_at": "2025-10-18T10:30:00.123Z",
  "ai_message_created_at": "2025-10-18T10:30:00.273Z",
  "attachments": [
    {
      "id": "01JC...",
      "filename": "test.txt",
      "mime_type": "text/plain",
      "sha256": "abc123...",
      "created_at": "2025-10-18T10:30:00.123Z",
      "storage_key": "finance-chatbot-attachments/attachments/test-user-123/01JCYYY.../1234567890.txt"
    }
  ]
}
```

### 6. Verify Upload in MinIO

1. Open MinIO Console: http://localhost:9001
2. Login with credentials (minioadmin/minioadmin)
3. Navigate to Buckets → `finance-chatbot-attachments`
4. Browse to `attachments/test-user-123/` to see uploaded files

### 7. Verify Database Records

```sql
-- Check messages
SELECT * FROM messages WHERE role = 'user' ORDER BY created_at DESC LIMIT 5;

-- Check attachments
SELECT * FROM message_attachments ORDER BY created_at DESC LIMIT 5;

-- Join query to see message with attachments
SELECT 
  m.id as message_id,
  m.content,
  m.created_at,
  a.id as attachment_id,
  a.filename,
  a.mime_type,
  a.sha256,
  a.storage_key
FROM messages m
LEFT JOIN message_attachments a ON a.message_id = m.id
WHERE m.role = 'user'
ORDER BY m.created_at DESC
LIMIT 10;
```

## Testing Scenarios

### Scenario 1: File Size Validation
```bash
# Create large file (>50MB)
dd if=/dev/zero of=large.bin bs=1M count=60

# Try to upload - should fail with error
curl -X POST http://localhost:8080/v1/chatbot/promt \
  -H "X-User-Id: test-user-123" \
  -F "content=Large file test" \
  -F "files=@large.bin"

# Expected: 400 Bad Request with error message
```

### Scenario 2: Multiple Files Limit
```bash
# Try uploading 11 files (limit is 10)
curl -X POST http://localhost:8080/v1/chatbot/promt \
  -H "X-User-Id: test-user-123" \
  -F "content=Too many files" \
  -F "files=@file1.txt" \
  -F "files=@file2.txt" \
  ... (11 files total)

# Expected: 400 Bad Request - "too many files"
```

### Scenario 3: File Without Content
```bash
# Missing content field
curl -X POST http://localhost:8080/v1/chatbot/promt \
  -H "X-User-Id: test-user-123" \
  -F "files=@test.txt"

# Expected: 400 Bad Request - "content is required"
```

### Scenario 4: Different File Types
```bash
# Test with different MIME types
curl -X POST http://localhost:8080/v1/chatbot/promt \
  -H "X-User-Id: test-user-123" \
  -F "content=Mixed file types" \
  -F "files=@document.pdf" \
  -F "files=@image.png" \
  -F "files=@data.json" \
  -F "files=@sheet.xlsx"
```

### Scenario 5: Duplicate File (SHA256 Check)
```bash
# Upload same file twice
curl -X POST http://localhost:8080/v1/chatbot/promt \
  -H "X-User-Id: test-user-123" \
  -F "content=First upload" \
  -F "files=@test.txt"

curl -X POST http://localhost:8080/v1/chatbot/promt \
  -H "X-User-Id: test-user-123" \
  -F "content=Second upload (same file)" \
  -F "files=@test.txt"

# Both should succeed but have same SHA256
# Query to check:
# SELECT filename, sha256, COUNT(*) FROM message_attachments GROUP BY sha256 HAVING COUNT(*) > 1;
```

### Scenario 6: MinIO Not Configured
```bash
# Stop MinIO or unset MINIO_ENDPOINT
unset MINIO_ENDPOINT

# Restart application
go run main.go

# Try to upload file
curl -X POST http://localhost:8080/v1/chatbot/promt \
  -H "X-User-Id: test-user-123" \
  -F "content=Test without MinIO" \
  -F "files=@test.txt"

# Expected: 500 Internal Server Error - "storage provider not configured"

# But text-only messages should still work:
curl -X POST http://localhost:8080/v1/chatbot/promt \
  -H "X-User-Id: test-user-123" \
  -H "Content-Type: application/json" \
  -d '{"content": "Hello without files"}'

# Expected: 200 OK
```

## Integration Testing with Postman

### Import this collection:
```json
{
  "info": {
    "name": "Chatbot File Upload Tests",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Send Message with File",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "X-User-Id",
            "value": "test-user-123"
          }
        ],
        "body": {
          "mode": "formdata",
          "formdata": [
            {
              "key": "content",
              "value": "Please analyze this document",
              "type": "text"
            },
            {
              "key": "files",
              "type": "file",
              "src": "/path/to/your/test.txt"
            }
          ]
        },
        "url": {
          "raw": "http://localhost:8080/v1/chatbot/promt",
          "protocol": "http",
          "host": ["localhost"],
          "port": "8080",
          "path": ["v1", "chatbot", "promt"]
        }
      }
    }
  ]
}
```

## Troubleshooting

### Issue: "storage provider not configured"
**Solution**: Ensure MINIO_ENDPOINT environment variable is set and MinIO is running.

### Issue: "failed to upload file"
**Causes**:
- MinIO not accessible
- Incorrect credentials
- Network issues

**Debug**:
```bash
# Test MinIO connectivity
curl http://localhost:9000/minio/health/live

# Check MinIO logs
docker logs minio
```

### Issue: "file size exceeds maximum"
**Solution**: File is larger than 50MB. Adjust `MaxFileSize` in `addon/common/file_utils.go` if needed.

### Issue: Database foreign key constraint error
**Solution**: Ensure `messages` table has the message_id that attachments reference. Check message was created successfully first.

### Issue: Files uploaded but not showing in MinIO console
**Solution**: 
- Refresh the bucket view
- Check the correct bucket name
- Verify storage_key format in database

## Performance Testing

### Load Test with multiple concurrent uploads
```bash
# Install Apache Bench
apt-get install apache2-utils

# Simple load test
ab -n 100 -c 10 -p test.txt -T "multipart/form-data; boundary=----WebKitFormBoundary" \
  http://localhost:8080/v1/chatbot/promt
```

## Clean Up

### Delete test data from database
```sql
DELETE FROM message_attachments WHERE filename LIKE 'test%';
DELETE FROM messages WHERE content LIKE '%test%';
```

### Delete files from MinIO
1. Via Console: Browse and delete files manually
2. Via CLI:
```bash
# Install mc (MinIO Client)
mc alias set local http://localhost:9000 minioadmin minioadmin
mc rm --recursive --force local/finance-chatbot-attachments/attachments/test-user-123/
```

## Notes
- Always test with authentication enabled in production
- Monitor storage usage and implement cleanup policies
- Consider implementing file retention policies
- Add monitoring for upload success/failure rates
- Implement rate limiting for file uploads in production

