package business

import (
	"context"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/user/entity"
	"strings"
)

type RBACStore interface {
	// Role operations
	CreateRole(ctx context.Context, role *entity.Role) error
	GetRoleByID(ctx context.Context, roleID int) (*entity.Role, error)
	GetRoleByCode(ctx context.Context, code string) (*entity.Role, error)
	ListRoles(ctx context.Context) ([]entity.Role, error)
	UpdateRole(ctx context.Context, roleID int, name, description string) error
	DeleteRole(ctx context.Context, roleID int) error

	// Permission operations
	CreatePermission(ctx context.Context, perm *entity.Permission) error
	GetPermissionByID(ctx context.Context, permID int) (*entity.Permission, error)
	GetPermissionByCode(ctx context.Context, code string) (*entity.Permission, error)
	ListPermissions(ctx context.Context) ([]entity.Permission, error)

	// Role-Permission operations
	GetRolePermissions(ctx context.Context, roleID int) ([]entity.Permission, error)
	AssignPermissionsToRole(ctx context.Context, roleID int, permissionIDs []int) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error

	// User-Role operations
	GetUserRoles(ctx context.Context, userID int) ([]entity.Role, error)
	AssignRolesToUser(ctx context.Context, userID int, roleIDs []int) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID int) error

	// Permission checking
	GetUserPermissionCodes(ctx context.Context, userID int) ([]string, error)
	HasAnyPermission(ctx context.Context, userID int, permCodes ...string) (bool, error)
}

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

	// Permission checking
	HasAnyPermission(ctx context.Context, userID int, permCodes ...string) (bool, error)
}

type rbacBusiness struct {
	store RBACStore
}

func NewRBACBusiness(store RBACStore) RBACBusiness {
	return &rbacBusiness{store: store}
}

// ========== Role management ==========

func (biz *rbacBusiness) CreateRole(ctx context.Context, req *entity.CreateRoleRequest) (*entity.Role, error) {
	// Validate input
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)

	if req.Code == "" || req.Name == "" {
		return nil, core.ErrBadRequest.WithError("code and name are required")
	}

	// Check if role code already exists
	existing, err := biz.store.GetRoleByCode(ctx, req.Code)
	if err != nil && err != core.ErrRecordNotFound {
		return nil, core.ErrInternalServerError.WithDebug(err.Error())
	}
	if existing != nil {
		return nil, core.ErrBadRequest.WithError("role code already exists")
	}

	role := &entity.Role{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := biz.store.CreateRole(ctx, role); err != nil {
		return nil, core.ErrInternalServerError.WithDebug(err.Error())
	}

	return role, nil
}

func (biz *rbacBusiness) ListRoles(ctx context.Context) ([]entity.Role, error) {
	roles, err := biz.store.ListRoles(ctx)
	if err != nil {
		return nil, core.ErrInternalServerError.WithDebug(err.Error())
	}
	return roles, nil
}

func (biz *rbacBusiness) GetRoleWithPermissions(ctx context.Context, roleID int) (*entity.RoleWithPermissions, error) {
	role, err := biz.store.GetRoleByID(ctx, roleID)
	if err != nil {
		if err == core.ErrRecordNotFound {
			return nil, core.ErrNotFound.WithError("role not found")
		}
		return nil, core.ErrInternalServerError.WithDebug(err.Error())
	}

	perms, err := biz.store.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, core.ErrInternalServerError.WithDebug(err.Error())
	}

	return &entity.RoleWithPermissions{
		Role:        *role,
		Permissions: perms,
	}, nil
}

func (biz *rbacBusiness) UpdateRole(ctx context.Context, roleID int, req *entity.UpdateRoleRequest) error {
	// Check if role exists
	_, err := biz.store.GetRoleByID(ctx, roleID)
	if err != nil {
		if err == core.ErrRecordNotFound {
			return core.ErrNotFound.WithError("role not found")
		}
		return core.ErrInternalServerError.WithDebug(err.Error())
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)

	if err := biz.store.UpdateRole(ctx, roleID, req.Name, req.Description); err != nil {
		return core.ErrInternalServerError.WithDebug(err.Error())
	}

	return nil
}

func (biz *rbacBusiness) DeleteRole(ctx context.Context, roleID int) error {
	// Check if role exists
	role, err := biz.store.GetRoleByID(ctx, roleID)
	if err != nil {
		if err == core.ErrRecordNotFound {
			return core.ErrNotFound.WithError("role not found")
		}
		return core.ErrInternalServerError.WithDebug(err.Error())
	}

	// Prevent deletion of system roles
	if role.Code == "sadmin" || role.Code == "admin" || role.Code == "user" {
		return core.ErrBadRequest.WithError("cannot delete system roles")
	}

	if err := biz.store.DeleteRole(ctx, roleID); err != nil {
		return core.ErrInternalServerError.WithDebug(err.Error())
	}

	return nil
}

// ========== Permission management ==========

