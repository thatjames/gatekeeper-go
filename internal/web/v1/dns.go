package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dns"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/service"
)

func getDNSConfig(c *gin.Context) {
	c.JSON(200, DNSConfigResponse{
		Upstreams:      config.Config.DNS.UpstreamServers,
		Interface:      config.Config.DNS.Interface,
		Blocklist:      config.Config.DNS.BlockLists,
		BlockedDomains: config.Config.DNS.BlockedDomains,
	})
}

func updateDNSConfig(c *gin.Context) {
	var req DNSConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	if validationErrors := req.Validate(); validationErrors != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:  "Unable to update DNS config",
			Fields: validationErrors,
		})
		return
	}
	log.Info("Updating DNS config: ", req)
	dnsService := service.GetService[*dns.DNSServer](service.DNS)
	oldOpts := new(dns.DNSServerOpts)
	*oldOpts = *dnsService.Options()
	dnsService.Options().Interface = req.Interface
	dnsService.Options().ResolverOpts.Upstreams = req.Upstreams
	dnsService.Options().BlockedDomains = req.BlockedDomains
	if err := dnsService.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		dnsService.Options().Interface = oldOpts.Interface
		dnsService.Options().ResolverOpts.Upstreams = oldOpts.ResolverOpts.Upstreams
		return
	} else if err = dnsService.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		dnsService.Options().Interface = oldOpts.Interface
		dnsService.Options().ResolverOpts.Upstreams = oldOpts.ResolverOpts.Upstreams
		dnsService.Start()
		return
	}
	config.Config.DNS.Interface = req.Interface
	config.Config.DNS.UpstreamServers = req.Upstreams
	if err := config.UpdateConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, DNSConfigResponse{
		Upstreams:      config.Config.DNS.UpstreamServers,
		Interface:      config.Config.DNS.Interface,
		Blocklist:      config.Config.DNS.BlockLists,
		BlockedDomains: config.Config.DNS.BlockedDomains,
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

func addBlocklist(c *gin.Context) {
	var req BlocklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	if validationErrors := req.Validate(); validationErrors != nil && len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:  "Unable to add blocklist",
			Fields: validationErrors,
		})
		return
	}

	dnsService := service.GetService[*dns.DNSServer](service.DNS)
	oldBlocklist := config.Config.DNS.BlockLists
	config.Config.DNS.BlockLists = append(config.Config.DNS.BlockLists, req.Url)
	if err := config.UpdateConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		config.Config.DNS.BlockLists = oldBlocklist
		return
	}

	if err := dnsService.AddBlocklistFromURL(req.Url); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dnsService.FlushBlocklist()
	dnsService.LoadBlocklistFromURLS(config.Config.DNS.BlockLists)
}

func deleteBlocklist(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid blocklist id"})
		return
	}
	if id < 0 || id >= len(config.Config.DNS.BlockLists) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid blocklist id"})
		return
	}
	log.Info("Deleting blocklist: ", config.Config.DNS.BlockLists[id])
	oldBlocklist := config.Config.DNS.BlockLists
	if id == len(config.Config.DNS.BlockLists)-1 {
		config.Config.DNS.BlockLists = config.Config.DNS.BlockLists[:id]
	} else {
		config.Config.DNS.BlockLists = append(config.Config.DNS.BlockLists[:id], config.Config.DNS.BlockLists[id+1:]...)
	}
	if err := config.UpdateConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		config.Config.DNS.BlockLists = oldBlocklist
		return
	}

	dnsService := service.GetService[*dns.DNSServer](service.DNS)
	dnsService.FlushBlocklist()
	dnsService.LoadBlocklistFromURLS(config.Config.DNS.BlockLists)
}

func deleteBlockedDomain(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid blocklist id"})
		return
	}
	if id < 0 || id >= len(config.Config.DNS.BlockedDomains) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid blocklist id"})
		return
	}
	log.Info("Deleting blocked domain: ", config.Config.DNS.BlockedDomains[id])
	dnsService := service.GetService[*dns.DNSServer](service.DNS)
	dnsService.DeleteBlockedDomain(config.Config.DNS.BlockedDomains[id])
	oldBlockedDomains := config.Config.DNS.BlockedDomains
	if id == len(config.Config.DNS.BlockedDomains)-1 {
		config.Config.DNS.BlockedDomains = config.Config.DNS.BlockedDomains[:id]
	} else {
		config.Config.DNS.BlockedDomains = append(config.Config.DNS.BlockedDomains[:id], config.Config.DNS.BlockedDomains[id+1:]...)
	}
	if err := config.UpdateConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		config.Config.DNS.BlockedDomains = oldBlockedDomains
		return
	}

	c.JSON(http.StatusOK, DNSConfigResponse{
		Upstreams:      config.Config.DNS.UpstreamServers,
		Interface:      config.Config.DNS.Interface,
		Blocklist:      config.Config.DNS.BlockLists,
		BlockedDomains: config.Config.DNS.BlockedDomains,
	})
}

func addBlockedDomain(c *gin.Context) {
	var req BlocklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	if validationErrors := req.Validate(); validationErrors != nil && len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:  "Unable to add blocked domain",
			Fields: validationErrors,
		})
		return
	}

	log.Info("Adding blocked domain: ", req)

	dnsService := service.GetService[*dns.DNSServer](service.DNS)
	oldBlockedDomains := config.Config.DNS.BlockedDomains
	config.Config.DNS.BlockedDomains = append(config.Config.DNS.BlockedDomains, req.Url)
	if err := config.UpdateConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		config.Config.DNS.BlockedDomains = oldBlockedDomains
		return
	}

	dnsService.AddBlockedDomain(req.Url)
	c.JSON(http.StatusOK, DNSConfigResponse{
		Upstreams:      config.Config.DNS.UpstreamServers,
		Interface:      config.Config.DNS.Interface,
		Blocklist:      config.Config.DNS.BlockLists,
		BlockedDomains: config.Config.DNS.BlockedDomains,
	})
}
