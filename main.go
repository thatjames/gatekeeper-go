package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	"gitlab.com/thatjames-go/gatekeeper-go/config"
	"gitlab.com/thatjames-go/gatekeeper-go/dhcp"
	"gitlab.com/thatjames-go/gatekeeper-go/service"
	"gitlab.com/thatjames-go/gatekeeper-go/web"
)

//Flags
var (
	configFile string
	debug      bool
	version    = "development-build"
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
	log.Info("Version ", version)
	log.Debugf("Config: %v", config.Config)

	if config.Config.DHCP != nil {
		log.Info("Registering DHCP server")
		dhcpServer := dhcp.NewDHCPServerFromConfig(config.Config.DHCP)
		service.Register(dhcpServer, service.DHCP)

		if config.Config.Web != nil {
			log.Debug("Registering web server")
			go func() {
				if err := web.Init(version, config.Config.Web, dhcpServer.LeaseDB()); err != nil {
					log.Error("unable to start web server:", err)
				}
			}()
		}

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
	msg := bytes.NewBuffer([]byte(fmt.Sprintf("%s %s - %s", e.Time.Format("2006-01-02 15:04:05"), strings.ToUpper(e.Level.String()), e.Message)))
	for key, dataField := range e.Data {
		msg.WriteString(fmt.Sprintf(" %s: %v", key, dataField))
	}
	msg.WriteString("\n")
	return msg.Bytes(), nil
}