func (biz *rbacBusiness) CreatePermission(ctx context.Context, req *entity.CreatePermissionRequest) (*entity.Permission, error) {
	// Validate input
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)

	if req.Code == "" || req.Name == "" {
		return nil, core.ErrBadRequest.WithError("code and name are required")
	}

	// Check if permission code already exists
	existing, err := biz.store.GetPermissionByCode(ctx, req.Code)
	if err != nil && err != core.ErrRecordNotFound {
		return nil, core.ErrInternalServerError.WithDebug(err.Error())
	}
	if existing != nil {
		return nil, core.ErrBadRequest.WithError("permission code already exists")
	}

	perm := &entity.Permission{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := biz.store.CreatePermission(ctx, perm); err != nil {
		return nil, core.ErrInternalServerError.WithDebug(err.Error())
	}

	return perm, nil
}

func (biz *rbacBusiness) ListPermissions(ctx context.Context) ([]entity.Permission, error) {
	perms, err := biz.store.ListPermissions(ctx)
	if err != nil {
		return nil, core.ErrInternalServerError.WithDebug(err.Error())
	}
	return perms, nil
}

// ========== Role-Permission assignment ==========

func (biz *rbacBusiness) AssignPermissionsToRole(ctx context.Context, roleID int, req *entity.AssignPermissionsRequest) error {
	// Check if role exists
	_, err := biz.store.GetRoleByID(ctx, roleID)
	if err != nil {
		if err == core.ErrRecordNotFound {
			return core.ErrNotFound.WithError("role not found")
		}
		return core.ErrInternalServerError.WithDebug(err.Error())
	}

	// Validate that all permissions exist
	for _, permID := range req.PermissionIDs {
		_, err := biz.store.GetPermissionByID(ctx, permID)
		if err != nil {
			if err == core.ErrRecordNotFound {
				return core.ErrBadRequest.WithError("one or more permissions not found")
			}
			return core.ErrInternalServerError.WithDebug(err.Error())
		}
	}

	if err := biz.store.AssignPermissionsToRole(ctx, roleID, req.PermissionIDs); err != nil {
		return core.ErrInternalServerError.WithDebug(err.Error())
	}

	return nil
}

func (biz *rbacBusiness) RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error {
	// Check if role exists
	_, err := biz.store.GetRoleByID(ctx, roleID)
	if err != nil {
		if err == core.ErrRecordNotFound {
			return core.ErrNotFound.WithError("role not found")
		}
		return core.ErrInternalServerError.WithDebug(err.Error())
	}

	// Check if permission exists
	_, err = biz.store.GetPermissionByID(ctx, permissionID)
	if err != nil {
		if err == core.ErrRecordNotFound {
			return core.ErrNotFound.WithError("permission not found")
		}
		return core.ErrInternalServerError.WithDebug(err.Error())
	}

	if err := biz.store.RemovePermissionFromRole(ctx, roleID, permissionID); err != nil {
		return core.ErrInternalServerError.WithDebug(err.Error())
	}

	return nil
}

// ========== User-Role assignment ==========

func (biz *rbacBusiness) AssignRolesToUser(ctx context.Context, userID int, req *entity.AssignRolesRequest) error {
	// Validate that all roles exist
	for _, roleID := range req.RoleIDs {
		_, err := biz.store.GetRoleByID(ctx, roleID)
		if err != nil {
			if err == core.ErrRecordNotFound {
				return core.ErrBadRequest.WithError("one or more roles not found")
			}
			return core.ErrInternalServerError.WithDebug(err.Error())
		}
	}

	if err := biz.store.AssignRolesToUser(ctx, userID, req.RoleIDs); err != nil {
		return core.ErrInternalServerError.WithDebug(err.Error())
	}

	return nil
}

func (biz *rbacBusiness) GetUserRoles(ctx context.Context, userID int) (*entity.UserWithRoles, error) {
	roles, err := biz.store.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, core.ErrInternalServerError.WithDebug(err.Error())
	}

	return &entity.UserWithRoles{
		UserID: userID,
		Roles:  roles,
	}, nil
}

func (biz *rbacBusiness) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	// Check if role exists
	_, err := biz.store.GetRoleByID(ctx, roleID)
	if err != nil {
		if err == core.ErrRecordNotFound {
			return core.ErrNotFound.WithError("role not found")
		}
		return core.ErrInternalServerError.WithDebug(err.Error())
	}

	if err := biz.store.RemoveRoleFromUser(ctx, userID, roleID); err != nil {
		return core.ErrInternalServerError.WithDebug(err.Error())
	}

	return nil
}

// ========== Permission checking ==========

func (biz *rbacBusiness) HasAnyPermission(ctx context.Context, userID int, permCodes ...string) (bool, error) {
	hasPermission, err := biz.store.HasAnyPermission(ctx, userID, permCodes...)
	if err != nil {
		return false, core.ErrInternalServerError.WithDebug(err.Error())
	}
	return hasPermission, nil
}
