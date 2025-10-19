# UID (Unique Identifier) System Explained

## 🤔 Why Does API Return `"id": "gGzTBURqhajG"` When DB Has `id = 1`?

This is **by design**! Your system uses a UID encoding system to hide real database IDs from API consumers.

## 📊 The System Architecture

### Real ID vs. Fake ID

```go
type SQLModel struct {
    Id     int   `json:"-" gorm:"column:id;" db:"id"`     // Hidden from JSON ❌
    FakeId *UID  `json:"id" gorm:"-"`                     // Exposed in JSON ✅
    CreatedAt *time.Time
    UpdatedAt *time.Time
}
```

**Key Points:**
- `Id` has `json:"-"` → **Never appears** in API responses
- `FakeId` has `json:"id"` → **Always appears** as "id" in API responses

## 🔄 How It Works

### Encoding Process (Output):

```
Database ID: 1
      ↓
Mask() function called
      ↓
UID Created: NewUID(1, objectType=1, shardID=1)
      ↓
Bit Composition: LocalID(32) | ObjectType(10) | ShardID(18)
      ↓
Base58 Encode
      ↓
API Response: "gGzTBURqhajG"
```

### Decoding Process (Input):

```
API Request: "gGzTBURqhajG"
      ↓
Base58 Decode
      ↓
Extract bits: LocalID | ObjectType | ShardID
      ↓
GetLocalID()
      ↓
Database Query: ID = 1
```

## 💻 Code Examples

### 1. Encoding (When Returning Data)

```go
// In your handler
func (api *adminUserAPI) GetUserByIDHdl() func(c *gin.Context) {
    return func(c *gin.Context) {
        // Get user from DB (ID = 1)
        user, err := api.business.GetUserByID(ctx, 1)
        
        // Mask the ID before returning
        user.Mask()  // Converts Id=1 to FakeId="gGzTBURqhajG"
        
        // Return to client
        c.JSON(http.StatusOK, core.ResponseData(user))
        // Response: {"data": {"id": "gGzTBURqhajG", ...}}
    }
}
```

### 2. Decoding (When Receiving Data)

```go
// In your handler
func (api *adminUserAPI) UpdateUserHdl() func(c *gin.Context) {
    return func(c *gin.Context) {
        // Client sends: PATCH /v1/admin/users/gGzTBURqhajG
        
        // Decode the UID to get real ID
        uid, err := core.FromBase58(c.Param("id"))
        if err != nil {
            // Invalid ID format
            return
        }
        
        userID := int(uid.GetLocalID())  // Gets 1
        
        // Use real ID for database query
        err = api.business.UpdateUser(ctx, userID, req)
    }
}
```

## 🎯 Benefits

### 1. **Security**
- Prevents enumeration attacks
- Hides database structure
- Makes it harder to guess valid IDs

**Example:**
```
Without UID: /api/users/1, /api/users/2, /api/users/3 (easy to enumerate)
With UID:    /api/users/gGzTBURqhajG (random-looking, hard to guess)
```

### 2. **Metadata Encoding**
The UID contains more than just the ID:
```
UID Components:
- Local ID (32 bits): The actual database ID
- Object Type (10 bits): What type of object (User, Role, etc.)
- Shard ID (18 bits): Which database shard/partition
```

### 3. **Distributed Systems**
Supports:
- Multiple database shards
- Multi-tenant architectures
- Microservices with separate databases

## 📝 Complete Example

### Create User Flow:

```bash
# 1. Create user
POST /v1/admin/users
{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com"
}

# Response (ID is encoded):
{
    "data": {
        "id": "gGzTBURqhajG",  ← Encoded ID
        "first_name": "John",
        "last_name": "Doe",
        "email": "john@example.com"
    }
}

# 2. Database stores:
INSERT INTO users (id, first_name, last_name, email) 
VALUES (1, 'John', 'Doe', 'john@example.com');
        ↑
     Real ID
```

### Update User Flow:

```bash
# 1. Client uses encoded ID from previous response
PATCH /v1/admin/users/gGzTBURqhajG
{
    "first_name": "Johnny"
}

# 2. Backend decodes:
"gGzTBURqhajG" → Base58 Decode → UID → GetLocalID() → 1

# 3. Database query:
UPDATE users SET first_name = 'Johnny' WHERE id = 1;
                                              ↑
                                          Real ID
```

## 🔍 Verifying the System

### Test Encoding/Decoding:

```go
package main

import (
    "fmt"
    "finance-chatbot/addon/core"
)

func main() {
    // Create UID from real ID
    realID := uint32(1)
    uid := core.NewUID(realID, 1, 1)
    
    // Encode to string
    encoded := uid.String()
    fmt.Printf("Real ID: %d\n", realID)
    fmt.Printf("Encoded: %s\n", encoded)  // "gGzTBURqhajG"
    
    // Decode back
    decoded, _ := core.FromBase58(encoded)
    fmt.Printf("Decoded: %d\n", decoded.GetLocalID())  // 1
}
```

