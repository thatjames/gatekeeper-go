package v1

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dhcp"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/service"
)

func getLeases(c *gin.Context) {
	dhcpService := service.GetService[*dhcp.DHCPServer](service.DHCP)
	activeLeases := dhcpService.LeaseDB().ActiveLeases()
	reservedLeases := dhcpService.LeaseDB().ReservedLeases()
	c.JSON(http.StatusOK, DhcpLeaseResponse{
		ActiveLeases:   MapLeases(activeLeases),
		ReservedLeases: MapLeases(reservedLeases),
	})
}

func deleteLease(c *gin.Context) {
	dhcpService := service.GetService[*dhcp.DHCPServer](service.DHCP)
	lease := dhcpService.LeaseDB().GetLease(c.Param("clientId"))
	if lease == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lease not found"})
		return
	}
	dhcpService.LeaseDB().ReleaseLease(lease)
	activeLeases := dhcpService.LeaseDB().ActiveLeases()
	reservedLeases := dhcpService.LeaseDB().ReservedLeases()
	c.JSON(http.StatusOK, DhcpLeaseResponse{
		ActiveLeases:   MapLeases(activeLeases),
		ReservedLeases: MapLeases(reservedLeases),
	})
}

func getDHCPOptions(c *gin.Context) {
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

func reserveLease(c *gin.Context) {
	dhcpService := service.GetService[*dhcp.DHCPServer](service.DHCP)
	var lease DhcpLeaseRequest
	if err := c.ShouldBindJSON(&lease); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	} else if validationErrors := lease.Validate(); validationErrors != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:  "Unable to reserve lease",
			Fields: validationErrors,
		})
		return
	}
	if err := dhcpService.LeaseDB().ReserveLease(lease.ClientId, net.ParseIP(lease.IP).To4()); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
			Fields: []ValidationError{
				{
					Field:   "ip",
					Message: "IP already reserved",
				},
			},
		})
		return
	}
	config.Config.DHCP.ReservedAddresses[lease.ClientId] = lease.IP
	if err := config.UpdateConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	activeLeases := dhcpService.LeaseDB().ActiveLeases()
	reservedLeases := dhcpService.LeaseDB().ReservedLeases()
	c.JSON(http.StatusOK, DhcpLeaseResponse{
		ActiveLeases:   MapLeases(activeLeases),
		ReservedLeases: MapLeases(reservedLeases),
	})
}

func updateLease(c *gin.Context) {
	dhcpService := service.GetService[*dhcp.DHCPServer](service.DHCP)
	var lease DhcpLeaseRequest
	if err := c.ShouldBindJSON(&lease); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	} else if validationErrors := lease.Validate(); validationErrors != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:  "Unable to update reserved lease",
			Fields: validationErrors,
		})
		return
	}
	if err := dhcpService.LeaseDB().UpdateLease(lease.ClientId, net.ParseIP(lease.IP).To4()); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	if err := config.UpdateConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	activeLeases := dhcpService.LeaseDB().ActiveLeases()
	reservedLeases := dhcpService.LeaseDB().ReservedLeases()
	c.JSON(http.StatusOK, DhcpLeaseResponse{
		ActiveLeases:   MapLeases(activeLeases),
		ReservedLeases: MapLeases(reservedLeases),
	})
}
