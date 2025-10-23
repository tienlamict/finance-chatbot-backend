# Conversation History API Documentation

## Overview

This document describes the comprehensive conversation history API implemented for the finance-chatbot-backend. The API provides advanced filtering, pagination, and access control for retrieving conversation messages and user conversation lists.

## Features

- **Advanced Filtering**: Time range, keyword search, cursor-based pagination
- **Access Control**: Only conversation participants or admin/superadmin can access
- **Attachment Support**: Optional inclusion of message attachments
- **Pagination**: Cursor-based and offset-based pagination
- **Performance**: Optimized queries with proper indexing

## API Endpoints

### 1. Get Conversation History

**Endpoint**: `GET /v1/chatbot/conversations/{conversationId}/messages`

**Description**: Retrieves messages from a specific conversation with advanced filtering options.

**Authentication**: Required (JWT Bearer token)

**Authorization**: Requires `chat.read` permission

**Path Parameters**:
- `conversationId` (string, required): The conversation ID

**Query Parameters**:
- `limit` (int, optional): Number of messages to return (default: 20, max: 100)
- `before` (string, optional): Message ID or timestamp to get messages before this point
- `after` (string, optional): Message ID or timestamp to get messages after this point
- `from` (string, optional): ISO 8601 timestamp - get messages from this time
- `to` (string, optional): ISO 8601 timestamp - get messages until this time
- `order` (string, optional): Sort order - "asc" or "desc" (default: "asc")
- `search` (string, optional): Keyword search in message content
- `include_attachments` (bool, optional): Include message attachments (default: true)

**Example Request**:
```bash
GET /v1/chatbot/conversations/conv-123/messages?limit=50&order=desc&search=finance&include_attachments=true
Authorization: Bearer <jwt_token>
```

**Response**:
```json
{
  "conversation_id": "conv-123",
  "total": 150,
  "items": [
    {
      "id": "msg-456",
      "conversation_id": "conv-123",
      "parent_id": null,
      "role": "user",
      "content": "What is the current market trend?",
      "content_type": "text",
      "meta": {},
      "tokens_input": 8,
      "tokens_output": 0,
      "latency_ms": 0,
      "model_name": null,
      "error_code": null,
      "created_at": "2024-01-15T10:30:00Z",
      "attachments": [
        {
          "id": "att-789",
          "filename": "market_report.pdf",
          "mime_type": "application/pdf",
          "size": 1024000,
          "sha256": "abc123...",
          "pages": 10,
          "created_at": "2024-01-15T10:30:00Z"
        }
      ]
    }
  ],
  "has_more": true,
  "next_cursor": "msg-455",
  "previous_cursor": "msg-456"
}
```

### 2. Get User Conversations

**Endpoint**: `GET /v1/chatbot/users/{userId}/conversations`

**Description**: Retrieves conversations for a specific user with filtering options.

**Authentication**: Required (JWT Bearer token)

**Authorization**: Requires `chat.read` permission (users can only access their own conversations)

**Path Parameters**:
- `userId` (string, required): The user ID

**Query Parameters**:
- `limit` (int, optional): Number of conversations to return (default: 20, max: 100)
- `offset` (int, optional): Number of conversations to skip (default: 0)
- `status` (string, optional): Filter by status - "active", "archived", "deleted"
- `from` (string, optional): ISO 8601 timestamp - get conversations from this time
- `to` (string, optional): ISO 8601 timestamp - get conversations until this time
- `search` (string, optional): Keyword search in conversation titles
- `include_last_message` (bool, optional): Include last message preview (default: true)

**Example Request**:
```bash
GET /v1/chatbot/users/user-123/conversations?limit=10&status=active&include_last_message=true
Authorization: Bearer <jwt_token>
```

**Response**:
```json
{
  "user_id": "user-123",
  "total": 25,
  "items": [
    {
      "id": "conv-123",
      "user_id": "user-123",
      "org_id": "org-456",
      "title": "Market Analysis Discussion",
      "status": "active",
      "created_at": "2024-01-15T09:00:00Z",
      "updated_at": "2024-01-15T10:30:00Z",
      "last_message": {
        "id": "msg-456",
        "role": "assistant",
        "content": "Based on the current market data...",
        "created_at": "2024-01-15T10:30:00Z"
      },
      "unread_count": 0,
      "total_messages": 15
    }
  ],
  "has_more": true,
  "next_offset": 10
}
```

### 3. Get Conversation Summary

**Endpoint**: `GET /v1/chatbot/conversations/{conversationId}`

**Description**: Retrieves a conversation with its summary information including last message.

**Authentication**: Required (JWT Bearer token)

**Authorization**: Requires `chat.read` permission

**Path Parameters**:
- `conversationId` (string, required): The conversation ID

**Example Request**:
```bash
GET /v1/chatbot/conversations/conv-123
Authorization: Bearer <jwt_token>
```

**Response**:
```json
{
  "id": "conv-123",
  "user_id": "user-123",
  "org_id": "org-456",
  "title": "Market Analysis Discussion",
  "status": "active",
  "created_at": "2024-01-15T09:00:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "last_message": {
    "id": "msg-456",
    "role": "assistant",
    "content": "Based on the current market data...",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "unread_count": 0,
  "total_messages": 15
}
```

## Access Control

### Authentication
All endpoints require JWT Bearer token authentication:
```
Authorization: Bearer <jwt_token>
```

