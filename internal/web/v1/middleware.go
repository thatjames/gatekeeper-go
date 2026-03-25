package v1

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var rawToken string

		if header := c.GetHeader("Authorization"); strings.HasPrefix(header, "Bearer ") {
			log.Trace("Bearer token found in header")
			rawToken = strings.TrimPrefix(header, "Bearer ")
		} else if cookie, err := c.Cookie("oauth_token"); err == nil {
			log.Trace("Bearer token found in cookie")
			rawToken = cookie
		} else {
			fmt.Println(err)
			log.Trace("No bearer token found")
		}

		if rawToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		claims, err := ParseAuthToken(rawToken)
		if err != nil {
			log.WithError(err).Error("Failed to parse token")
			c.SetCookie("oauth_token", "", -1, "/", "", false, false)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}

func loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		log.WithFields(log.Fields{
			"status":     param.StatusCode,
			"method":     param.Method,
			"path":       param.Path,
			"ip":         param.ClientIP,
			"user_agent": param.Request.UserAgent(),
			"duration":   param.Latency,
		}).Info("HTTP Request")
		return ""
	})
}
