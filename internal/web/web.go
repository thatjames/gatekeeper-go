package web

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	v1 "gitlab.com/thatjames-go/gatekeeper-go/internal/web/v1"
)

func Init(ver string, cfg *config.Web) error {
	version = ver
	if cfg.Address == "" {
		cfg.Address = ":8085"
	}
	if cfg.WebURL == "" {
		cfg.WebURL = "http://localhost:5173"
	}

	r := gin.New()
	r.Use(gin.Recovery())

	var corsConfig cors.Config
	if cfg.CORS != nil {
		corsConfig = cors.DefaultConfig()
		if len(cfg.CORS.AllowOrigins) > 0 {
			corsConfig.AllowOrigins = cfg.CORS.AllowOrigins
		} else {
			corsConfig.AllowOrigins = []string{"*"}
		}
		if len(cfg.CORS.AllowMethods) > 0 {
			corsConfig.AllowMethods = cfg.CORS.AllowMethods
		}
		if len(cfg.CORS.AllowHeaders) > 0 {
			corsConfig.AllowHeaders = cfg.CORS.AllowHeaders
		} else {
			corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
		}
		corsConfig.AllowCredentials = cfg.CORS.AllowCredentials
		if len(cfg.CORS.ExposeHeaders) > 0 {
			corsConfig.ExposeHeaders = cfg.CORS.ExposeHeaders
		}
		if cfg.CORS.MaxAge > 0 {
			corsConfig.MaxAge = time.Duration(cfg.CORS.MaxAge) * time.Second
		}
	} else {
		corsConfig = cors.DefaultConfig()
		corsConfig.AllowOrigins = []string{cfg.WebURL}
		corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
		corsConfig.AllowCredentials = true
	}
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
