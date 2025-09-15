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
	c.JSON(http.StatusOK, DhcpOptions{
		Interface:  opts.Interface,
		StartAddr:  opts.StartFrom.String(),
		EndAddr:    opts.EndAt.String(),
		LeaseTTL:   opts.LeaseTTL,
		Gateway:    opts.Gateway.String(),
		SubnetMask: opts.SubnetMask.String(),
		DomainName: opts.DomainName,
		LeaseFile:  opts.LeaseFile,
	})
}

func updateDHCPOptions(c *gin.Context) {
	dhcpService := service.GetService[*dhcp.DHCPServer](service.DHCP)
	var opts DhcpOptions
	if err := c.ShouldBindJSON(&opts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	} else if validationErrors := opts.Validate(); validationErrors != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:  "Unable to update options",
			Fields: validationErrors,
		})
		return
	}
	var oldOpts = new(dhcp.DHCPServerOpts)
	*oldOpts = *dhcpService.Options()
	dhcpService.UpdateOptions(&dhcp.DHCPServerOpts{
		Interface:      opts.Interface,
		StartFrom:      net.ParseIP(opts.StartAddr).To4(),
		EndAt:          net.ParseIP(opts.EndAddr).To4(),
		LeaseTTL:       opts.LeaseTTL,
		Gateway:        net.ParseIP(opts.Gateway).To4(),
		SubnetMask:     net.ParseIP(opts.SubnetMask).To4(),
		DomainName:     opts.DomainName,
		LeaseFile:      opts.LeaseFile,
		ReservedLeases: dhcpService.Options().ReservedLeases,
	})

	if err := dhcpService.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		dhcpService.UpdateOptions(oldOpts)
		return
	} else if err = dhcpService.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		dhcpService.UpdateOptions(oldOpts)
		// recover the service
		dhcpService.Start()
		return
	}
	config.Config.DHCP = &config.DHCP{
		Interface:         opts.Interface,
		StartAddr:         opts.StartAddr,
		EndAddr:           opts.EndAddr,
		LeaseTTL:          opts.LeaseTTL,
		Gateway:           opts.Gateway,
		SubnetMask:        opts.SubnetMask,
		DomainName:        opts.DomainName,
		LeaseFile:         opts.LeaseFile,
		ReservedAddresses: dhcpService.Options().ReservedLeases,
	}
	if err := config.UpdateConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		dhcpService.UpdateOptions(oldOpts)
		return
	}
	c.JSON(http.StatusOK, DhcpOptions{
		Interface:  opts.Interface,
		StartAddr:  opts.StartAddr,
		EndAddr:    opts.EndAddr,
		LeaseTTL:   opts.LeaseTTL,
		Gateway:    opts.Gateway,
		SubnetMask: opts.SubnetMask,
		DomainName: opts.DomainName,
		LeaseFile:  opts.LeaseFile,
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
