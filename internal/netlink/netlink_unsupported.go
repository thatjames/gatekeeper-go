//go:build !linux
// +build !linux

package netlink

import "errors"

func NewNetlinkService() NetlinkService {
	return nil, errors.New("unsupported platform")
}
