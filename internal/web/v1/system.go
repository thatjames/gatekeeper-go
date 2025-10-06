package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/service"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/system"
)

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	})
}

func getSystemInfo(c *gin.Context) {
	sysInfo, err := system.GetSystemInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sysInfo)
}

func getInterfaces(c *gin.Context) {
	interfaces, err := system.GetNetworkInterfaces()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, interfaces)
}

func getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": system.Version,
	})
}

func getModules(c *gin.Context) {
	services := service.GetActiveServices()
	c.JSON(http.StatusOK, services)
}
