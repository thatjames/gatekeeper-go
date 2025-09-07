package v1

import (
	"github.com/golang-jwt/jwt/v5"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dhcp"
)

type Role string

const (
	AdminRole Role = "admin"
	UserRole  Role = "user"
)

type UserLoginRequest struct {
	Username string
	Password string
}

type PageData struct {
	Leases []dhcp.Lease
}

type User struct {
	Username string
	Role     Role
}

type UserClaims struct {
	User
	jwt.RegisteredClaims
}

type DhcpLeaseResponse struct {
	ActiveLeases   []Lease `json:"active,omitempty"`
	ReservedLeases []Lease `json:"reserved,omitempty"`
}

type Lease struct {
	ClientId string `json:"clientId"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	State    string `json:"state"`
	Expiry   string `json:"expiry"`
}

type DhcpOptionsResponse struct {
	Interface      string            `json:"interface"`
	StartAddr      string            `json:"startAddr"`
	EndAddr        string            `json:"endAddr"`
	LeaseTTL       int               `json:"leaseTTL"`
	Router         string            `json:"router"`
	SubnetMask     string            `json:"subnetMask"`
	DomainName     string            `json:"domainName"`
	ReservedLeases map[string]string `json:"reservedLeases"`
	LeaseFile      string            `json:"leaseFile"`
}

func MapLease(lease dhcp.Lease) Lease {
	return Lease{
		ClientId: lease.ClientId,
		Hostname: lease.Hostname,
		IP:       lease.IP.String(),
		State:    lease.State.String(),
		Expiry:   lease.Expiry.Format("15:04:05"),
	}
}

func MapLeases(leases []dhcp.Lease) []Lease {
	var leaseList []Lease
	for _, lease := range leases {
		leaseList = append(leaseList, MapLease(lease))
	}
	return leaseList
}
