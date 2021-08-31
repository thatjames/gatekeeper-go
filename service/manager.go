package service

type ServiceKey string

var (
	DHCP ServiceKey = "dchp"
	Web  ServiceKey = "web"
)

type serviceManager map[ServiceKey]interface{}

var instance = make(serviceManager)

type Service interface {
	Type() ServiceKey
}

func Register(s Service) {
	instance[s.Type()] = s
}
