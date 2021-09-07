module gitlab.com/thatjames-go/gatekeeper-go

go 1.16

replace gitlab.com/thatjames-go/netlink-go => ../netlink-go

require (
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
	golang.org/x/sys v0.0.0-20210902050250-f475640dd07b
	gopkg.in/yaml.v2 v2.4.0
)
