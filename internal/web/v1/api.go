package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/tg123/go-htpasswd"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dhcp"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/service"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/system"
)

func LoginHandler(c *gin.Context) {
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

func VerifyHandler(c *gin.Context) {
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

func GetLeases(c *gin.Context) {
	dhcpService := service.GetService[*dhcp.DHCPServer](service.DHCP)
	activeLeases := dhcpService.LeaseDB().ActiveLeases()
	reservedLeases := dhcpService.LeaseDB().ReservedLeases()
	c.JSON(http.StatusOK, DhcpLeaseResponse{
		ActiveLeases:   MapLeases(activeLeases),
		ReservedLeases: MapLeases(reservedLeases),
	})
}

func GetDHCPOptions(c *gin.Context) {
	dhcpService := service.GetService[*dhcp.DHCPServer](service.DHCP)
	opts := dhcpService.Options()
	c.JSON(http.StatusOK, DhcpOptionsResponse{
		Interface:      opts.Interface,
		StartAddr:      opts.StartFrom.String(),
		EndAddr:        opts.EndAt.String(),
		LeaseTTL:       opts.LeaseTTL,
		Router:         opts.Router.String(),
		SubnetMask:     opts.SubnetMask.String(),
		DomainName:     opts.DomainName,
		ReservedLeases: opts.ReservedLeases,
		LeaseFile:      opts.LeaseFile,
	})
}

func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	})
}

func GetSystemInfo(c *gin.Context) {
	sysInfo, err := system.GetSystemInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sysInfo)
}
