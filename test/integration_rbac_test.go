package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// Integration tests for RBAC functionality
// These tests should be run against a running instance with the database seeded

const (
	baseURL            = "http://localhost:8080/v1"
	superadminEmail    = "superadmin@system.local"
	superadminPassword = "superadmin"
)

type LoginResponse struct {
	Data struct {
		AccessToken struct {
			Token     string `json:"token"`
			ExpiredIn int    `json:"expired_in"`
		} `json:"access_token"`
	} `json:"data"`
}

type APIResponse struct {
	Data interface{} `json:"data"`
}

// Helper function to login and get token
func loginAsSuperAdmin(t *testing.T) string {
	loginPayload := map[string]string{
		"email":    superadminEmail,
		"password": superadminPassword,
	}

	body, _ := json.Marshal(loginPayload)
	resp, err := http.Post(fmt.Sprintf("%s/authenticate", baseURL), "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login failed with status: %d", resp.StatusCode)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}

	return loginResp.Data.AccessToken.Token
}

// Helper function to make authenticated request
func makeAuthRequest(t *testing.T, method, endpoint, token string, body interface{}) (*http.Response, error) {
	var reqBody *bytes.Buffer
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = bytes.NewBuffer([]byte{})
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", baseURL, endpoint), reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	return client.Do(req)
}

func TestSuperAdminLogin(t *testing.T) {
	t.Skip("This is an integration test, run manually with database")

	token := loginAsSuperAdmin(t)
	if token == "" {
		t.Fatal("Expected non-empty token")
	}
	t.Logf("Successfully logged in as superadmin, token: %s", token[:20]+"...")
}

func TestListRolesAsSuperAdmin(t *testing.T) {
	t.Skip("This is an integration test, run manually with database")

	token := loginAsSuperAdmin(t)

	resp, err := makeAuthRequest(t, http.MethodGet, "/rbac/roles", token, nil)
	if err != nil {
		t.Fatalf("Failed to list roles: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	roles := apiResp.Data.([]interface{})
	if len(roles) < 3 {
		t.Errorf("Expected at least 3 roles (sadmin, admin, user), got %d", len(roles))
	}

	t.Logf("Found %d roles", len(roles))
}

func TestCreateRoleAsSuperAdmin(t *testing.T) {
	t.Skip("This is an integration test, run manually with database")

	token := loginAsSuperAdmin(t)

	newRole := map[string]string{
		"code":        "test-role",
		"name":        "Test Role",
		"description": "A test role created during integration test",
	}

	resp, err := makeAuthRequest(t, http.MethodPost, "/rbac/roles", token, newRole)
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 201 or 200, got %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	role := apiResp.Data.(map[string]interface{})
	if role["code"] != "test-role" {
		t.Errorf("Expected role code 'test-role', got '%s'", role["code"])
	}

	t.Logf("Successfully created role: %+v", role)
}

func TestListPermissionsAsSuperAdmin(t *testing.T) {
	t.Skip("This is an integration test, run manually with database")

	token := loginAsSuperAdmin(t)

	resp, err := makeAuthRequest(t, http.MethodGet, "/rbac/permissions", token, nil)
	if err != nil {
		t.Fatalf("Failed to list permissions: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	perms := apiResp.Data.([]interface{})
	t.Logf("Found %d permissions", len(perms))
}

func TestCreatePermissionAsSuperAdmin(t *testing.T) {
	t.Skip("This is an integration test, run manually with database")

	token := loginAsSuperAdmin(t)

	newPerm := map[string]string{
		"code":        "test.permission",
		"name":        "Test Permission",
		"description": "A test permission",
	}

	resp, err := makeAuthRequest(t, http.MethodPost, "/rbac/permissions", token, newPerm)
	if err != nil {
		t.Fatalf("Failed to create permission: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 201 or 200, got %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	perm := apiResp.Data.(map[string]interface{})
	if perm["code"] != "test.permission" {
		t.Errorf("Expected permission code 'test.permission', got '%s'", perm["code"])
	}

	t.Logf("Successfully created permission: %+v", perm)
}

func TestNormalUserCannotAccessRBAC(t *testing.T) {
	t.Skip("This is an integration test, run manually with database")

	// This test requires a normal user to be created
	// For now, we just document the expected behavior

	// 1. Register a normal user
	// 2. Login with normal user credentials
	// 3. Try to access /rbac/roles endpoint
	// 4. Expect 403 Forbidden
}

func TestAssignRolesToUser(t *testing.T) {
	t.Skip("This is an integration test, run manually with database")

	token := loginAsSuperAdmin(t)

	// Assuming user ID 2 exists and role ID 2 (admin) exists
	assignReq := map[string][]int{
		"role_ids": {2},
	}

	resp, err := makeAuthRequest(t, http.MethodPost, "/rbac/users/2/roles", token, assignReq)
	if err != nil {
		t.Fatalf("Failed to assign roles: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	t.Log("Successfully assigned roles to user")
}

func TestGetUserRoles(t *testing.T) {
	t.Skip("This is an integration test, run manually with database")

	token := loginAsSuperAdmin(t)

	// Get roles for user ID 1 (superadmin)
	resp, err := makeAuthRequest(t, http.MethodGet, "/rbac/users/1/roles", token, nil)
	if err != nil {
		t.Fatalf("Failed to get user roles: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	userRoles := apiResp.Data.(map[string]interface{})
	roles := userRoles["roles"].([]interface{})

	t.Logf("User has %d roles", len(roles))
}
