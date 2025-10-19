# Admin User Management - Quick Start Guide

This guide will help you quickly get started with the Admin User Management API.

## Prerequisites

1. Backend server running
2. Superadmin user account
3. Valid JWT authentication token

## Getting Your Admin Token

First, you need to authenticate as a superadmin user:

```bash
# Login as superadmin
curl -X POST "http://localhost:8080/v1/authenticate" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "your_password"
  }'
```

Response:
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer"
  }
}
```

Save the `access_token` for subsequent requests.

## Common Operations

### 1. Create a New User

```bash
curl -X POST "http://localhost:8080/v1/admin/users" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Alice",
    "last_name": "Johnson",
    "email": "alice.johnson@example.com",
    "phone": "+1234567890",
    "gender": "female",
    "system_role": "user",
    "status": "active"
  }'
```

**Success Response (201 Created):**
```json
{
  "data": {
    "id": "encoded_user_id",
    "first_name": "Alice",
    "last_name": "Johnson",
    "email": "alice.johnson@example.com",
    "phone": "+1234567890",
    "avatar": null,
    "gender": "female",
    "system_role": "user",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### 2. List All Users (Paginated)

```bash
curl -X GET "http://localhost:8080/v1/admin/users?page=1&limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**Success Response (200 OK):**
```json
{
  "data": {
    "data": [
      {
        "id": "encoded_user_id_1",
        "first_name": "Alice",
        "last_name": "Johnson",
        "email": "alice.johnson@example.com",
        "system_role": "user",
        "status": "active",
        "created_at": "2024-01-15T10:30:00Z"
      }
    ],
    "paging": {
      "page": 1,
      "limit": 10,
      "total": 1
    }
  }
}
```

### 3. Search for Users

Search in first name, last name, or email:

```bash
curl -X GET "http://localhost:8080/v1/admin/users?search=alice" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 4. Filter Users by Role

Get all admin users:

```bash
curl -X GET "http://localhost:8080/v1/admin/users?role=admin" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

Available roles: `user`, `admin`, `sadmin`

### 5. Filter Users by Status

Get all active users:

```bash
curl -X GET "http://localhost:8080/v1/admin/users?status=active" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

Available statuses: `active`, `waiting_verify`, `banned`

### 6. Combine Multiple Filters

Get active admin users with search:

```bash
curl -X GET "http://localhost:8080/v1/admin/users?search=alice&role=admin&status=active&page=1&limit=20" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 7. Get User by ID

```bash
curl -X GET "http://localhost:8080/v1/admin/users/123" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**Success Response (200 OK):**
```json
{
  "data": {
    "id": "encoded_user_id",
    "first_name": "Alice",
    "last_name": "Johnson",
    "email": "alice.johnson@example.com",
    "phone": "+1234567890",
    "avatar": null,
    "gender": "female",
    "system_role": "user",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### 8. Update User Information

Update specific fields (all fields are optional):

```bash
curl -X PATCH "http://localhost:8080/v1/admin/users/123" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Alicia",
    "phone": "+9876543210"
  }'
```

**Success Response (200 OK):**
```json
{
  "data": true
}
```

### 9. Change User Role

Promote user to admin:

```bash
curl -X PATCH "http://localhost:8080/v1/admin/users/123" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "system_role": "admin"
  }'
```

### 10. Ban/Unban User

Ban a user:

```bash
curl -X PATCH "http://localhost:8080/v1/admin/users/123" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "banned"
  }'
```

Unban a user:

```bash
curl -X PATCH "http://localhost:8080/v1/admin/users/123" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "active"
  }'
```

### 11. Delete User

```bash
curl -X DELETE "http://localhost:8080/v1/admin/users/123" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**Success Response (200 OK):**
```json
{
  "data": true
}
```

**Note:** This is a soft delete. The user record remains in the database but is marked as deleted.

## Common Error Responses

### Unauthorized (401)

Missing or invalid token:

```json
{
  "error": {
    "message": "missing access token",
    "code": 401
  }
}
```

**Fix:** Add valid `Authorization: Bearer <token>` header

### Forbidden (403)

User doesn't have admin/superadmin role:

