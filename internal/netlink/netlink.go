//go:build linux
// +build linux

package netlink

import (
	"golang.org/x/sys/unix"

	log "github.com/sirupsen/logrus"
)

type NetlinkService struct {
	fd int
}

func NewNetlinkService() (*NetlinkService, error) {
	return &NetlinkService{}, nil
}

func (z *NetlinkService) Start() error {
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW, unix.NETLINK_ROUTE)
	if err != nil {
		return err
	} else if err := unix.Bind(fd, &unix.SockaddrNetlink{Family: unix.AF_NETLINK}); err != nil {
		return err
	}
	z.fd = fd
	log.Info("netlink service started")
	return nil
}

func (z *NetlinkService) Stop() error {
	return nil
}
