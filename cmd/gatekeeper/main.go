package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dhcp"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dns"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/service"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/system"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/web"
)

// Flags
var (
	configFile   string
	debug, trace bool
	version      = "development-build"
)

func main() {
	flag.StringVar(&configFile, "config", "config.yml", "config file")
	flag.BoolVar(&debug, "debug", false, "verbose printout")
	flag.BoolVar(&trace, "trace", false, "very verbose printout")
	flag.Parse()
	if err := config.LoadConfig(configFile); err != nil {
		panic(err)
	}
	log.SetReportCaller(true)
	log.SetFormatter(logFormatFunc(formatLogEntry))
	if debug {
		log.SetLevel(log.DebugLevel)
	} else if trace {
		log.SetLevel(log.TraceLevel)
	}
	log.Info("Starting gatekeeper")
	log.Info("Version ", version)
	system.Version = version
	log.Debugf("Config: %v", config.Config)

	if config.Config.DHCP != nil {
		log.Info("Registering DHCP server")
		dhcpServer := dhcp.NewDHCPServerFromConfig(config.Config.DHCP)
		service.Register(dhcpServer, service.DHCP)

	}

	if config.Config.DNS != nil {
		log.Info("Registering DNS server")
		localDomains := make(map[string]net.IP)
		for domain, ip := range config.Config.DNS.LocalDomains {
			localDomains[domain] = net.ParseIP(ip).To4()
		}

		dnsServer := dns.NewDNSServerWithOpts(dns.DNSServerOpts{
			Interface:      config.Config.DNS.Interface,
			BlocklistUrls:  config.Config.DNS.BlockLists,
			BlockedDomains: config.Config.DNS.BlockedDomains,
			ResolverOpts: &dns.ResolverOpts{
				LocalDomains: localDomains,
				Upstreams:    config.Config.DNS.UpstreamServers,
			},
			Port: config.Config.DNS.Port,
		})
		service.Register(dnsServer, service.DNS)
	}

	if config.Config.Web != nil {
		log.Info("Registering web server")
		go func() {
			if err := web.Init(version, config.Config.Web); err != nil {
				log.Error("unable to start web server:", err)
			}
		}()
	}

	// routingMan, err := routing.New()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// service.Register(routingMan)

	log.Debug("Starting registered services")
	if err := service.Start(); err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	sig := <-sigChan
	log.Infof("caught signal %v", sig)
	if err := service.Stop(); err != nil {
		log.Warning("unclean service exit: ", err.Error())
	}
}

type logFormatFunc func(*log.Entry) ([]byte, error)

func (fn logFormatFunc) Format(e *log.Entry) ([]byte, error) {
	return fn(e)
}

func formatLogEntry(e *log.Entry) ([]byte, error) {
	funcName := ""
	if e.Caller != nil {
		funcName = filepath.Base(e.Caller.Function)
	}
	msg := bytes.NewBuffer([]byte(fmt.Sprintf("%s %s %s - %s", e.Time.Format("2006-01-02 15:04:05"), funcName, strings.ToUpper(e.Level.String()), e.Message)))
	for key, dataField := range e.Data {
		msg.WriteString(fmt.Sprintf(" %s: %v", key, dataField))
	}
	msg.WriteString("\n")
	return msg.Bytes(), nil
}
