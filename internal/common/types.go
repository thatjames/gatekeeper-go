package common

import (
	"fmt"
	"net"
	"time"
)

type LeaseState int

func (ls LeaseState) String() string {
	switch ls {
	case LeaseAvailable:
		return "Available"
	case LeaseOffered:
		return "Offered"
	case LeaseReserved:
		return "Reserved"
	case LeaseActive:
		return "Active"
	default:
		return "unknown"
	}
}

const (
	LeaseAvailable LeaseState = iota
	LeaseOffered
	LeaseReserved
	LeaseActive
)

type Lease struct {
	Id       int
	ClientId string
	Hostname string
	IP       net.IP
	Expiry   time.Time
	State    LeaseState
}

func (l *Lease) String() string {
	return fmt.Sprintf("%s: %s - %s: %s expiring at %s", l.Hostname, l.IP.String(), l.ClientId, l.State, l.Expiry.Format("15:04:05"))
}

func (l *Lease) Clear() {
	l.ClientId = ""
	l.Hostname = ""
	l.Expiry = time.Time{}
	l.State = LeaseAvailable
}
