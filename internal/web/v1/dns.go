package v1

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dns"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/service"
)

func getDNSConfig(c *gin.Context) {
	c.JSON(200, DNSConfigResponse{
		Upstreams: strings.Join(config.Config.DNS.UpstreamServers, ","),
		Interface: config.Config.DNS.Interface,
	})
}

func getLocalDomains(c *gin.Context) {
	c.JSON(200, config.Config.DNS.LocalDomains)
}

func addLocalDomain(c *gin.Context) {
	var req LocalDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if validationErrors := req.Validate(); validationErrors != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:  "Unable to add local domain",
			Fields: validationErrors,
		})
		return
	}

	log.Info("Adding local domain: ", req)

	dnsService := service.GetService[*dns.DNSServer](service.DNS)
	if err := dnsService.AddLocalDomain(req.Domain, req.IP); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	config.Config.DNS.LocalDomains[req.Domain] = req.IP
	if err := config.UpdateConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config.Config.DNS.LocalDomains)
}

func deleteLocalDomain(c *gin.Context) {
	domain := c.Param("domain")
	log.Info("Deleting local domain: ", domain)

	delete(config.Config.DNS.LocalDomains, domain)
	if err := config.UpdateConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	dnsService := service.GetService[*dns.DNSServer](service.DNS)
	dnsService.DeleteLocalDomain(domain)
	c.JSON(http.StatusOK, config.Config.DNS.LocalDomains)
}

func updateLocalDomain(c *gin.Context) {
	var req LocalDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	if validationErrors := req.Validate(); validationErrors != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:  "Unable to update local domain",
			Fields: validationErrors,
		})
		return
	}
	originalDomain := c.Param("domain")
	log.Info("Updating local domain: ", originalDomain)
	dnsService := service.GetService[*dns.DNSServer](service.DNS)
	dnsService.DeleteLocalDomain(originalDomain)
	if err := dnsService.AddLocalDomain(req.Domain, req.IP); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	delete(config.Config.DNS.LocalDomains, originalDomain)
	config.Config.DNS.LocalDomains[req.Domain] = req.IP
	if err := config.UpdateConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, config.Config.DNS.LocalDomains)
}
