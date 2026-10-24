package dhcp

import (
	"bytes"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/thatjames/gatekeeper-go/internal/config"
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

func TestDHCPServer_LeaseFilePath(t *testing.T) {
	server := NewDHCPServerWithOpts(&DHCPServerOpts{
		StartFrom: net.ParseIP("10.0.0.10").To4(),
		EndAt:     net.ParseIP("10.0.0.20").To4(),
	})
	if server.leaseFile != "/var/lib/gatekeeper/leases" {
		t.Errorf("empty LeaseFile resolved to %q, want the default", server.leaseFile)
	}
	opts := *server.Options()
	opts.LeaseFile = "/data/leases"
	server.UpdateOptions(&opts)
	if server.leaseFile != "/data/leases" {
		t.Errorf("after UpdateOptions the lease file is %q, want /data/leases", server.leaseFile)
	}
}

// testServer is a DHCP server on a synthetic range that is not started; requests go straight to handleRequest.
func testServer(start, end string) *DHCPServer {
	z := NewDHCPServerWithOpts(&DHCPServerOpts{
		StartFrom: net.ParseIP(start).To4(),
		EndAt:     net.ParseIP(end).To4(),
		LeaseTTL:  300,
	})
	z.interfaceAddr = net.ParseIP("10.0.0.254").To4()
	return z
}

func testRequest(t *testing.T, msgType DHCPMessageType, mac string, extra Options) *DHCPPacket {
	t.Helper()
	hw, err := net.ParseMAC(mac)
	if err != nil {
		t.Fatal(err)
	}
	m := NewMessage(OpRequest)[:240]
	m.SetHLen(len(hw))
	m.SetCHAddr(hw)
	opts := Options{OptionDHCPMessageType: []byte{byte(msgType)}}
	for tag, val := range extra {
		opts[tag] = val
	}
	return &DHCPPacket{Message: append(m, MarshalOptions(opts)...)}
}

func TestDHCPServer_ReservedClientReleaseKeepsReservation(t *testing.T) {
	const mac = "02:00:00:00:00:01"
	const reservedIP = "10.0.0.5"
	cfgFile := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(cfgFile, []byte("DHCP:\n  ReservedAddresses:\n    \""+mac+"\": "+reservedIP+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	saved := config.Config
	t.Cleanup(func() { config.Config = saved })
	if err := config.LoadConfig(cfgFile); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(cfgFile)

	z := testServer("10.0.0.1", "10.0.0.10")
	for clientID, ip := range config.Config.DHCP.ReservedAddresses {
		if err := z.issuedLeases.ReserveLease(clientID, net.ParseIP(ip).To4()); err != nil {
			t.Fatal(err)
		}
	}

	if resp := z.handleRequest(testRequest(t, DHCPRelease, mac, nil)); resp != nil {
		t.Fatalf("expected no reply to RELEASE, got %v", resp)
	}
	resp := z.handleRequest(testRequest(t, DHCPDiscover, mac, nil))
	if resp == nil {
		t.Fatalf("expected OFFER of %s after RELEASE, got none", reservedIP)
	} else if !resp.YIAddr().Equal(net.ParseIP(reservedIP)) {
		t.Fatalf("expected OFFER of %s after RELEASE, got %s", reservedIP, resp.YIAddr())
	}
	resp = z.handleRequest(testRequest(t, DHCPRequest, mac, Options{OptionRequestedIPAddress: net.ParseIP(reservedIP).To4()}))
	if resp == nil || DHCPMessageType(ParseOptions(resp)[OptionDHCPMessageType][0]) != DHCPAck {
		t.Fatalf("expected ACK of %s after RELEASE", reservedIP)
	}
	if got := config.Config.DHCP.ReservedAddresses[mac]; got != reservedIP {
		t.Errorf("config reservation changed to %q", got)
	}
	if after, _ := os.ReadFile(cfgFile); !bytes.Equal(before, after) {
		t.Errorf("config file was rewritten on RELEASE")
	}
}

func TestDHCPServer_FullPoolNeverOffersReservedIP(t *testing.T) {
	z := testServer("10.0.0.1", "10.0.0.3")
	reservedIP := net.ParseIP("10.0.0.2").To4()
	if err := z.issuedLeases.ReserveLease("02:00:00:00:00:10", reservedIP); err != nil {
		t.Fatal(err)
	}
	offered := 0
	for _, mac := range []string{"02:00:00:00:00:11", "02:00:00:00:00:12", "02:00:00:00:00:13"} {
		resp := z.handleRequest(testRequest(t, DHCPDiscover, mac, nil))
		if resp == nil {
			continue
		}
		offered++
		if resp.YIAddr().Equal(reservedIP) {
			t.Errorf("client %s was offered reserved IP %s", mac, reservedIP)
		}
		z.handleRequest(testRequest(t, DHCPRequest, mac, Options{OptionRequestedIPAddress: resp.YIAddr().To4()}))
	}
	if offered != 2 {
		t.Errorf("expected 2 offers from a 3-address pool with one reserved, got %d", offered)
	}
}
