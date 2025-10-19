package business

import (
	"context"
	"crypto/rand"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/user/entity"
	"math/big"
)

type AuthRepository interface {
	CreateAuthForUser(ctx context.Context, userID int, email, salt, hashedPassword string) error
	UpdateUserPassword(ctx context.Context, userID int, salt, hashedPassword string) error
	GetAuthByUserID(ctx context.Context, userID int) (interface{}, error)
}

type Hasher interface {
	RandomStr(length int) (string, error)
	HashPassword(salt, password string) (string, error)
}

// GenerateTemporaryPassword generates a secure random password
func GenerateTemporaryPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)

	for i := range password {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password[i] = charset[num.Int64()]
	}

	return string(password), nil
}

type AdminUserPasswordBusinessInterface interface {
	CreateUserWithPassword(ctx context.Context, req *entity.AdminCreateUserRequest) (*entity.AdminCreateUserResponse, error)
	SetUserPassword(ctx context.Context, userID int, req *entity.AdminSetUserPasswordRequest) error
}

type AdminUserPasswordBusiness struct {
	userRepo AdminUserRepository
	authRepo AuthRepository
	hasher   Hasher
}

func NewAdminUserPasswordBusiness(userRepo AdminUserRepository, authRepo AuthRepository, hasher Hasher) *AdminUserPasswordBusiness {
	return &AdminUserPasswordBusiness{
		userRepo: userRepo,
		authRepo: authRepo,
		hasher:   hasher,
	}
}

// CreateUserWithPassword creates user with auto-generated password
func (biz *AdminUserPasswordBusiness) CreateUserWithPassword(
	ctx context.Context,
	req *entity.AdminCreateUserRequest,
) (*entity.AdminCreateUserResponse, error) {

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, core.ErrBadRequest.WithError(err.Error())
	}

	// Create user in users table
	user, err := biz.userRepo.AdminCreateUser(ctx, req)
	if err != nil {
		if err == entity.ErrUserAlreadyExists {
			return nil, core.ErrBadRequest.WithError(entity.ErrUserAlreadyExists.Error())
		}
		return nil, core.ErrInternalServerError.
			WithError(entity.ErrCannotCreateUser.Error()).
			WithDebug(err.Error())
	}

	// Generate temporary password
	tempPassword, err := GenerateTemporaryPassword(12)
	if err != nil {
		return nil, core.ErrInternalServerError.
			WithError("failed to generate password").
			WithDebug(err.Error())
	}

	// Generate salt
	salt, err := biz.hasher.RandomStr(16)
	if err != nil {
		return nil, core.ErrInternalServerError.
			WithError("failed to generate salt").
			WithDebug(err.Error())
	}

	// Hash password
	hashedPassword, err := biz.hasher.HashPassword(salt, tempPassword)
	if err != nil {
		return nil, core.ErrInternalServerError.
			WithError("failed to hash password").
			WithDebug(err.Error())
	}

	// Create auth record
	err = biz.authRepo.CreateAuthForUser(ctx, user.Id, user.Email, salt, hashedPassword)
	if err != nil {
		return nil, core.ErrInternalServerError.
			WithError("failed to create authentication").
			WithDebug(err.Error())
	}

	user.Mask()

	return &entity.AdminCreateUserResponse{
		User:              *user,
		TemporaryPassword: tempPassword,
		Message:           "User created successfully. Please provide the temporary password to the user.",
	}, nil
}

// SetUserPassword allows admin to set/reset a user's password
func (biz *AdminUserPasswordBusiness) SetUserPassword(
	ctx context.Context,
	userID int,
	req *entity.AdminSetUserPasswordRequest,
) error {

	if err := req.Validate(); err != nil {
		return core.ErrBadRequest.WithError(err.Error())
	}

	// Verify user exists
	user, err := biz.userRepo.AdminGetUserByID(ctx, userID)
	if err != nil {
		if err == entity.ErrUserNotFound {
			return core.ErrNotFound.WithError(entity.ErrUserNotFound.Error())
		}
		return core.ErrInternalServerError.
			WithError("failed to get user").
			WithDebug(err.Error())
	}

	// Generate new salt
	salt, err := biz.hasher.RandomStr(16)
	if err != nil {
		return core.ErrInternalServerError.
			WithError("failed to generate salt").
			WithDebug(err.Error())
	}

	// Hash new password
	hashedPassword, err := biz.hasher.HashPassword(salt, req.NewPassword)
	if err != nil {
		return core.ErrInternalServerError.
			WithError("failed to hash password").
			WithDebug(err.Error())
	}

	// Check if auth exists
	_, err = biz.authRepo.GetAuthByUserID(ctx, userID)
	if err != nil {
		// Auth doesn't exist, create new one
		err = biz.authRepo.CreateAuthForUser(ctx, userID, user.Email, salt, hashedPassword)
		if err != nil {
			return core.ErrInternalServerError.
				WithError("failed to create authentication").
				WithDebug(err.Error())
		}
	} else {
		// Auth exists, update password
		err = biz.authRepo.UpdateUserPassword(ctx, userID, salt, hashedPassword)
		if err != nil {
			return core.ErrInternalServerError.
				WithError("failed to update password").
				WithDebug(err.Error())
		}
	}

	// TODO: Send email to user if req.SendEmail is true

	return nil
}
