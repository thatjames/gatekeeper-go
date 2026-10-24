package v1

import (
	"maps"
	"net/http"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/thatjames/gatekeeper-go/internal/config"
	"github.com/thatjames/gatekeeper-go/internal/dns"
	"github.com/thatjames/gatekeeper-go/internal/service"
)

// dnsConfigResponse is the DNS section as the API answers it. The caller holds the config lock.
func dnsConfigResponse() DNSConfigResponse {
	port := config.Config.DNS.Port
	if port == 0 {
		port = 53
	}
	return DNSConfigResponse{
		Upstreams:      config.Config.DNS.UpstreamServers,
		Interface:      config.Config.DNS.Interface,
		Port:           port,
		Blocklist:      config.Config.DNS.BlockLists,
		BlockedDomains: config.Config.DNS.BlockedDomains,
	}
}

// localDomains returns a copy of the local domains to change, so the old map stays intact for a rollback.
func localDomains() map[string]string {
	domains := maps.Clone(config.Config.DNS.LocalDomains)
	if domains == nil {
		domains = make(map[string]string)
	}
	return domains
}

func getDNSConfig(c *gin.Context) {
	config.RLock()
	defer config.RUnlock()
	c.JSON(http.StatusOK, dnsConfigResponse())
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
	config.Lock()
	defer config.Unlock()
	dnsService := service.GetService[*dns.DNSServer](service.DNS)
	oldOpts := new(dns.DNSServerOpts)
	*oldOpts = *dnsService.Options()
	port := req.Port
	if port == 0 {
		port = oldOpts.Port
	}
	restore := func() {
		dnsService.Options().Interface = oldOpts.Interface
		dnsService.Options().Port = oldOpts.Port
		dnsService.Options().ResolverOpts.Upstreams = oldOpts.ResolverOpts.Upstreams
	}
	dnsService.Options().Interface = req.Interface
	dnsService.Options().Port = port
	dnsService.Options().ResolverOpts.Upstreams = req.Upstreams
	dnsService.Options().BlockedDomains = req.BlockedDomains
	if err := dnsService.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		restore()
		return
	} else if err = dnsService.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		restore()
		dnsService.Start()
		return
	}
	dnsCfg := config.Config.DNS
	oldInterface, oldPort, oldUpstreams := dnsCfg.Interface, dnsCfg.Port, dnsCfg.UpstreamServers
	dnsCfg.Interface = req.Interface
	dnsCfg.Port = port
	dnsCfg.UpstreamServers = req.Upstreams
	if err := config.UpdateConfig(); err != nil {
		dnsCfg.Interface, dnsCfg.Port, dnsCfg.UpstreamServers = oldInterface, oldPort, oldUpstreams
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dnsConfigResponse())
}

func getLocalDomains(c *gin.Context) {
	config.RLock()
	defer config.RUnlock()
	c.JSON(http.StatusOK, config.Config.DNS.LocalDomains)
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

	config.Lock()
	defer config.Unlock()
	dnsService := service.GetService[*dns.DNSServer](service.DNS)
	if err := dnsService.AddLocalDomain(req.Domain, req.IP); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oldDomains := config.Config.DNS.LocalDomains
	domains := localDomains()
	domains[req.Domain] = req.IP
	config.Config.DNS.LocalDomains = domains
	if err := config.UpdateConfig(); err != nil {
		config.Config.DNS.LocalDomains = oldDomains
		restoreLocalDomain(dnsService, oldDomains, req.Domain)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config.Config.DNS.LocalDomains)
}

// restoreLocalDomain sets domain in the resolver back to its entry in domains, or removes it.
func restoreLocalDomain(dnsService *dns.DNSServer, domains map[string]string, domain string) {
	if ip, ok := domains[domain]; ok {
		dnsService.AddLocalDomain(domain, ip)
	} else {
		dnsService.DeleteLocalDomain(domain)
	}
}

