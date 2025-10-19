# Admin User Management API Documentation

## Overview

The Admin User Management API provides full CRUD operations for managing users in the system. These endpoints are restricted to administrators (users with `admin` or `superadmin` system roles) and require proper authentication and authorization.

## Authentication & Authorization

- **Authentication**: All endpoints require a valid JWT token in the Authorization header
  - Header format: `Authorization: Bearer <token>`
- **Authorization**: Only users with `superadmin` system role can access these endpoints
- Unauthorized access will return:
  - `401 Unauthorized` - Missing or invalid token
  - `403 Forbidden` - User doesn't have admin/superadmin role

## Base URL

```
/v1/admin/users
```

## Endpoints

### 1. Create User

Creates a new user account.

**Endpoint**: `POST /v1/admin/users`

**Request Body**:
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "phone": "+1234567890",
  "gender": "male",
  "system_role": "user",
  "status": "active"
}
```

**Request Fields**:
- `first_name` (required, string): User's first name (max 30 characters)
- `last_name` (required, string): User's last name (max 30 characters)
- `email` (required, string): Valid email address (must be unique)
- `phone` (optional, string): Phone number
- `gender` (optional, string): One of: `male`, `female`, `unknown` (default: `unknown`)
- `system_role` (optional, string): One of: `user`, `admin`, `sadmin` (default: `user`)
- `status` (optional, string): One of: `active`, `waiting_verify`, `banned` (default: `active`)

**Success Response** (201 Created):
```json
{
  "data": {
    "id": "encoded_user_id",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com",
    "phone": "+1234567890",
    "avatar": null,
    "gender": "male",
    "system_role": "user",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

**Error Responses**:
- `400 Bad Request`: Validation error or duplicate email
  ```json
  {
    "error": {
      "message": "email is not valid",
      "code": 400
    }
  }
  ```

### 2. Update User

Updates an existing user's information.

**Endpoint**: `PUT /v1/admin/users/:id` or `PATCH /v1/admin/users/:id`

**URL Parameters**:
- `id` (required, string): Encoded user ID (e.g., "gGzTBURqhajG") from previous API response

**Request Body** (all fields optional):
```json
{
  "first_name": "Jane",
  "last_name": "Smith",
  "email": "jane.smith@example.com",
  "phone": "+9876543210",
  "gender": "female",
  "system_role": "admin",
  "status": "banned"
}
```

**Request Fields** (all optional):
- `first_name` (string): User's first name (max 30 characters)
- `last_name` (string): User's last name (max 30 characters)
- `email` (string): Valid email address (must be unique)
- `phone` (string): Phone number
- `gender` (string): One of: `male`, `female`, `unknown`
- `system_role` (string): One of: `user`, `admin`, `sadmin`
- `status` (string): One of: `active`, `waiting_verify`, `banned`

**Success Response** (200 OK):
```json
{
  "data": true
}
```

**Error Responses**:
- `400 Bad Request`: Validation error
- `404 Not Found`: User not found
  ```json
  {
    "error": {
      "message": "user not found",
      "code": 404
    }
  }
  ```

### 3. Delete User

Soft deletes a user from the system.

**Endpoint**: `DELETE /v1/admin/users/:id`

**URL Parameters**:
- `id` (required, string): Encoded user ID (e.g., "gGzTBURqhajG")

**Success Response** (200 OK):
```json
{
  "data": true
}
```

**Error Responses**:
- `404 Not Found`: User not found

### 4. Get User by ID

Retrieves detailed information about a specific user.

**Endpoint**: `GET /v1/admin/users/:id`

**URL Parameters**:
- `id` (required, string): Encoded user ID (e.g., "gGzTBURqhajG")

**Success Response** (200 OK):
```json
{
  "data": {
    "id": "encoded_user_id",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com",
    "phone": "+1234567890",
    "avatar": null,
    "gender": "male",
    "system_role": "user",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

**Error Responses**:
- `400 Bad Request`: Invalid user ID format
- `404 Not Found`: User not found

### 5. List Users

Retrieves a paginated list of users with optional filters.

**Endpoint**: `GET /v1/admin/users`

**Query Parameters**:
- `page` (optional, integer): Page number (default: 1)
- `limit` (optional, integer): Items per page (default: 10, max: 200)
- `search` (optional, string): Search in first name, last name, or email
- `role` (optional, string): Filter by system role (`user`, `admin`, `sadmin`)
- `status` (optional, string): Filter by status (`active`, `waiting_verify`, `banned`)
- `email` (optional, string): Filter by exact email match

**Examples**:
```
GET /v1/admin/users?page=1&limit=20
GET /v1/admin/users?search=john
GET /v1/admin/users?role=admin&status=active
GET /v1/admin/users?email=john.doe@example.com
GET /v1/admin/users?search=john&role=user&page=1&limit=10
```

**Success Response** (200 OK):
```json
{
  "data": {
    "data": [
      {
        "id": "encoded_user_id_1",
        "first_name": "John",
        "last_name": "Doe",
        "email": "john.doe@example.com",
        "phone": "+1234567890",
        "avatar": null,
        "gender": "male",
        "system_role": "user",
        "status": "active",
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      },
      {
        "id": "encoded_user_id_2",
        "first_name": "Jane",
        "last_name": "Smith",
        "email": "jane.smith@example.com",
        "phone": "+9876543210",
        "avatar": null,
        "gender": "female",
        "system_role": "admin",
        "status": "active",
        "created_at": "2024-01-15T11:00:00Z",
        "updated_at": "2024-01-15T11:00:00Z"
      }
    ],
    "paging": {
      "page": 1,
      "limit": 10,
      "total": 2
    }
  }
}
```

**Error Responses**:
- `400 Bad Request`: Invalid query parameters

## Data Types

### SystemRole
- `user`: Regular user
- `admin`: Administrator
- `sadmin`: Super administrator

### Status
- `active`: User account is active
- `waiting_verify`: User account is pending verification
- `banned`: User account is banned

### Gender
- `male`: Male
- `female`: Female
- `unknown`: Unknown/Prefer not to say

## Error Handling

All error responses follow this format:

```json
{
  "error": {
    "message": "Error description",
    "code": 400,
    "debug": "Internal error details (only in development)"
  }
}
```

Common HTTP status codes:
- `200 OK`: Successful request
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## Examples

### Example 1: Creating a New Admin User

```bash
curl -X POST "http://localhost:8080/v1/admin/users" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Admin",
    "last_name": "User",
    "email": "admin@example.com",
    "system_role": "admin",
    "status": "active"
  }'
```

### Example 2: Searching for Users

```bash
curl -X GET "http://localhost:8080/v1/admin/users?search=john&page=1&limit=10" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

### Example 3: Updating User Status

```bash
# Note: Use the encoded ID from the API response, not the database ID
curl -X PATCH "http://localhost:8080/v1/admin/users/gGzTBURqhajG" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "banned"
  }'
```

### Example 4: Listing Active Admin Users

```bash
curl -X GET "http://localhost:8080/v1/admin/users?role=admin&status=active" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

### Example 5: Deleting a User

```bash
# Note: Use the encoded ID from the API response, not the database ID
curl -X DELETE "http://localhost:8080/v1/admin/users/gGzTBURqhajG" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

## Best Practices

1. **Use Encoded IDs**: Always use the encoded ID (e.g., "gGzTBURqhajG") returned by the API, never use database IDs directly
2. **Email Uniqueness**: Always check for duplicate emails before creating users
3. **Soft Delete**: Users are soft-deleted, maintaining data integrity and audit trails
4. **Pagination**: Always use pagination for list endpoints to avoid performance issues
5. **Filtering**: Combine multiple filters for more precise queries
6. **Role Management**: Be cautious when changing user roles, especially to `sadmin`
7. **Status Management**: Use `banned` status instead of deletion for temporarily restricting users

## Architecture

This feature follows Clean Architecture principles:

- **Entity Layer**: Domain models and validation logic
  - `microservice/user/entity/admin_user_models.go`
  
- **Repository Layer**: Data access and persistence
  - `microservice/user/repository/mysql/admin_user_store.go`
  
- **Business Layer**: Business logic and rules
  - `microservice/user/business/admin_user_business.go`
  
- **Transport Layer**: HTTP handlers
  - `microservice/user/transport/api/admin_user_api.go`

## Testing

Unit tests and integration tests are available:

- **Unit Tests**: `microservice/user/business/admin_user_business_test.go`
- **Integration Tests**: `test/integration_admin_user_test.go`

Run tests with:
```bash
go test ./microservice/user/business/... -v
go test ./test/... -v
```

## Important Notes

### ID Encoding

All user IDs in API responses are **encoded strings** (not integers) for security purposes:

```json
{
  "id": "gGzTBURqhajG"  // ← This is an encoded ID representing database ID = 1
}
```

**Key Points:**
- IDs are base58-encoded UIDs containing metadata
- Always use the ID exactly as returned by the API
- Never try to decode or manipulate IDs client-side
- IDs are NOT sequential or predictable (security feature)

**For more details**, see `docs/UID_SYSTEM_EXPLANATION.md`

## Future Enhancements

Potential improvements for future versions:

1. Bulk user operations (create, update, delete multiple users)
2. User import/export functionality (CSV, Excel)
3. User activity logs and audit trails
4. Advanced filtering options (created date range, last login, etc.)
5. User profile image upload support
6. Email notification on user creation
7. Password reset functionality for admin
8. Two-factor authentication management

