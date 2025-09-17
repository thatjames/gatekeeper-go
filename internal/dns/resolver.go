package dns

import (
	"errors"
	"net"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	ErrNxDomain = errors.New("domain unavailable/blocked")
)

type DNSResolver struct {
	cache     map[string]*DNSCacheItem
	upstream  []net.IP
	blacklist []string
}

type ResolverOpts struct {
	Upstreams []string
	Blacklist []string
	TTL       time.Duration
}

var defaultResolverOpts = ResolverOpts{
	Upstreams: []string{"1.1.1.1", "9.9.9.9"},
	Blacklist: []string{},
	TTL:       300,
}

type DNSCacheItem struct {
	record *DNSRecord
	ttl    time.Time
}

func NewDNSResolverWithDefaultOpts() *DNSResolver {
	return NewDNSResolverWithOpts(defaultResolverOpts)
}

func NewDNSResolverWithOpts(options ResolverOpts) *DNSResolver {
	sort.Strings(options.Blacklist)
	var upstreamAddrs []net.IP
	for _, upstream := range options.Upstreams {
		ip := net.ParseIP(upstream).To4()
		if ip == nil {
			log.Warnf("unable to parse upstream %s", upstream)
			continue
		}
		upstreamAddrs = append(upstreamAddrs, ip)
	}
	return &DNSResolver{
		cache:     make(map[string]*DNSCacheItem),
		upstream:  upstreamAddrs,
		blacklist: options.Blacklist,
	}
}

func (r *DNSResolver) Resolve(domain string) (*DNSRecord, error) {
	log.Debugf("resolving %s", domain)
	if index := sort.SearchStrings(r.blacklist, domain); index < len(r.blacklist) && r.blacklist[index] == domain {
		log.Debugf("found %s in blacklist", domain)
		return nil, ErrNxDomain
	}
	if cacheItem, ok := r.cache[domain]; ok {
		if cacheItem.ttl.After(time.Now()) {
			log.Debugf("found %v in cache", cacheItem)
			return cacheItem.record, nil
		} else {

		}
	}
	//TODO return NXDOMAIN
	return &DNSRecord{
		Name:  domain,
		Type:  DNSTypeA,
		Class: 1,
		TTL:   300,
		RData: net.ParseIP("10.0.0.1"),
	}, nil
}

func (r *DNSResolver) Lookup(domain string) (*DNSRecord, error) {
	return nil, nil
}
