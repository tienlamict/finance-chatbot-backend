# RBAC System - ID Usage Guide

## 🔑 Understanding ID Types in RBAC

When working with the RBAC system, **different entities use different ID formats**:

| Entity | ID Type | Example | Format |
|--------|---------|---------|--------|
| **User** | ✅ Encoded String | `"gGzTBURqhajG"` | Base58 UID |
| **Role** | ❌ Plain Integer | `1`, `2`, `3` | Integer |
| **Permission** | ❌ Plain Integer | `1`, `2`, `3` | Integer |

## ⚠️ Critical: User IDs Are Encoded!

### Why This Matters:

When assigning roles to users, you **MUST use the encoded user ID**, not the database ID:

```bash
# ❌ WRONG - Using database ID
POST /v1/rbac/user-roles/1
# Error: "invalid user ID format"

# ✅ CORRECT - Using encoded ID from API
POST /v1/rbac/user-roles/gGzTBURqhajG
# Success!
```

## 📝 How to Get the Correct User ID

### Method 1: From User List API

```bash
# Get users list
curl -X GET "http://localhost:8080/v1/admin/users" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Response includes encoded IDs:
{
  "data": {
    "data": [
      {
        "id": "gGzTBURqhajG",  ← Use this ID!
        "first_name": "John",
        "email": "john@example.com"
      }
    ]
  }
}
```

### Method 2: From Create User Response

```bash
# Create a user
curl -X POST "http://localhost:8080/v1/admin/users" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"first_name":"Jane","last_name":"Doe","email":"jane@example.com"}'

# Response includes encoded ID:
{
  "data": {
    "id": "kNpRx3mQvWc7",  ← Use this ID!
    "first_name": "Jane",
    "email": "jane@example.com"
  }
}
```

## 🎯 RBAC Endpoints - ID Usage

### User-Role Assignments (Use Encoded User IDs)

#### 1. Assign Roles to User
```bash
# Use encoded user ID in URL
POST /v1/rbac/user-roles/{encodedUserId}
```

**Example:**
```bash
curl -X POST "http://localhost:8080/v1/rbac/user-roles/gGzTBURqhajG" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role_ids": [1, 2]
  }'
```

**Note**: 
- `gGzTBURqhajG` = Encoded **user ID** ✅
- `[1, 2]` = Plain **role IDs** (integers) ✅

#### 2. Get User's Roles
```bash
# Use encoded user ID in URL
GET /v1/rbac/user-roles/{encodedUserId}
```

**Example:**
```bash
curl -X GET "http://localhost:8080/v1/rbac/user-roles/gGzTBURqhajG" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 3. Remove Role from User
```bash
# Use encoded user ID and plain role ID
DELETE /v1/rbac/user-roles/{encodedUserId}/{roleId}
```

**Example:**
```bash
curl -X DELETE "http://localhost:8080/v1/rbac/user-roles/gGzTBURqhajG/1" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Note**:
- `gGzTBURqhajG` = Encoded **user ID** ✅
- `1` = Plain **role ID** (integer) ✅

### Role Management (Use Plain Integer IDs)

All role-related endpoints use **plain integer IDs**:

```bash
# Create role
POST /v1/rbac/roles

# Get all roles
GET /v1/rbac/roles

# Get specific role
GET /v1/rbac/roles/1  ← Plain integer ID

# Update role
PATCH /v1/rbac/roles/1  ← Plain integer ID

# Delete role
DELETE /v1/rbac/roles/1  ← Plain integer ID

# Assign permissions to role
POST /v1/rbac/role-permissions/1  ← Plain integer ID

# Remove permission from role
DELETE /v1/rbac/role-permissions/1/2  ← Both plain integers
```

### Permission Management (Use Plain Integer IDs)

All permission-related endpoints use **plain integer IDs**:

```bash
# Create permission
POST /v1/rbac/permissions

# Get all permissions
GET /v1/rbac/permissions
```

## 🔄 Complete Workflow Example

### Scenario: Create a user and assign them admin role

