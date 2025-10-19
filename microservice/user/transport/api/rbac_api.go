package api

import (
	"context"
	"finance-chatbot/addon/common"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/user/entity"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RBACBusiness interface {
	// Role management
	CreateRole(ctx context.Context, req *entity.CreateRoleRequest) (*entity.Role, error)
	ListRoles(ctx context.Context) ([]entity.Role, error)
	GetRoleWithPermissions(ctx context.Context, roleID int) (*entity.RoleWithPermissions, error)
	UpdateRole(ctx context.Context, roleID int, req *entity.UpdateRoleRequest) error
	DeleteRole(ctx context.Context, roleID int) error

	// Permission management
	CreatePermission(ctx context.Context, req *entity.CreatePermissionRequest) (*entity.Permission, error)
	ListPermissions(ctx context.Context) ([]entity.Permission, error)

	// Role-Permission assignment
	AssignPermissionsToRole(ctx context.Context, roleID int, req *entity.AssignPermissionsRequest) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error

	// User-Role assignment
	AssignRolesToUser(ctx context.Context, userID int, req *entity.AssignRolesRequest) error
	GetUserRoles(ctx context.Context, userID int) (*entity.UserWithRoles, error)
	RemoveRoleFromUser(ctx context.Context, userID, roleID int) error
}

type rbacAPI struct {
	business RBACBusiness
}

func NewRBACAPI(business RBACBusiness) *rbacAPI {
	return &rbacAPI{business: business}
}

// ========== Role management handlers ==========

// CreateRoleHdl creates a new role
// POST /api/v1/rbac/roles
func (api *rbacAPI) CreateRoleHdl() func(c *gin.Context) {
	return func(c *gin.Context) {
		var req entity.CreateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError(err.Error()))
			return
		}

		role, err := api.business.CreateRole(c.Request.Context(), &req)
		if err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusCreated, core.ResponseData(role))
	}
}

// ListRolesHdl lists all roles
// GET /api/v1/rbac/roles
func (api *rbacAPI) ListRolesHdl() func(c *gin.Context) {
	return func(c *gin.Context) {
		roles, err := api.business.ListRoles(c.Request.Context())
		if err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, core.ResponseData(roles))
	}
}

// GetRoleWithPermissionsHdl gets a role with its permissions
// GET /api/v1/rbac/roles/:id
func (api *rbacAPI) GetRoleWithPermissionsHdl() func(c *gin.Context) {
	return func(c *gin.Context) {
		roleID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError("invalid role ID"))
			return
		}

		role, err := api.business.GetRoleWithPermissions(c.Request.Context(), roleID)
		if err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, core.ResponseData(role))
	}
}

// UpdateRoleHdl updates a role
// PATCH /api/v1/rbac/roles/:id
func (api *rbacAPI) UpdateRoleHdl() func(c *gin.Context) {
	return func(c *gin.Context) {
		roleID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError("invalid role ID"))
			return
		}

		var req entity.UpdateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError(err.Error()))
			return
		}

		if err := api.business.UpdateRole(c.Request.Context(), roleID, &req); err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, core.SimpleSuccessResponse(true))
	}
}

// DeleteRoleHdl deletes a role
// DELETE /api/v1/rbac/roles/:id
func (api *rbacAPI) DeleteRoleHdl() func(c *gin.Context) {
	return func(c *gin.Context) {
		roleID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError("invalid role ID"))
			return
		}

		if err := api.business.DeleteRole(c.Request.Context(), roleID); err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, core.SimpleSuccessResponse(true))
	}
}

// ========== Permission management handlers ==========

// CreatePermissionHdl creates a new permission
// POST /api/v1/rbac/permissions
func (api *rbacAPI) CreatePermissionHdl() func(c *gin.Context) {
	return func(c *gin.Context) {
		var req entity.CreatePermissionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError(err.Error()))
			return
		}

		perm, err := api.business.CreatePermission(c.Request.Context(), &req)
		if err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusCreated, core.ResponseData(perm))
	}
}

// ListPermissionsHdl lists all permissions
// GET /api/v1/rbac/permissions
func (api *rbacAPI) ListPermissionsHdl() func(c *gin.Context) {
	return func(c *gin.Context) {
		perms, err := api.business.ListPermissions(c.Request.Context())
		if err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, core.ResponseData(perms))
	}
}

// ========== Role-Permission assignment handlers ==========

// AssignPermissionsToRoleHdl assigns permissions to a role
// POST /api/v1/rbac/roles/:id/permissions
func (api *rbacAPI) AssignPermissionsToRoleHdl() func(c *gin.Context) {
	return func(c *gin.Context) {
		roleID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError("invalid role ID"))
			return
		}

		var req entity.AssignPermissionsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError(err.Error()))
			return
		}

		if err := api.business.AssignPermissionsToRole(c.Request.Context(), roleID, &req); err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, core.SimpleSuccessResponse(true))
	}
}

// RemovePermissionFromRoleHdl removes a permission from a role
// DELETE /api/v1/rbac/roles/:roleId/permissions/:permId
func (api *rbacAPI) RemovePermissionFromRoleHdl() func(c *gin.Context) {
	return func(c *gin.Context) {
		roleID, err := strconv.Atoi(c.Param("roleId"))
		if err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError("invalid role ID"))
			return
		}

		permID, err := strconv.Atoi(c.Param("permId"))
		if err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError("invalid permission ID"))
			return
		}

		if err := api.business.RemovePermissionFromRole(c.Request.Context(), roleID, permID); err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, core.SimpleSuccessResponse(true))
	}
}

// ========== User-Role assignment handlers ==========

// AssignRolesToUserHdl assigns roles to a user
// POST /api/v1/rbac/users/:userId/roles
func (api *rbacAPI) AssignRolesToUserHdl() func(c *gin.Context) {
	return func(c *gin.Context) {
		// Decode base58 UID to real user ID
		uid, err := core.FromBase58(c.Param("userId"))
		if err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError("invalid user ID format"))
			return
		}
		userID := int(uid.GetLocalID())

		var req entity.AssignRolesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError(err.Error()))
			return
		}

		if err := api.business.AssignRolesToUser(c.Request.Context(), userID, &req); err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, core.SimpleSuccessResponse(true))
	}
}

// GetUserRolesHdl gets all roles assigned to a user
// GET /api/v1/rbac/users/:userId/roles
func (api *rbacAPI) GetUserRolesHdl() func(c *gin.Context) {
	return func(c *gin.Context) {
		// Decode base58 UID to real user ID
		uid, err := core.FromBase58(c.Param("userId"))
		if err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError("invalid user ID format"))
			return
		}
		userID := int(uid.GetLocalID())

		userRoles, err := api.business.GetUserRoles(c.Request.Context(), userID)
		if err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, core.ResponseData(userRoles))
	}
}

// RemoveRoleFromUserHdl removes a role from a user
// DELETE /api/v1/rbac/users/:userId/roles/:roleId
func (api *rbacAPI) RemoveRoleFromUserHdl() func(c *gin.Context) {
	return func(c *gin.Context) {
		// Decode base58 UID to real user ID
		uid, err := core.FromBase58(c.Param("userId"))
		if err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError("invalid user ID format"))
			return
		}
		userID := int(uid.GetLocalID())

		// Role IDs are plain integers, not encoded
		roleID, err := strconv.Atoi(c.Param("roleId"))
		if err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError("invalid role ID"))
			return
		}

		if err := api.business.RemoveRoleFromUser(c.Request.Context(), userID, roleID); err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, core.SimpleSuccessResponse(true))
	}
}
