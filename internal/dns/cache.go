package dns

import (
	"container/heap"
	"encoding/binary"
	"sync"
	"time"
)

const (
	defaultCacheLimit  = 10000
	cacheSweepInterval = time.Minute
	maxNegativeTTL     = 60 * time.Second
)

// DNSCacheItem is one cached result: the records of an upstream answer, or
// an NXDOMAIN with its SOA.
type DNSCacheItem struct {
	answers     []*DNSRecord
	authorities []*DNSRecord
	nxDomain    bool
	expires     time.Time

	key   string
	index int // position in dnsCache.byExpiry
}

// dnsCache is a TTL cache with an entry limit. A heap ordered by expiry
// gives the entry to evict when it is full and the expired ones to sweep.
type dnsCache struct {
	lock     sync.RWMutex
	items    map[string]*DNSCacheItem
	byExpiry expiryHeap
	limit    int
}

func newDNSCache(limit int) *dnsCache {
	return &dnsCache{
		items: make(map[string]*DNSCacheItem),
		limit: limit,
	}
}

// get returns the item for key when it has not expired. The item must not
// be changed by the caller.
func (c *dnsCache) get(key string, now time.Time) (*DNSCacheItem, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	item, ok := c.items[key]
	if !ok || !item.expires.After(now) {
		return nil, false
	}
	return item, true
}

// set stores item under key. A new key in a full cache first evicts the
// entry nearest expiry.
func (c *dnsCache) set(key string, item *DNSCacheItem) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if old, ok := c.items[key]; ok {
		heap.Remove(&c.byExpiry, old.index)
		delete(c.items, key)
	} else if len(c.items) >= c.limit {
		evicted := heap.Pop(&c.byExpiry).(*DNSCacheItem)
		delete(c.items, evicted.key)
	}
	item.key = key
	c.items[key] = item
	heap.Push(&c.byExpiry, item)
}

// sweep deletes every entry that has expired by now.
func (c *dnsCache) sweep(now time.Time) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for len(c.byExpiry) > 0 && !c.byExpiry[0].expires.After(now) {
		expired := heap.Pop(&c.byExpiry).(*DNSCacheItem)
		delete(c.items, expired.key)
	}
}

func (c *dnsCache) len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.items)
}

type expiryHeap []*DNSCacheItem

func (h expiryHeap) Len() int           { return len(h) }
func (h expiryHeap) Less(i, j int) bool { return h[i].expires.Before(h[j].expires) }

func (h expiryHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *expiryHeap) Push(x any) {
	item := x.(*DNSCacheItem)
	item.index = len(*h)
	*h = append(*h, item)
}

func (h *expiryHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*h = old[:n-1]
	return item
}

// soaMinimum returns the first SOA record in records and the negative
// caching TTL it allows: min(SOA TTL, SOA minimum, maxNegativeTTL).
func soaMinimum(records []*DNSRecord) (*DNSRecord, time.Duration) {
	for _, record := range records {
		// The SOA RDATA ends in five 32-bit fields; MINIMUM is the last.
		if record.Type != DNSTypeSOA || len(record.RData) < 20 {
			continue
		}
		minimum := binary.BigEndian.Uint32(record.RData[len(record.RData)-4:])
		ttl := time.Duration(min(record.TTL, minimum)) * time.Second
		return record, min(ttl, maxNegativeTTL)
	}
	return nil, 0
}
