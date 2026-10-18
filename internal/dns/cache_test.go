package dns

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestDNSCache_StaysAtLimit(t *testing.T) {
	resolver := NewDNSResolverWithDefaultOpts()
	defer resolver.Close()

	now := time.Now()
	for i := 0; i < defaultCacheLimit+500; i++ {
		resolver.cache.set(fmt.Sprintf("host%d.example.test|A", i), &DNSCacheItem{expires: now.Add(time.Hour)})
		if n := resolver.cache.len(); n > defaultCacheLimit {
			t.Fatalf("cache holds %d entries after %d inserts, limit %d", n, i+1, defaultCacheLimit)
		}
	}
	if n := resolver.cache.len(); n != defaultCacheLimit {
		t.Errorf("expected %d entries, got %d", defaultCacheLimit, n)
	}
}

func TestDNSCache_EvictsNearestExpiry(t *testing.T) {
	cache := newDNSCache(100)
	now := time.Now()
	for _, i := range rand.Perm(250) {
		cache.set(fmt.Sprintf("host%d|A", i), &DNSCacheItem{expires: now.Add(time.Duration(i+1) * time.Second)})
	}
	if n := cache.len(); n != 100 {
		t.Fatalf("expected 100 entries, got %d", n)
	}
	// Every insert into the full cache evicts the entry nearest expiry, so
	// what is left is a subset of the latest expiries; the very latest
	// inserted after the cache filled must have survived.
	if _, ok := cache.get("host249|A", now); !ok {
		t.Error("the entry with the latest expiry was evicted")
	}
	for key, item := range cache.items {
		for _, other := range cache.byExpiry {
			if other.key != key && other.index == item.index {
				t.Fatalf("heap index %d shared by %s and %s", item.index, key, other.key)
			}
		}
	}
}

func TestDNSCache_FullCacheEvictsEarliest(t *testing.T) {
	cache := newDNSCache(3)
	now := time.Now()
	cache.set("late|A", &DNSCacheItem{expires: now.Add(3 * time.Minute)})
	cache.set("early|A", &DNSCacheItem{expires: now.Add(1 * time.Minute)})
	cache.set("middle|A", &DNSCacheItem{expires: now.Add(2 * time.Minute)})
	cache.set("new|A", &DNSCacheItem{expires: now.Add(4 * time.Minute)})

	if _, ok := cache.get("early|A", now); ok {
		t.Error("expected the entry nearest expiry to be evicted")
	}
	for _, key := range []string{"late|A", "middle|A", "new|A"} {
		if _, ok := cache.get(key, now); !ok {
			t.Errorf("expected %s to stay", key)
		}
	}
}

func TestDNSCache_SetReplacesKey(t *testing.T) {
	cache := newDNSCache(10)
	now := time.Now()
	cache.set("a|A", &DNSCacheItem{expires: now.Add(time.Minute)})
	cache.set("a|A", &DNSCacheItem{expires: now.Add(time.Hour), nxDomain: true})

	if n := cache.len(); n != 1 || len(cache.byExpiry) != 1 {
		t.Fatalf("expected 1 entry, got %d (heap %d)", n, len(cache.byExpiry))
	}
	if item, ok := cache.get("a|A", now); !ok || !item.nxDomain {
		t.Error("expected the second set to replace the first")
	}
}

func TestDNSCache_ExpiredIsMiss(t *testing.T) {
	cache := newDNSCache(10)
	now := time.Now()
	cache.set("old|A", &DNSCacheItem{expires: now.Add(-time.Second)})
	if _, ok := cache.get("old|A", now); ok {
		t.Error("expected an expired entry to be a miss")
	}
}

func TestDNSCache_SweepDeletesExpired(t *testing.T) {
	cache := newDNSCache(10)
	now := time.Now()
	cache.set("gone1|A", &DNSCacheItem{expires: now.Add(-time.Minute)})
	cache.set("gone2|A", &DNSCacheItem{expires: now})
	cache.set("kept|A", &DNSCacheItem{expires: now.Add(time.Minute)})

	cache.sweep(now)

	if n := cache.len(); n != 1 {
		t.Fatalf("expected 1 entry after the sweep, got %d", n)
	}
	if _, ok := cache.get("kept|A", now); !ok {
		t.Error("the sweep deleted an entry that had not expired")
	}
}

func TestDNSResolver_SweepRunsOnTimer(t *testing.T) {
	resolver := NewDNSResolverWithDefaultOpts()
	resolver.Close() // stop the minute sweep; run a fast one instead
	stop := make(chan struct{})
	defer close(stop)

	resolver.cache.set("gone|A", &DNSCacheItem{expires: time.Now().Add(20 * time.Millisecond)})
	go resolver.sweepCache(10*time.Millisecond, stop)

	deadline := time.Now().Add(time.Second)
	for resolver.cache.len() > 0 {
		if time.Now().After(deadline) {
			t.Fatal("expired entry was not swept")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestDNSResolver_ZeroTTLNotCached(t *testing.T) {
	resolver := NewDNSResolverWithDefaultOpts()
	defer resolver.Close()
	resolver.cacheResult("zero|A", &DNSCacheItem{}, 0)
	if n := resolver.cache.len(); n != 0 {
		t.Errorf("expected nothing cached, got %d entries", n)
	}
}

func TestSoaMinimum(t *testing.T) {
	if soa, ttl := soaMinimum([]*DNSRecord{{Type: DNSTypeNS}}); soa != nil || ttl != 0 {
		t.Errorf("expected no SOA, got %v %v", soa, ttl)
	}
	if _, ttl := soaMinimum([]*DNSRecord{soaRecord(300, 45)}); ttl != 45*time.Second {
		t.Errorf("expected 45s, got %v", ttl)
	}
	if _, ttl := soaMinimum([]*DNSRecord{{Type: DNSTypeSOA, TTL: 300, RData: []byte{1, 2}}}); ttl != 0 {
		t.Errorf("expected a short SOA to be ignored, got %v", ttl)
	}
}
