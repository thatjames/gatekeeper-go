package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	"gitlab.com/thatjames-go/gatekeeper-go/config"
	"gitlab.com/thatjames-go/gatekeeper-go/dhcp"
)

//Flags
var (
	configFile string
	debug      bool
)

func main() {
	flag.StringVar(&configFile, "c", "config.yml", "config file")
	flag.BoolVar(&debug, "v", false, "verbose printout")
	flag.Parse()
	if err := config.LoadConfig(configFile); err != nil {
		panic(err)
	}
	log.SetFormatter(logFormatFunc(formatLogEntry))
	if debug {
		log.SetLevel(log.DebugLevel)
	}
	log.Info("Starting gatekeeper")
	log.Info("Starting DHCP server")
	log.Debugf("%+v", config.Config)
	nameServers := make([]net.IP, 0, len(config.Config.DHCP.NameServers))
	for _, nameServer := range config.Config.DHCP.NameServers {
		nameServers = append(nameServers, net.ParseIP(nameServer).To4())
	}
	options := &dhcp.DHCPServerOpts{
		Interface:   config.Config.DHCP.Interface,
		StartFrom:   net.ParseIP(config.Config.DHCP.StartAddr).To4(),
		EndAt:       net.ParseIP(config.Config.DHCP.EndAddr).To4(),
		NameServers: nameServers,
		LeaseTTL:    config.Config.DHCP.LeaseTTL,
		Router:      net.ParseIP(config.Config.DHCP.Router).To4(),
		SubnetMask:  net.ParseIP(config.Config.DHCP.SubnetMask).To4(),
	}
	dhcpServer := dhcp.NewDHCPServerWithOpts(options)
	if err := dhcpServer.Start(); err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	sig := <-sigChan
	log.Infof("caught signal %v", sig)
}

type logFormatFunc func(*log.Entry) ([]byte, error)

func (fn logFormatFunc) Format(e *log.Entry) ([]byte, error) {
	return fn(e)
}

func formatLogEntry(e *log.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("%s %s - %s\n", e.Time.Format("2006-01-02 15:04:05"), strings.ToUpper(e.Level.String()), e.Message)), nil
}