### Authorization
- **Permission Required**: `chat.read`
- **User Access**: Users can only access their own conversations
- **Admin Access**: Admin/superadmin users can access any conversation (TODO: implement admin check)

### Access Control Logic
1. Extract user ID from JWT token
2. Verify user has `chat.read` permission
3. Check if user is the conversation owner
4. Allow access if user owns the conversation or is admin/superadmin

## Error Responses

### 400 Bad Request
```json
{
  "error": "conversation_id is required"
}
```

### 401 Unauthorized
```json
{
  "error": "missing access token"
}
```

### 403 Forbidden
```json
{
  "error": "access denied to conversation"
}
```

### 404 Not Found
```json
{
  "error": "conversation not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error"
}
```

## Pagination

### Cursor-based Pagination (Conversation History)
- Use `before` or `after` parameters with message IDs or timestamps
- Response includes `next_cursor` and `previous_cursor` for navigation
- More efficient for large datasets

### Offset-based Pagination (User Conversations)
- Use `limit` and `offset` parameters
- Response includes `next_offset` for navigation
- Suitable for smaller datasets

## Filtering Options

### Time Range Filtering
- `from`: Get messages/conversations created after this timestamp
- `to`: Get messages/conversations created before this timestamp
- Supports ISO 8601 format timestamps

### Keyword Search
- `search`: Search in message content or conversation titles
- Case-insensitive partial matching
- Uses SQL LIKE operator with wildcards

### Status Filtering
- `status`: Filter conversations by status (active, archived, deleted)

## Performance Considerations

### Database Indexes
- `idx_msg_conv_created`: (conversation_id, created_at) for message queries
- `idx_conv_user_created`: (user_id, created_at) for user conversation queries
- `idx_msg_role_created`: (role, created_at) for role-based queries

### Query Optimization
- Limit maximum results to 100 per request
- Use cursor-based pagination for better performance
- Lazy load attachments only when requested

## Usage Examples

### Frontend Integration

#### React/JavaScript Example
```javascript
// Get conversation history
const getConversationHistory = async (conversationId, options = {}) => {
  const params = new URLSearchParams({
    limit: options.limit || 20,
    order: options.order || 'asc',
    include_attachments: options.includeAttachments !== false
  });
  
  if (options.search) params.append('search', options.search);
  if (options.before) params.append('before', options.before);
  if (options.after) params.append('after', options.after);
  
  const response = await fetch(
    `/v1/chatbot/conversations/${conversationId}/messages?${params}`,
    {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    }
  );
  
  return response.json();
};

// Get user conversations
const getUserConversations = async (userId, options = {}) => {
  const params = new URLSearchParams({
    limit: options.limit || 20,
    offset: options.offset || 0,
    include_last_message: options.includeLastMessage !== false
  });
  
  if (options.status) params.append('status', options.status);
  if (options.search) params.append('search', options.search);
  
  const response = await fetch(
    `/v1/chatbot/users/${userId}/conversations?${params}`,
    {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    }
  );
  
  return response.json();
};
```

### cURL Examples

#### Get conversation history with search
```bash
curl -X GET \
  "http://localhost:3001/v1/chatbot/conversations/conv-123/messages?limit=50&search=finance&order=desc" \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json"
```

#### Get user conversations
```bash
curl -X GET \
  "http://localhost:3001/v1/chatbot/users/user-123/conversations?limit=10&status=active" \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json"
```

## Implementation Details

### Database Schema
The API uses the existing database schema with these tables:
- `conversations`: Stores conversation metadata
- `messages`: Stores message content and metadata
- `message_attachments`: Stores file attachments

### Business Logic
- **Validation**: Request parameters are validated before processing
- **Access Control**: User permissions are checked before data access
- **Filtering**: Advanced filtering is applied at the database level
- **Pagination**: Efficient pagination using cursors and offsets

### Repository Layer
- **MessageStore**: Handles message-related database operations
- **ConversationStore**: Handles conversation-related database operations
- **Optimized Queries**: Uses proper indexing and query optimization

## Future Enhancements

### Planned Features
1. **Admin Access**: Allow admin/superadmin to access any conversation
2. **Real-time Updates**: WebSocket support for real-time message updates
3. **Advanced Search**: Full-text search with Elasticsearch integration
4. **Message Threading**: Support for message threads and replies
5. **Unread Counts**: Track and return unread message counts
6. **Message Reactions**: Support for message reactions and emojis

### Performance Improvements
1. **Caching**: Redis caching for frequently accessed conversations
2. **Database Sharding**: Horizontal scaling for large datasets
3. **CDN Integration**: CDN for attachment file delivery
4. **Query Optimization**: Further database query optimizations

## Testing

### Unit Tests
- Repository layer tests for database operations
- Business logic tests for access control and validation
- API handler tests for request/response handling

### Integration Tests
- End-to-end API testing with authentication
- Database integration testing
- Performance testing with large datasets

### Manual Testing
Use the provided test scripts or Postman collection to test the API endpoints with various parameter combinations.

## Security Considerations

1. **Input Validation**: All input parameters are validated
2. **SQL Injection**: Uses parameterized queries to prevent SQL injection
3. **Access Control**: Proper authorization checks at multiple layers
4. **Rate Limiting**: Consider implementing rate limiting for production
5. **Audit Logging**: Log access attempts for security monitoring

This API provides a comprehensive solution for conversation history management with proper security, performance, and usability considerations.
