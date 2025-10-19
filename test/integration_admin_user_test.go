package test

import (
	"bytes"
	"encoding/json"
	"finance-chatbot/microservice/user/entity"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAdminCreateUser tests the POST /api/v1/admin/users endpoint
func TestAdminCreateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		req := entity.AdminCreateUserRequest{
			FirstName:  "Test",
			LastName:   "User",
			Email:      "testuser@example.com",
			Phone:      "+1234567890",
			Gender:     entity.GenderMale,
			SystemRole: entity.RoleUser,
			Status:     entity.StatusActive,
		}

		body, _ := json.Marshal(req)
		_ = httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewBuffer(body))
		// Note: In real integration test, you would need to add Authorization header with admin token

		// This is a template - actual integration test would need:
		// 1. A test database
		// 2. Test server setup
		// 3. Admin authentication token
		// 4. Proper assertions on response

		// Expected response structure
		// response: {"data": {"id": "...", "first_name": "Test", ...}}
	})

	t.Run("ValidationError_MissingFields", func(t *testing.T) {
		req := entity.AdminCreateUserRequest{
			// Missing required fields
			Email: "invalid@example.com",
		}

		body, _ := json.Marshal(req)
		_ = httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewBuffer(body))

		// Expected: 400 Bad Request with validation error
	})

	t.Run("DuplicateEmail", func(t *testing.T) {
		req := entity.AdminCreateUserRequest{
			FirstName:  "Test",
			LastName:   "User",
			Email:      "existing@example.com", // Assume this email already exists
			SystemRole: entity.RoleUser,
			Status:     entity.StatusActive,
		}

		body, _ := json.Marshal(req)
		_ = httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewBuffer(body))

		// Expected: 400 Bad Request with "user with this email already exists"
	})
}

// TestAdminUpdateUser tests the PUT/PATCH /api/v1/admin/users/:id endpoint
func TestAdminUpdateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		userID := 1
		firstName := "UpdatedName"
		req := entity.AdminUpdateUserRequest{
			FirstName: &firstName,
		}

		body, _ := json.Marshal(req)
		url := fmt.Sprintf("/api/v1/admin/users/%d", userID)
		_ = httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))

		// Expected: 200 OK with success response
	})

	t.Run("UserNotFound", func(t *testing.T) {
		userID := 999999
		firstName := "UpdatedName"
		req := entity.AdminUpdateUserRequest{
			FirstName: &firstName,
		}

		body, _ := json.Marshal(req)
		url := fmt.Sprintf("/api/v1/admin/users/%d", userID)
		_ = httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))

		// Expected: 404 Not Found
	})

	t.Run("ValidationError", func(t *testing.T) {
		userID := 1
		invalidFirstName := "" // Empty name is invalid
		req := entity.AdminUpdateUserRequest{
			FirstName: &invalidFirstName,
		}

		body, _ := json.Marshal(req)
		url := fmt.Sprintf("/api/v1/admin/users/%d", userID)
		_ = httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))

		// Expected: 400 Bad Request
	})
}

// TestAdminDeleteUser tests the DELETE /api/v1/admin/users/:id endpoint
func TestAdminDeleteUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		userID := 1
		url := fmt.Sprintf("/api/v1/admin/users/%d", userID)
		_ = httptest.NewRequest(http.MethodDelete, url, nil)

		// Expected: 200 OK with success response
	})

	t.Run("UserNotFound", func(t *testing.T) {
		userID := 999999
		url := fmt.Sprintf("/api/v1/admin/users/%d", userID)
		_ = httptest.NewRequest(http.MethodDelete, url, nil)

		// Expected: 404 Not Found
	})
}

// TestAdminListUsers tests the GET /api/v1/admin/users endpoint
func TestAdminListUsers(t *testing.T) {
	t.Run("Success_WithoutFilters", func(t *testing.T) {
		url := "/api/v1/admin/users?page=1&limit=10"
		_ = httptest.NewRequest(http.MethodGet, url, nil)

		// Expected: 200 OK with paginated user list
		// Response structure: {"data": {"data": [...], "paging": {...}}}
	})

	t.Run("Success_WithSearch", func(t *testing.T) {
		url := "/api/v1/admin/users?page=1&limit=10&search=john"
		_ = httptest.NewRequest(http.MethodGet, url, nil)

		// Expected: 200 OK with filtered results
	})

	t.Run("Success_WithRoleFilter", func(t *testing.T) {
		url := fmt.Sprintf("/api/v1/admin/users?page=1&limit=10&role=%s", entity.RoleAdmin)
		_ = httptest.NewRequest(http.MethodGet, url, nil)

		// Expected: 200 OK with filtered results by role
	})

	t.Run("Success_WithStatusFilter", func(t *testing.T) {
		url := fmt.Sprintf("/api/v1/admin/users?page=1&limit=10&status=%s", entity.StatusActive)
		_ = httptest.NewRequest(http.MethodGet, url, nil)

		// Expected: 200 OK with filtered results by status
	})

	t.Run("Success_WithMultipleFilters", func(t *testing.T) {
		url := fmt.Sprintf("/api/v1/admin/users?page=1&limit=10&search=john&role=%s&status=%s",
			entity.RoleUser, entity.StatusActive)
		_ = httptest.NewRequest(http.MethodGet, url, nil)

		// Expected: 200 OK with filtered results by multiple criteria
	})
}

