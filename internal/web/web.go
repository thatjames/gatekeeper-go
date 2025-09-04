package web

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/tg123/go-htpasswd"

	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dhcp"
)

func setupAPIRoutes(r *gin.Engine) {
	api := r.Group("/api")
	api.POST("/login", loginHandler)
	api.GET("/health", healthHandler)

	protected := api.Group("/", authMiddleware())
	protected.GET("/verify", verifyHandler)
	protected.GET("/version", versionHandler)
}

func Init(ver string, cfg *config.Web, leases *dhcp.LeaseDB) error {
	version = ver
	leaseDB = leases

	r := gin.New()

	r.Use(loggingMiddleware())
	r.Use(gin.Recovery())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	r.Use(cors.New(corsConfig))

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	setupAPIRoutes(r)

	r.Use(spaMiddleware())
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
	var req UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	log.Info("Login request: ", req)

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

	token, err := CreateAuthToken(req.Username)
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
