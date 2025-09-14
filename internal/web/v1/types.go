package v1

import (
	"net"
	"regexp"

	"github.com/golang-jwt/jwt/v5"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dhcp"
)

type Role string

const (
	AdminRole Role = "admin"
	UserRole  Role = "user"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error  string            `json:"error"`
	Fields []ValidationError `json:"fields,omitempty"`
}

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

type DhcpLeaseRequest struct {
	ClientId string `json:"clientId"`
	IP       string `json:"ip"`
}

func (z *DhcpLeaseRequest) Validate() []ValidationError {
	validationErrors := make([]ValidationError, 0)
	if z.ClientId == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "clientId",
			Message: "Client ID is required",
		})
	} else if _, err := net.ParseMAC(z.ClientId); err != nil {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "clientId",
			Message: "Client ID must be a valid MAC address",
		})
	}

	if z.IP == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "ip",
			Message: "IP address is required",
		})
	} else {
		ipRegex := regexp.MustCompile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
		if !ipRegex.MatchString(z.IP) {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "ip",
				Message: "IP address must be a valid IP address",
			})
		} else if net.ParseIP(z.IP).To4() == nil {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "ip",
				Message: "IP address must be a valid IPv4 address",
			})
		}
	}
	if len(validationErrors) > 0 {
		return validationErrors
	}
	return nil
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
