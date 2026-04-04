package dhcp

import (
	"net"
	"testing"

	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
)

func TestNewDHCPServer_DefaultOpts(t *testing.T) {
	server := NewDHCPServer()

	if server.opts == nil {
		t.Error("expected opts to be set")
	}

	if server.opts.Interface != "eth0" {
		t.Errorf("expected interface 'eth0', got '%s'", server.opts.Interface)
	}

	if server.opts.LeaseTTL != 300 {
		t.Errorf("expected LeaseTTL 300, got %d", server.opts.LeaseTTL)
	}

	if server.issuedLeases == nil {
		t.Error("expected lease pool to be created")
	}
}

func TestNewDHCPServerFromConfig(t *testing.T) {
	cfg := &config.DHCP{
		Interface:  "wlan0",
		StartAddr:  "192.168.1.10",
		EndAddr:    "192.168.1.50",
		LeaseTTL:   600,
		SubnetMask: "255.255.255.0",
		Gateway:    "192.168.1.1",
		DomainName: "home.local",
		LeaseFile:  "/tmp/leases",
	}

	server := NewDHCPServerFromConfig(cfg)

	if server.opts.Interface != "wlan0" {
		t.Errorf("expected interface 'wlan0', got '%s'", server.opts.Interface)
	}

	if server.opts.StartFrom == nil {
		t.Error("expected StartFrom to be set")
	} else if server.opts.StartFrom.String() != "192.168.1.10" {
		t.Errorf("expected StartFrom 192.168.1.10, got %s", server.opts.StartFrom)
	}

	if server.opts.EndAt == nil {
		t.Error("expected EndAt to be set")
	} else if server.opts.EndAt.String() != "192.168.1.50" {
		t.Errorf("expected EndAt 192.168.1.50, got %s", server.opts.EndAt)
	}

	if server.opts.LeaseTTL != 600 {
		t.Errorf("expected LeaseTTL 600, got %d", server.opts.LeaseTTL)
	}
}

func TestNewDHCPServerWithOpts(t *testing.T) {
	opts := &DHCPServerOpts{
		Interface:  "lo",
		StartFrom:  net.ParseIP("10.0.0.10").To4(),
		EndAt:      net.ParseIP("10.0.0.20").To4(),
		LeaseTTL:   1800,
		SubnetMask: net.ParseIP("255.255.255.0").To4(),
		Gateway:    net.ParseIP("10.0.0.1").To4(),
	}

	server := NewDHCPServerWithOpts(opts)

	if server.opts.Interface != "lo" {
		t.Errorf("expected interface 'lo', got '%s'", server.opts.Interface)
	}

	if server.opts.LeaseTTL != 1800 {
		t.Errorf("expected LeaseTTL 1800, got %d", server.opts.LeaseTTL)
	}
}

func TestDHCPServer_LeaseDB(t *testing.T) {
	server := NewDHCPServer()

	leaseDB := server.LeaseDB()
	if leaseDB == nil {
		t.Error("expected lease DB to be returned")
	}
}

func TestDHCPServer_Options(t *testing.T) {
	opts := &DHCPServerOpts{
		Interface:  "eth1",
		StartFrom:  net.ParseIP("172.16.0.10").To4(),
		EndAt:      net.ParseIP("172.16.0.100").To4(),
		LeaseTTL:   3600,
		SubnetMask: net.ParseIP("255.255.0.0").To4(),
		Gateway:    net.ParseIP("172.16.0.1").To4(),
		DomainName: "internal",
		LeaseFile:  "/data/leases",
	}

	server := NewDHCPServerWithOpts(opts)

	result := server.Options()

	if result.Interface != "eth1" {
		t.Errorf("expected interface 'eth1', got '%s'", result.Interface)
	}

	if result.LeaseTTL != 3600 {
		t.Errorf("expected LeaseTTL 3600, got %d", result.LeaseTTL)
	}

	if result.DomainName != "internal" {
		t.Errorf("expected DomainName 'internal', got '%s'", result.DomainName)
	}
}

func TestDHCPServer_UpdateOptions(t *testing.T) {
	server := NewDHCPServer()

	newOpts := &DHCPServerOpts{
		Interface:  "wlan0",
		StartFrom:  net.ParseIP("192.168.0.10").To4(),
		EndAt:      net.ParseIP("192.168.0.50").To4(),
		LeaseTTL:   7200,
		SubnetMask: net.ParseIP("255.255.255.0").To4(),
		Gateway:    net.ParseIP("192.168.0.1").To4(),
	}

	err := server.UpdateOptions(newOpts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if server.opts.Interface != "wlan0" {
		t.Errorf("expected interface 'wlan0', got '%s'", server.opts.Interface)
	}

	if server.opts.LeaseTTL != 7200 {
		t.Errorf("expected LeaseTTL 7200, got %d", server.opts.LeaseTTL)
	}
}
