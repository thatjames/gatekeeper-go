package web

import "gitlab.com/thatjames-go/gatekeeper-go/dhcp"

type LeasePage struct {
	Start          string
	End            string
	Nameservers    []string
	ActiveLeases   []dhcp.Lease
	ReservedLeases []dhcp.Lease
	DomainName     string
}
