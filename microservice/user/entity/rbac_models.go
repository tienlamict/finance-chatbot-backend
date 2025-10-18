package entity

import "time"

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	UserID    int       `gorm:"column:user_id;primaryKey"`
	RoleID    int       `gorm:"column:role_id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (UserRole) TableName() string { return "user_roles" }

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	RoleID       int       `gorm:"column:role_id;primaryKey"`
	PermissionID int       `gorm:"column:permission_id;primaryKey"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (RolePermission) TableName() string { return "role_permissions" }

// CreateRoleRequest represents a request to create a new role
type CreateRoleRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateRoleRequest represents a request to update a role
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreatePermissionRequest represents a request to create a new permission
type CreatePermissionRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// AssignPermissionsRequest represents a request to assign permissions to a role
type AssignPermissionsRequest struct {
	PermissionIDs []int `json:"permission_ids" binding:"required"`
}

// AssignRolesRequest represents a request to assign roles to a user
type AssignRolesRequest struct {
	RoleIDs []int `json:"role_ids" binding:"required"`
}

// RoleWithPermissions represents a role with its permissions
type RoleWithPermissions struct {
	Role
	Permissions []Permission `json:"permissions"`
}

// UserWithRoles represents a user with their assigned roles
type UserWithRoles struct {
	UserID int    `json:"user_id"`
	Roles  []Role `json:"roles"`
}
