package middleware

import (
	"context"
	"finance-chatbot/addon/common"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/user/entity"

	"github.com/gin-gonic/gin"
)

// RBACClient defines capability to verify if a user has at least one of required permissions.
type RBACClient interface {
	HasAnyPermission(userID int, permCodes ...string) (bool, error)
}

// UserStore defines capability to get user information
type UserStore interface {
	GetUserByID(ctx context.Context, userID int) (*entity.User, error)
}

// RequirePermissions ensures the requester has ANY of the provided permission codes.
func RequirePermissions(rbac RBACClient, permCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requester, ok := c.Get(core.KeyRequester)
		if !ok {
			common.WriteErrorResponse(c, core.ErrUnauthorized.WithError("missing requester"))
			c.Abort()
			return
		}

		req := requester.(core.Requester)

		uid, err := core.FromBase58(req.GetSubject())
		if err != nil {
			common.WriteErrorResponse(c, core.ErrUnauthorized.WithDebug(err.Error()))
			c.Abort()
			return
		}

		userID := int(uid.GetLocalID())

		okPerm, err := rbac.HasAnyPermission(userID, permCodes...)
		if err != nil {
			common.WriteErrorResponse(c, core.ErrInternalServerError.WithDebug(err.Error()))
			c.Abort()
			return
		}

		if !okPerm {
			common.WriteErrorResponse(c, core.ErrForbidden.WithError("insufficient permissions"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSuperAdmin ensures the requester has superadmin role
func RequireSuperAdmin(userStore UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		requester, ok := c.Get(core.KeyRequester)
		if !ok {
			common.WriteErrorResponse(c, core.ErrUnauthorized.WithError("missing requester"))
			c.Abort()
			return
		}

		req := requester.(core.Requester)

		uid, err := core.FromBase58(req.GetSubject())
		if err != nil {
			common.WriteErrorResponse(c, core.ErrUnauthorized.WithDebug(err.Error()))
			c.Abort()
			return
		}

		userID := int(uid.GetLocalID())

		user, err := userStore.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			common.WriteErrorResponse(c, core.ErrUnauthorized.WithDebug(err.Error()))
			c.Abort()
			return
		}

		if user.SystemRole != entity.RoleSuperAdmin {
			common.WriteErrorResponse(c, core.ErrForbidden.WithError("superadmin access required"))
			c.Abort()
			return
		}

		c.Next()
	}
}
