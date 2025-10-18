# RBAC API Documentation

## Overview

The RBAC (Role-Based Access Control) API provides comprehensive management of roles, permissions, and user-role assignments. All RBAC endpoints require authentication and superadmin privileges.

## Base URL

```
/v1/rbac
```

## Authentication

All RBAC endpoints require:
1. Valid JWT token in the Authorization header: `Authorization: Bearer <token>`
2. Superadmin role (`system_role = 'sadmin'`)

If these requirements are not met, the API will return a `403 Forbidden` error.

## Default Superadmin

A default superadmin user is seeded in the database:
- **Email**: `superadmin@system.local`
- **Password**: `superadmin`
- **Role**: `sadmin`

⚠️ **Important**: Change this password in production!

## Endpoints

### Role Management

#### Create Role

Creates a new role in the system.

**Request:**
```http
POST /v1/rbac/roles
Content-Type: application/json
Authorization: Bearer <token>

{
  "code": "moderator",
  "name": "Moderator",
  "description": "User moderator role"
}
```

**Response:**
```json
{
  "data": {
    "id": 4,
    "code": "moderator",
    "name": "Moderator",
    "description": "User moderator role"
  }
}
```

**Status Codes:**
- `201 Created`: Role created successfully
- `400 Bad Request`: Invalid request body or duplicate role code
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: User does not have superadmin privileges
- `500 Internal Server Error`: Server error

---

#### List Roles

Retrieves all roles in the system.

**Request:**
```http
GET /v1/rbac/roles
Authorization: Bearer <token>
```

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "code": "sadmin",
      "name": "Super Admin",
      "description": ""
    },
    {
      "id": 2,
      "code": "admin",
      "name": "Admin",
      "description": ""
    },
    {
      "id": 3,
      "code": "user",
      "name": "User",
      "description": ""
    }
  ]
}
```

**Status Codes:**
- `200 OK`: Success
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: User does not have superadmin privileges
- `500 Internal Server Error`: Server error

---

#### Get Role with Permissions

Retrieves a specific role with its assigned permissions.

**Request:**
```http
GET /v1/rbac/roles/{roleId}
Authorization: Bearer <token>
```

**Path Parameters:**
- `roleId` (integer, required): The ID of the role

**Response:**
```json
{
  "data": {
    "id": 2,
    "code": "admin",
    "name": "Admin",
    "description": "",
    "permissions": [
      {
        "id": 1,
        "code": "chat.send",
        "name": "Chat: send message",
        "description": ""
      },
      {
        "id": 2,
        "code": "chat.read",
        "name": "Chat: read history",
        "description": ""
      }
    ]
  }
}
```

**Status Codes:**
- `200 OK`: Success
- `400 Bad Request`: Invalid role ID
- `404 Not Found`: Role not found
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: User does not have superadmin privileges
- `500 Internal Server Error`: Server error

---

#### Update Role

Updates an existing role's name and/or description.

**Request:**
```http
PATCH /v1/rbac/roles/{roleId}
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "Updated Admin Name",
  "description": "Updated description"
}
```

**Path Parameters:**
- `roleId` (integer, required): The ID of the role

**Response:**
```json
{
  "data": {
    "success": true
  }
}
```

**Status Codes:**
- `200 OK`: Role updated successfully
- `400 Bad Request`: Invalid request body or role ID
- `404 Not Found`: Role not found
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: User does not have superadmin privileges
- `500 Internal Server Error`: Server error

---

#### Delete Role

Deletes a role from the system. System roles (`sadmin`, `admin`, `user`) cannot be deleted.

**Request:**
```http
DELETE /v1/rbac/roles/{roleId}
Authorization: Bearer <token>
```

**Path Parameters:**
- `roleId` (integer, required): The ID of the role

**Response:**
```json
{
  "data": {
    "success": true
  }
}
```

**Status Codes:**
- `200 OK`: Role deleted successfully
- `400 Bad Request`: Cannot delete system roles
- `404 Not Found`: Role not found
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: User does not have superadmin privileges
- `500 Internal Server Error`: Server error

---

### Permission Management

#### Create Permission

Creates a new permission in the system.

**Request:**
```http
POST /v1/rbac/permissions
Content-Type: application/json
Authorization: Bearer <token>

{
  "code": "report.generate",
  "name": "Generate Reports",
  "description": "Permission to generate system reports"
}
```

**Response:**
```json
{
  "data": {
    "id": 15,
    "code": "report.generate",
    "name": "Generate Reports",
    "description": "Permission to generate system reports"
  }
}
```

**Status Codes:**
- `201 Created`: Permission created successfully
- `400 Bad Request`: Invalid request body or duplicate permission code
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: User does not have superadmin privileges
- `500 Internal Server Error`: Server error

---

#### List Permissions

Retrieves all permissions in the system.

**Request:**
```http
GET /v1/rbac/permissions
Authorization: Bearer <token>
```

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "code": "chat.send",
      "name": "Chat: send message",
      "description": ""
    },
    {
      "id": 2,
      "code": "chat.read",
      "name": "Chat: read history",
      "description": ""
    }
  ]
}
```

**Status Codes:**
- `200 OK`: Success
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: User does not have superadmin privileges
- `500 Internal Server Error`: Server error

---

### Role-Permission Assignment

#### Assign Permissions to Role

Assigns one or more permissions to a role.

**Request:**
```http
POST /v1/rbac/role-permissions/{id}
Content-Type: application/json
Authorization: Bearer <token>

{
  "permission_ids": [1, 2, 3]
}
```

**Path Parameters:**
- `id` (integer, required): The ID of the role

