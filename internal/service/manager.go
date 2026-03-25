package service

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
)

type ServiceKey string

var (
	DHCP ServiceKey = "dhcp"
	Web  ServiceKey = "web"
	DNS  ServiceKey = "dns"
)

type serviceManager map[ServiceKey]Service

var (
	instance = make(serviceManager)
	lock     = &sync.Mutex{}
)

type Service interface {
	Start() error
	Stop() error
}

func Register(s Service, stype ServiceKey) {
	instance[stype] = s
}

func Start() error {
	fns := make([]func() error, 0, len(instance))
	for _, service := range instance {
		fns = append(fns, service.Start)
	}
	return checkErrors(fns...)
}

func Stop() error {
	fns := make([]func() error, 0, len(instance))
	for name, service := range instance {
		log.Debug("stopping service ", name)
		fns = append(fns, service.Stop)
	}
	return checkErrors(fns...)
}

func GetService[T Service](stype ServiceKey) T {
	lock.Lock()
	defer lock.Unlock()
	return instance[stype].(T)
}

func IsRegistered(stype ServiceKey) bool {
	_, ok := instance[stype]
	return ok
}

func ActiveServices() []string {
	lock.Lock()
	defer lock.Unlock()
	services := make([]string, 0)
	for name := range instance {
		services = append(services, string(name))
	}
	return services
}

func ExtraFeatures() []string {
	lock.Lock()
	defer lock.Unlock()
	features := make([]string, 0)
	if config.Config.Auth.Type() == "oidc" {
		features = append(features, "oidc")
	}
	return features
}

func checkErrors(fns ...func() error) error {
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
