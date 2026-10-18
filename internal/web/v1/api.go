package v1

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
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

	// fallbackPassword is set at startup when no htpasswd file can be read;
	// admin logs in with it for the life of the process.
	fallbackPassword string
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
	initLogin()
	v1Group.POST("/login", defaultLoginHandler)

	v1Group.GET("/health", healthHandler)
	v1Group.GET("/version", getVersion)

	protected := v1Group.Group("/", loggingMiddleware(), AuthMiddleware())
	protected.GET("/me", meHandler)
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

	// Issue app token so AuthMiddleware is unchanged
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
	setTokenCookie(c, token, int(tokenLife.Seconds()))

	c.Redirect(http.StatusFound, webURL)
}

// initLogin checks once whether the htpasswd file can be used. Without one,
// it makes a password for admin and logs it, once.
func initLogin() {
	fallbackPassword = ""
	file := ""
	if config.Config.Web != nil {
		file = config.Config.Web.HTPasswdFile
	}
	_, err := htpasswd.New(file, htpasswd.DefaultSystems, nil)
	if err == nil {
		return
	}
	log.Warn("Unable to read htpasswd file: ", err.Error())
	password, err := generateRandomString(16)
	if err != nil {
		log.WithError(err).Error("Failed to generate fallback password; password login is disabled")
		return
	}
	fallbackPassword = password
	log.Warn("No htpasswd file; log in as admin with password: ", fallbackPassword)
}

func defaultLoginHandler(c *gin.Context) {
	var req UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	logger := log.WithField("username", req.Username)
	logger.Info("Login attempt")

	authenticated := false
	if fallbackPassword != "" {
		authenticated = req.Username == "admin" &&
			subtle.ConstantTimeCompare([]byte(req.Password), []byte(fallbackPassword)) == 1
	} else if passwd, err := htpasswd.New(config.Config.Web.HTPasswdFile, htpasswd.DefaultSystems, nil); err == nil {
		authenticated = passwd.Match(req.Username, req.Password)
	} else {
		logger.Error("Unable to read htpasswd file: ", err.Error())
	}

	if !authenticated {
		logger.Warn("Login failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := CreateAuthToken(req.Username)
	if err != nil {
		logger.Error("Failed to create auth token: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}
	logger.Info("Login succeeded")
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func meHandler(c *gin.Context) {
	c.JSON(http.StatusOK, c.MustGet("user"))
}

func logoutHandler(c *gin.Context) {
	setTokenCookie(c, "", -1)
}

// setTokenCookie sets the oauth_token cookie. Scripts cannot read it; it is
// Secure when the UI is served over https.
func setTokenCookie(c *gin.Context, value string, maxAge int) {
	secure := config.Config.Web != nil && strings.HasPrefix(config.Config.Web.WebURL, "https://")
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_token", value, maxAge, "/", "", secure, true)
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateRandomString(length int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	ret := make([]byte, length)
	for i := range ret {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[n.Int64()]
	}
	return string(ret), nil
}
