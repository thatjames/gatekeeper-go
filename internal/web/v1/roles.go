package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const defaultRoleClaim = "groups"

// RoleFromClaims maps the values of the named claim to a gatekeeper role.
// Without a roles map, the values "admin" and "viewer" map to themselves.
// When several values map, admin wins. ok is false when none maps.
func RoleFromClaims(claims map[string]any, claim string, roles map[string]string) (Role, bool) {
	if claim == "" {
		claim = defaultRoleClaim
	}
	var values []string
	switch v := claims[claim].(type) {
	case string:
		values = []string{v}
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok {
				values = append(values, s)
			}
		}
	}

	found := false
	for _, value := range values {
		mapped := value
		if roles != nil {
			mapped = roles[value]
		}
		switch Role(mapped) {
		case AdminRole:
			return AdminRole, true
		case ViewerRole:
			found = true
		}
	}
	if found {
		return ViewerRole, true
	}
	return "", false
}

// RoleMiddleware runs after AuthMiddleware. Both roles may read (GET, HEAD);
// any other method needs admin. A token with an unknown role is refused.
func RoleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("user")
		role := user.(User).Role
		if role != AdminRole && role != ViewerRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead && role != AdminRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		c.Next()
	}
}
