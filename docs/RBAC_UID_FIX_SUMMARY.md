# RBAC UID Fix Summary

## 🐛 Issue Reported

**Problem**: When trying to assign roles to a user using encoded user ID (`"gGzTBURqhajG"`), getting error:

```json
{
  "code": 400,
  "status": "Bad Request",
  "message": "invalid user ID"
}
```

**Endpoint**: `POST /v1/rbac/user-roles/gGzTBURqhajG`

## ✅ Root Cause

The RBAC API handlers were using `strconv.Atoi()` to parse user IDs, expecting integers. However, **user IDs are encoded as base58 strings** for security, not plain integers.

## 🔧 What Was Fixed

### Files Modified:

1. **`microservice/user/transport/api/rbac_api.go`**
   - Fixed `AssignRolesToUserHdl()` - Decode user ID
   - Fixed `GetUserRolesHdl()` - Decode user ID
   - Fixed `RemoveRoleFromUserHdl()` - Decode user ID (role ID stays as integer)

2. **`microservice/user/transport/api/rbac_api_test.go`**
   - Updated tests to use encoded user IDs
   - Added `core` package import

3. **`docs/RBAC_UID_USAGE.md`** (NEW)
   - Complete guide on ID usage in RBAC system
   - Examples and common errors

### Code Changes:

**Before (Broken):**
```go
// ❌ Expects integer, fails with encoded ID
userID, err := strconv.Atoi(c.Param("userId"))
if err != nil {
    common.WriteErrorResponse(c, core.ErrBadRequest.WithError("invalid user ID"))
    return
}
```

**After (Fixed):**
```go
// ✅ Decodes base58 UID to integer
uid, err := core.FromBase58(c.Param("userId"))
if err != nil {
    common.WriteErrorResponse(c, core.ErrBadRequest.WithError("invalid user ID format"))
    return
}
userID := int(uid.GetLocalID())
```

## 📊 ID Types Summary

| Entity | ID Type | Example | Notes |
|--------|---------|---------|-------|
| **User** | ✅ Encoded String | `"gGzTBURqhajG"` | Base58 UID |
| **Role** | ❌ Plain Integer | `1`, `2`, `3` | Not encoded |
| **Permission** | ❌ Plain Integer | `1`, `2`, `3` | Not encoded |

## 🎯 How to Use RBAC Endpoints Now

### 1. Assign Roles to User

```bash
# ✅ CORRECT - Use encoded user ID
curl -X POST "http://localhost:8080/v1/rbac/user-roles/gGzTBURqhajG" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role_ids": [1, 2]}'

# ❌ WRONG - Using database ID
curl -X POST "http://localhost:8080/v1/rbac/user-roles/1" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"role_ids": [1, 2]}'
# Error: "invalid user ID format"
```

### 2. Get User's Roles

```bash
# ✅ CORRECT
curl -X GET "http://localhost:8080/v1/rbac/user-roles/gGzTBURqhajG" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 3. Remove Role from User

```bash
# ✅ CORRECT - Encoded user ID + plain role ID
curl -X DELETE "http://localhost:8080/v1/rbac/user-roles/gGzTBURqhajG/1" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Note**: 
- First parameter (`gGzTBURqhajG`) = Encoded user ID ✅
- Second parameter (`1`) = Plain role ID (integer) ✅

## 🔄 Complete Workflow

### Scenario: Assign admin role to a user

```bash
# Step 1: Get the user's encoded ID
curl -X GET "http://localhost:8080/v1/admin/users?search=john" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Response includes encoded ID:
{
  "data": {
    "data": [{
      "id": "gGzTBURqhajG",  ← Save this!
      "first_name": "John"
    }]
  }
}

# Step 2: Get available roles
curl -X GET "http://localhost:8080/v1/rbac/roles" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Response:
{
  "data": [
    {"id": 1, "code": "superadmin"},
    {"id": 2, "code": "admin"},  ← Use ID: 2
    {"id": 3, "code": "user"}
  ]
}

# Step 3: Assign admin role to user
curl -X POST "http://localhost:8080/v1/rbac/user-roles/gGzTBURqhajG" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role_ids": [2]}'

# ✅ Success!
```

## 🧪 Test Results

**All tests passing:**
```
✅ TestAssignRolesToUserAPI - PASS
✅ TestGetUserRolesAPI - PASS
✅ All 27 tests - PASS
```

**Build:**
```
✅ go build - Success
✅ No linter errors
```

## 📚 Documentation Created

1. **`docs/RBAC_UID_USAGE.md`** - Complete guide with examples
2. **`RBAC_UID_FIX_SUMMARY.md`** - This file
3. **`docs/UID_SYSTEM_EXPLANATION.md`** - UID system explained

## ⚠️ Important Notes

### For API Consumers:

1. **Always use encoded IDs for users**
   - Get from API responses (`/v1/admin/users`)
   - Treat as opaque strings
   - Don't try to decode client-side

2. **Use plain integers for roles/permissions**
   - Get from API responses (`/v1/rbac/roles`, `/v1/rbac/permissions`)
   - Can use directly in requests

### For Developers:

When creating new RBAC endpoints:
- ✅ Decode user IDs: `core.FromBase58()`
- ❌ Don't decode role/permission IDs: use `strconv.Atoi()`

## 🎉 Summary

**Problem**: RBAC endpoints failed with encoded user IDs

**Root Cause**: Handlers expected integer IDs, but users use encoded IDs

**Solution**: Updated handlers to decode user IDs properly

**Result**: 
- ✅ All RBAC endpoints work with encoded user IDs
- ✅ All tests passing
- ✅ Complete documentation provided

---

**Status**: ✅ **FIXED AND TESTED**

**Affected Endpoints**:
- POST `/v1/rbac/user-roles/:userId` ✅
- GET `/v1/rbac/user-roles/:userId` ✅  
- DELETE `/v1/rbac/user-roles/:userId/:roleId` ✅

**Documentation**: See `docs/RBAC_UID_USAGE.md` for complete usage guide.

