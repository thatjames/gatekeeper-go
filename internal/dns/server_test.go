package dns

import (
	"net"
	"strings"
	"testing"
	"time"
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

	if server.opts.Interface != "" {
		t.Errorf("expected no interface, got '%s'", server.opts.Interface)
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

func TestDNSServerOpts_ReadDeadline(t *testing.T) {
	opts := DNSServerOpts{
		ReadDeadline: time.Second * 5,
	}
	server := NewDNSServerWithOpts(opts, nil, nil)

	if server.readDeadline != time.Second*5 {
		t.Errorf("expected readDeadline 5s, got %v", server.readDeadline)
	}
}

func TestDNSServerOpts_DefaultReadDeadline(t *testing.T) {
	opts := DNSServerOpts{}
	server := NewDNSServerWithOpts(opts, nil, nil)

	if server.readDeadline != time.Second*2 {
		t.Errorf("expected default readDeadline 2s, got %v", server.readDeadline)
	}
}

func TestNewDNSServerWithOptsPassesResolverOptsThrough(t *testing.T) {
	server := NewDNSServerWithOpts(DNSServerOpts{
		Upstream: []string{"9.9.9.9"},
		ResolverOpts: &ResolverOpts{
			LocalDomains: map[string]net.IP{
				"router": net.ParseIP("192.168.1.1").To4(),
			},
		},
	}, nil, nil)

	resolver, ok := server.resolver.(*DNSResolver)
	if !ok {
		t.Fatalf("expected a *DNSResolver, got %T", server.resolver)
	}

	if got := resolver.localDomains["router"]; !got.Equal(net.ParseIP("192.168.1.1")) {
		t.Fatalf("expected local domain router to resolve to 192.168.1.1, got %v", got)
	}

	// Upstreams fall back to the server's when the resolver opts omit them.
	if len(resolver.upstream) != 1 || !resolver.upstream[0].Equal(net.ParseIP("9.9.9.9")) {
		t.Fatalf("expected upstream 9.9.9.9, got %v", resolver.upstream)
	}

	answers, _, err := server.resolver.Resolve("router", DNSTypeA)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(answers) != 1 || !net.IP(answers[0].RData).Equal(net.ParseIP("192.168.1.1")) {
		t.Fatalf("expected a local answer of 192.168.1.1, got %v", answers)
	}
}

// answeredQuery is a query for big.test with count A answers filled in.
func answeredQuery(count int, additionals ...*DNSRecord) *DNSMessage {
	msg := NewDnsMessage()
	msg.Questions = append(msg.Questions, &DNSQuestion{Name: stringToDNSWireFormat("big.test"), Type: DNSTypeA, Class: DNSClassIN, ParsedName: "big.test"})
	for i := 0; i < count; i++ {
		msg.Answers = append(msg.Answers, &DNSRecord{Name: compressedDomainVal, Type: DNSTypeA, Class: DNSClassIN, TTL: 60, RData: []byte{10, 0, byte(i >> 8), byte(i)}})
	}
	msg.Additionals = append(msg.Additionals, additionals...)
	return msg
}

func TestUDPSizeLimit(t *testing.T) {
	cases := map[string]struct {
		additionals []*DNSRecord
		want        int
	}{
		"no EDNS":    {nil, 512},
		"EDNS small": {[]*DNSRecord{{Type: DNSTypeOPT, Class: 256}}, 512},
		"EDNS 1000":  {[]*DNSRecord{{Type: DNSTypeOPT, Class: 1000}}, 1000},
		"EDNS 4096":  {[]*DNSRecord{{Type: DNSTypeOPT, Class: 4096}}, 1232},
	}
	for name, c := range cases {
		if got := udpSizeLimit(answeredQuery(0, c.additionals...)); got != c.want {
			t.Errorf("%s: expected %d, got %d", name, c.want, got)
		}
	}
}

func TestPrepareReplyTruncatesOversizedUDP(t *testing.T) {
	// 12 header + 14 question + 40 answers of 16 bytes = 666 bytes
	data, err := prepareReply(answeredQuery(40), 512)
	if err != nil {
		t.Fatal(err)
	}
	reply, err := ParseDNSMessage(data)
	if err != nil {
		t.Fatal(err)
	}
	if reply.Header.Flags&0x0200 == 0 {
		t.Error("expected TC set")
	}
	if len(reply.Answers) != 0 || len(reply.Questions) != 1 {
		t.Errorf("expected question only, got %d questions and %d answers", len(reply.Questions), len(reply.Answers))
	}
}

func TestPrepareReplyKeepsFittingAndTCPReplies(t *testing.T) {
	for name, limit := range map[string]int{"fits UDP": 1232, "TCP": 0} {
		data, err := prepareReply(answeredQuery(40), limit)
		if err != nil {
			t.Fatal(err)
		}
		reply, err := ParseDNSMessage(data)
		if err != nil {
			t.Fatal(err)
		}
		if reply.Header.Flags&0x0200 != 0 || len(reply.Answers) != 40 {
			t.Errorf("%s: expected 40 answers without TC, got %d (flags %#x)", name, len(reply.Answers), reply.Header.Flags)
		}
	}
}

func TestPrepareReplyAnswersEDNS(t *testing.T) {
	cases := map[string]struct {
		query    *DNSMessage
		limit    int
		wantOPT  bool
		wantTC   bool
		wantAnsw int
	}{
		"EDNS query":       {answeredQuery(1, &DNSRecord{Type: DNSTypeOPT, Class: 4096, TTL: 0x8000}), 1232, true, false, 1},
		"plain query":      {answeredQuery(1), 512, false, false, 1},
		"EDNS truncated":   {answeredQuery(100, &DNSRecord{Type: DNSTypeOPT, Class: 1232}), 1232, true, true, 0},
		"other additional": {answeredQuery(1, &DNSRecord{Name: compressedDomainVal, Type: DNSTypeA, Class: DNSClassIN, RData: []byte{1, 2, 3, 4}}), 512, false, false, 1},
	}
	for name, c := range cases {
		data, err := prepareReply(c.query, c.limit)
		if err != nil {
			t.Fatal(err)
		}
		reply, err := ParseDNSMessage(data)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		opt := queryOPT(reply)
		if (opt != nil) != c.wantOPT || len(reply.Additionals) > 1 {
			t.Errorf("%s: expected OPT %v, got additionals %v", name, c.wantOPT, reply.Additionals)
		}
		if opt != nil && (opt.Class != 1232 || opt.TTL != 0 || len(opt.RData) != 0) {
			t.Errorf("%s: unexpected OPT size %d ttl %#x rdata %v", name, opt.Class, opt.TTL, opt.RData)
		}
		if (reply.Header.Flags&0x0200 != 0) != c.wantTC || len(reply.Answers) != c.wantAnsw {
			t.Errorf("%s: expected TC %v and %d answers, got flags %#x and %d answers", name, c.wantTC, c.wantAnsw, reply.Header.Flags, len(reply.Answers))
		}
	}
}

func TestListenHost(t *testing.T) {
	if host, err := listenHost(""); err != nil || host != "" {
		t.Errorf("no interface: expected all addresses, got %q %v", host, err)
	}
	if host, err := listenHost("lo"); err != nil || host != "127.0.0.1" {
		t.Errorf("lo: expected 127.0.0.1, got %q %v", host, err)
	}
	if _, err := listenHost("gk-missing0"); err == nil {
		t.Error("missing interface: expected an error")
	}
}

func TestStartBindsToInterfaceAddress(t *testing.T) {
	server := NewDNSServerWithOpts(DNSServerOpts{Interface: "lo", Port: 0}, NewDNSResolverWithDefaultOpts(), nil)
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	defer server.Stop()
	if addr := server.packetConn.LocalAddr().String(); !strings.HasPrefix(addr, "127.0.0.1:") {
		t.Errorf("expected UDP on 127.0.0.1, got %s", addr)
	}
	if addr := server.tcpListener.Addr().String(); addr != server.packetConn.LocalAddr().String() {
		t.Errorf("expected TCP on the UDP address %s, got %s", server.packetConn.LocalAddr(), addr)
	}

	missing := NewDNSServerWithOpts(DNSServerOpts{Interface: "gk-missing0", Port: 0}, NewDNSResolverWithDefaultOpts(), nil)
	if err := missing.Start(); err == nil {
		missing.Stop()
		t.Error("expected Start to fail for a missing interface")
	}
}
