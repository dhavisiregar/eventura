package middleware

import (
	"net/http"
	"strings"

	"eventman/backend/internal/models"
	"eventman/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

const (
	CtxUserID = "userID"
	CtxRole   = "userRole"
)

// RequireAuth validates the Bearer JWT and stores userID/role in the context.
func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			utils.Error(c, http.StatusUnauthorized, "missing or invalid authorization header")
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := utils.ParseToken(secret, tokenStr)
		if err != nil {
			utils.Error(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxRole, claims.Role)
		c.Next()
	}
}

// RequireRole restricts an already-authenticated request to the given roles.
func RequireRole(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(CtxRole)
		if !exists {
			utils.Error(c, http.StatusUnauthorized, "unauthenticated")
			return
		}
		role := roleVal.(models.Role)
		for _, r := range roles {
			if r == role {
				c.Next()
				return
			}
		}
		utils.Error(c, http.StatusForbidden, "you do not have permission to perform this action")
	}
}

func UserID(c *gin.Context) uint {
	v, _ := c.Get(CtxUserID)
	id, _ := v.(uint)
	return id
}

func UserRole(c *gin.Context) models.Role {
	v, _ := c.Get(CtxRole)
	role, _ := v.(models.Role)
	return role
}
