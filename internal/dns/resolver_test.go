package dns

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

func TestDNSResolver_Blacklist_LookupBlockedDomainA(t *testing.T) {
	resolver := NewDNSResolverWithOpts(ResolverOpts{
		Upstreams:    []string{"1.1.1.1"},
		LocalDomains: make(map[string]net.IP),
	})
	resolver.blacklist["test.com"] = struct{}{}

	answers, authorities, err := resolver.Resolve("test.com", DNSTypeA)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(answers) != 1 {
		t.Errorf("expected 1 answer, got %d", len(answers))
	}
	if answers[0].Type != DNSTypeA {
		t.Errorf("expected A record, got %s", answers[0].Type)
	}
	if authorities != nil {
		t.Error("expected no authorities")
	}
}

func TestDNSResolver_Blacklist_LookupBlockedDomainAAAA(t *testing.T) {
	resolver := NewDNSResolverWithOpts(ResolverOpts{
		Upstreams:    []string{"1.1.1.1"},
		LocalDomains: make(map[string]net.IP),
	})
	resolver.blacklist["test.com"] = struct{}{}

	answers, _, err := resolver.Resolve("test.com", DNSTypeAAAA)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(answers) != 1 {
		t.Errorf("expected 1 answer, got %d", len(answers))
	}
	if answers[0].Type != DNSTypeAAAA {
		t.Errorf("expected AAAA record, got %s", answers[0].Type)
	}
}

func TestDNSResolver_Blacklist_LookupNotBlocked(t *testing.T) {
	resolver := NewDNSResolverWithOpts(ResolverOpts{
		Upstreams:    []string{"1.1.1.1"},
		LocalDomains: make(map[string]net.IP),
	})
	resolver.blacklist["blocked.com"] = struct{}{}

	_, _, err := resolver.Resolve("test.com", DNSTypeA)

	if err != ErrNxDomain && err != nil {
		t.Errorf("expected error or NXDOMAIN, got %v", err)
	}
}

func TestDNSResolver_Blacklist_AddEntries(t *testing.T) {
	resolver := NewDNSResolverWithOpts(ResolverOpts{
		Upstreams:    []string{"1.1.1.1"},
		LocalDomains: make(map[string]net.IP),
	})

	resolver.AddBlocklistEntries([]string{"a.com", "b.com", "c.com"})

	if _, found := resolver.blacklist["a.com"]; !found {
		t.Error("expected a.com to be in blacklist")
	}
	if _, found := resolver.blacklist["b.com"]; !found {
		t.Error("expected b.com to be in blacklist")
	}
	if _, found := resolver.blacklist["c.com"]; !found {
		t.Error("expected c.com to be in blacklist")
	}
	if len(resolver.blacklist) != 3 {
		t.Errorf("expected 3 entries, got %d", len(resolver.blacklist))
	}
}

func TestDNSResolver_Blacklist_DeleteEntry(t *testing.T) {
	resolver := NewDNSResolverWithOpts(ResolverOpts{
		Upstreams:    []string{"1.1.1.1"},
		LocalDomains: make(map[string]net.IP),
	})
	resolver.blacklist["a.com"] = struct{}{}
	resolver.blacklist["b.com"] = struct{}{}

	resolver.DeleteBlocklistEntry("a.com")

	if _, found := resolver.blacklist["a.com"]; found {
		t.Error("expected a.com to be removed from blacklist")
	}
	if _, found := resolver.blacklist["b.com"]; !found {
		t.Error("expected b.com to still be in blacklist")
	}
}

func TestDNSResolver_Blacklist_Flush(t *testing.T) {
	resolver := NewDNSResolverWithOpts(ResolverOpts{
		Upstreams:    []string{"1.1.1.1"},
		LocalDomains: make(map[string]net.IP),
	})
	resolver.blacklist["a.com"] = struct{}{}
	resolver.blacklist["b.com"] = struct{}{}
	resolver.blacklist["c.com"] = struct{}{}

	resolver.FlushBlocklist()

	if len(resolver.blacklist) != 0 {
		t.Errorf("expected blacklist to be empty, got %d entries", len(resolver.blacklist))
	}
}

