# Understanding the UID System - Quick Reference

## 🤔 Your Question

> "Why does the API return `"id": "gGzTBURqhajG"` when in my database `id = 1`?"

## ✅ Short Answer

**This is by design for security!** Your system encodes database IDs to prevent enumeration attacks and hide your database structure.

## 🔄 Simple Explanation

### What Happens:

```
Database stores:    id = 1
API returns:        "id": "gGzTBURqhajG"
```

### Why:

1. **Security**: Prevents attackers from guessing valid IDs
2. **Privacy**: Hides how many users/resources you have
3. **Flexibility**: Supports distributed systems and sharding

## 🎯 How to Use It

### For API Consumers (Frontend):

```javascript
// ✅ CORRECT Usage

// 1. Get user from API
const response = await fetch('/v1/admin/users');
const user = response.data;
console.log(user.id);  // "gGzTBURqhajG"

// 2. Use ID as-is (DON'T convert to number!)
await fetch(`/v1/admin/users/${user.id}`, {
    method: 'PATCH',
    body: JSON.stringify({ first_name: 'New Name' })
});

// ❌ WRONG Usage
const numericId = parseInt(user.id);  // DON'T DO THIS!
```

### For Backend Developers:

```go
// ✅ Returning data to client
user.Mask()  // Encode ID
c.JSON(200, core.ResponseData(user))

// ✅ Receiving ID from client
uid, err := core.FromBase58(c.Param("id"))  // Decode
userID := int(uid.GetLocalID())
```

## 📖 Full Documentation

- **Complete Guide**: [`docs/UID_SYSTEM_EXPLANATION.md`](docs/UID_SYSTEM_EXPLANATION.md)
- **Fix Details**: [`docs/UID_FIX_SUMMARY.md`](docs/UID_FIX_SUMMARY.md)
- **API Docs**: [`docs/ADMIN_USER_MANAGEMENT_API.md`](docs/ADMIN_USER_MANAGEMENT_API.md)

## 🔧 What Was Fixed

The handlers now properly **decode** the encoded IDs back to database IDs:

```go
// BEFORE (Broken) ❌
userID, err := strconv.Atoi(c.Param("id"))  // Fails with "gGzTBURqhajG"

// AFTER (Fixed) ✅
uid, err := core.FromBase58(c.Param("id"))  // Decodes "gGzTBURqhajG" → 1
userID := int(uid.GetLocalID())
```

## 🎉 Bottom Line

- ✅ Everything is working as designed
- ✅ System is more secure because of this
- ✅ Just use IDs as strings, don't worry about the encoding
- ✅ All documentation has been updated

**Need more details?** Read [`docs/UID_SYSTEM_EXPLANATION.md`](docs/UID_SYSTEM_EXPLANATION.md)

