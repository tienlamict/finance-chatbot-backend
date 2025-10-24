package business

import (
	"context"
	"finance-chatbot/addon/common"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/auth/entity"
	userEntity "finance-chatbot/microservice/user/entity"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthRepository interface {
	AddNewAuth(ctx context.Context, data *entity.Auth) error
	GetAuth(ctx context.Context, email string) (*entity.Auth, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, firstName, lastName, email string) (newId int, err error)
}

type RBACBusiness interface {
	AssignRolesToUser(ctx context.Context, userID int, req *userEntity.AssignRolesRequest) error
}

type Hasher interface {
	RandomStr(length int) (string, error)
	HashPassword(salt, password string) (string, error)
	CompareHashPassword(hashedPassword, salt, password string) bool
}

type business struct {
	repository     AuthRepository
	userRepository UserRepository
	rbacBusiness   RBACBusiness
	jwtProvider    common.JWTProvider
	hasher         Hasher
}

func NewBusiness(repository AuthRepository, userRepository UserRepository,
	rbacBusiness RBACBusiness, jwtProvider common.JWTProvider, hasher Hasher) *business {
	return &business{
		repository:     repository,
		userRepository: userRepository,
		rbacBusiness:   rbacBusiness,
		jwtProvider:    jwtProvider,
		hasher:         hasher,
	}
}

func (biz *business) Login(ctx context.Context, data *entity.AuthEmailPassword) (*entity.TokenResponse, error) {
	if err := data.Validate(); err != nil {
		return nil, core.ErrBadRequest.WithError(err.Error())
	}

	authData, err := biz.repository.GetAuth(ctx, data.Email)

	if err != nil {
		return nil, core.ErrBadRequest.WithError(entity.ErrLoginFailed.Error()).WithDebug(err.Error())
	}

	if !biz.hasher.CompareHashPassword(authData.Password, authData.Salt, data.Password) {
		return nil, core.ErrBadRequest.WithError(entity.ErrLoginFailed.Error())
	}

	uid := core.NewUID(uint32(authData.UserId), 1, 1)
	sub := uid.String()
	tid := uuid.New().String()

	tokenStr, expSecs, err := biz.jwtProvider.IssueToken(ctx, tid, sub)

	if err != nil {
		return nil, core.ErrInternalServerError.WithError(entity.ErrLoginFailed.Error()).WithDebug(err.Error())
	}

	return &entity.TokenResponse{
		AccessToken: entity.Token{
			Token:     tokenStr,
			ExpiredIn: expSecs,
		},
	}, nil
}

func (biz *business) Register(ctx context.Context, data *entity.AuthRegister) error {
	if err := data.Validate(); err != nil {
		return core.ErrBadRequest.WithError(err.Error())
	}

	_, err := biz.repository.GetAuth(ctx, data.Email)

	if err == nil {
		return core.ErrBadRequest.WithError(entity.ErrEmailHasExisted.Error())
	} else if err != core.ErrRecordNotFound {
		return core.ErrInternalServerError.WithError(entity.ErrCannotRegister.Error()).WithDebug(err.Error())
	}

	newUserId, err := biz.userRepository.CreateUser(ctx, data.FirstName, data.LastName, data.Email)

	if err != nil {
		return core.ErrInternalServerError.WithError(entity.ErrCannotRegister.Error()).WithDebug(err.Error())
	}

	salt, err := biz.hasher.RandomStr(16)

	if err != nil {
		return core.ErrInternalServerError.WithError(entity.ErrCannotRegister.Error()).WithDebug(err.Error())
	}

	passHashed, err := biz.hasher.HashPassword(salt, data.Password)

	if err != nil {
		return core.ErrInternalServerError.WithError(entity.ErrCannotRegister.Error()).WithDebug(err.Error())
	}

	newAuth := entity.NewAuthWithEmailPassword(newUserId, data.Email, salt, passHashed)

	if err := biz.repository.AddNewAuth(ctx, &newAuth); err != nil {
		return core.ErrInternalServerError.WithError(entity.ErrCannotRegister.Error()).WithDebug(err.Error())
	}

	// Assign default role (role_id = 3) to the new user
	assignRolesReq := &userEntity.AssignRolesRequest{
		RoleIDs: []int{3}, // Default role for regular users
	}

	if err := biz.rbacBusiness.AssignRolesToUser(ctx, newUserId, assignRolesReq); err != nil {
		// Log the error but don't fail the registration since user is already created
		// In production, you might want to handle this differently (e.g., retry or rollback)
		return core.ErrInternalServerError.WithError("User registered but failed to assign default role").WithDebug(err.Error())
	}

	return nil
}

func (biz *business) IntrospectToken(ctx context.Context, accessToken string) (*jwt.RegisteredClaims, error) {
	claims, err := biz.jwtProvider.ParseToken(ctx, accessToken)

	if err != nil {
		return nil, core.ErrUnauthorized.WithDebug(err.Error())
	}

	return claims, nil
}