func TestDNSResolver_Blacklist_EmptyLookup(t *testing.T) {
	resolver := NewDNSResolverWithOpts(ResolverOpts{
		Upstreams:    []string{"1.1.1.1"},
		LocalDomains: make(map[string]net.IP),
	})

	_, _, err := resolver.Resolve("test.com", DNSTypeA)

	if err != ErrNxDomain && err != nil {
		t.Errorf("expected error or NXDOMAIN, got %v", err)
	}
}

func TestResolverOpts_Defaults(t *testing.T) {
	opts := ResolverOpts{
		Upstreams: []string{"1.1.1.1"},
	}

	resolver := NewDNSResolverWithOpts(opts)

	if resolver.cacheTTL != time.Minute*5 {
		t.Errorf("expected default cacheTTL 5m, got %v", resolver.cacheTTL)
	}
	if resolver.upstreamTimeout != time.Second*5 {
		t.Errorf("expected default upstreamTimeout 5s, got %v", resolver.upstreamTimeout)
	}
	if resolver.dialTimeout != time.Second*2 {
		t.Errorf("expected default dialTimeout 2s, got %v", resolver.dialTimeout)
	}
	if resolver.readTimeout != time.Second*2 {
		t.Errorf("expected default readTimeout 2s, got %v", resolver.readTimeout)
	}
}

func TestResolverOpts_CustomTimeouts(t *testing.T) {
	opts := ResolverOpts{
		Upstreams:       []string{"1.1.1.1"},
		CacheTTL:        time.Minute * 10,
		UpstreamTimeout: time.Second * 10,
		DialTimeout:     time.Second * 3,
		ReadTimeout:     time.Second * 3,
	}

	resolver := NewDNSResolverWithOpts(opts)

	if resolver.cacheTTL != time.Minute*10 {
		t.Errorf("expected cacheTTL 10m, got %v", resolver.cacheTTL)
	}
	if resolver.upstreamTimeout != time.Second*10 {
		t.Errorf("expected upstreamTimeout 10s, got %v", resolver.upstreamTimeout)
	}
	if resolver.dialTimeout != time.Second*3 {
		t.Errorf("expected dialTimeout 3s, got %v", resolver.dialTimeout)
	}
	if resolver.readTimeout != time.Second*3 {
		t.Errorf("expected readTimeout 3s, got %v", resolver.readTimeout)
	}
}

// fakeUpstream is a UDP DNS server on 127.0.0.1 that answers each query
// with reply after delay, each in its own goroutine.
type fakeUpstream struct {
	conn    net.PacketConn
	queries chan string
	reply   func(query *DNSMessage) (*DNSMessage, time.Duration)
}

func startFakeUpstream(t *testing.T, reply func(query *DNSMessage) (*DNSMessage, time.Duration)) *fakeUpstream {
	t.Helper()
	conn, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	up := &fakeUpstream{conn: conn, queries: make(chan string, 100), reply: reply}
	t.Cleanup(func() { conn.Close() })
	go up.serve()
	return up
}

func (u *fakeUpstream) port() int {
	return u.conn.LocalAddr().(*net.UDPAddr).Port
}

func (u *fakeUpstream) serve() {
	for {
		buff := make([]byte, 1500)
		n, addr, err := u.conn.ReadFrom(buff)
		if err != nil {
			return
		}
		query, err := ParseDNSMessage(buff[:n])
		if err != nil {
			continue
		}
		select {
		case u.queries <- query.Questions[0].ParsedName:
		default:
		}
		go func() {
			response, delay := u.reply(query)
			if response == nil {
				return
			}
			time.Sleep(delay)
			dat, err := MarshalDNSMessage(response)
			if err != nil {
				return
			}
			u.conn.WriteTo(dat, addr)
		}()
	}
}