```bash
# Step 1: Authenticate as superadmin
curl -X POST "http://localhost:8080/v1/authenticate" \
  -H "Content-Type: application/json" \
  -d '{"email":"superadmin@system.local","password":"superadmin"}'

# Save the token from response
TOKEN="eyJhbGciOiJIUzI1NiIs..."

# Step 2: Create a new user
curl -X POST "http://localhost:8080/v1/admin/users" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Alice",
    "last_name": "Admin",
    "email": "alice@company.com",
    "system_role": "user",
    "status": "active"
  }'

# Response includes encoded user ID:
{
  "data": {
    "id": "mXtY9pLqR2vN",  ← SAVE THIS!
    "first_name": "Alice",
    ...
  }
}

# Step 3: Get list of roles to find admin role ID
curl -X GET "http://localhost:8080/v1/rbac/roles" \
  -H "Authorization: Bearer $TOKEN"

# Response:
{
  "data": [
    {"id": 1, "code": "superadmin", "name": "Super Administrator"},
    {"id": 2, "code": "admin", "name": "Administrator"},  ← Use ID: 2
    {"id": 3, "code": "user", "name": "Regular User"}
  ]
}

# Step 4: Assign admin role to the user
curl -X POST "http://localhost:8080/v1/rbac/user-roles/mXtY9pLqR2vN" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role_ids": [2]
  }'

# ✅ Success! Alice now has admin role

# Step 5: Verify the assignment
curl -X GET "http://localhost:8080/v1/rbac/user-roles/mXtY9pLqR2vN" \
  -H "Authorization: Bearer $TOKEN"

# Response:
{
  "data": {
    "user_id": 1,
    "roles": [
      {"id": 2, "code": "admin", "name": "Administrator"}
    ]
  }
}
```

## ❌ Common Errors

### Error 1: "invalid user ID format"

```json
{
  "error": {
    "message": "invalid user ID format",
    "code": 400
  }
}
```

**Cause**: Using database ID instead of encoded ID

**Fix**:
```bash
# ❌ Wrong
POST /v1/rbac/user-roles/1

# ✅ Correct
POST /v1/rbac/user-roles/gGzTBURqhajG
```

### Error 2: "invalid role ID"

```json
{
  "error": {
    "message": "invalid role ID",
    "code": 400
  }
}
```

**Cause**: Using encoded ID for role (roles use plain integers)

**Fix**:
```bash
# ❌ Wrong
DELETE /v1/rbac/user-roles/gGzTBURqhajG/gGzTBURqhajG

# ✅ Correct
DELETE /v1/rbac/user-roles/gGzTBURqhajG/2
```

## 🎓 Best Practices

1. **Always Get IDs from API Responses**
   - Don't hardcode IDs
   - Don't use database IDs directly
   - Store IDs from API responses

2. **Remember the Format**
   - Users: Encoded strings
   - Roles: Plain integers
   - Permissions: Plain integers

3. **Store Both If Needed**
   ```javascript
   // In your frontend
   const user = {
     id: "gGzTBURqhajG",  // Use for API calls
     displayName: "John Doe",
     email: "john@example.com"
   };
   
   const role = {
     id: 2,  // Use for API calls (integer)
     name: "Administrator",
     code: "admin"
   };
   ```

4. **Test with Postman/curl**
   - Always test endpoints with actual encoded IDs
   - Save example IDs for testing
   - Verify responses

## 📚 Reference

### Quick ID Format Check:

```javascript
// Pseudo-code for validation
function isValidUserId(id) {
  return typeof id === 'string' && id.length > 10;  // Encoded
}

function isValidRoleId(id) {
  return typeof id === 'number' && id > 0;  // Integer
}

function isValidPermissionId(id) {
  return typeof id === 'number' && id > 0;  // Integer
}
```

### Related Documentation:

- **UID System**: `docs/UID_SYSTEM_EXPLANATION.md`
- **RBAC API**: `docs/RBAC_API.md`
- **Admin User Management**: `docs/ADMIN_USER_MANAGEMENT_API.md`

## 🎉 Summary

**Remember**:
- 👤 **User IDs** → Encoded strings (`"gGzTBURqhajG"`)
- 🎭 **Role IDs** → Plain integers (`1`, `2`, `3`)
- 🔑 **Permission IDs** → Plain integers (`1`, `2`, `3`)

**Always use the ID format from API responses!**

---

**Need Help?** Check `docs/UID_SYSTEM_EXPLANATION.md` for more details on the UID encoding system.

