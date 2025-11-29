//go:build router
// +build router

package main

import (
	log "github.com/sirupsen/logrus"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dhcp"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/netlink"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/service"
)

func loadModules() {
	loadBaseModules()
	if config.Config.DHCP != nil {
		moduleHooks = append(moduleHooks, func(conf *config.ConfigInstance) error {
			log.Info("Registering DHCP server")
			dhcpServer := dhcp.NewDHCPServerFromConfig(conf.DHCP)
			service.Register(dhcpServer, service.DHCP)
			return nil
		})
	}

	if config.Config.Router != nil {
		moduleHooks = append(moduleHooks, func(conf *config.ConfigInstance) error {
			log.Info("Registering router service")
			routerService := netlink.NewNetlinkService(netlink.NetlinkServiceOpts{
				WANInterface: conf.Router.WANInterface,
				LANInterface: conf.Router.LANInterface,
			})
			service.Register(routerService, service.Router)
			return nil
		})
	}
}