// answerA answers an A query with 192.0.2.1 and a 300 s TTL.
func answerA(query *DNSMessage) *DNSMessage {
	query.Header.SetQR(true)
	query.Header.SetRCODE(RCODESuccess)
	query.Answers = []*DNSRecord{{
		Name:  compressedDomainVal,
		Type:  DNSTypeA,
		Class: DNSClassIN,
		TTL:   300,
		RData: net.IPv4(192, 0, 2, 1).To4(),
	}}
	return query
}

// soaRecord is an SOA for example.test with the given TTL and MINIMUM.
func soaRecord(ttl, minimum uint32) *DNSRecord {
	rdata := append(stringToDNSWireFormat("ns.example.test"), stringToDNSWireFormat("hostmaster.example.test")...)
	fields := make([]byte, 20)
	binary.BigEndian.PutUint32(fields[0:4], 1)     // serial
	binary.BigEndian.PutUint32(fields[4:8], 7200)  // refresh
	binary.BigEndian.PutUint32(fields[8:12], 3600) // retry
	binary.BigEndian.PutUint32(fields[12:16], 86400)
	binary.BigEndian.PutUint32(fields[16:20], minimum)
	return &DNSRecord{
		Name:  stringToDNSWireFormat("example.test"),
		Type:  DNSTypeSOA,
		Class: DNSClassIN,
		TTL:   ttl,
		RData: append(rdata, fields...),
	}
}

func newTestResolver(t *testing.T, port int, opts ResolverOpts) *DNSResolver {
	t.Helper()
	if opts.Upstreams == nil {
		opts.Upstreams = []string{"127.0.0.1"}
	}
	opts.LocalDomains = make(map[string]net.IP)
	resolver := NewDNSResolverWithOpts(opts)
	resolver.upstreamPort = port
	t.Cleanup(resolver.Close)
	return resolver
}

func TestDNSResolver_ParallelQueriesResolveInParallel(t *testing.T) {
	up := startFakeUpstream(t, func(q *DNSMessage) (*DNSMessage, time.Duration) {
		return answerA(q), 300 * time.Millisecond
	})
	resolver := newTestResolver(t, up.port(), ResolverOpts{})

	start := time.Now()
	var wg sync.WaitGroup
	errs := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			answers, _, err := resolver.Resolve(fmt.Sprintf("host%d.example.test", i), DNSTypeA)
			if err == nil && len(answers) != 1 {
				err = fmt.Errorf("host%d: expected 1 answer, got %d", i, len(answers))
			}
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Error(err)
		}
	}
	// One at a time this takes 10 x 300 ms.
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("10 parallel queries took %v, expected well under 1s", elapsed)
	}
}

func TestDNSResolver_CachedNameAnswersDuringSlowLookup(t *testing.T) {
	up := startFakeUpstream(t, func(q *DNSMessage) (*DNSMessage, time.Duration) {
		return answerA(q), 3 * time.Second
	})
	resolver := newTestResolver(t, up.port(), ResolverOpts{ReadTimeout: 4 * time.Second})
	resolver.cache.set("cached.example.test|A", &DNSCacheItem{
		answers: []*DNSRecord{{Type: DNSTypeA, Class: DNSClassIN, TTL: 300, RData: net.IPv4(192, 0, 2, 2).To4()}},
		expires: time.Now().Add(time.Minute),
	})

	slowDone := make(chan error, 1)
	go func() {
		_, _, err := resolver.Resolve("slow.example.test", DNSTypeA)
		slowDone <- err
	}()
	select {
	case <-up.queries:
	case <-time.After(time.Second):
		t.Fatal("slow lookup never reached the upstream")
	}

	start := time.Now()
	answers, _, err := resolver.Resolve("cached.example.test", DNSTypeA)
	elapsed := time.Since(start)
	if err != nil || len(answers) != 1 {
		t.Fatalf("cached lookup: answers %d, err %v", len(answers), err)
	}
	if elapsed > 50*time.Millisecond {
		t.Errorf("cached lookup took %v while another lookup was in flight, expected under 50ms", elapsed)
	}

	if err := <-slowDone; err != nil {
		t.Errorf("slow lookup: %v", err)
	}
}

