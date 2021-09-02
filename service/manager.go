package service

type ServiceKey string

var (
	DHCP ServiceKey = "dchp"
	Web  ServiceKey = "web"
)

type serviceManager map[ServiceKey]Service

var instance = make(serviceManager)

type Service interface {
	Type() ServiceKey
	Start() error
	Stop() error
}

func Register(s Service) {
	instance[s.Type()] = s
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
	for _, service := range instance {
		fns = append(fns, service.Stop)
	}
	return checkErrors(fns...)
}

func checkErrors(fns ...func() error) error {
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
