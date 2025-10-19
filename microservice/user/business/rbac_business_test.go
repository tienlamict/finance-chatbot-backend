package business

import (
	"context"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/user/entity"
	"testing"
)

// Mock RBAC Store
type mockRBACStore struct {
	roles           map[int]*entity.Role
	rolesByCode     map[string]*entity.Role
	permissions     map[int]*entity.Permission
	permsByCode     map[string]*entity.Permission
	rolePermissions map[int][]int // roleID -> []permissionID
	userRoles       map[int][]int // userID -> []roleID
}

func newMockRBACStore() *mockRBACStore {
	return &mockRBACStore{
		roles:           make(map[int]*entity.Role),
		rolesByCode:     make(map[string]*entity.Role),
		permissions:     make(map[int]*entity.Permission),
		permsByCode:     make(map[string]*entity.Permission),
		rolePermissions: make(map[int][]int),
		userRoles:       make(map[int][]int),
	}
}

func (m *mockRBACStore) CreateRole(ctx context.Context, role *entity.Role) error {
	role.ID = len(m.roles) + 1
	m.roles[role.ID] = role
	m.rolesByCode[role.Code] = role
	return nil
}

func (m *mockRBACStore) GetRoleByID(ctx context.Context, roleID int) (*entity.Role, error) {
	role, ok := m.roles[roleID]
	if !ok {
		return nil, core.ErrRecordNotFound
	}
	return role, nil
}

func (m *mockRBACStore) GetRoleByCode(ctx context.Context, code string) (*entity.Role, error) {
	role, ok := m.rolesByCode[code]
	if !ok {
		return nil, core.ErrRecordNotFound
	}
	return role, nil
}

func (m *mockRBACStore) ListRoles(ctx context.Context) ([]entity.Role, error) {
	roles := make([]entity.Role, 0, len(m.roles))
	for _, role := range m.roles {
		roles = append(roles, *role)
	}
	return roles, nil
}

func (m *mockRBACStore) UpdateRole(ctx context.Context, roleID int, name, description string) error {
	role, ok := m.roles[roleID]
	if !ok {
		return core.ErrRecordNotFound
	}
	if name != "" {
		role.Name = name
	}
	if description != "" {
		role.Description = description
	}
	return nil
}

func (m *mockRBACStore) DeleteRole(ctx context.Context, roleID int) error {
	role, ok := m.roles[roleID]
	if !ok {
		return core.ErrRecordNotFound
	}
	delete(m.roles, roleID)
	delete(m.rolesByCode, role.Code)
	return nil
}

func (m *mockRBACStore) CreatePermission(ctx context.Context, perm *entity.Permission) error {
	perm.ID = len(m.permissions) + 1
	m.permissions[perm.ID] = perm
	m.permsByCode[perm.Code] = perm
	return nil
}

func (m *mockRBACStore) GetPermissionByID(ctx context.Context, permID int) (*entity.Permission, error) {
	perm, ok := m.permissions[permID]
	if !ok {
		return nil, core.ErrRecordNotFound
	}
	return perm, nil
}

func (m *mockRBACStore) GetPermissionByCode(ctx context.Context, code string) (*entity.Permission, error) {
	perm, ok := m.permsByCode[code]
	if !ok {
		return nil, core.ErrRecordNotFound
	}
	return perm, nil
}

func (m *mockRBACStore) ListPermissions(ctx context.Context) ([]entity.Permission, error) {
	perms := make([]entity.Permission, 0, len(m.permissions))
	for _, perm := range m.permissions {
		perms = append(perms, *perm)
	}
	return perms, nil
}

func (m *mockRBACStore) GetRolePermissions(ctx context.Context, roleID int) ([]entity.Permission, error) {
	permIDs, ok := m.rolePermissions[roleID]
	if !ok {
		return []entity.Permission{}, nil
	}

	perms := make([]entity.Permission, 0, len(permIDs))
	for _, permID := range permIDs {
		if perm, ok := m.permissions[permID]; ok {
			perms = append(perms, *perm)
		}
	}
	return perms, nil
}

func (m *mockRBACStore) AssignPermissionsToRole(ctx context.Context, roleID int, permissionIDs []int) error {
	if _, ok := m.roles[roleID]; !ok {
		return core.ErrRecordNotFound
	}

	existing := m.rolePermissions[roleID]
	for _, permID := range permissionIDs {
		// Check if already exists
		found := false
		for _, existingID := range existing {
			if existingID == permID {
				found = true
				break
			}
		}
		if !found {
			existing = append(existing, permID)
		}
	}
	m.rolePermissions[roleID] = existing
	return nil
}

func (m *mockRBACStore) RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error {
	permIDs, ok := m.rolePermissions[roleID]
	if !ok {
		return nil
	}

	newPermIDs := make([]int, 0, len(permIDs))
	for _, id := range permIDs {
		if id != permissionID {
			newPermIDs = append(newPermIDs, id)
		}
	}
	m.rolePermissions[roleID] = newPermIDs
	return nil
}

func (m *mockRBACStore) GetUserRoles(ctx context.Context, userID int) ([]entity.Role, error) {
	roleIDs, ok := m.userRoles[userID]
	if !ok {
		return []entity.Role{}, nil
	}

	roles := make([]entity.Role, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		if role, ok := m.roles[roleID]; ok {
			roles = append(roles, *role)
		}
	}
	return roles, nil
}

func (m *mockRBACStore) AssignRolesToUser(ctx context.Context, userID int, roleIDs []int) error {
	existing := m.userRoles[userID]
	for _, roleID := range roleIDs {
		found := false
		for _, existingID := range existing {
			if existingID == roleID {
				found = true
				break
			}
		}
		if !found {
			existing = append(existing, roleID)
		}
	}
	m.userRoles[userID] = existing
	return nil
}

