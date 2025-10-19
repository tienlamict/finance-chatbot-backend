package business

import (
	"context"
	"errors"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/user/entity"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAdminUserRepository is a mock implementation of AdminUserRepository
type MockAdminUserRepository struct {
	mock.Mock
}

type MockAdminUserPasswordBusiness struct {
	mock.Mock
}

func (m *MockAdminUserPasswordBusiness) CreateUserWithPassword(ctx context.Context, req *entity.AdminCreateUserRequest) (*entity.AdminCreateUserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AdminCreateUserResponse), args.Error(1)
}

func (m *MockAdminUserPasswordBusiness) SetUserPassword(ctx context.Context, userID int, req *entity.AdminSetUserPasswordRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}

func (m *MockAdminUserRepository) AdminCreateUser(ctx context.Context, req *entity.AdminCreateUserRequest) (*entity.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockAdminUserRepository) AdminUpdateUser(ctx context.Context, userID int, req *entity.AdminUpdateUserRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}

func (m *MockAdminUserRepository) AdminDeleteUser(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAdminUserRepository) AdminListUsers(ctx context.Context, filter *entity.ListUsersFilter) ([]entity.User, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.User), args.Error(1)
}

func (m *MockAdminUserRepository) AdminGetUserByID(ctx context.Context, userID int) (*entity.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func TestCreateUser_Success(t *testing.T) {
	mockRepo := new(MockAdminUserRepository)
	mockPasswordBiz := new(MockAdminUserPasswordBusiness)
	biz := NewAdminUserBusiness(mockRepo, mockPasswordBiz)

	ctx := context.Background()
	req := &entity.AdminCreateUserRequest{
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john.doe@example.com",
		Phone:      "+1234567890",
		Gender:     entity.GenderMale,
		SystemRole: entity.RoleUser,
		Status:     entity.StatusActive,
	}

	expectedResponse := &entity.AdminCreateUserResponse{
		User: entity.User{
			SQLModel:   core.NewSQLModel(),
			FirstName:  req.FirstName,
			LastName:   req.LastName,
			Email:      req.Email,
			Phone:      req.Phone,
			Gender:     req.Gender,
			SystemRole: req.SystemRole,
			Status:     req.Status,
		},
		TemporaryPassword: "tempPassword123",
		Message:           "User created successfully",
	}
	expectedResponse.User.Id = 1

	mockPasswordBiz.On("CreateUserWithPassword", ctx, req).Return(expectedResponse, nil)

	response, err := biz.CreateUser(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, expectedResponse.User.Email, response.User.Email)
	assert.NotEmpty(t, response.TemporaryPassword)
	mockPasswordBiz.AssertExpectations(t)
}

func TestCreateUser_ValidationError(t *testing.T) {
	mockRepo := new(MockAdminUserRepository)
	mockPasswordBiz := new(MockAdminUserPasswordBusiness)
	biz := NewAdminUserBusiness(mockRepo, mockPasswordBiz)

	ctx := context.Background()
	req := &entity.AdminCreateUserRequest{
		FirstName:  "", // Invalid: empty first name
		LastName:   "Doe",
		Email:      "john.doe@example.com",
		SystemRole: entity.RoleUser,
		Status:     entity.StatusActive,
	}

	user, err := biz.CreateUser(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, user)
	// Should not call repository if validation fails
	mockRepo.AssertNotCalled(t, "AdminCreateUser")
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	mockRepo := new(MockAdminUserRepository)
	mockPasswordBiz := new(MockAdminUserPasswordBusiness)
	biz := NewAdminUserBusiness(mockRepo, mockPasswordBiz)

	ctx := context.Background()
	req := &entity.AdminCreateUserRequest{
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john.doe@example.com",
		SystemRole: entity.RoleUser,
		Status:     entity.StatusActive,
	}

	mockPasswordBiz.On("CreateUserWithPassword", ctx, req).Return(nil, entity.ErrUserAlreadyExists)

	response, err := biz.CreateUser(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, response)
	mockPasswordBiz.AssertExpectations(t)
}

func TestUpdateUser_Success(t *testing.T) {
	mockRepo := new(MockAdminUserRepository)
	mockPasswordBiz := new(MockAdminUserPasswordBusiness)
	biz := NewAdminUserBusiness(mockRepo, mockPasswordBiz)

	ctx := context.Background()
	userID := 1
	firstName := "Jane"
	req := &entity.AdminUpdateUserRequest{
		FirstName: &firstName,
	}

	mockRepo.On("AdminUpdateUser", ctx, userID, req).Return(nil)

	err := biz.UpdateUser(ctx, userID, req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUser_ValidationError(t *testing.T) {
	mockRepo := new(MockAdminUserRepository)
	mockPasswordBiz := new(MockAdminUserPasswordBusiness)
	biz := NewAdminUserBusiness(mockRepo, mockPasswordBiz)

	ctx := context.Background()
	userID := 1
	invalidFirstName := "" // Invalid: empty first name
	req := &entity.AdminUpdateUserRequest{
		FirstName: &invalidFirstName,
	}

	err := biz.UpdateUser(ctx, userID, req)

	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "AdminUpdateUser")
}

func TestUpdateUser_NotFound(t *testing.T) {
	mockRepo := new(MockAdminUserRepository)
	mockPasswordBiz := new(MockAdminUserPasswordBusiness)
	biz := NewAdminUserBusiness(mockRepo, mockPasswordBiz)

	ctx := context.Background()
	userID := 999
	firstName := "Jane"
	req := &entity.AdminUpdateUserRequest{
		FirstName: &firstName,
	}

	mockRepo.On("AdminUpdateUser", ctx, userID, req).Return(entity.ErrUserNotFound)

	err := biz.UpdateUser(ctx, userID, req)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteUser_Success(t *testing.T) {
	mockRepo := new(MockAdminUserRepository)
	mockPasswordBiz := new(MockAdminUserPasswordBusiness)
	biz := NewAdminUserBusiness(mockRepo, mockPasswordBiz)

	ctx := context.Background()
	userID := 1

	mockRepo.On("AdminDeleteUser", ctx, userID).Return(nil)

	err := biz.DeleteUser(ctx, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteUser_NotFound(t *testing.T) {
	mockRepo := new(MockAdminUserRepository)
	mockPasswordBiz := new(MockAdminUserPasswordBusiness)
	biz := NewAdminUserBusiness(mockRepo, mockPasswordBiz)

	ctx := context.Background()
	userID := 999

	mockRepo.On("AdminDeleteUser", ctx, userID).Return(entity.ErrUserNotFound)

	err := biz.DeleteUser(ctx, userID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestListUsers_Success(t *testing.T) {
	mockRepo := new(MockAdminUserRepository)
	mockPasswordBiz := new(MockAdminUserPasswordBusiness)
	biz := NewAdminUserBusiness(mockRepo, mockPasswordBiz)

	ctx := context.Background()
	filter := &entity.ListUsersFilter{
		Search: "john",
		Paging: core.Paging{
			Page:  1,
			Limit: 10,
		},
	}

	users := []entity.User{
		{
			SQLModel:   core.NewSQLModel(),
			FirstName:  "John",
			LastName:   "Doe",
			Email:      "john.doe@example.com",
			SystemRole: entity.RoleUser,
			Status:     entity.StatusActive,
		},
	}
	users[0].Id = 1

	mockRepo.On("AdminListUsers", ctx, filter).Return(users, nil)

	response, err := biz.ListUsers(ctx, filter)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Data, 1)
	assert.Equal(t, "John", response.Data[0].FirstName)
	mockRepo.AssertExpectations(t)
}

func TestListUsers_RepositoryError(t *testing.T) {
	mockRepo := new(MockAdminUserRepository)
	mockPasswordBiz := new(MockAdminUserPasswordBusiness)
	biz := NewAdminUserBusiness(mockRepo, mockPasswordBiz)

	ctx := context.Background()
	filter := &entity.ListUsersFilter{
		Paging: core.Paging{
			Page:  1,
			Limit: 10,
		},
	}

	mockRepo.On("AdminListUsers", ctx, filter).Return(nil, errors.New("database error"))

	response, err := biz.ListUsers(ctx, filter)

	assert.Error(t, err)
	assert.Nil(t, response)
	mockRepo.AssertExpectations(t)
}

func TestGetUserByID_Success(t *testing.T) {
	mockRepo := new(MockAdminUserRepository)
	mockPasswordBiz := new(MockAdminUserPasswordBusiness)
	biz := NewAdminUserBusiness(mockRepo, mockPasswordBiz)

	ctx := context.Background()
	userID := 1

	expectedUser := &entity.User{
		SQLModel:   core.NewSQLModel(),
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john.doe@example.com",
		SystemRole: entity.RoleUser,
		Status:     entity.StatusActive,
	}
	expectedUser.Id = userID

	mockRepo.On("AdminGetUserByID", ctx, userID).Return(expectedUser, nil)

	user, err := biz.GetUserByID(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.Email, user.Email)
	mockRepo.AssertExpectations(t)
}

func TestGetUserByID_NotFound(t *testing.T) {
	mockRepo := new(MockAdminUserRepository)
	mockPasswordBiz := new(MockAdminUserPasswordBusiness)
	biz := NewAdminUserBusiness(mockRepo, mockPasswordBiz)

	ctx := context.Background()
	userID := 999

	mockRepo.On("AdminGetUserByID", ctx, userID).Return(nil, entity.ErrUserNotFound)

	user, err := biz.GetUserByID(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}
