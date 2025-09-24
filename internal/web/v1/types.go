package v1

import (
	"encoding/binary"
	"fmt"
	"net"
	"regexp"
	"strings"

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

type DNSConfigResponse struct {
	Upstreams string `json:"upstreams"`
	Interface string `json:"interface"`
}

type LocalDomainRequest struct {
	Domain string `json:"domain"`
	IP     string `json:"ip"`
}

func (z *LocalDomainRequest) Validate() []ValidationError {
	validationErrors := make([]ValidationError, 0)
	if z.Domain == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "domain",
			Message: "Domain is required",
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

type DhcpOptions struct {
	Interface   string `json:"interface"`
	StartAddr   string `json:"startAddr"`
	EndAddr     string `json:"endAddr"`
	LeaseTTL    int    `json:"leaseTTL"`
	Gateway     string `json:"gateway"`
	SubnetMask  string `json:"subnetMask"`
	DomainName  string `json:"domainName"`
	LeaseFile   string `json:"leaseFile"`
	NameServers string `json:"nameServers"`
}

func (opts *DhcpOptions) Validate() []ValidationError {
	validationErrors := make([]ValidationError, 0)
	if opts.Interface == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "interface",
			Message: "Interface is required",
		})
	}
	if opts.StartAddr == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "startAddr",
			Message: "Start address is required",
		})
	} else {
		if net.ParseIP(opts.StartAddr).To4() == nil {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "startAddr",
				Message: "Start address must be a valid IPv4 address",
			})
		} else if opts.StartAddr == opts.EndAddr {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "startAddr",
				Message: "Start address must not be the same as the end address",
			})
		} else if opts.StartAddr == opts.Gateway {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "startAddr",
				Message: "Start address must not be the same as the gateway address",
			})
		} else {
			startIP := net.ParseIP(opts.StartAddr).To4()
			endIP := net.ParseIP(opts.EndAddr).To4()

			if endIP == nil {
				validationErrors = append(validationErrors, ValidationError{
					Field:   "endAddr",
					Message: "End address must be a valid IPv4 address",
				})
			} else {
				startUint32 := binary.BigEndian.Uint32(startIP)
				endUint32 := binary.BigEndian.Uint32(endIP)

				if startUint32 >= endUint32 {
					validationErrors = append(validationErrors, ValidationError{
						Field:   "startAddr",
						Message: "Start address must be less than end address",
					})
				}
			}
		}
	}
	if opts.EndAddr == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "endAddr",
			Message: "End address is required",
		})
	} else {
		if net.ParseIP(opts.EndAddr).To4() == nil {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "endAddr",
				Message: "End address must be a valid IPv4 address",
			})
		} else if opts.EndAddr == opts.StartAddr {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "endAddr",
				Message: "End address must not be the same as the start address",
			})
		} else if opts.EndAddr == opts.Gateway {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "endAddr",
				Message: "End address must not be the same as the gateway address",
			})
		} else {
			startIP := net.ParseIP(opts.StartAddr).To4()
			endIP := net.ParseIP(opts.EndAddr).To4()

			if startIP == nil {
				validationErrors = append(validationErrors, ValidationError{
					Field:   "startAddr",
					Message: "Start address must be a valid IPv4 address",
				})
			} else {
				startUint32 := binary.BigEndian.Uint32(startIP)
				endUint32 := binary.BigEndian.Uint32(endIP)

				if endUint32 <= startUint32 {
					validationErrors = append(validationErrors, ValidationError{
						Field:   "endAddr",
						Message: "End address must be greater than start address",
					})
				}
			}
		}
	}

	if opts.LeaseTTL == 0 {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "leaseTTL",
			Message: "Lease TTL is required",
		})
	}
	if opts.Gateway == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "gateway",
			Message: "Gateway is required",
		})
	} else {
		ipRegex := regexp.MustCompile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
		if !ipRegex.MatchString(opts.Gateway) {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "gateway",
				Message: "Gateway must be a valid IP address",
			})
		} else if net.ParseIP(opts.Gateway).To4() == nil {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "gateway",
				Message: "Gateway must be a valid IPv4 address",
			})
		}
	}
	if opts.SubnetMask == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "subnetMask",
			Message: "Subnet mask is required",
		})
	} else {
		ipRegex := regexp.MustCompile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
		if !ipRegex.MatchString(opts.SubnetMask) {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "subnetMask",
				Message: "Subnet mask must be a valid IP address",
			})
		} else if net.ParseIP(opts.SubnetMask).To4() == nil {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "subnetMask",
				Message: "Subnet mask must be a valid IPv4 address",
			})
		}
	}

	if nameServers := strings.Split(opts.NameServers, ","); len(nameServers) > 0 {
		for i, nameServer := range nameServers {
			if net.ParseIP(nameServer).To4() == nil {
				validationErrors = append(validationErrors, ValidationError{
					Field:   "nameServers",
					Message: fmt.Sprintf("Name server %d must be a valid IPv4 address", i+1),
				})
			}
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}
	return nil
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
