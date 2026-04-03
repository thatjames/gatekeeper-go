package v1

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	badrand "math/rand"
	"net/http"
	"os"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/tg123/go-htpasswd"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/service"
	"golang.org/x/oauth2"
)

var (
	oidcProvider     *oidc.Provider
	oidcVerifier     *oidc.IDTokenVerifier
	oidcOAuth2Config oauth2.Config
	oidcOnce         sync.Once
	oidcInitErr      error
)

func initOIDC() error {
	oidcOnce.Do(func() {
		cfg, ok := config.Config.Auth.(*config.OIDCAuth)
		if !ok {
			oidcInitErr = fmt.Errorf("auth config is not OIDCAuth")
			return
		}

		provider, err := oidc.NewProvider(context.Background(), cfg.IssuerURL)
		if err != nil {
			oidcInitErr = fmt.Errorf("failed to create OIDC provider: %w", err)
			return
		}

		scopes := []string{oidc.ScopeOpenID, "profile", "email"}
		if len(cfg.Scopes) > 0 {
			scopes = cfg.Scopes
		}

		oidcProvider = provider
		oidcVerifier = provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})
		oidcOAuth2Config = oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: os.Getenv(cfg.ClientSecretVar),
			RedirectURL:  cfg.RedirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       scopes,
		}
	})
	return oidcInitErr
}

func SetupV1Endpoints(r *gin.RouterGroup) {
	r.GET("/health", healthHandler)

	v1Group := r.Group("/v1")
	v1Group.GET("/features", getFeatures)

	if config.Config.Auth != nil {
		switch config.Config.Auth.Type() {
		case "oidc":
			if err := initOIDC(); err != nil {
				log.WithError(err).Error("Failed to initialise OIDC")
			} else {
				v1Group.GET("/oidc-login", oidcLoginHandler)
				v1Group.GET("/auth/callback", oidcCallbackHandler)
				v1Group.POST("/logout", logoutHandler)
				log.Debug("OIDC login enabled")
			}
		default:
			log.Debug("Using default login handler")
		}
	}
	v1Group.POST("/login", defaultLoginHandler)

	v1Group.GET("/health", healthHandler)
	v1Group.GET("/version", getVersion)

	protected := v1Group.Group("/", loggingMiddleware(), authMiddleware())
	if service.IsRegistered(service.DHCP) {
		log.Info("Registering DHCP endpoints")
		setupDHCPRoutes(protected)
	}
	if service.IsRegistered(service.DNS) {
		log.Info("Registering DNS endpoints")
		setupDNSRoutes(protected)
	}
	setupSystemRoutes(protected)
}

func setupDHCPRoutes(g *gin.RouterGroup) {
	dhcp := g.Group("/dhcp")
	dhcp.GET("/leases", getLeases)
	dhcp.DELETE("/leases/:clientId", deleteLease)
	dhcp.POST("/leases/reserve", reserveLease)
	dhcp.PUT("/leases", updateLease)
	dhcp.GET("/options", getDHCPOptions)
	dhcp.PUT("/options", updateDHCPOptions)
}

func setupDNSRoutes(g *gin.RouterGroup) {
	dns := g.Group("/dns")
	dns.GET("/config", getDNSConfig)
	dns.PUT("/config", updateDNSConfig)
	dns.GET("/local-domains", getLocalDomains)
	dns.POST("/local-domains", addLocalDomain)
	dns.PUT("/local-domains/:domain", updateLocalDomain)
	dns.DELETE("/local-domains/:domain", deleteLocalDomain)
	dns.POST("/blocklist", addBlocklist)
	dns.DELETE("/blocklist/:id", deleteBlocklist)
	dns.PUT("/blockeddomains", addBlockedDomain)
	dns.DELETE("/blockeddomains/:id", deleteBlockedDomain)
}

func setupSystemRoutes(g *gin.RouterGroup) {
	system := g.Group("/system")
	system.GET("/info", getSystemInfo)
	system.GET("/interfaces", getInterfaces)
	system.GET("/modules", getModules)
}

func oidcLoginHandler(c *gin.Context) {
	state, err := generateState()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state"})
		return
	}
	c.SetCookie("oauth_state", state, 600, "/", "", true, true)
	c.Redirect(http.StatusFound, oidcOAuth2Config.AuthCodeURL(state))
}

func oidcCallbackHandler(c *gin.Context) {
	// Validate state
	cookieState, err := c.Cookie("oauth_state")
	if err != nil || cookieState != c.Query("state") {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}
	c.SetCookie("oauth_state", "", 600, "/", "", true, true)
	log.Debug("OIDC callback received")

	// Exchange code for tokens
	oauth2Token, err := oidcOAuth2Config.Exchange(c.Request.Context(), c.Query("code"))
	if err != nil {
		log.WithError(err).Error("Failed to exchange OIDC code")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to exchange token"})
		return
	}

	// Extract and verify ID token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing ID token"})
		return
	}
	idToken, err := oidcVerifier.Verify(c.Request.Context(), rawIDToken)
	if err != nil {
		log.WithError(err).Error("Failed to verify ID token")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid ID token"})
		return
	}

	// Extract claims
	var claims struct {
		Email   string   `json:"email"`
		Name    string   `json:"name"`
		Subject string   `json:"sub"`
		Roles   []string `json:"roles"`
	}
	if err := idToken.Claims(&claims); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse claims"})
		return
	}

	// Issue app token so authMiddleware is unchanged
	token, err := CreateAuthToken(claims.Email)
	if err != nil {
		log.WithError(err).Error("Failed to create auth token after OIDC")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}
	log.Debug("OIDC callback returning token")
	webURL := "http://localhost:5173"
	if config.Config.Web != nil && config.Config.Web.WebURL != "" {
		webURL = config.Config.Web.WebURL
	}
	c.SetCookie(
		"oauth_token",
		token,
		3600, // max age in seconds, match your JWT expiry
		"/",
		"",
		false,
		false,
	)

	c.Redirect(http.StatusFound, webURL)
}

func defaultLoginHandler(c *gin.Context) {
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
		password := generateRandomString(16)
		log.Warn("Using random password: ", password)
		authenticated = (req.Username == "admin" && req.Password == password)
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
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func logoutHandler(c *gin.Context) {
	c.SetCookie("oauth_token", "", -1, "/", "", false, false)
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateRandomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		ret[i] = letters[badrand.Intn(len(letters))]
	}
	return string(ret)
}