```json
{
  "error": {
    "message": "superadmin access required",
    "code": 403
  }
}
```

**Fix:** Use account with superadmin role

### Bad Request (400)

Validation error:

```json
{
  "error": {
    "message": "email is not valid",
    "code": 400
  }
}
```

Duplicate email:

```json
{
  "error": {
    "message": "user with this email already exists",
    "code": 400
  }
}
```

**Fix:** Check input data and correct errors

### Not Found (404)

User doesn't exist:

```json
{
  "error": {
    "message": "user not found",
    "code": 404
  }
}
```

**Fix:** Verify user ID is correct

## Using Postman

### 1. Create a Collection

1. Open Postman
2. Create new collection: "Admin User Management"
3. Add collection variable: `baseUrl` = `http://localhost:8080/v1`
4. Add collection variable: `token` = (your JWT token)

### 2. Set Authorization

In collection settings:
- Type: Bearer Token
- Token: `{{token}}`

This will automatically add the token to all requests.

### 3. Create Requests

**Create User:**
- Method: POST
- URL: `{{baseUrl}}/admin/users`
- Body: Raw JSON

**List Users:**
- Method: GET
- URL: `{{baseUrl}}/admin/users?page=1&limit=10`

**Get User:**
- Method: GET
- URL: `{{baseUrl}}/admin/users/:id`

**Update User:**
- Method: PATCH
- URL: `{{baseUrl}}/admin/users/:id`
- Body: Raw JSON

**Delete User:**
- Method: DELETE
- URL: `{{baseUrl}}/admin/users/:id`

## Using JavaScript/Fetch

```javascript
// Base configuration
const baseURL = 'http://localhost:8080/v1';
const token = 'YOUR_JWT_TOKEN';

const headers = {
  'Authorization': `Bearer ${token}`,
  'Content-Type': 'application/json'
};

// Create user
async function createUser(userData) {
  const response = await fetch(`${baseURL}/admin/users`, {
    method: 'POST',
    headers,
    body: JSON.stringify(userData)
  });
  return response.json();
}

// List users
async function listUsers(page = 1, limit = 10, filters = {}) {
  const params = new URLSearchParams({ page, limit, ...filters });
  const response = await fetch(`${baseURL}/admin/users?${params}`, {
    headers
  });
  return response.json();
}

// Update user
async function updateUser(userId, updates) {
  const response = await fetch(`${baseURL}/admin/users/${userId}`, {
    method: 'PATCH',
    headers,
    body: JSON.stringify(updates)
  });
  return response.json();
}

// Delete user
async function deleteUser(userId) {
  const response = await fetch(`${baseURL}/admin/users/${userId}`, {
    method: 'DELETE',
    headers
  });
  return response.json();
}

// Example usage
const newUser = await createUser({
  first_name: 'Bob',
  last_name: 'Smith',
  email: 'bob.smith@example.com',
  system_role: 'user',
  status: 'active'
});

const users = await listUsers(1, 10, { search: 'bob' });
```

## Tips & Best Practices

1. **Always validate input** before sending requests
2. **Handle errors gracefully** in your client application
3. **Use pagination** for large user lists to avoid performance issues
4. **Combine filters** for more precise queries
5. **Cache tokens** but refresh them when they expire
6. **Log admin actions** for audit purposes
7. **Use HTTPS** in production
8. **Implement rate limiting** for admin endpoints
9. **Back up data** before bulk operations
10. **Test on staging** before production deployment

## Troubleshooting

### Token Expired

If you get 401 errors after some time, your token has expired. Login again to get a new token.

### CORS Issues

If making requests from browser, ensure CORS is configured correctly on the backend.

### Connection Refused

Verify the backend server is running and the port is correct.

### Database Errors

Check database connection and ensure migrations are up to date.

## Next Steps

- Read the full [API Documentation](ADMIN_USER_MANAGEMENT_API.md)
- Learn about [RBAC System](RBAC_API.md)
- Review [Implementation Details](ADMIN_USER_MANAGEMENT_IMPLEMENTATION.md)

## Support

For issues or questions:
1. Check the API documentation
2. Review error messages carefully
3. Check backend logs
4. Consult the implementation guide

Happy coding! 🚀

