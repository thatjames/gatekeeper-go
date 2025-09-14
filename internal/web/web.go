package web

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dhcp"
	v1 "gitlab.com/thatjames-go/gatekeeper-go/internal/web/v1"
)

var (
	leasePool *dhcp.LeasePool
)

func Init(ver string, cfg *config.Web, leases *dhcp.LeasePool) error {
	version = ver
	leasePool = leases

	r := gin.New()
	r.Use(gin.Recovery())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	r.Use(cors.New(corsConfig))

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := r.Group("/api")
	v1.SetupV1Endpoints(api)

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

func spaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") ||
			strings.HasPrefix(c.Request.URL.Path, "/metrics") {
			c.Next()
			return
		}

		if strings.Contains(c.Request.URL.Path, ".") {
			c.Next()
			return
		}

		c.Request.URL.Path = "/"
		c.Next()
	}
}
