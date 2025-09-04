package web

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func spaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") ||
			strings.HasPrefix(c.Request.URL.Path, "/metrics") {
			c.Next()
			return
		}

		if strings.Contains(c.Request.URL.Path, ".") {
			c.Next()
			return
		}

		c.Request.URL.Path = "/"
		c.Next()
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, err := ParseAuthToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
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
