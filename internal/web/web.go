package web

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dhcp"
	v1 "gitlab.com/thatjames-go/gatekeeper-go/internal/web/v1"
)

func setupAPIRoutes(r *gin.Engine) {
	api := r.Group("/api")
	v1Group := api.Group("/v1")
	v1Group.POST("/login", v1.LoginHandler)
	v1Group.GET("/health", v1.HealthHandler)

	protected := v1Group.Group("/", v1.AuthMiddleware())
	protected.GET("/verify", v1.VerifyHandler)
	protected.GET("/leases", v1.GetLeases)
}

func Init(ver string, cfg *config.Web, leases *dhcp.LeaseDB) error {
	version = ver
	leaseDB = leases

	r := gin.New()

	r.Use(v1.LoggingMiddleware())
	r.Use(gin.Recovery())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	r.Use(cors.New(corsConfig))

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	setupAPIRoutes(r)

	r.Use(v1.SpaMiddleware())
	r.Use(static.Serve("/", NewEmbeddedFS()))

	if cfg.TLS != nil {
		log.Info("Starting TLS server on ", cfg.Address)
		return r.RunTLS(cfg.Address, cfg.TLS.PublicKey, cfg.TLS.PrivateKey)
	} else {
		log.Info("Starting HTTP server on ", cfg.Address)
		return r.Run(cfg.Address)
	}
}
