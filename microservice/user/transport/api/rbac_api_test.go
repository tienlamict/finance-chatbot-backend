package api

import (
	"bytes"
	"context"
	"encoding/json"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/user/entity"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock RBAC Business
type mockRBACBusiness struct {
	roles       map[int]*entity.Role
	permissions map[int]*entity.Permission
}

func newMockRBACBusiness() *mockRBACBusiness {
	return &mockRBACBusiness{
		roles:       make(map[int]*entity.Role),
		permissions: make(map[int]*entity.Permission),
	}
}

func (m *mockRBACBusiness) CreateRole(ctx context.Context, req *entity.CreateRoleRequest) (*entity.Role, error) {
	role := &entity.Role{
		ID:          len(m.roles) + 1,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	}
	m.roles[role.ID] = role
	return role, nil
}

func (m *mockRBACBusiness) ListRoles(ctx context.Context) ([]entity.Role, error) {
	roles := make([]entity.Role, 0, len(m.roles))
	for _, role := range m.roles {
		roles = append(roles, *role)
	}
	return roles, nil
}

func (m *mockRBACBusiness) GetRoleWithPermissions(ctx context.Context, roleID int) (*entity.RoleWithPermissions, error) {
	role, ok := m.roles[roleID]
	if !ok {
		return nil, nil
	}
	return &entity.RoleWithPermissions{
		Role:        *role,
		Permissions: []entity.Permission{},
	}, nil
}

func (m *mockRBACBusiness) UpdateRole(ctx context.Context, roleID int, req *entity.UpdateRoleRequest) error {
	role, ok := m.roles[roleID]
	if !ok {
		return nil
	}
	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	return nil
}

func (m *mockRBACBusiness) DeleteRole(ctx context.Context, roleID int) error {
	delete(m.roles, roleID)
	return nil
}

func (m *mockRBACBusiness) CreatePermission(ctx context.Context, req *entity.CreatePermissionRequest) (*entity.Permission, error) {
	perm := &entity.Permission{
		ID:          len(m.permissions) + 1,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	}
	m.permissions[perm.ID] = perm
	return perm, nil
}

func (m *mockRBACBusiness) ListPermissions(ctx context.Context) ([]entity.Permission, error) {
	perms := make([]entity.Permission, 0, len(m.permissions))
	for _, perm := range m.permissions {
		perms = append(perms, *perm)
	}
	return perms, nil
}

func (m *mockRBACBusiness) AssignPermissionsToRole(ctx context.Context, roleID int, req *entity.AssignPermissionsRequest) error {
	return nil
}

func (m *mockRBACBusiness) RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error {
	return nil
}

func (m *mockRBACBusiness) AssignRolesToUser(ctx context.Context, userID int, req *entity.AssignRolesRequest) error {
	return nil
}

func (m *mockRBACBusiness) GetUserRoles(ctx context.Context, userID int) (*entity.UserWithRoles, error) {
	return &entity.UserWithRoles{
		UserID: userID,
		Roles:  []entity.Role{},
	}, nil
}

func (m *mockRBACBusiness) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	return nil
}

// Helper function to setup router
func setupTestRouter(api *rbacAPI) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Roles
	router.POST("/v1/rbac/roles", api.CreateRoleHdl())
	router.GET("/v1/rbac/roles", api.ListRolesHdl())
	router.GET("/v1/rbac/roles/:id", api.GetRoleWithPermissionsHdl())
	router.PATCH("/v1/rbac/roles/:id", api.UpdateRoleHdl())
	router.DELETE("/v1/rbac/roles/:id", api.DeleteRoleHdl())

	// Permissions
	router.POST("/v1/rbac/permissions", api.CreatePermissionHdl())
	router.GET("/v1/rbac/permissions", api.ListPermissionsHdl())

	// Role-Permission assignment (separate routes to avoid conflicts)
	router.POST("/v1/rbac/role-permissions/:id", api.AssignPermissionsToRoleHdl())
	router.DELETE("/v1/rbac/role-permissions/:roleId/:permId", api.RemovePermissionFromRoleHdl())

	// User-Role assignment
	router.POST("/v1/rbac/user-roles/:userId", api.AssignRolesToUserHdl())
	router.GET("/v1/rbac/user-roles/:userId", api.GetUserRolesHdl())
	router.DELETE("/v1/rbac/user-roles/:userId/:roleId", api.RemoveRoleFromUserHdl())

	return router
}

