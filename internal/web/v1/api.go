package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/tg123/go-htpasswd"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/service"
)

func SetupV1Endpoints(r *gin.RouterGroup) {
	r.POST("/login", loginHandler)
	r.GET("/health", healthHandler)

	v1Group := r.Group("/v1")
	v1Group.POST("/login", loginHandler)
	v1Group.GET("/health", healthHandler)
	v1Group.GET("/version", getVersion)

	protected := v1Group.Group("/", authMiddleware(), loggingMiddleware())
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
}

func setupSystemRoutes(g *gin.RouterGroup) {
	system := g.Group("/system")
	system.GET("/info", getSystemInfo)
	system.GET("/interfaces", getInterfaces)
	system.GET("/modules", getModules)
}

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
