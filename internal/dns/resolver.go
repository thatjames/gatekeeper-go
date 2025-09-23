package dns

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	ErrNxDomain = errors.New("domain unavailable/blocked")
)

var (
	compressedDomainVal = []byte{0xc0, 0x0c}
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

func (r *DNSResolver) Resolve(domain string, class DNSType) (*DNSRecord, error) {
	log.Debugf("resolving %s", domain)
	if index := sort.SearchStrings(r.blacklist, domain); index < len(r.blacklist) && r.blacklist[index] == domain {
		log.Debugf("found %s in blacklist", domain)
		return nil, ErrNxDomain
	}
	if cacheItem, ok := r.cache[fmt.Sprintf("%s%s", domain, class)]; ok {
		if cacheItem.ttl.After(time.Now()) {
			log.Debugf("found %s - %s in cache", cacheItem.record.ParsedName, cacheItem.record.Type)
			return cacheItem.record, nil
		} else {
			log.Debugf("removing expired cache item for %s", domain)
			delete(r.cache, domain)
		}
	}
	for _, upstream := range r.upstream {
		record, err := r.lookup(domain, class, upstream)
		if err != nil {
			log.Error("unable to lookup: ", err.Error())
			continue
		}
		log.Debugf("cache %s%s with value %s", domain, class, record.ParsedName)
		r.cache[fmt.Sprintf("%s%s", domain, class)] = &DNSCacheItem{
			record: record,
			ttl:    time.Now().Add(time.Duration(record.TTL) * time.Second),
		}
		return record, nil
	}
	return nil, ErrNxDomain
}

func (r *DNSResolver) lookup(domain string, dnsType DNSType, upstream net.IP) (*DNSRecord, error) {
	log.Debugf("looking up %s in %s", domain, upstream.String())
	message := NewDnsMessage()
	message.Header.ID = uint16(rand.Intn(65535))
	message.Header.SetRD(true)

	question := new(DNSQuestion)
	question.Name = stringToDNSWireFormat(domain)
	question.Type = dnsType
	question.Class = DNSClassIN //probably going to regret hardcoding this one day
	message.Questions = append(message.Questions, question)

	dat, err := MarshalDNSMessage(message)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: upstream,
		Port: 53,
	})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_, err = conn.Write(dat)
	if err != nil {
		return nil, err
	}

	buff := make([]byte, 1500)
	n, err := conn.Read(buff)
	if err != nil {
		return nil, err
	}

	msg, err := ParseDNSMessage(buff[:n])
	if err != nil {
		return nil, err
	}

	if len(msg.Answers) == 0 {
		return nil, ErrNxDomain
	}

	return msg.Answers[0], nil
}
