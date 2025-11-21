package main

import (
	"net"

	log "github.com/sirupsen/logrus"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dns"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/service"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/web"
)

func loadBaseModules() {
	if config.Config.DNS != nil {
		log.Info("Registering DNS server")
		moduleHooks = append(moduleHooks, loadDNSModule)
	}

	if config.Config.Web != nil {
		log.Info("Registering web server")
		moduleHooks = append(moduleHooks, loadWebModule)
	}
}

func loadDNSModule(conf *config.ConfigInstance) error {
	localDomains := make(map[string]net.IP)
	for domain, ip := range conf.DNS.LocalDomains {
		localDomains[domain] = net.ParseIP(ip).To4()
	}

	dnsServer := dns.NewDNSServerWithOpts(dns.DNSServerOpts{
		Interface:      conf.DNS.Interface,
		BlocklistUrls:  conf.DNS.BlockLists,
		BlockedDomains: conf.DNS.BlockedDomains,
		ResolverOpts: &dns.ResolverOpts{
			LocalDomains: localDomains,
			Upstreams:    conf.DNS.UpstreamServers,
		},
		Port: conf.DNS.Port,
	})
	service.Register(dnsServer, service.DNS)
	return nil
}

func loadWebModule(conf *config.ConfigInstance) error {
	log.Info("Registering web server")
	go func() {
		if err := web.Init(version, config.Config.Web); err != nil {
			log.Error("unable to start web server:", err)
		}
	}()
	return nil
}