func (m *mockRBACStore) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	roleIDs, ok := m.userRoles[userID]
	if !ok {
		return nil
	}

	newRoleIDs := make([]int, 0, len(roleIDs))
	for _, id := range roleIDs {
		if id != roleID {
			newRoleIDs = append(newRoleIDs, id)
		}
	}
	m.userRoles[userID] = newRoleIDs
	return nil
}

func (m *mockRBACStore) GetUserPermissionCodes(ctx context.Context, userID int) ([]string, error) {
	roleIDs, ok := m.userRoles[userID]
	if !ok {
		return []string{}, nil
	}

	permCodes := make(map[string]bool)
	for _, roleID := range roleIDs {
		permIDs, ok := m.rolePermissions[roleID]
		if !ok {
			continue
		}
		for _, permID := range permIDs {
			if perm, ok := m.permissions[permID]; ok {
				permCodes[perm.Code] = true
			}
		}
	}

	result := make([]string, 0, len(permCodes))
	for code := range permCodes {
		result = append(result, code)
	}
	return result, nil
}

func (m *mockRBACStore) HasAnyPermission(ctx context.Context, userID int, permCodes ...string) (bool, error) {
	userPerms, err := m.GetUserPermissionCodes(ctx, userID)
	if err != nil {
		return false, err
	}

	permMap := make(map[string]bool)
	for _, code := range userPerms {
		permMap[code] = true
	}

	for _, code := range permCodes {
		if permMap[code] {
			return true, nil
		}
	}
	return false, nil
}

// Tests

func TestCreateRole(t *testing.T) {
	store := newMockRBACStore()
	biz := NewRBACBusiness(store)
	ctx := context.Background()

	req := &entity.CreateRoleRequest{
		Code:        "test-role",
		Name:        "Test Role",
		Description: "A test role",
	}

	role, err := biz.CreateRole(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if role.Code != "test-role" {
		t.Errorf("Expected role code 'test-role', got '%s'", role.Code)
	}

	if role.Name != "Test Role" {
		t.Errorf("Expected role name 'Test Role', got '%s'", role.Name)
	}
}

func TestCreateRoleDuplicate(t *testing.T) {
	store := newMockRBACStore()
	biz := NewRBACBusiness(store)
	ctx := context.Background()

	req := &entity.CreateRoleRequest{
		Code: "test-role",
		Name: "Test Role",
	}

	_, err := biz.CreateRole(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error on first create, got %v", err)
	}

	_, err = biz.CreateRole(ctx, req)
	if err == nil {
		t.Fatal("Expected error on duplicate role, got nil")
	}
}

func TestDeleteSystemRole(t *testing.T) {
	store := newMockRBACStore()
	biz := NewRBACBusiness(store)
	ctx := context.Background()

	// Create system role
	role := &entity.Role{ID: 1, Code: "sadmin", Name: "Super Admin"}
	store.roles[1] = role
	store.rolesByCode["sadmin"] = role

	err := biz.DeleteRole(ctx, 1)
	if err == nil {
		t.Fatal("Expected error when deleting system role, got nil")
	}
}

func TestAssignPermissionsToRole(t *testing.T) {
	store := newMockRBACStore()
	biz := NewRBACBusiness(store)
	ctx := context.Background()

	// Create role and permissions
	role := &entity.Role{ID: 1, Code: "test", Name: "Test"}
	perm1 := &entity.Permission{ID: 1, Code: "perm1", Name: "Permission 1"}
	perm2 := &entity.Permission{ID: 2, Code: "perm2", Name: "Permission 2"}

	store.roles[1] = role
	store.permissions[1] = perm1
	store.permissions[2] = perm2

	req := &entity.AssignPermissionsRequest{
		PermissionIDs: []int{1, 2},
	}

	err := biz.AssignPermissionsToRole(ctx, 1, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	perms, _ := store.GetRolePermissions(ctx, 1)
	if len(perms) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(perms))
	}
}

func TestAssignRolesToUser(t *testing.T) {
	store := newMockRBACStore()
	biz := NewRBACBusiness(store)
	ctx := context.Background()

	// Create roles
	role1 := &entity.Role{ID: 1, Code: "role1", Name: "Role 1"}
	role2 := &entity.Role{ID: 2, Code: "role2", Name: "Role 2"}

	store.roles[1] = role1
	store.roles[2] = role2

	req := &entity.AssignRolesRequest{
		RoleIDs: []int{1, 2},
	}

	err := biz.AssignRolesToUser(ctx, 1, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	roles, _ := store.GetUserRoles(ctx, 1)
	if len(roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(roles))
	}
}

func TestHasAnyPermission(t *testing.T) {
	store := newMockRBACStore()
	biz := NewRBACBusiness(store)
	ctx := context.Background()

	// Setup: role with permissions
	role := &entity.Role{ID: 1, Code: "test", Name: "Test"}
	perm := &entity.Permission{ID: 1, Code: "test.read", Name: "Test Read"}

	store.roles[1] = role
	store.permissions[1] = perm
	store.rolePermissions[1] = []int{1}
	store.userRoles[1] = []int{1}

	has, err := biz.HasAnyPermission(ctx, 1, "test.read")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !has {
		t.Error("Expected user to have permission, but didn't")
	}

	has, err = biz.HasAnyPermission(ctx, 1, "test.write")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if has {
		t.Error("Expected user to not have permission, but did")
	}
}
