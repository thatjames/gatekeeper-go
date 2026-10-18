package dns

import (
	"net"
	"sync/atomic"
	"testing"
	"time"
)

func TestParseForwarders(t *testing.T) {
	parsed := parseForwarders([]ConditionalForwarder{
		{Zone: "styx", Server: "10.20.1.1"},
		{Zone: "1.20.10.in-addr.arpa.", Server: "10.20.1.1:5353"},
		{Zone: "", Server: "10.20.1.1"},
		{Zone: "bad", Server: "hecate"},
		{Zone: "bad6", Server: "[::1]:53"},
		{Zone: "badport", Server: "10.20.1.1:0"},
		{Zone: "Lab.Styx", Server: "10.20.2.1"},
	})
	want := []conditionalForwarder{
		{zone: "1.20.10.in-addr.arpa", addr: "10.20.1.1:5353"},
		{zone: "lab.styx", addr: "10.20.2.1:53"},
		{zone: "styx", addr: "10.20.1.1:53"},
	}
	if len(parsed) != len(want) {
		t.Fatalf("expected %v, got %v", want, parsed)
	}
	for i := range want {
		if parsed[i] != want[i] {
			t.Errorf("forwarder %d: expected %v, got %v", i, want[i], parsed[i])
		}
	}
}

func TestForwarderFor(t *testing.T) {
	r := NewDNSResolverWithOpts(ResolverOpts{ConditionalForwarders: []ConditionalForwarder{
		{Zone: "styx", Server: "10.20.1.1"},
		{Zone: "lab.styx", Server: "10.20.2.1"},
	}})
	cases := map[string]string{
		"styx":          "10.20.1.1:53",
		"atlas.styx":    "10.20.1.1:53",
		"ATLAS.Styx.":   "10.20.1.1:53",
		"pi.lab.styx":   "10.20.2.1:53",
		"notstyx":       "",
		"styx.example":  "",
		"www.google.nl": "",
	}
	for name, want := range cases {
		got := ""
		if f := r.forwarderFor(name); f != nil {
			got = f.addr
		}
		if got != want {
			t.Errorf("%s: expected %q, got %q", name, want, got)
		}
	}
}

// forwardingResolver has one forwarder for styx and the reverse zone of
// 10.20.1.0/24 at forwarder, and upstream as its only upstream.
func forwardingResolver(forwarder, upstream *fakeDNSServer) *DNSResolver {
	upstreamHost, _, _ := net.SplitHostPort(upstream.addr)
	r := NewDNSResolverWithOpts(ResolverOpts{
		Upstreams: []string{upstreamHost},
		ConditionalForwarders: []ConditionalForwarder{
			{Zone: "styx", Server: forwarder.addr},
			{Zone: "1.20.10.in-addr.arpa", Server: forwarder.addr},
		},
		LocalDomains: map[string]net.IP{"gatekeeper.local": net.ParseIP("10.20.1.5").To4()},
		ReadTimeout:  time.Millisecond * 200,
	})
	return r
}

func TestForwardAnswersFromForwarderOnly(t *testing.T) {
	forwarder := newFakeDNSServer(t, replyWith(RCODESuccess, aRecord("10.20.1.10", 60)))
	upstream := newFakeDNSServer(t, replyWith(RCODESuccess, aRecord("1.2.3.4", 60)))
	r := forwardingResolver(forwarder, upstream)

	answers, _, err := r.Resolve("atlas.styx", DNSTypeA)
	if err != nil || len(answers) != 1 || net.IP(answers[0].RData).String() != "10.20.1.10" {
		t.Fatalf("expected 10.20.1.10, got %v %v", answers, err)
	}
	if _, _, err := r.Resolve("atlas.styx", DNSTypeA); err != nil {
		t.Fatal(err)
	}
	if n := forwarder.queries.Load(); n != 1 {
		t.Errorf("expected 1 forwarder query (second one cached), got %d", n)
	}
	if n := upstream.queries.Load(); n != 0 {
		t.Errorf("expected no upstream queries, got %d", n)
	}
}

