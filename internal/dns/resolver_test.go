package dns

import (
	"net"
	"testing"
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
