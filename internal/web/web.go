package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/tg123/go-htpasswd"

	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dhcp"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/web/domain"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/web/security"
)

// Embed the entire built Svelte app
//
//go:embed ui/dist/*
var staticFiles embed.FS

var (
	leaseDB *dhcp.LeaseDB
	version string
)

type EmbeddedFileSystem struct {
	http.FileSystem
}

func (e EmbeddedFileSystem) Exists(prefix string, path string) bool {
	_, err := e.Open(path)
	return err == nil
}

func NewEmbeddedFS() static.ServeFileSystem {
	sub, err := fs.Sub(staticFiles, "ui/dist")
	if err != nil {
		log.Fatal("Failed to create embedded filesystem:", err)
	}

	return EmbeddedFileSystem{
		FileSystem: http.FS(sub),
	}
}

func SPAMiddleware() gin.HandlerFunc {
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

func AuthMiddleware() gin.HandlerFunc {
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

		// TODO: Implement token validation
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Store a placeholder username - TODO decode this from the token
		c.Set("username", "authenticated_user")
		c.Next()
	}
}

func LoggingMiddleware() gin.HandlerFunc {
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

func setupAPIRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/login", loginHandler)
		api.GET("/health", healthHandler)

		protected := api.Group("/", AuthMiddleware())
		{
			protected.GET("/verify", verifyHandler)
			protected.GET("/version", versionHandler)
		}
	}
}

func Init(ver string, cfg *config.Web, leases *dhcp.LeaseDB) error {
	version = ver
	leaseDB = leases

	r := gin.New()

	r.Use(LoggingMiddleware())
	r.Use(gin.Recovery())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	r.Use(cors.New(corsConfig))

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	setupAPIRoutes(r)

	r.Use(SPAMiddleware())
	r.Use(static.Serve("/", NewEmbeddedFS()))

	if cfg.TLS != nil {
		log.Info("Starting TLS server on ", cfg.Address)
		return r.RunTLS(cfg.Address, cfg.TLS.PublicKey, cfg.TLS.PrivateKey)
	} else {
		log.Info("Starting HTTP server on ", cfg.Address)
		return r.Run(cfg.Address)
	}
}

// API Handlers

func loginHandler(c *gin.Context) {
	var req domain.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	authenticated := false
	if passwd, err := htpasswd.New(config.Config.Web.HTPasswdFile, htpasswd.DefaultSystems, nil); err == nil {
		authenticated = passwd.Match(req.Username, req.Password)
	} else {
		log.Warn("Unable to read htpasswd file: ", err.Error())
		log.Warn("Defaulting to default username/password")
		authenticated = (req.Username == "admin" && req.Password == "admin")
	}

	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := security.CreateAuthToken(req.Username)
	if err != nil {
		log.Error("Failed to create auth token: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func verifyHandler(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No username in context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":    true,
		"username": username,
	})
}

func versionHandler(c *gin.Context) {
	c.String(http.StatusOK, version)
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   version,
	})
}
