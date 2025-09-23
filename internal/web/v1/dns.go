package v1

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
)

type DNSConfigResponse struct {
	Upstreams    []string          `json:"upstreams"`
	LocalDomains map[string]string `json:"localDomains"`
	Interface    string            `json:"interface"`
}

func getDNSConfig(c *gin.Context) {
	c.JSON(200, DNSConfigResponse{
		Upstreams:    config.Config.DNS.UpstreamServers,
		LocalDomains: config.Config.DNS.LocalDomains,
		Interface:    config.Config.DNS.Interface,
	})
}