func TestDNSResolver_OneDeadlineForAllUpstreams(t *testing.T) {
	up := startFakeUpstream(t, func(q *DNSMessage) (*DNSMessage, time.Duration) {
		return nil, 0 // never answers
	})
	resolver := newTestResolver(t, up.port(), ResolverOpts{
		Upstreams:       []string{"127.0.0.1", "127.0.0.1"},
		UpstreamTimeout: 400 * time.Millisecond,
	})

	start := time.Now()
	_, _, err := resolver.Resolve("silent.example.test", DNSTypeA)
	elapsed := time.Since(start)
	if err != ErrNxDomain {
		t.Errorf("expected ErrNxDomain after the deadline, got %v", err)
	}
	// A fresh timeout per wait would take 2 x 400 ms.
	if elapsed > 650*time.Millisecond {
		t.Errorf("Resolve took %v, expected one 400ms deadline", elapsed)
	}
}

func TestDNSResolver_CachesAuthorities(t *testing.T) {
	up := startFakeUpstream(t, func(q *DNSMessage) (*DNSMessage, time.Duration) {
		q.Header.SetQR(true)
		q.Header.SetRCODE(RCODESuccess)
		q.Authorities = []*DNSRecord{soaRecord(120, 30)}
		return q, 0
	})
	resolver := newTestResolver(t, up.port(), ResolverOpts{})

	for i := 0; i < 2; i++ {
		answers, authorities, err := resolver.Resolve("nodata.example.test", DNSTypeAAAA)
		if err != nil || len(answers) != 0 || len(authorities) != 1 || authorities[0].Type != DNSTypeSOA {
			t.Fatalf("lookup %d: answers %d, authorities %d, err %v", i, len(answers), len(authorities), err)
		}
	}
	if n := len(up.queries); n != 1 {
		t.Errorf("expected 1 upstream query, got %d", n)
	}
}

func TestDNSResolver_NegativeCaching(t *testing.T) {
	tests := []struct {
		name    string
		soa     *DNSRecord
		wantTTL time.Duration // 0: not cached
	}{
		{"SOA minimum", soaRecord(3600, 30), 30 * time.Second},
		{"SOA TTL", soaRecord(10, 3600), 10 * time.Second},
		{"capped at 60s", soaRecord(3600, 3600), 60 * time.Second},
		{"no SOA", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			up := startFakeUpstream(t, func(q *DNSMessage) (*DNSMessage, time.Duration) {
				q.Header.SetQR(true)
				q.Header.SetRCODE(RCODENameFailure)
				if tt.soa != nil {
					q.Authorities = []*DNSRecord{tt.soa}
				}
				return q, 0
			})
			resolver := newTestResolver(t, up.port(), ResolverOpts{})

			for i := 0; i < 2; i++ {
				_, authorities, err := resolver.Resolve("missing.example.test", DNSTypeA)
				if err != ErrNxDomain {
					t.Fatalf("lookup %d: expected ErrNxDomain, got %v", i, err)
				}
				if tt.soa != nil && (len(authorities) != 1 || authorities[0].Type != DNSTypeSOA) {
					t.Fatalf("lookup %d: expected the SOA as authority, got %v", i, authorities)
				}
			}

			wantQueries := 1
			if tt.wantTTL == 0 {
				wantQueries = 2
			}
			if n := len(up.queries); n != wantQueries {
				t.Errorf("expected %d upstream queries, got %d", wantQueries, n)
			}
			item, ok := resolver.cache.get("missing.example.test|A", time.Now())
			if tt.wantTTL == 0 {
				if ok {
					t.Error("NXDOMAIN without SOA was cached")
				}
				return
			}
			if !ok {
				t.Fatal("NXDOMAIN was not cached")
			}
			if left := time.Until(item.expires); left > tt.wantTTL || left < tt.wantTTL-5*time.Second {
				t.Errorf("cached for %v, expected %v", left, tt.wantTTL)
			}
		})
	}
}