func TestForwardReverseLookup(t *testing.T) {
	forwarder := newFakeDNSServer(t, replyWith(RCODESuccess, ptrRecord("atlas.styx", 60)))
	upstream := newFakeDNSServer(t, replyWith(RCODENameFailure))
	r := forwardingResolver(forwarder, upstream)

	answers, _, err := r.Resolve("10.1.20.10.in-addr.arpa", DNSTypePTR)
	if err != nil || len(answers) != 1 {
		t.Fatalf("expected one PTR answer, got %v %v", answers, err)
	}
	if forwarder.queries.Load() != 1 || upstream.queries.Load() != 0 {
		t.Errorf("expected only the forwarder asked, got forwarder %d upstream %d", forwarder.queries.Load(), upstream.queries.Load())
	}

	// a local domain's own address is answered locally
	answers, _, err = r.Resolve("5.1.20.10.in-addr.arpa", DNSTypePTR)
	if err != nil || len(answers) != 1 || answers[0].ParsedName != "gatekeeper.local" {
		t.Fatalf("expected gatekeeper.local, got %v %v", answers, err)
	}
	if forwarder.queries.Load() != 1 {
		t.Errorf("expected the local PTR not forwarded, got %d forwarder queries", forwarder.queries.Load())
	}
}

func TestForwardFailures(t *testing.T) {
	upstream := newFakeDNSServer(t, replyWith(RCODESuccess, aRecord("1.2.3.4", 60)))
	cases := map[string]struct {
		handler func(*DNSMessage, bool) *DNSMessage
		want    error
	}{
		"NXDOMAIN passed on": {replyWith(RCODENameFailure), ErrNxDomain},
		"SERVFAIL":           {replyWith(RCODEServerFailure), ErrDNSServerFailure},
		"no answer":          {func(*DNSMessage, bool) *DNSMessage { return nil }, ErrDNSServerFailure},
		"empty NOERROR":      {replyWith(RCODESuccess), nil},
	}
	for name, c := range cases {
		forwarder := newFakeDNSServer(t, c.handler)
		answers, _, err := forwardingResolver(forwarder, upstream).Resolve("atlas.styx", DNSTypeAAAA)
		if err != c.want || len(answers) != 0 {
			t.Errorf("%s: expected %v and no answers, got %v %v", name, c.want, answers, err)
		}
	}
	if n := upstream.queries.Load(); n != 0 {
		t.Errorf("expected no upstream queries, got %d", n)
	}
}

func TestUnforwardedNamesGoUpstream(t *testing.T) {
	forwarder := newFakeDNSServer(t, replyWith(RCODESuccess, aRecord("10.20.1.10", 60)))
	upstream := newFakeDNSServer(t, replyWith(RCODESuccess, aRecord("1.2.3.4", 60)))
	r := forwardingResolver(forwarder, upstream)
	// the fake upstream listens on a random port, but upstreams are asked on 53
	r.upstream = nil
	if _, _, err := r.Resolve("example.com", DNSTypeA); err != ErrNxDomain {
		t.Errorf("expected the upstream path (no upstreams: NXDOMAIN), got %v", err)
	}
	if n := forwarder.queries.Load(); n != 0 {
		t.Errorf("expected no forwarder queries, got %d", n)
	}
}

func TestExchangeSendsEDNSAndRetriesTruncatedOverTCP(t *testing.T) {
	many := make([]*DNSRecord, 0, 100)
	for i := 0; i < 100; i++ {
		many = append(many, aRecord(net.IPv4(10, 0, 0, byte(i)).String(), 60))
	}
	var sawOPT atomic.Bool
	server := newFakeDNSServer(t, func(query *DNSMessage, tcp bool) *DNSMessage {
		if opt := queryOPT(query); opt != nil && opt.Class == maxUDPSize {
			sawOPT.Store(true)
		}
		reply := &DNSMessage{Header: &DNSHeader{}, Questions: query.Questions}
		if tcp {
			reply.Answers = many
		} else {
			reply.Header.SetTC(true)
		}
		return reply
	})
	r := NewDNSResolverWithOpts(ResolverOpts{ReadTimeout: time.Second})

	answers, _, err := r.exchange("big.test", DNSTypeA, server.addr)
	if err != nil || len(answers) != 100 {
		t.Fatalf("expected 100 answers over TCP, got %d %v", len(answers), err)
	}
	if !sawOPT.Load() {
		t.Error("expected an EDNS OPT with UDP size 1232 in the query")
	}
	if n := server.queries.Load(); n != 2 {
		t.Errorf("expected a UDP and a TCP query, got %d", n)
	}
}
