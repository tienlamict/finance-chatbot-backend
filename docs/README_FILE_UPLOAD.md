# File Upload Feature for Chatbot - Feature overview and usage

## Overview

This implementation adds file upload capability to the chatbot messaging flow. Users can now attach one or more files (documents, images, PDFs, etc.) when sending messages to the chatbot.

## Features

1. **Multi-file Upload**: Support uploading multiple files (up to 10) per message
2. **Object Storage**: Files are stored in MinIO/S3-compatible object storage
3. **File Metadata**: Tracks filename, MIME type, SHA256 hash, size, and storage location
4. **File Validation**: Validates file size (max 50MB) and optionally restricts file types
5. **Database Integration**: Stores file metadata in `message_attachments` table linked to messages

## Architecture

### Components Added

1. **Entity Layer** (`microservice/chatbot/entity/attachment.go`)
   - `MessageAttachment`: Database model for attachment metadata
   - `FileUploadInfo`: DTO for file upload information
   - `AttachmentDTO`: Response DTO for attachments

2. **Storage Component** (`addon/component/minioc/minio.go`)
   - MinIO/S3 client wrapper
   - File upload/download/delete operations
   - Presigned URL generation for temporary access
   - Storage key generation

3. **Storage Interface** (`addon/common/storage_interface.go`)
   - Common interface for storage providers
   - Allows easy switching between MinIO, S3, or other storage backends

4. **File Utilities** (`addon/common/file_utils.go`)
   - File validation
   - MIME type detection
   - SHA256 hash calculation
   - Filename sanitization

5. **Repository Layer** (`microservice/chatbot/repository/mysql/attachment_store.go`)
   - CRUD operations for attachments
   - Batch creation support
   - Query by message ID or attachment ID

6. **Business Logic** (`microservice/chatbot/business/send_message.go`)
   - Integrated attachment processing into message flow
   - Handles file upload to storage
   - Creates attachment records in database
   - Rollback support on errors

7. **Transport Layer** (`microservice/chatbot/transport/api/`)
   - `send_message_hdl.go`: Updated to handle multipart/form-data
   - `file_handler.go`: File parsing and validation logic
   - Supports both JSON and multipart requests

## API Usage

### Send Message with Files

**Endpoint**: `POST /v1/chatbot/promt`

**Content-Type**: `multipart/form-data`

**Form Fields**:
- `content` (required): Message text
- `conversation_id` (optional): Existing conversation ID
- `user_id` (optional): User ID (if not from auth context)
- `org_id` (optional): Organization ID
- `title` (optional): Conversation title
- `files` (optional): One or more files to upload

**Example using curl**:
```bash
curl -X POST http://localhost:8080/v1/chatbot/promt \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "content=Please analyze this document" \
  -F "conversation_id=abc123" \
  -F "files=@/path/to/document.pdf" \
  -F "files=@/path/to/image.png"
```

**Example using JavaScript (FormData)**:
```javascript
const formData = new FormData();
formData.append('content', 'Please analyze this document');
formData.append('conversation_id', 'abc123');
formData.append('files', fileInput.files[0]);
formData.append('files', fileInput.files[1]);

fetch('/v1/chatbot/promt', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_TOKEN'
  },
  body: formData
});
```

**Response**:
```json
{
  "conversation_id": "abc123",
  "user_message_id": "msg-user-xyz",
  "assistant_message_id": "msg-ai-xyz",
  "assistant_content": "I've analyzed your documents...",
  "model_name": "gpt-4",
  "tokens_input": 1500,
  "tokens_output": 500,
  "latency_ms": 2500,
  "user_message_created_at": "2025-10-18T10:30:00Z",
  "ai_message_created_at": "2025-10-18T10:30:02Z",
  "attachments": [
    {
      "id": "att-123",
      "filename": "document.pdf",
      "mime_type": "application/pdf",
      "size": 1024000,
      "sha256": "abc123...",
      "pages": 5,
      "created_at": "2025-10-18T10:30:00Z",
      "storage_key": "finance-chatbot-attachments/attachments/user123/msg-user-xyz/1234567890.pdf"
    }
  ]
}
```