// ptrResolver has upstream as its only upstream, the given local domains and
// no conditional forwarders.
func ptrResolver(t *testing.T, upstream *fakeDNSServer, localDomains map[string]string) *DNSResolver {
	t.Helper()
	host, port, _ := net.SplitHostPort(upstream.addr)
	locals := make(map[string]net.IP, len(localDomains))
	for name, ip := range localDomains {
		locals[name] = net.ParseIP(ip).To4()
	}
	r := NewDNSResolverWithOpts(ResolverOpts{
		Upstreams:    []string{host},
		LocalDomains: locals,
		ReadTimeout:  time.Millisecond * 200,
	})
	fmt.Sscan(port, &r.upstreamPort)
	t.Cleanup(r.Close)
	return r
}

func TestPrivatePTRChecksEveryLocalDomain(t *testing.T) {
	upstream := newFakeDNSServer(t, replyWith(RCODESuccess, ptrRecord("public.example", 60)))
	locals := map[string]string{}
	for i := 1; i <= 8; i++ {
		locals[fmt.Sprintf("host%d.lan", i)] = fmt.Sprintf("192.168.1.%d", i)
	}
	r := ptrResolver(t, upstream, locals)

	for i := 1; i <= 8; i++ {
		for _, name := range []string{
			fmt.Sprintf("%d.1.168.192.in-addr.arpa", i),
			fmt.Sprintf("%d.1.168.192.IN-ADDR.ARPA.", i),
		} {
			answers, _, err := r.Resolve(name, DNSTypePTR)
			want := fmt.Sprintf("host%d.lan", i)
			if err != nil || len(answers) != 1 || answers[0].ParsedName != want {
				t.Errorf("%s: expected %s, got %v %v", name, want, answers, err)
			}
		}
	}
	if n := upstream.queries.Load(); n != 0 {
		t.Errorf("expected no upstream queries, got %d", n)
	}
}

func TestUnmatchedPrivatePTRIsNXDOMAIN(t *testing.T) {
	for name, locals := range map[string]map[string]string{
		"no local domains":    {},
		"other local domains": {"box.lan": "10.0.0.2", "nas.lan": "10.0.0.3"},
	} {
		t.Run(name, func(t *testing.T) {
			upstream := newFakeDNSServer(t, replyWith(RCODESuccess, ptrRecord("public.example", 60)))
			r := ptrResolver(t, upstream, locals)
			answers, authorities, err := r.Resolve("99.0.0.10.in-addr.arpa", DNSTypePTR)
			if err != ErrNxDomain || len(answers) != 0 || len(authorities) != 0 {
				t.Fatalf("expected NXDOMAIN without records, got %v %v %v", answers, authorities, err)
			}
			if n := upstream.queries.Load(); n != 0 {
				t.Errorf("expected no upstream queries, got %d", n)
			}
		})
	}
}

func TestOtherPTRNamesGoUpstream(t *testing.T) {
	upstream := newFakeDNSServer(t, replyWith(RCODESuccess, ptrRecord("upstream.example", 60)))
	r := ptrResolver(t, upstream, map[string]string{"box.lan": "192.168.0.1"})

	names := []string{
		"8.8.8.8.in-addr.arpa",                    // public address
		"1.0.0.127.in-addr.arpa",                  // loopback, not private
		"0.168.192.in-addr.arpa",                  // a zone, not an address
		"b._dns-sd._udp.0.0.168.192.in-addr.arpa", // DNS-SD browse
		"1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.ip6.arpa",
		"lb._dns-sd._udp.lan",
	}
	for i, name := range names {
		answers, _, err := r.Resolve(name, DNSTypePTR)
		if err != nil || len(answers) != 1 || string(answers[0].RData) != string(stringToDNSWireFormat("upstream.example")) {
			t.Errorf("%s: expected the upstream answer, got %v %v", name, answers, err)
		}
		if n := upstream.queries.Load(); int(n) != i+1 {
			t.Errorf("%s: expected %d upstream queries, got %d", name, i+1, n)
		}
	}
}