**Response:**
```json
{
  "data": {
    "success": true
  }
}
```

**Status Codes:**
- `200 OK`: Permissions assigned successfully
- `400 Bad Request`: Invalid request body, role not found, or permissions not found
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: User does not have superadmin privileges
- `500 Internal Server Error`: Server error

---

#### Remove Permission from Role

Removes a permission from a role.

**Request:**
```http
DELETE /v1/rbac/role-permissions/{id}/{permId}
Authorization: Bearer <token>
```

**Path Parameters:**
- `id` (integer, required): The ID of the role
- `permId` (integer, required): The ID of the permission

**Response:**
```json
{
  "data": {
    "success": true
  }
}
```

**Status Codes:**
- `200 OK`: Permission removed successfully
- `400 Bad Request`: Invalid role ID or permission ID
- `404 Not Found`: Role or permission not found
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: User does not have superadmin privileges
- `500 Internal Server Error`: Server error

---

### User-Role Assignment

#### Assign Roles to User

Assigns one or more roles to a user.

**Request:**
```http
POST /v1/rbac/user-roles/{userId}
Content-Type: application/json
Authorization: Bearer <token>

{
  "role_ids": [2, 3]
}
```

**Path Parameters:**
- `userId` (integer, required): The ID of the user

**Response:**
```json
{
  "data": {
    "success": true
  }
}
```

**Status Codes:**
- `200 OK`: Roles assigned successfully
- `400 Bad Request`: Invalid request body, user not found, or roles not found
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: User does not have superadmin privileges
- `500 Internal Server Error`: Server error

---

#### Get User Roles

Retrieves all roles assigned to a user.

**Request:**
```http
GET /v1/rbac/user-roles/{userId}
Authorization: Bearer <token>
```

**Path Parameters:**
- `userId` (integer, required): The ID of the user

**Response:**
```json
{
  "data": {
    "user_id": 1,
    "roles": [
      {
        "id": 1,
        "code": "sadmin",
        "name": "Super Admin",
        "description": ""
      }
    ]
  }
}
```

**Status Codes:**
- `200 OK`: Success
- `400 Bad Request`: Invalid user ID
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: User does not have superadmin privileges
- `500 Internal Server Error`: Server error

---

#### Remove Role from User

Removes a role from a user.

**Request:**
```http
DELETE /v1/rbac/user-roles/{userId}/{roleId}
Authorization: Bearer <token>
```

**Path Parameters:**
- `userId` (integer, required): The ID of the user
- `roleId` (integer, required): The ID of the role

**Response:**
```json
{
  "data": {
    "success": true
  }
}
```

**Status Codes:**
- `200 OK`: Role removed successfully
- `400 Bad Request`: Invalid user ID or role ID
- `404 Not Found`: Role not found
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: User does not have superadmin privileges
- `500 Internal Server Error`: Server error

---

## Default Roles

The system comes with three default roles:

| Code | Name | Description |
|------|------|-------------|
| `sadmin` | Super Admin | Full system access, can manage RBAC |
| `admin` | Admin | Administrative access to most features |
| `user` | User | Basic user access |

## Default Permissions

The system comes with pre-defined permissions:

### Chat Permissions
- `chat.send`: Send chat messages
- `chat.read`: Read chat history

### RBAC Permissions
- `rbac.role.create`: Create roles
- `rbac.role.read`: Read roles
- `rbac.role.update`: Update roles
- `rbac.role.delete`: Delete roles
- `rbac.permission.create`: Create permissions
- `rbac.permission.read`: Read permissions
- `rbac.permission.assign`: Assign permissions to roles
- `rbac.permission.revoke`: Revoke permissions from roles
- `rbac.user.assign`: Assign roles to users
- `rbac.user.read`: Read user roles

## Error Responses

All endpoints return errors in a consistent format:

```json
{
  "error": {
    "status_code": 403,
    "message": "superadmin access required",
    "log": "debug information (only in development)"
  }
}
```

Common error codes:
- `400 Bad Request`: Invalid input data
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: Insufficient privileges
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## Usage Examples

### Example 1: Login as Superadmin

```bash
curl -X POST http://localhost:8080/v1/authenticate \
  -H "Content-Type: application/json" \
  -d '{
    "email": "superadmin@system.local",
    "password": "superadmin"
  }'
```

### Example 2: Create a New Role

```bash
curl -X POST http://localhost:8080/v1/rbac/roles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{
    "code": "moderator",
    "name": "Moderator",
    "description": "Content moderator role"
  }'
```

### Example 3: Assign Permissions to Role

```bash
curl -X POST http://localhost:8080/v1/rbac/role-permissions/2 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{
    "permission_ids": [1, 2, 3, 4, 5, 6]
  }'
```

### Example 4: Assign Roles to User

```bash
curl -X POST http://localhost:8080/v1/rbac/user-roles/5 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{
    "role_ids": [2, 4]
  }'
```

## Security Considerations

1. **Change Default Password**: Always change the superadmin password in production
2. **Principle of Least Privilege**: Assign only necessary permissions to roles
3. **Audit Logs**: Consider implementing audit logging for RBAC changes
4. **Token Security**: Store JWT tokens securely, never expose them in logs or URLs
5. **HTTPS Only**: Always use HTTPS in production to protect tokens in transit
6. **Regular Review**: Periodically review role assignments and permissions

## Testing

Run the integration tests with:

```bash
go test ./test/integration_rbac_test.go -v
```

Run unit tests with:

```bash
go test ./microservice/user/business/... -v
go test ./microservice/user/transport/api/... -v
```

