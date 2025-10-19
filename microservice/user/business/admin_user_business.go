package business

import (
	"context"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/user/entity"
)

type AdminUserRepository interface {
	AdminCreateUser(ctx context.Context, req *entity.AdminCreateUserRequest) (*entity.User, error)
	AdminUpdateUser(ctx context.Context, userID int, req *entity.AdminUpdateUserRequest) error
	AdminDeleteUser(ctx context.Context, userID int) error
	AdminListUsers(ctx context.Context, filter *entity.ListUsersFilter) ([]entity.User, error)
	AdminGetUserByID(ctx context.Context, userID int) (*entity.User, error)
}

type adminUserBusiness struct {
	repository       AdminUserRepository
	passwordBusiness AdminUserPasswordBusinessInterface
}

func NewAdminUserBusiness(repository AdminUserRepository, passwordBusiness AdminUserPasswordBusinessInterface) *adminUserBusiness {
	return &adminUserBusiness{
		repository:       repository,
		passwordBusiness: passwordBusiness,
	}
}

// CreateUser creates a new user with auto-generated password
func (biz *adminUserBusiness) CreateUser(ctx context.Context, req *entity.AdminCreateUserRequest) (*entity.AdminCreateUserResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, core.ErrBadRequest.
			WithError(err.Error())
	}

	// Create user with password using password business
	response, err := biz.passwordBusiness.CreateUserWithPassword(ctx, req)
	if err != nil {
		if err == entity.ErrUserAlreadyExists {
			return nil, core.ErrBadRequest.
				WithError(entity.ErrUserAlreadyExists.Error())
		}

		return nil, core.ErrInternalServerError.
			WithError(entity.ErrCannotCreateUser.Error()).
			WithDebug(err.Error())
	}

	return response, nil
}

// CreateUserWithoutPassword creates a new user without password (legacy method)
func (biz *adminUserBusiness) CreateUserWithoutPassword(ctx context.Context, req *entity.AdminCreateUserRequest) (*entity.User, error) {
	if err := req.Validate(); err != nil {
		return nil, core.ErrBadRequest.
			WithError(err.Error())
	}

	user, err := biz.repository.AdminCreateUser(ctx, req)
	if err != nil {
		if err == entity.ErrUserAlreadyExists {
			return nil, core.ErrBadRequest.
				WithError(entity.ErrUserAlreadyExists.Error())
		}

		return nil, core.ErrInternalServerError.
			WithError(entity.ErrCannotCreateUser.Error()).
			WithDebug(err.Error())
	}

	return user, nil
}

// UpdateUser updates an existing user
func (biz *adminUserBusiness) UpdateUser(ctx context.Context, userID int, req *entity.AdminUpdateUserRequest) error {
	if err := req.Validate(); err != nil {
		return core.ErrBadRequest.
			WithError(err.Error())
	}

	err := biz.repository.AdminUpdateUser(ctx, userID, req)
	if err != nil {
		if err == entity.ErrUserNotFound {
			return core.ErrNotFound.
				WithError(entity.ErrUserNotFound.Error())
		}

		if err == entity.ErrUserAlreadyExists {
			return core.ErrBadRequest.
				WithError(entity.ErrUserAlreadyExists.Error())
		}

		return core.ErrInternalServerError.
			WithError(entity.ErrCannotUpdateUser.Error()).
			WithDebug(err.Error())
	}

	return nil
}

// DeleteUser deletes a user
func (biz *adminUserBusiness) DeleteUser(ctx context.Context, userID int) error {
	err := biz.repository.AdminDeleteUser(ctx, userID)
	if err != nil {
		if err == entity.ErrUserNotFound {
			return core.ErrNotFound.
				WithError(entity.ErrUserNotFound.Error())
		}

		return core.ErrInternalServerError.
			WithError(entity.ErrCannotDeleteUser.Error()).
			WithDebug(err.Error())
	}

	return nil
}

// ListUsers lists all users with pagination and filters
func (biz *adminUserBusiness) ListUsers(ctx context.Context, filter *entity.ListUsersFilter) (*entity.UserListResponse, error) {
	filter.Process()

	users, err := biz.repository.AdminListUsers(ctx, filter)
	if err != nil {
		return nil, core.ErrInternalServerError.
			WithError(entity.ErrCannotGetUsers.Error()).
			WithDebug(err.Error())
	}

	// Convert to list items
	items := make([]entity.UserListItem, len(users))
	for i, user := range users {
		items[i] = entity.UserListItem{User: user}
	}

	response := &entity.UserListResponse{
		Data:   items,
		Paging: filter.Paging,
	}

	return response, nil
}

// GetUserByID gets a user by ID
func (biz *adminUserBusiness) GetUserByID(ctx context.Context, userID int) (*entity.User, error) {
	user, err := biz.repository.AdminGetUserByID(ctx, userID)
	if err != nil {
		if err == entity.ErrUserNotFound {
			return nil, core.ErrNotFound.
				WithError(entity.ErrUserNotFound.Error())
		}

		return nil, core.ErrInternalServerError.
			WithError(entity.ErrCannotGetUser.Error()).
			WithDebug(err.Error())
	}

	return user, nil
}
