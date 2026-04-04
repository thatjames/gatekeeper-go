package v1

import (
	"testing"
)

func TestDhcpLeaseRequestValidate_Valid(t *testing.T) {
	req := &DhcpLeaseRequest{
		ClientId: "00:11:22:33:44:55",
		IP:       "192.168.1.100",
	}
	errs := req.Validate()
	if errs != nil {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestDhcpLeaseRequestValidate_InvalidMAC(t *testing.T) {
	req := &DhcpLeaseRequest{
		ClientId: "invalid-mac",
		IP:       "192.168.1.100",
	}
	errs := req.Validate()
	if errs == nil {
		t.Error("expected validation error for invalid MAC")
	}
	if len(errs) < 1 || errs[0].Field != "clientId" {
		t.Errorf("expected clientId error, got %v", errs)
	}
}

func TestDhcpLeaseRequestValidate_InvalidIP(t *testing.T) {
	req := &DhcpLeaseRequest{
		ClientId: "00:11:22:33:44:55",
		IP:       "invalid-ip",
	}
	errs := req.Validate()
	if errs == nil {
		t.Error("expected validation error for invalid IP")
	}
	if len(errs) < 1 || errs[0].Field != "ip" {
		t.Errorf("expected ip error, got %v", errs)
	}
}

func TestDhcpLeaseRequestValidate_IPv6(t *testing.T) {
	req := &DhcpLeaseRequest{
		ClientId: "00:11:22:33:44:55",
		IP:       "::1",
	}
	errs := req.Validate()
	if errs == nil {
		t.Error("expected validation error for IPv6")
	}
	if len(errs) < 1 || errs[0].Field != "ip" {
		t.Errorf("expected ip error, got %v", errs)
	}
}

func TestDhcpLeaseRequestValidate_EmptyClientId(t *testing.T) {
	req := &DhcpLeaseRequest{
		ClientId: "",
		IP:       "192.168.1.100",
	}
	errs := req.Validate()
	if errs == nil {
		t.Error("expected validation error for empty clientId")
	}
}

func TestDhcpLeaseRequestValidate_EmptyIP(t *testing.T) {
	req := &DhcpLeaseRequest{
		ClientId: "00:11:22:33:44:55",
		IP:       "",
	}
	errs := req.Validate()
	if errs == nil {
		t.Error("expected validation error for empty IP")
	}
}

func TestLocalDomainRequestValidate_Valid(t *testing.T) {
	req := &LocalDomainRequest{
		Domain: "example.com",
		IP:     "192.168.1.1",
	}
	errs := req.Validate()
	if errs != nil {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestLocalDomainRequestValidate_EmptyDomain(t *testing.T) {
	req := &LocalDomainRequest{
		Domain: "",
		IP:     "192.168.1.1",
	}
	errs := req.Validate()
	if errs == nil {
		t.Error("expected validation error for empty domain")
	}
}

func TestLocalDomainRequestValidate_EmptyIP(t *testing.T) {
	req := &LocalDomainRequest{
		Domain: "example.com",
		IP:     "",
	}
	errs := req.Validate()
	if errs == nil {
		t.Error("expected validation error for empty IP")
	}
}

func TestDhcpOptionsValidate_Valid(t *testing.T) {
	opts := &DhcpOptions{
		Interface:   "eth0",
		StartAddr:   "192.168.1.100",
		EndAddr:     "192.168.1.200",
		LeaseTTL:    300,
		Gateway:     "192.168.1.1",
		SubnetMask:  "255.255.255.0",
		DomainName:  "local",
		NameServers: []string{"8.8.8.8"},
	}
	errs := opts.Validate()
	if errs != nil {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestDhcpOptionsValidate_MissingInterface(t *testing.T) {
	opts := &DhcpOptions{
		StartAddr: "192.168.1.100",
		EndAddr:   "192.168.1.200",
	}
	errs := opts.Validate()
	if errs == nil {
		t.Error("expected validation error for missing interface")
	}
}

func TestDhcpOptionsValidate_StartEqualsEnd(t *testing.T) {
	opts := &DhcpOptions{
		Interface: "eth0",
		StartAddr: "192.168.1.100",
		EndAddr:   "192.168.1.100",
	}
	errs := opts.Validate()
	if errs == nil {
		t.Error("expected validation error for same start/end")
	}
}

func TestDhcpOptionsValidate_StartGreaterThanEnd(t *testing.T) {
	opts := &DhcpOptions{
		Interface: "eth0",
		StartAddr: "192.168.1.200",
		EndAddr:   "192.168.1.100",
	}
	errs := opts.Validate()
	if errs == nil {
		t.Error("expected validation error for start > end")
	}
}

func TestDhcpOptionsValidate_GatewayConflict(t *testing.T) {
	opts := &DhcpOptions{
		Interface: "eth0",
		StartAddr: "192.168.1.100",
		EndAddr:   "192.168.1.200",
		Gateway:   "192.168.1.100",
	}
	errs := opts.Validate()
	if errs == nil {
		t.Error("expected validation error for gateway conflict")
	}
}

func TestDhcpOptionsValidate_InvalidGateway(t *testing.T) {
	opts := &DhcpOptions{
		Interface: "eth0",
		StartAddr: "192.168.1.100",
		EndAddr:   "192.168.1.200",
		Gateway:   "invalid",
	}
	errs := opts.Validate()
	if errs == nil {
		t.Error("expected validation error for invalid gateway")
	}
}

func TestDhcpOptionsValidate_InvalidSubnetMask(t *testing.T) {
	opts := &DhcpOptions{
		Interface:  "eth0",
		StartAddr:  "192.168.1.100",
		EndAddr:    "192.168.1.200",
		SubnetMask: "invalid",
	}
	errs := opts.Validate()
	if errs == nil {
		t.Error("expected validation error for invalid subnet mask")
	}
}

func TestDhcpOptionsValidate_MissingLeaseTTL(t *testing.T) {
	opts := &DhcpOptions{
		Interface: "eth0",
		StartAddr: "192.168.1.100",
		EndAddr:   "192.168.1.200",
	}
	errs := opts.Validate()
	if errs == nil {
		t.Error("expected validation error for missing lease TTL")
	}
}

func TestDhcpOptionsValidate_InvalidNameServer(t *testing.T) {
	opts := &DhcpOptions{
		Interface:   "eth0",
		StartAddr:   "192.168.1.100",
		EndAddr:     "192.168.1.200",
		LeaseTTL:    300,
		Gateway:     "192.168.1.1",
		SubnetMask:  "255.255.255.0",
		NameServers: []string{"invalid"},
	}
	errs := opts.Validate()
	if errs == nil {
		t.Error("expected validation error for invalid nameserver")
	}
}

func TestDNSConfigRequestValidate_Valid(t *testing.T) {
	req := &DNSConfigRequest{
		Interface: "eth0",
		Upstreams: []string{"8.8.8.8", "1.1.1.1"},
	}
	errs := req.Validate()
	if errs != nil {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestDNSConfigRequestValidate_MissingInterface(t *testing.T) {
	req := &DNSConfigRequest{
		Upstreams: []string{"8.8.8.8"},
	}
	errs := req.Validate()
	if errs == nil {
		t.Error("expected validation error for missing interface")
	}
}

func TestDNSConfigRequestValidate_MissingUpstreams(t *testing.T) {
	req := &DNSConfigRequest{
		Interface: "eth0",
	}
	errs := req.Validate()
	if errs == nil {
		t.Error("expected validation error for missing upstreams")
	}
}

func TestDNSConfigRequestValidate_InvalidUpstream(t *testing.T) {
	req := &DNSConfigRequest{
		Interface: "eth0",
		Upstreams: []string{"invalid"},
	}
	errs := req.Validate()
	if errs == nil {
		t.Error("expected validation error for invalid upstream")
	}
}

func TestBlocklistRequestValidate_Valid(t *testing.T) {
	req := &BlocklistRequest{
		Url: "https://example.com/blocklist",
	}
	errs := req.Validate()
	if len(errs) > 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestBlocklistRequestValidate_EmptyUrl(t *testing.T) {
	req := &BlocklistRequest{
		Url: "",
	}
	errs := req.Validate()
	if errs == nil {
		t.Error("expected validation error for empty URL")
	}
}
