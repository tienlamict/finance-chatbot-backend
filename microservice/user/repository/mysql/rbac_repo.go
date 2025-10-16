package mysql

import (
	"context"

	"gorm.io/gorm"
)

type RBACRepository interface {
	GetUserPermissionCodes(ctx context.Context, userID int) ([]string, error)
	HasPermission(ctx context.Context, userID int, permCode string) (bool, error)
}

type rbacRepo struct{ db *gorm.DB }

func NewRBACRepository(db *gorm.DB) RBACRepository { return &rbacRepo{db: db} }

// Lấy toàn bộ permission code của user theo bảng user_roles -> role_permissions -> permissions
func (r *rbacRepo) GetUserPermissionCodes(ctx context.Context, userID int) ([]string, error) {
	var codes []string
	raw := `
SELECT DISTINCT p.code
FROM user_roles ur
JOIN role_permissions rp ON rp.role_id = ur.role_id
JOIN permissions p       ON p.id       = rp.permission_id
WHERE ur.user_id = ?`
	if err := r.db.WithContext(ctx).Raw(raw, userID).Scan(&codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}

func (r *rbacRepo) HasPermission(ctx context.Context, userID int, permCode string) (bool, error) {
	var exists int
	raw := `
SELECT 1
FROM user_roles ur
JOIN role_permissions rp ON rp.role_id = ur.role_id
JOIN permissions p       ON p.id       = rp.permission_id
WHERE ur.user_id = ? AND p.code = ?
LIMIT 1`
	if err := r.db.WithContext(ctx).Raw(raw, userID, permCode).Scan(&exists).Error; err != nil {
		return false, err
	}
	return exists == 1, nil
}
