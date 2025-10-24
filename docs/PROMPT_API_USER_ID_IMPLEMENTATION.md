# Prompt API User ID Implementation

## Overview

This document describes the implementation of automatic user ID extraction and storage for the `/chatbot/prompt` API endpoint. When a new conversation is created, the system now automatically retrieves the authenticated user's ID from the JWT token and stores it in the `conversations.user_id` field.

## Changes Made

### 1. Updated Authentication Context Extraction

**File**: `microservice/chatbot/transport/api/send_message_hdl.go`

- Modified `resolveUserID()` function to properly extract user ID from authentication context
- Added import for `finance-chatbot/addon/core` to access the `KeyRequester` constant
- Updated logic to get user ID from `core.KeyRequester` (JWT token context) instead of generic `user_id` key
- Added validation to ensure user ID is not empty before proceeding
- Updated comments to clarify that client-provided user_id is ignored for security

### 2. Enhanced Business Logic Validation

**File**: `microservice/chatbot/business/send_message.go`

- Added validation in `ensureConversation()` function to ensure user ID is not empty when creating new conversations
- Added error handling for missing user ID with descriptive error message

### 3. Fixed Route Typo

**File**: `cmd/root.go`

- Fixed typo in route definition: changed `/promt` to `/prompt`

### 4. Updated Entity Documentation

**File**: `microservice/chatbot/entity/chat.go`

- Updated comment for `UserID` field in `SendMessageRequest` to clarify that it's ignored in favor of authenticated user ID
- Maintained backward compatibility by keeping the field in the struct

## Implementation Details

### Authentication Flow

1. **JWT Token Processing**: The authentication middleware (`middleware.RequireAuth`) processes the JWT token and extracts the user ID from the token's subject field
2. **Context Storage**: The user ID is stored in the Gin context under `core.KeyRequester` as a `Requester` object
3. **User ID Extraction**: The `resolveUserID()` function extracts the user ID from the `Requester.GetSubject()` method
4. **Conversation Creation**: When creating a new conversation, the extracted user ID is automatically assigned to the `conversations.user_id` field

### Security Considerations

- **Client Input Ignored**: The `user_id` field in the request body is ignored to prevent users from impersonating other users
- **Authentication Required**: User ID can only come from the authenticated JWT token context
- **Validation**: The system validates that a user ID is present before creating conversations

### Database Schema

The existing `conversations` table schema is used without modifications:

```sql
CREATE TABLE conversations (
  `id`            CHAR(36) NOT NULL,
  `user_id`       VARCHAR(50) NOT NULL,
  `org_id`        VARCHAR(50) NULL,
  `title`         VARCHAR(255) NULL,
  `status`        ENUM('active','archived','deleted') DEFAULT 'active',
  `created_at`    DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at`    DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  INDEX idx_conv_user_created (user_id, created_at),
  INDEX idx_conv_org_created  (org_id, created_at)
);
```

## API Behavior

### New Conversation Creation

When a user sends a message to the `/chatbot/prompt` endpoint without providing a `conversation_id`:

1. System extracts user ID from JWT token
2. Creates new conversation with the authenticated user's ID
3. Stores the conversation in the database with proper `user_id` assignment

### Existing Conversation

When a user sends a message to an existing conversation:

1. System uses the existing conversation
2. Does not update the `user_id` field (as required)
3. Processes the message normally

## Testing

### Test Scripts Created

1. **`test_prompt_user_id.sh`** - Bash script for testing the implementation
2. **`test_prompt_user_id.ps1`** - PowerShell script for Windows users

### Test Cases

The test scripts cover:

- ✅ Authenticated message sending (new conversation creation)
- ✅ Authenticated message sending (existing conversation)
- ✅ Unauthenticated requests (should fail with 401)
- ✅ Invalid token requests (should fail with 401)
- ✅ Empty content validation (should fail with 400)
- ✅ Database verification queries

### Running Tests

```bash
# For Linux/Mac
chmod +x test_prompt_user_id.sh
./test_prompt_user_id.sh

# For Windows PowerShell
.\test_prompt_user_id.ps1
```

**Note**: Replace `your-jwt-token-here` with a valid JWT token from the authentication endpoint.

## Expected Results

### Successful Requests

- **Status**: 200 OK
- **Response**: Contains `conversation_id`, message IDs, and AI response
- **Database**: New conversations have `user_id` set from authentication context

### Error Cases

- **401 Unauthorized**: Missing or invalid authentication token
- **400 Bad Request**: Empty content or validation errors
- **500 Internal Server Error**: Server-side processing errors

## Backward Compatibility

- The `user_id` field in the request body is maintained for backward compatibility
- Existing client applications will continue to work without modifications
- The field is ignored in favor of the authenticated user ID for security

## Future Considerations

1. **Database Constraints**: Consider adding foreign key constraints to ensure `user_id` references valid users
2. **Audit Logging**: Consider adding audit logs for conversation creation with user tracking
3. **Rate Limiting**: Consider implementing rate limiting per user for conversation creation
4. **Conversation Ownership**: Consider implementing conversation ownership validation for access control

## Files Modified

1. `microservice/chatbot/transport/api/send_message_hdl.go` - Updated user ID extraction logic
2. `microservice/chatbot/business/send_message.go` - Added validation for user ID
3. `cmd/root.go` - Fixed route typo
4. `microservice/chatbot/entity/chat.go` - Updated documentation

## Files Created

1. `test_prompt_user_id.sh` - Bash test script
2. `test_prompt_user_id.ps1` - PowerShell test script
3. `docs/PROMPT_API_USER_ID_IMPLEMENTATION.md` - This documentation

## Conclusion

The implementation successfully addresses the requirements:

- ✅ User ID is automatically extracted from authentication context
- ✅ New conversations are created with the authenticated user's ID
- ✅ Existing conversations are not modified
- ✅ Client input is ignored for security
- ✅ Backward compatibility is maintained
- ✅ Proper error handling and validation is implemented
- ✅ Test scripts are provided for verification

The system now properly associates conversations with authenticated users while maintaining security and backward compatibility.
