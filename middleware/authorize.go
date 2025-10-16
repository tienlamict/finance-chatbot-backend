package middleware

import (
	"finance-chatbot/addon/common"
	"finance-chatbot/addon/core"

	"github.com/gin-gonic/gin"
)

// RBACClient defines capability to verify if a user has at least one of required permissions.
type RBACClient interface {
	HasAnyPermission(userID int, permCodes ...string) (bool, error)
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