**Output:**
```
Real ID: 1
Encoded: gGzTBURqhajG
Decoded: 1
```

## 🛠️ Implementation Details

### Where Masking Happens:

1. **Entity Level** (`microservice/user/entity/user.go`):
```go
func (u *User) Mask() {
    u.SQLModel.Mask(common.MaskTypeUser)
}
```

2. **SQLModel Level** (`addon/core/sql_model.go`):
```go
func (sqlModel *SQLModel) Mask(objectId int) {
    uid := NewUID(uint32(sqlModel.Id), objectId, 1)
    sqlModel.FakeId = &uid
}
```

3. **Handler Level** (`microservice/user/transport/api/admin_user_api.go`):
```go
// Before returning user
user.Mask()
c.JSON(http.StatusOK, core.ResponseData(user))
```

### Where Decoding Happens:

```go
// In handlers that receive ID parameter
uid, err := core.FromBase58(c.Param("id"))
if err != nil {
    // Handle invalid ID format
}
userID := int(uid.GetLocalID())
```

## ⚠️ Common Issues & Solutions

### Issue 1: "Invalid User ID" Error

**Problem:**
```bash
GET /v1/admin/users/1
# Error: invalid user ID format
```

**Solution:**
Use the encoded ID, not the database ID:
```bash
GET /v1/admin/users/gGzTBURqhajG
```

### Issue 2: Mixing Real and Encoded IDs

**Problem:**
```go
// Wrong - using real ID in URL
userID := 1
url := fmt.Sprintf("/api/users/%d", userID)  // /api/users/1
```

**Solution:**
```go
// Correct - use encoded ID
user.Mask()
url := fmt.Sprintf("/api/users/%s", user.FakeId.String())  // /api/users/gGzTBURqhajG
```

### Issue 3: Frontend Confusion

**Frontend developer sees:**
```json
{"id": "gGzTBURqhajG"}
```

**They should:**
1. **Store** the ID as-is (string)
2. **Use** it exactly as received in API calls
3. **Never** try to convert it to integer
4. **Never** try to decode it (backend handles this)

## 🔒 Security Considerations

### What UID System Prevents:

1. **Sequential ID Enumeration:**
   ```
   Bad:  /api/users/1, /api/users/2, /api/users/3
   Good: /api/users/gGzTBURqhajG (can't guess next ID)
   ```

2. **Resource Count Estimation:**
   - Can't estimate total users by looking at ID
   - Can't know registration order

3. **Database Structure Exposure:**
   - Real table structure is hidden
   - Shard distribution is obfuscated

### What UID System Does NOT Prevent:

1. **Unauthorized Access** (use proper auth/authz)
2. **SQL Injection** (use parameterized queries)
3. **Brute Force** (implement rate limiting)

## 📚 Reference

### Related Files:
- `addon/core/uid.go` - UID implementation
- `addon/core/sql_model.go` - SQLModel with masking
- `addon/common/const.go` - Object type constants

### Related Functions:
- `NewUID(localID, objectType, shardID)` - Create UID
- `FromBase58(encoded)` - Decode UID
- `GetLocalID()` - Extract real ID
- `Mask(objectId)` - Encode ID

### Constants (in `addon/common/const.go`):
```go
const (
    MaskTypeUser       = 1
    MaskTypeRole       = 2
    MaskTypePermission = 3
    // ... more types
)
```

## 🎓 Best Practices

1. **Always mask** before returning entities to clients
2. **Always decode** when receiving IDs from clients
3. **Never expose** real database IDs in APIs
4. **Use consistent** object type IDs across the system
5. **Handle errors** gracefully when decoding fails
6. **Document** the UID system for frontend developers

## 🤝 Working with Frontend

### Documentation for Frontend Developers:

```markdown
# API IDs

All IDs in our API are encoded strings (e.g., "gGzTBURqhajG"), not integers.

## Do's ✅
- Treat IDs as opaque strings
- Store them as strings
- Use them exactly as received
- Pass them to API calls as-is

## Don'ts ❌
- Don't convert to integer
- Don't try to decode
- Don't assume sequential order
- Don't try to generate IDs client-side
```

## 🎉 Summary

The `"id": "gGzTBURqhajG"` in your API response is an **intentional security and architecture feature**:

1. **Database stores**: `id = 1` (integer)
2. **API returns**: `"id": "gGzTBURqhajG"` (base58-encoded UID)
3. **Backend automatically**:
   - Encodes on output (Mask)
   - Decodes on input (FromBase58)
4. **Clients should**: Use IDs as opaque strings

This system provides **security through obscurity** while supporting **distributed architecture** and **multi-tenancy**.

---

**Questions?** Check the implementation in `addon/core/uid.go` or reach out to the development team!

