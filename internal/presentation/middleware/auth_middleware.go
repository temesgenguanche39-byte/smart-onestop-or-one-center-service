package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smart-onestop/platform/internal/domain"
	"github.com/smart-onestop/platform/internal/infrastructure/security"
)

const (
	ContextKeyClaims = "jwt_claims"
	ContextKeyUserID = "user_id"
	ContextKeyRole   = "user_role"
)

// AuthRequired verifies Bearer JWT token in Authorization header
func AuthRequired(jwtManager *security.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Expected 'Bearer <token>'",
			})
			return
		}

		claims, err := jwtManager.Verify(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		c.Set(ContextKeyClaims, claims)
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}

// RequireRoles declarative RBAC authorization filter matching Section 6.2 of specification
func RequireRoles(allowedRoles ...domain.UserRole) gin.HandlerFunc {
	roleMap := make(map[domain.UserRole]bool)
	for _, r := range allowedRoles {
		roleMap[r] = true
	}

	return func(c *gin.Context) {
		roleVal, exists := c.Get(ContextKeyRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			return
		}

		userRole, ok := roleVal.(domain.UserRole)
		if !ok || (!roleMap[userRole] && userRole != domain.RoleSuperAdmin) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied: insufficient administrative permissions",
			})
			return
		}

		c.Next()
	}
}

func GetCurrentUserID(c *gin.Context) uuid.UUID {
	val, exists := c.Get(ContextKeyUserID)
	if !exists {
		return uuid.Nil
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return id
}
