package domain

import "gitlab.com/thatjames-go/gatekeeper-go/dhcp"

type UserLoginRequest struct {
	Username string
	Password string
}

type PageData struct {
	Leases []dhcp.Lease
}