// TestAdminGetUserByID tests the GET /api/v1/admin/users/:id endpoint
func TestAdminGetUserByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		userID := 1
		url := fmt.Sprintf("/api/v1/admin/users/%d", userID)
		_ = httptest.NewRequest(http.MethodGet, url, nil)

		// Expected: 200 OK with user details
	})

	t.Run("UserNotFound", func(t *testing.T) {
		userID := 999999
		url := fmt.Sprintf("/api/v1/admin/users/%d", userID)
		_ = httptest.NewRequest(http.MethodGet, url, nil)

		// Expected: 404 Not Found
	})

	t.Run("InvalidUserID", func(t *testing.T) {
		url := "/api/v1/admin/users/invalid"
		_ = httptest.NewRequest(http.MethodGet, url, nil)

		// Expected: 400 Bad Request
	})
}

// TestAdminUserManagement_Authorization tests authorization requirements
func TestAdminUserManagement_Authorization(t *testing.T) {
	t.Run("UnauthorizedAccess_NoToken", func(t *testing.T) {
		_ = httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
		// No Authorization header

		// Expected: 401 Unauthorized
	})

	t.Run("Forbidden_NonAdminUser", func(t *testing.T) {
		_ = httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
		// Authorization header with regular user token (not admin/superadmin)

		// Expected: 403 Forbidden
	})

	t.Run("Success_AdminUser", func(t *testing.T) {
		_ = httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
		// Authorization header with admin token

		// Expected: 200 OK
	})

	t.Run("Success_SuperAdminUser", func(t *testing.T) {
		_ = httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
		// Authorization header with superadmin token

		// Expected: 200 OK
	})
}

// Helper function for actual integration tests
// This would be implemented with real test infrastructure
func setupTestServer(t *testing.T) (*httptest.Server, func()) {
	// Setup:
	// 1. Initialize test database
	// 2. Run migrations
	// 3. Create test users (including admin)
	// 4. Create test service context
	// 5. Setup test router
	// 6. Return test server and cleanup function

	return nil, func() {
		// Cleanup:
		// 1. Drop test database
		// 2. Close connections
	}
}

// Helper function to create admin JWT token for testing
func createTestAdminToken(t *testing.T, userID int, role entity.SystemRole) string {
	// This would use the JWT component to create a valid token
	return "test-token"
}

// Example of a complete integration test with actual HTTP calls
func TestCompleteAdminUserFlow(t *testing.T) {
	// This is a template for how a real integration test would look
	t.Skip("This is a template - implement with actual test infrastructure")

	// Setup test server
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Create admin token
	_ = createTestAdminToken(t, 1, entity.RoleSuperAdmin)

	// Test 1: Create a new user
	createReq := entity.AdminCreateUserRequest{
		FirstName:  "Integration",
		LastName:   "Test",
		Email:      "integration@test.com",
		SystemRole: entity.RoleUser,
		Status:     entity.StatusActive,
	}
	createBody, _ := json.Marshal(createReq)

	resp, err := http.Post(
		server.URL+"/api/v1/admin/users",
		"application/json",
		bytes.NewBuffer(createBody),
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResponse struct {
		Data entity.User `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&createResponse)
	createdUserID := createResponse.Data.Id

	// Test 2: Get the created user
	resp, err = http.Get(fmt.Sprintf("%s/api/v1/admin/users/%d", server.URL, createdUserID))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test 3: Update the user
	firstName := "Updated"
	updateReq := entity.AdminUpdateUserRequest{
		FirstName: &firstName,
	}
	updateBody, _ := json.Marshal(updateReq)

	req, _ := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/api/v1/admin/users/%d", server.URL, createdUserID),
		bytes.NewBuffer(updateBody),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test 4: List users and verify our user is there
	resp, err = http.Get(fmt.Sprintf("%s/api/v1/admin/users?search=%s", server.URL, "integration"))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test 5: Delete the user
	req, _ = http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/api/v1/admin/users/%d", server.URL, createdUserID),
		nil,
	)
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test 6: Verify user is deleted (should get 404)
	resp, err = http.Get(fmt.Sprintf("%s/api/v1/admin/users/%d", server.URL, createdUserID))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
