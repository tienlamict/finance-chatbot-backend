# Send Message Handler Bug Fix

## Issue Description

The `send_message_hdl.go` file had a bug where the user ID was being incorrectly extracted and converted to an integer, but the system expects user IDs as strings throughout the chatbot module.

## Problems Identified

1. **Incorrect User ID Source**: The code was trying to extract user ID from URL parameters (`c.Param("userId")`) instead of from the authentication context (JWT token)
2. **Type Mismatch**: The user ID was being converted to `int` but the business logic expects `string`
3. **Security Issue**: Extracting user ID from URL parameters instead of authenticated context could allow user impersonation

## Root Cause

The user modified the code to extract user ID from URL parameters and convert it to an integer:

```go
// INCORRECT CODE
uid, err := core.FromBase58(c.Param("userId"))
if err != nil {
    common.WriteErrorResponse(c, core.ErrBadRequest.WithError("invalid user ID format"))
    return
}
userID := int(uid.GetLocalID())  // Converting to int - WRONG!
```

## Fix Applied

Restored the correct implementation that extracts user ID from authentication context:

```go
// CORRECT CODE
userID := a.resolveUserID(c, req.UserID)
// UserID is now extracted from authentication context
if userID == "" {
    common.WriteErrorResponse(c, core.ErrUnauthorized.WithError("user ID is required"))
    return
}
```

## Why This Fix is Correct

1. **Security**: User ID comes from authenticated JWT token context, not from client input
2. **Type Consistency**: User ID remains as string, matching the business logic expectations
3. **Proper Authentication**: Uses the existing `resolveUserID()` function that correctly extracts from `core.KeyRequester`

## System Architecture

The correct flow is:

1. **Authentication Middleware**: Processes JWT token and stores user info in context under `core.KeyRequester`
2. **Handler**: Extracts user ID from authentication context using `resolveUserID()`
3. **Business Logic**: Receives user ID as string and stores it in database
4. **Database**: Stores user ID as VARCHAR(50) string

## Type Consistency Across the System

- **Authentication Context**: User ID stored as string in JWT token subject
- **Handler Layer**: User ID extracted as string
- **Business Layer**: User ID processed as string
- **Database Layer**: User ID stored as VARCHAR(50) string

## Files Modified

- `microservice/chatbot/transport/api/send_message_hdl.go` - Fixed user ID extraction and type handling

## Testing

The fix ensures that:

1. ✅ User ID is extracted from authentication context (JWT token)
2. ✅ User ID remains as string type throughout the system
3. ✅ Security is maintained by not accepting user ID from client input
4. ✅ Type consistency is maintained across all layers

## Prevention

To prevent similar issues in the future:

1. **Always use authentication context** for user ID extraction
2. **Maintain type consistency** - user IDs should be strings in the chatbot module
3. **Follow the established pattern** used in other handlers like `conversation_history_hdl.go`
4. **Test type compatibility** when making changes to ID handling

## Related Files

- `microservice/chatbot/transport/api/conversation_history_hdl.go` - Shows correct pattern for user ID extraction
- `middleware/authen.go` - Shows how authentication context is set up
- `addon/core/requester.go` - Shows the Requester interface structure