### Send Message without Files (JSON)

The API still supports JSON requests without files:

```bash
curl -X POST http://localhost:8080/v1/chatbot/promt \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Hello chatbot",
    "conversation_id": "abc123"
  }'
```

## Configuration

### Environment Variables

Add these environment variables to configure MinIO/S3 storage:

```env
# MinIO/S3 Configuration
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY_ID=minioadmin
MINIO_SECRET_ACCESS_KEY=minioadmin
MINIO_USE_SSL=false
MINIO_BUCKET_NAME=finance-chatbot-attachments
MINIO_REGION=us-east-1
```

**Note**: If `MINIO_ENDPOINT` is not set, the file upload feature will be disabled (but the API will still work for text-only messages).

### Docker Compose Setup

Add MinIO service to your `docker-compose.yaml`:

```yaml
services:
  minio:
    image: minio/minio:latest
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
```

Access MinIO console at: http://localhost:9001

## Database Schema

The `message_attachments` table is already defined in `data.sql`:

```sql
CREATE TABLE message_attachments (
  `id`            CHAR(36) PRIMARY KEY,
  `message_id`    CHAR(36) NOT NULL,
  `filename`      VARCHAR(255) NOT NULL,
  `storage_key`   VARCHAR(512) NOT NULL,
  `mime_type`     VARCHAR(64)  NULL,
  `sha256`        CHAR(64) NULL,
  `pages`         INT NULL,
  `meta`          JSON NULL,
  `created_at`    DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  FOREIGN KEY (message_id) REFERENCES messages(id),
  INDEX idx_att_msg (message_id),
  INDEX idx_att_sha (sha256)
);
```

## File Validation

- **Max File Size**: 50MB per file
- **Max Files**: 10 files per request
- **Allowed Types**: By default, all file types are allowed. You can restrict by modifying `ValidateFileUpload()` in `addon/common/file_utils.go`

## Security Considerations

1. **File Size Limits**: Prevents large file attacks
2. **MIME Type Detection**: Uses magic bytes detection, not just file extensions
3. **SHA256 Hashing**: Detects duplicate files and ensures integrity
4. **Filename Sanitization**: Removes dangerous characters from filenames
5. **Storage Keys**: Uses unique paths to prevent collisions

## Future Enhancements

1. **PDF Page Counting**: Implement PDF analysis to count pages
2. **Image Processing**: Add thumbnail generation for images
3. **Virus Scanning**: Integrate antivirus scanning before storage
4. **File Size Optimization**: Compress large files before storage
5. **Download API**: Add endpoints to download attachments
6. **List Attachments**: Add endpoint to list all attachments for a message
7. **Delete Attachments**: Add endpoint to delete specific attachments

## Dependencies

Add to `go.mod`:
```
github.com/minio/minio-go/v7 v7.0.66
github.com/gabriel-vasile/mimetype v1.4.10
```

Run:
```bash
go get github.com/minio/minio-go/v7
go get github.com/gabriel-vasile/mimetype
```

## Testing

### Manual Testing with curl

```bash
# Create a test file
echo "Test content" > test.txt

# Send message with attachment
curl -X POST http://localhost:8080/v1/chatbot/promt \
  -H "X-User-Id: user123" \
  -F "content=Test message with file" \
  -F "files=@test.txt"
```

### Testing MinIO Connection

```bash
# Check MinIO health
curl http://localhost:9000/minio/health/live

# List buckets (requires credentials)
mc alias set local http://localhost:9000 minioadmin minioadmin
mc ls local
```

## Troubleshooting

1. **"Storage provider not configured"**: Set `MINIO_ENDPOINT` environment variable
2. **"Failed to upload file"**: Check MinIO is running and credentials are correct
3. **"File size exceeds maximum"**: File is larger than 50MB, adjust `MaxFileSize` in `file_utils.go`
4. **"Too many files"**: Trying to upload more than 10 files, adjust `MaxUploadFiles` in `file_handler.go`

## Example Integration

See the updated `microservice/chatbot/transport/api/send_message_hdl.go` for implementation details.

The feature is backward compatible - existing JSON-based message sending will continue to work without modification.

