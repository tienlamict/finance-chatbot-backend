package api

import (
	"context"
	"finance-chatbot/addon/common"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/user/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminUserBusiness interface {
	CreateUser(ctx context.Context, req *entity.AdminCreateUserRequest) (*entity.AdminCreateUserResponse, error)
	UpdateUser(ctx context.Context, userID int, req *entity.AdminUpdateUserRequest) error
	DeleteUser(ctx context.Context, userID int) error
	ListUsers(ctx context.Context, filter *entity.ListUsersFilter) (*entity.UserListResponse, error)
	GetUserByID(ctx context.Context, userID int) (*entity.User, error)
}

type adminUserAPI struct {
	business AdminUserBusiness
}

func NewAdminUserAPI(business AdminUserBusiness) *adminUserAPI {
	return &adminUserAPI{business: business}
}

// CreateUserHdl creates a new user with auto-generated password
// POST /api/v1/admin/users
func (api *adminUserAPI) CreateUserHdl() func(c *gin.Context) {
	return func(c *gin.Context) {
		var req entity.AdminCreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError(err.Error()))
			return
		}

		response, err := api.business.CreateUser(c.Request.Context(), &req)
		if err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		// Mask the user ID in response
		response.User.Mask()

		c.JSON(http.StatusCreated, core.ResponseData(response))
	}
}

// UpdateUserHdl updates an existing user
// PUT /api/v1/admin/users/:id or PATCH /api/v1/admin/users/:id
func (api *adminUserAPI) UpdateUserHdl() func(c *gin.Context) {
	return func(c *gin.Context) {
		// Decode base58 UID to real ID
		uid, err := core.FromBase58(c.Param("id"))
		if err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError("invalid user ID format"))
			return
		}
		userID := int(uid.GetLocalID())

		var req entity.AdminUpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError(err.Error()))
			return
		}

		if err := api.business.UpdateUser(c.Request.Context(), userID, &req); err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, core.SimpleSuccessResponse(true))
	}
}

// DeleteUserHdl deletes a user
// DELETE /api/v1/admin/users/:id
func (api *adminUserAPI) DeleteUserHdl() func(c *gin.Context) {
	return func(c *gin.Context) {
		// Decode base58 UID to real ID
		uid, err := core.FromBase58(c.Param("id"))
		if err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError("invalid user ID format"))
			return
		}
		userID := int(uid.GetLocalID())

		if err := api.business.DeleteUser(c.Request.Context(), userID); err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, core.SimpleSuccessResponse(true))
	}
}

// ListUsersHdl lists all users with pagination and filters
// GET /api/v1/admin/users
func (api *adminUserAPI) ListUsersHdl() func(c *gin.Context) {
	return func(c *gin.Context) {
		var filter entity.ListUsersFilter
		if err := c.ShouldBindQuery(&filter); err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError(err.Error()))
			return
		}

		response, err := api.business.ListUsers(c.Request.Context(), &filter)
		if err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		// Mask all user IDs
		for i := range response.Data {
			response.Data[i].Mask()
		}

		c.JSON(http.StatusOK, core.ResponseData(response))
	}
}

// GetUserByIDHdl gets a user by ID
// GET /api/v1/admin/users/:id
func (api *adminUserAPI) GetUserByIDHdl() func(c *gin.Context) {
	return func(c *gin.Context) {
		// Decode base58 UID to real ID
		uid, err := core.FromBase58(c.Param("id"))
		if err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError("invalid user ID format"))
			return
		}
		userID := int(uid.GetLocalID())

		user, err := api.business.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		user.Mask()

		c.JSON(http.StatusOK, core.ResponseData(user))
	}
}