func deleteLocalDomain(c *gin.Context) {
	domain := c.Param("domain")
	log.Info("Deleting local domain: ", domain)

	config.Lock()
	defer config.Unlock()
	oldDomains := config.Config.DNS.LocalDomains
	domains := localDomains()
	delete(domains, domain)
	config.Config.DNS.LocalDomains = domains
	if err := config.UpdateConfig(); err != nil {
		config.Config.DNS.LocalDomains = oldDomains
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
	config.Lock()
	defer config.Unlock()
	dnsService := service.GetService[*dns.DNSServer](service.DNS)
	oldDomains := config.Config.DNS.LocalDomains
	undo := func() {
		restoreLocalDomain(dnsService, oldDomains, req.Domain)
		restoreLocalDomain(dnsService, oldDomains, originalDomain)
	}
	dnsService.DeleteLocalDomain(originalDomain)
	if err := dnsService.AddLocalDomain(req.Domain, req.IP); err != nil {
		undo()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	domains := localDomains()
	delete(domains, originalDomain)
	domains[req.Domain] = req.IP
	config.Config.DNS.LocalDomains = domains
	if err := config.UpdateConfig(); err != nil {
		config.Config.DNS.LocalDomains = oldDomains
		undo()
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
	if validationErrors := req.Validate(); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:  "Unable to add blocklist",
			Fields: validationErrors,
		})
		return
	}

	// Fetched before the lock is taken: a slow list must not hold up other config changes.
	dnsService := service.GetService[*dns.DNSServer](service.DNS)
	hosts, err := dnsService.FetchBlocklist(req.Url)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.Lock()
	defer config.Unlock()
	oldBlocklist := config.Config.DNS.BlockLists
	config.Config.DNS.BlockLists = append(slices.Clone(oldBlocklist), req.Url)
	if err := config.UpdateConfig(); err != nil {
		config.Config.DNS.BlockLists = oldBlocklist
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dnsService.AddBlockedDomains(hosts)
	c.JSON(http.StatusOK, dnsConfigResponse())
}

func deleteBlocklist(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid blocklist id"})
		return
	}
	config.Lock()
	defer config.Unlock()
	oldBlocklist := config.Config.DNS.BlockLists
	if id < 0 || id >= len(oldBlocklist) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid blocklist id"})
		return
	}
	log.Info("Deleting blocklist: ", oldBlocklist[id])
	config.Config.DNS.BlockLists = slices.Delete(slices.Clone(oldBlocklist), id, id+1)
	if err := config.UpdateConfig(); err != nil {
		config.Config.DNS.BlockLists = oldBlocklist
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dnsService := service.GetService[*dns.DNSServer](service.DNS)
	dnsService.FlushBlocklist()
	dnsService.LoadBlocklistFromURLS(config.Config.DNS.BlockLists)
	c.JSON(http.StatusOK, dnsConfigResponse())
}

func deleteBlockedDomain(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid blocklist id"})
		return
	}
	config.Lock()
	defer config.Unlock()
	oldBlockedDomains := config.Config.DNS.BlockedDomains
	if id < 0 || id >= len(oldBlockedDomains) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid blocklist id"})
		return
	}
	domain := oldBlockedDomains[id]
	log.Info("Deleting blocked domain: ", domain)
	config.Config.DNS.BlockedDomains = slices.Delete(slices.Clone(oldBlockedDomains), id, id+1)
	if err := config.UpdateConfig(); err != nil {
		config.Config.DNS.BlockedDomains = oldBlockedDomains
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dnsService := service.GetService[*dns.DNSServer](service.DNS)
	dnsService.DeleteBlockedDomain(domain)
	c.JSON(http.StatusOK, dnsConfigResponse())
}

func addBlockedDomain(c *gin.Context) {
	var req BlocklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	if validationErrors := req.Validate(); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:  "Unable to add blocked domain",
			Fields: validationErrors,
		})
		return
	}

	log.Info("Adding blocked domain: ", req)

	config.Lock()
	defer config.Unlock()
	oldBlockedDomains := config.Config.DNS.BlockedDomains
	config.Config.DNS.BlockedDomains = append(slices.Clone(oldBlockedDomains), req.Url)
	if err := config.UpdateConfig(); err != nil {
		config.Config.DNS.BlockedDomains = oldBlockedDomains
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dnsService := service.GetService[*dns.DNSServer](service.DNS)
	dnsService.AddBlockedDomain(req.Url)
	c.JSON(http.StatusOK, dnsConfigResponse())
}
