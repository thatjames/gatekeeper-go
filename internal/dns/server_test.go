package dns

import (
	"net"
	"testing"
)

type mockResolver struct {
	localDomains      map[string]net.IP
	blocklist         map[string]struct{}
	addLocalDomainErr error
}

func newMockResolver() *mockResolver {
	return &mockResolver{
		localDomains: make(map[string]net.IP),
		blocklist:    make(map[string]struct{}),
	}
}

func (m *mockResolver) Resolve(domain string, dnsType DNSType) (answers, authorities []*DNSRecord, err error) {
	return nil, nil, nil
}

func (m *mockResolver) AddLocalDomain(domain string, ip net.IP) error {
	if m.addLocalDomainErr != nil {
		return m.addLocalDomainErr
	}
	m.localDomains[domain] = ip
	return nil
}

func (m *mockResolver) DeleteLocalDomain(domain string) {
	delete(m.localDomains, domain)
}

func (m *mockResolver) AddBlocklistEntries(entries []string) {
	for _, entry := range entries {
		m.blocklist[entry] = struct{}{}
	}
}

func (m *mockResolver) DeleteBlocklistEntry(domain string) {
	delete(m.blocklist, domain)
}

func (m *mockResolver) FlushBlocklist() {
	m.blocklist = make(map[string]struct{})
}

func TestNewDNSServer_DefaultOpts(t *testing.T) {
	server := NewDNSServer()

	if server.opts == nil {
		t.Error("expected opts to be set")
	}

	if server.opts.Interface != "eth0" {
		t.Errorf("expected interface 'eth0', got '%s'", server.opts.Interface)
	}

	if server.opts.Port != 53 {
		t.Errorf("expected port 53, got %d", server.opts.Port)
	}

	if server.resolver == nil {
		t.Error("expected resolver to be created")
	}

	if server.blocklistFetcher == nil {
		t.Error("expected blocklistFetcher to be created")
	}
}

func TestNewDNSServer_WithResolver(t *testing.T) {
	mock := newMockResolver()
	opts := DNSServerOpts{
		Interface: "lo",
		Port:      5353,
	}

	server := NewDNSServerWithOpts(opts, mock, nil)

	if server.resolver != mock {
		t.Error("expected resolver to be the mock resolver")
	}
}

func TestNewDNSServer_WithFetcher(t *testing.T) {
	fetcher := NewHTTPBlocklistFetcher()
	opts := DNSServerOpts{}

	server := NewDNSServerWithOpts(opts, nil, fetcher)

	if server.blocklistFetcher != fetcher {
		t.Error("expected blocklistFetcher to be the provided fetcher")
	}
}

func TestDNSServer_AddLocalDomain(t *testing.T) {
	mock := newMockResolver()
	server := NewDNSServerWithOpts(DNSServerOpts{}, mock, nil)

	err := server.AddLocalDomain("test.example.com", "192.168.1.1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if _, ok := mock.localDomains["test.example.com"]; !ok {
		t.Error("expected domain to be added to resolver")
	}
}

func TestDNSServer_AddLocalDomain_InvalidIP(t *testing.T) {
	mock := newMockResolver()
	server := NewDNSServerWithOpts(DNSServerOpts{}, mock, nil)

	err := server.AddLocalDomain("test.example.com", "invalid-ip")
	if err == nil {
		t.Error("expected error for invalid IP")
	}
}

func TestDNSServer_DeleteLocalDomain(t *testing.T) {
	mock := newMockResolver()
	mock.localDomains["test.example.com"] = net.ParseIP("192.168.1.1")
	server := NewDNSServerWithOpts(DNSServerOpts{}, mock, nil)

	server.DeleteLocalDomain("test.example.com")

	if _, ok := mock.localDomains["test.example.com"]; ok {
		t.Error("expected domain to be deleted from resolver")
	}
}

func TestDNSServer_AddBlockedDomain(t *testing.T) {
	mock := newMockResolver()
	server := NewDNSServerWithOpts(DNSServerOpts{}, mock, nil)

	server.AddBlockedDomain("malicious.com")

	if _, found := mock.blocklist["malicious.com"]; !found {
		t.Error("expected domain to be added to blocklist")
	}
}

func TestDNSServer_DeleteBlockedDomain(t *testing.T) {
	mock := newMockResolver()
	mock.blocklist = map[string]struct{}{"malicious.com": {}, "ads.com": {}}
	server := NewDNSServerWithOpts(DNSServerOpts{}, mock, nil)

	server.DeleteBlockedDomain("malicious.com")

	if _, found := mock.blocklist["malicious.com"]; found {
		t.Error("expected domain to be removed from blocklist")
	}
}

func TestDNSServer_FlushBlocklist(t *testing.T) {
	mock := newMockResolver()
	mock.blocklist = map[string]struct{}{"a.com": {}, "b.com": {}, "c.com": {}}
	server := NewDNSServerWithOpts(DNSServerOpts{}, mock, nil)

	server.FlushBlocklist()

	if len(mock.blocklist) != 0 {
		t.Errorf("expected blocklist to be empty, got %d entries", len(mock.blocklist))
	}
}

func TestDNSServer_Options(t *testing.T) {
	opts := DNSServerOpts{
		Interface: "lo",
		Port:      5353,
		Upstream:  []string{"8.8.8.8"},
	}
	server := NewDNSServerWithOpts(opts, nil, nil)

	result := server.Options()

	if result.Interface != "lo" {
		t.Errorf("expected interface 'lo', got '%s'", result.Interface)
	}

	if result.Port != 5353 {
		t.Errorf("expected port 5353, got %d", result.Port)
	}
}

func TestDNSServer_UpdateOptions(t *testing.T) {
	server := NewDNSServer()

	newOpts := &DNSServerOpts{
		Interface: "wlan0",
		Port:      5353,
	}
	server.UpdateOptions(newOpts)

	if server.opts.Interface != "wlan0" {
		t.Errorf("expected interface 'wlan0', got '%s'", server.opts.Interface)
	}
}
