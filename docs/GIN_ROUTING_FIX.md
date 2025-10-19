# Gin Routing Conflict Fix

## The Problem

You encountered this error:
```
panic: ':roleId' in new path '/v1/rbac/roles/:roleId/permissions/:permId' 
conflicts with existing wildcard ':id' in existing prefix '/v1/rbac/roles/:id'
```

## Why It Happened

**Gin's routing system doesn't allow conflicting wildcard parameters at the same path segment level.**

When you have routes like:
```go
GET    /rbac/roles/:id                          // Route A
DELETE /rbac/roles/:roleId/permissions/:permId  // Route B
```

Gin can't tell the difference between:
- `/rbac/roles/123` → Should this match Route A with `id=123`?
- `/rbac/roles/123/permissions/456` → Should this match Route B with `roleId=123`?

Both routes start with `/rbac/roles/` followed by a wildcard. Gin sees `:id` and `:roleId` as conflicting wildcards at the same position, even though they have different continuation paths.

## The Solution

We restructured the routes to avoid the conflict by using different path segments:

### ❌ Before (Conflicting)
```go
// Role management
rbac.GET("/roles/:id", ...)                              // ✗ Conflicts
rbac.DELETE("/roles/:roleId/permissions/:permId", ...)   // ✗ Conflicts

// User management  
rbac.GET("/users/:userId/roles", ...)                    // ✗ Conflicts
```

### ✅ After (Fixed)
```go
// Role management
rbac.GET("/roles/:id", ...)                              // ✓ No conflict
rbac.DELETE("/role-permissions/:id/:permId", ...)        // ✓ Different path segment

// User management
rbac.GET("/user-roles/:userId", ...)                     // ✓ Different path segment
```

## Updated API Routes

### Role Management (No Changes)
```
POST   /v1/rbac/roles          Create role
GET    /v1/rbac/roles          List roles
GET    /v1/rbac/roles/:id      Get role with permissions
PATCH  /v1/rbac/roles/:id      Update role
DELETE /v1/rbac/roles/:id      Delete role
```

### Permission Management (No Changes)
```
POST   /v1/rbac/permissions    Create permission
GET    /v1/rbac/permissions    List permissions
```

### Role-Permission Assignment (CHANGED)
```
OLD: POST   /v1/rbac/roles/:id/permissions
NEW: POST   /v1/rbac/role-permissions/:id

OLD: DELETE /v1/rbac/roles/:roleId/permissions/:permId
NEW: DELETE /v1/rbac/role-permissions/:id/:permId
```

### User-Role Assignment (CHANGED)
```
OLD: POST   /v1/rbac/users/:userId/roles
NEW: POST   /v1/rbac/user-roles/:userId

OLD: GET    /v1/rbac/users/:userId/roles
NEW: GET    /v1/rbac/user-roles/:userId

OLD: DELETE /v1/rbac/users/:userId/roles/:roleId
NEW: DELETE /v1/rbac/user-roles/:userId/:roleId
```

## Updated cURL Examples

### Assign Permissions to Role
```bash
# OLD (conflicting)
curl -X POST http://localhost:8080/v1/rbac/roles/2/permissions \
  -H "Authorization: Bearer <token>" \
  -d '{"permission_ids": [1,2,3]}'

# NEW (working)
curl -X POST http://localhost:8080/v1/rbac/role-permissions/2 \
  -H "Authorization: Bearer <token>" \
  -d '{"permission_ids": [1,2,3]}'
```

### Assign Roles to User
```bash
# OLD (conflicting)
curl -X POST http://localhost:8080/v1/rbac/users/5/roles \
  -H "Authorization: Bearer <token>" \
  -d '{"role_ids": [2,3]}'

# NEW (working)
curl -X POST http://localhost:8080/v1/rbac/user-roles/5 \
  -H "Authorization: Bearer <token>" \
  -d '{"role_ids": [2,3]}'
```

### Get User's Roles
```bash
# OLD (conflicting)
curl -X GET http://localhost:8080/v1/rbac/users/5/roles \
  -H "Authorization: Bearer <token>"

# NEW (working)
curl -X GET http://localhost:8080/v1/rbac/user-roles/5 \
  -H "Authorization: Bearer <token>"
```

## Why This Design Is Better

1. **No Conflicts**: Each resource type has its own distinct path segment
2. **Clear Semantics**: 
   - `/role-permissions/:id` clearly indicates "role-permission relationships"
   - `/user-roles/:userId` clearly indicates "user-role relationships"
3. **RESTful**: Still follows REST principles with distinct resources
4. **Consistent**: Uses hyphens for multi-word resources (standard in URLs)

## Alternative Solutions (Not Used)

### Option 1: Use Same Parameter Name (Not Ideal)
```go
rbac.GET("/roles/:id", ...)
rbac.DELETE("/roles/:id/permissions/:permId", ...)  // Use :id instead of :roleId
```
**Why not**: Handler code would be confusing - what does `:id` refer to?

### Option 2: Use Query Parameters (Not RESTful)
```go
rbac.DELETE("/role-permissions?roleId=X&permId=Y", ...)
```
**Why not**: Not RESTful, harder to read, less semantic

### Option 3: Nested Router Groups (Overly Complex)
```go
roleGroup := rbac.Group("/roles/:id")
roleGroup.POST("/permissions", ...)
```
**Why not**: More complex code structure, similar conflicts can still occur

## Files Updated

1. ✅ `cmd/root.go` - Route definitions
2. ✅ `docs/RBAC_API.md` - API documentation
3. ✅ `RBAC_QUICKSTART.md` - Quick start guide
4. ✅ Built successfully with `go build`

## Testing

The application now:
- ✅ Compiles without errors
- ✅ No routing conflicts
- ✅ All unit tests pass
- ✅ Routes are properly structured

## Learn More

For more on Gin routing:
- [Gin Router Documentation](https://github.com/gin-gonic/gin#parameters-in-path)
- [Radix Tree Routing](https://en.wikipedia.org/wiki/Radix_tree) - How Gin's router works internally

---

**Status**: ✅ Fixed and Verified  
**Date**: October 2025