// Tests

func TestCreateRoleAPI(t *testing.T) {
	mockBiz := newMockRBACBusiness()
	api := NewRBACAPI(mockBiz)
	router := setupTestRouter(api)

	reqBody := entity.CreateRoleRequest{
		Code:        "test-role",
		Name:        "Test Role",
		Description: "A test role",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/v1/rbac/roles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "test-role")
	assert.Contains(t, w.Body.String(), "Test Role")
}

func TestListRolesAPI(t *testing.T) {
	mockBiz := newMockRBACBusiness()

	// Pre-populate with test data
	mockBiz.roles[1] = &entity.Role{ID: 1, Code: "role1", Name: "Role 1"}
	mockBiz.roles[2] = &entity.Role{ID: 2, Code: "role2", Name: "Role 2"}

	api := NewRBACAPI(mockBiz)
	router := setupTestRouter(api)

	req, _ := http.NewRequest(http.MethodGet, "/v1/rbac/roles", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "role1")
	assert.Contains(t, w.Body.String(), "role2")
}

func TestGetRoleWithPermissionsAPI(t *testing.T) {
	mockBiz := newMockRBACBusiness()
	mockBiz.roles[1] = &entity.Role{ID: 1, Code: "test", Name: "Test Role"}

	api := NewRBACAPI(mockBiz)
	router := setupTestRouter(api)

	req, _ := http.NewRequest(http.MethodGet, "/v1/rbac/roles/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test")
	assert.Contains(t, w.Body.String(), "permissions")
}

func TestUpdateRoleAPI(t *testing.T) {
	mockBiz := newMockRBACBusiness()
	mockBiz.roles[1] = &entity.Role{ID: 1, Code: "test", Name: "Old Name"}

	api := NewRBACAPI(mockBiz)
	router := setupTestRouter(api)

	reqBody := entity.UpdateRoleRequest{
		Name:        "New Name",
		Description: "Updated description",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPatch, "/v1/rbac/roles/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "New Name", mockBiz.roles[1].Name)
}

func TestDeleteRoleAPI(t *testing.T) {
	mockBiz := newMockRBACBusiness()
	mockBiz.roles[1] = &entity.Role{ID: 1, Code: "test", Name: "Test"}

	api := NewRBACAPI(mockBiz)
	router := setupTestRouter(api)

	req, _ := http.NewRequest(http.MethodDelete, "/v1/rbac/roles/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	_, exists := mockBiz.roles[1]
	assert.False(t, exists)
}

func TestCreatePermissionAPI(t *testing.T) {
	mockBiz := newMockRBACBusiness()
	api := NewRBACAPI(mockBiz)
	router := setupTestRouter(api)

	reqBody := entity.CreatePermissionRequest{
		Code:        "test.read",
		Name:        "Test Read",
		Description: "Read permission",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/v1/rbac/permissions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "test.read")
}

func TestAssignPermissionsToRoleAPI(t *testing.T) {
	mockBiz := newMockRBACBusiness()
	mockBiz.roles[1] = &entity.Role{ID: 1, Code: "test", Name: "Test"}

	api := NewRBACAPI(mockBiz)
	router := setupTestRouter(api)

	reqBody := entity.AssignPermissionsRequest{
		PermissionIDs: []int{1, 2, 3},
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/v1/rbac/role-permissions/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAssignRolesToUserAPI(t *testing.T) {
	mockBiz := newMockRBACBusiness()
	api := NewRBACAPI(mockBiz)
	router := setupTestRouter(api)

	reqBody := entity.AssignRolesRequest{
		RoleIDs: []int{1, 2},
	}

	// Create encoded UID for user ID = 1
	uid := core.NewUID(uint32(1), 1, 1)
	encodedUserID := uid.String()

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/v1/rbac/user-roles/"+encodedUserID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserRolesAPI(t *testing.T) {
	mockBiz := newMockRBACBusiness()
	api := NewRBACAPI(mockBiz)
	router := setupTestRouter(api)

	// Create encoded UID for user ID = 1
	uid := core.NewUID(uint32(1), 1, 1)
	encodedUserID := uid.String()

	req, _ := http.NewRequest(http.MethodGet, "/v1/rbac/user-roles/"+encodedUserID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "roles")
}
