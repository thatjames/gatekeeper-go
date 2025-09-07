package service

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type ServiceKey string

var (
	DHCP ServiceKey = "dchp"
	Web  ServiceKey = "web"
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

func checkErrors(fns ...func() error) error {
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
