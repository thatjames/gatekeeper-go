package v1

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
)

type DNSConfigResponse struct {
	Upstreams string `json:"upstreams"`
	Interface string `json:"interface"`
}

func getDNSConfig(c *gin.Context) {
	c.JSON(200, DNSConfigResponse{
		Upstreams: strings.Join(config.Config.DNS.UpstreamServers, ","),
		Interface: config.Config.DNS.Interface,
	})
}

func getLocalDomains(c *gin.Context) {
	c.JSON(200, config.Config.DNS.LocalDomains)
}
