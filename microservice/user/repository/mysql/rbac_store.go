package mysql

import (
	"context"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/user/entity"

	"gorm.io/gorm"
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

type rbacStore struct {
	db *gorm.DB
}

func NewRBACStore(db *gorm.DB) RBACStore {
	return &rbacStore{db: db}
}

// ========== Role operations ==========

func (s *rbacStore) CreateRole(ctx context.Context, role *entity.Role) error {
	if err := s.db.WithContext(ctx).Create(role).Error; err != nil {
		return err
	}
	return nil
}

func (s *rbacStore) GetRoleByID(ctx context.Context, roleID int) (*entity.Role, error) {
	var role entity.Role
	if err := s.db.WithContext(ctx).Where("id = ?", roleID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, core.ErrRecordNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (s *rbacStore) GetRoleByCode(ctx context.Context, code string) (*entity.Role, error) {
	var role entity.Role
	if err := s.db.WithContext(ctx).Where("code = ?", code).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, core.ErrRecordNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (s *rbacStore) ListRoles(ctx context.Context) ([]entity.Role, error) {
	var roles []entity.Role
	if err := s.db.WithContext(ctx).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (s *rbacStore) UpdateRole(ctx context.Context, roleID int, name, description string) error {
	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if description != "" {
		updates["description"] = description
	}

	if len(updates) == 0 {
		return nil
	}

	return s.db.WithContext(ctx).Model(&entity.Role{}).Where("id = ?", roleID).Updates(updates).Error
}

func (s *rbacStore) DeleteRole(ctx context.Context, roleID int) error {
	return s.db.WithContext(ctx).Delete(&entity.Role{}, roleID).Error
}

// ========== Permission operations ==========

func (s *rbacStore) CreatePermission(ctx context.Context, perm *entity.Permission) error {
	if err := s.db.WithContext(ctx).Create(perm).Error; err != nil {
		return err
	}
	return nil
}

func (s *rbacStore) GetPermissionByID(ctx context.Context, permID int) (*entity.Permission, error) {
	var perm entity.Permission
	if err := s.db.WithContext(ctx).Where("id = ?", permID).First(&perm).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, core.ErrRecordNotFound
		}
		return nil, err
	}
	return &perm, nil
}

func (s *rbacStore) GetPermissionByCode(ctx context.Context, code string) (*entity.Permission, error) {
	var perm entity.Permission
	if err := s.db.WithContext(ctx).Where("code = ?", code).First(&perm).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, core.ErrRecordNotFound
		}
		return nil, err
	}
	return &perm, nil
}

func (s *rbacStore) ListPermissions(ctx context.Context) ([]entity.Permission, error) {
	var perms []entity.Permission
	if err := s.db.WithContext(ctx).Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// ========== Role-Permission operations ==========

func (s *rbacStore) GetRolePermissions(ctx context.Context, roleID int) ([]entity.Permission, error) {
	var perms []entity.Permission
	query := `
		SELECT p.*
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = ?
	`
	if err := s.db.WithContext(ctx).Raw(query, roleID).Scan(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

func (s *rbacStore) AssignPermissionsToRole(ctx context.Context, roleID int, permissionIDs []int) error {
	if len(permissionIDs) == 0 {
		return nil
	}

	// Use transaction to ensure atomicity
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, permID := range permissionIDs {
			// Use INSERT IGNORE pattern
			if err := tx.Exec(`
				INSERT IGNORE INTO role_permissions (role_id, permission_id, created_at)
				VALUES (?, ?, NOW())
			`, roleID, permID).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *rbacStore) RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error {
	return s.db.WithContext(ctx).Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&entity.RolePermission{}).Error
}

// ========== User-Role operations ==========

func (s *rbacStore) GetUserRoles(ctx context.Context, userID int) ([]entity.Role, error) {
	var roles []entity.Role
	query := `
		SELECT r.*
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = ?
	`
	if err := s.db.WithContext(ctx).Raw(query, userID).Scan(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (s *rbacStore) AssignRolesToUser(ctx context.Context, userID int, roleIDs []int) error {
	if len(roleIDs) == 0 {
		return nil
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, roleID := range roleIDs {
			// Use INSERT IGNORE pattern
			if err := tx.Exec(`
				INSERT IGNORE INTO user_roles (user_id, role_id, created_at)
				VALUES (?, ?, NOW())
			`, userID, roleID).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *rbacStore) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	return s.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&entity.UserRole{}).Error
}

// ========== Permission checking ==========

func (s *rbacStore) GetUserPermissionCodes(ctx context.Context, userID int) ([]string, error) {
	var codes []string
	raw := `
		SELECT DISTINCT p.code
		FROM user_roles ur
		JOIN role_permissions rp ON rp.role_id = ur.role_id
		JOIN permissions p       ON p.id       = rp.permission_id
		WHERE ur.user_id = ?
	`
	if err := s.db.WithContext(ctx).Raw(raw, userID).Scan(&codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}

func (s *rbacStore) HasAnyPermission(ctx context.Context, userID int, permCodes ...string) (bool, error) {
	if len(permCodes) == 0 {
		return false, nil
	}

	var exists int
	query := `
		SELECT 1
		FROM user_roles ur
		JOIN role_permissions rp ON rp.role_id = ur.role_id
		JOIN permissions p       ON p.id       = rp.permission_id
		WHERE ur.user_id = ? AND p.code IN (?)
		LIMIT 1
	`
	if err := s.db.WithContext(ctx).Raw(query, userID, permCodes).Scan(&exists).Error; err != nil {
		return false, err
	}
	return exists == 1, nil
}
