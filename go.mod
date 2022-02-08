module gitlab.com/thatjames-go/gatekeeper-go

go 1.16

replace gitlab.com/thatjames-go/netlink-go => ../netlink-go

require (
	github.com/prometheus/client_golang v1.12.1 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/tg123/go-htpasswd v1.1.0 // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9
	gopkg.in/yaml.v2 v2.4.0
)
