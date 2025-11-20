package dns

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	ErrNxDomain         = errors.New("domain unavailable/blocked")
	ErrDNSFormatError   = errors.New("DNS packet format error")
	ErrDNSServerFailure = errors.New("DNS packet server failure")
)

var (
	compressedDomainVal = []byte{0xc0, 0x0c}
)

type DNSResolver struct {
	cache        map[string]*DNSCacheItem
	upstream     []net.IP
	blacklist    []string
	localDomains map[string]net.IP
	domainLock   *sync.RWMutex
}

type ResolverOpts struct {
	Upstreams    []string
	LocalDomains map[string]net.IP
}

var defaultResolverOpts = ResolverOpts{
	Upstreams:    []string{"1.1.1.1", "9.9.9.9"},
	LocalDomains: make(map[string]net.IP),
}

type DNSCacheItem struct {
	records []*DNSRecord //Handle multiple answers, e.g. CNAME
	ttl     time.Time
}

func NewDNSResolverWithDefaultOpts() *DNSResolver {
	return NewDNSResolverWithOpts(defaultResolverOpts)
}

func NewDNSResolverWithOpts(options ResolverOpts) *DNSResolver {
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
		cache:        make(map[string]*DNSCacheItem),
		upstream:     upstreamAddrs,
		localDomains: options.LocalDomains,
		domainLock:   new(sync.RWMutex),
		blacklist:    make([]string, 0),
	}
}

func (r *DNSResolver) Resolve(domain string, dnsType DNSType) (answers, authorities []*DNSRecord, err error) {
	r.domainLock.Lock()
	defer r.domainLock.Unlock()

	answers, authorities = make([]*DNSRecord, 0), make([]*DNSRecord, 0)
	log.Debugf("resolving %s", domain)

	if r.blacklist != nil && len(r.blacklist) > 0 {
		if index := sort.SearchStrings(r.blacklist, domain); index < len(r.blacklist) && r.blacklist[index] == domain {
			log.Infof("rejected %s in blacklist", domain)
			var result []byte
			if dnsType == DNSTypeA {
				result = make([]byte, 4)
			} else if dnsType == DNSTypeAAAA {
				result = net.IPv6zero
			}
			blockedDomainCounter.With(prometheus.Labels{"domain": domain}).Inc()
			answers = append(answers, &DNSRecord{
				Name:       compressedDomainVal,
				Type:       dnsType,
				Class:      DNSClassIN,
				TTL:        uint32((time.Second * 300).Seconds()),
				ParsedName: domain,
				RData:      result,
			})
			return answers, nil, nil
		}
	}

	// Generate cache key
	keyBuff := bytes.NewBufferString(domain)
	binary.Write(keyBuff, binary.BigEndian, dnsType)
	cacheKey := fmt.Sprintf("%x", keyBuff.Bytes())

	// Check cache
	if cacheItem, ok := r.cache[cacheKey]; ok {
		if cacheItem.ttl.After(time.Now()) {
			queryCounter.With(prometheus.Labels{"domain": domain, "upstream": "cache", "result": "success"}).Inc()
			answers = cacheItem.records
			return answers, nil, nil
		} else {
			log.Debugf("removing expired cache item for %s", domain)
			delete(r.cache, cacheKey)
		}
	}

	if responseIP, ok := r.localDomains[domain]; ok {
		log.Debugf("found %s in local domains", domain)
		if dnsType != DNSTypeA {
			return nil, nil, nil
		}
		queryCounter.With(prometheus.Labels{"domain": domain, "upstream": "local-domain", "result": "success"}).Inc()
		answers = append(answers, &DNSRecord{
			Name:       compressedDomainVal,
			Type:       dnsType,
			Class:      DNSClassIN,
			TTL:        uint32((time.Second * 300).Seconds()),
			ParsedName: domain,
			RData:      responseIP.To4(),
		})
		return answers, nil, nil
	}

	type result struct {
		answers     []*DNSRecord
		authorities []*DNSRecord
		upstream    net.IP
		err         error
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results := make(chan result, len(r.upstream))

	for _, upstream := range r.upstream {
		go func(up net.IP) {
			ans, auth, err := r.lookup(domain, dnsType, up)

			select {
			case results <- result{ans, auth, up, err}:
			case <-ctx.Done():
			}
		}(upstream)
	}

	var lastErr error

	for i := 0; i < len(r.upstream); i++ {
		var res result
		select {
		case res = <-results:
		case <-time.After(time.Second * 5):
			log.Warn("timeout waiting for an upstream response")
			continue
		}

		if res.err != nil {
			log.Error("unable to lookup: ", res.err.Error())
			lastErr = res.err

			if res.err != ErrNxDomain {
				queryCounter.With(prometheus.Labels{
					"domain":   domain,
					"upstream": res.upstream.String(),
					"result":   "failed",
				}).Inc()
			}
			continue
		}

		if res.answers != nil || res.authorities != nil {
			cancel()

			queryCounter.With(prometheus.Labels{
				"domain":   domain,
				"upstream": res.upstream.String(),
				"result":   "success",
			}).Inc()

			var ttl time.Time
			if res.answers != nil {
				ttl = time.Now().Add(time.Duration(res.answers[0].TTL) * time.Second)
			} else {
				ttl = time.Now().Add(time.Duration(res.authorities[0].TTL) * time.Second)
				log.Debugf("TTL: %v", ttl)
			}
			r.cache[cacheKey] = &DNSCacheItem{
				records: res.answers,
				ttl:     ttl,
			}
			log.Debugf("Processing response from upstream %v", res.upstream)

			return res.answers, res.authorities, nil
		}
	}

	// All upstreams failed
	if lastErr != nil {
		return nil, nil, lastErr
	}
	return nil, nil, ErrNxDomain
}

func (r *DNSResolver) AddLocalDomain(domain string, ip net.IP) error {
	defer r.domainLock.Unlock()
	r.domainLock.Lock()
	if localAddr := net.ParseIP(ip.String()).To4(); localAddr == nil || localAddr.Equal(net.IPv4zero) {
		return errors.New("invalid IP address")
	} else {
		r.localDomains[domain] = localAddr
	}
	return nil
}

func (r *DNSResolver) DeleteLocalDomain(domain string) {
	defer r.domainLock.Unlock()
	r.domainLock.Lock()
	delete(r.localDomains, domain)
}

func (r *DNSResolver) AddBlocklistEntries(blacklist []string) {
	defer r.domainLock.Unlock()
	r.domainLock.Lock()
	r.blacklist = append(r.blacklist, blacklist...)
	slices.Sort(r.blacklist)
}

func (r *DNSResolver) DeleteBlocklistEntry(domain string) {
	defer r.domainLock.Unlock()
	r.domainLock.Lock()
	if index := sort.SearchStrings(r.blacklist, domain); index < len(r.blacklist) && r.blacklist[index] == domain {
		if index == len(r.blacklist)-1 {
			r.blacklist = r.blacklist[:index]
		} else {
			r.blacklist = append(r.blacklist[:index], r.blacklist[index+1:]...)
		}
	}
}

func (r *DNSResolver) FlushBlocklist() {
	defer r.domainLock.Unlock()
	r.domainLock.Lock()
	r.blacklist = make([]string, 0)
}

func (r *DNSResolver) lookup(domain string, dnsType DNSType, upstream net.IP) (answers, authorities []*DNSRecord, err error) {
	log.Debugf("looking up %s in %s", domain, upstream.String())
	if dnsType == DNSTypePTR {
		reverseDNS := strings.TrimSuffix(domain, ".in-addr.arpa.")
		reverseDNS = strings.TrimSuffix(domain, ".in-addr.arpa")
		octets := strings.Split(reverseDNS, ".")
		if len(octets) != 4 {
			return nil, nil, errors.New("invalid reverse DNS")
		}
		ip := octets[3] + "." + octets[2] + "." + octets[1] + "." + octets[0]
		log.Debugf("reverse lookup %s", ip)
		if net.ParseIP(ip).IsPrivate() {
			for host, localIp := range r.localDomains {
				log.Tracef("checking %s against %v", host, localIp)
				if localIp.String() == ip {
					log.Tracef("found %s in local domains", host)
					answers = append(answers, &DNSRecord{
						Name:       compressedDomainVal,
						Type:       dnsType,
						Class:      DNSClassIN,
						TTL:        uint32((time.Second * 300).Seconds()),
						ParsedName: host,
						RData:      stringToDNSWireFormat(host),
					})
					return answers, nil, nil
				}
				return nil, nil, ErrNxDomain // we don't have this local domain defined, so we treat it as a bad name
			}
		}
	}
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
		return nil, nil, err
	}

	dialer := net.Dialer{
		Timeout: time.Second * 2,
	}

	raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", upstream, 53))
	if err != nil {
		return nil, nil, err
	}

	conn, err := dialer.Dial("udp", raddr.String())
	if err != nil {
		return nil, nil, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(time.Second * 2))

	_, err = conn.Write(dat)
	if err != nil {
		return nil, nil, err
	}

	buff := make([]byte, 1500)
	n, err := conn.Read(buff)
	if err != nil {
		return nil, nil, err
	}

	msg, err := ParseDNSMessage(buff[:n])
	if err != nil {
		return nil, nil, err
	}

	log.Tracef("received DNS packet from %s", conn.RemoteAddr().String())
	switch msg.Header.RCODE() {
	case RCODESuccess:
		log.Tracef("DNS packet from %s successful", conn.RemoteAddr().String())
	case RCODEFormatError:
		log.Errorf("DNS packet from %s format error", conn.RemoteAddr().String())
		return nil, nil, ErrDNSFormatError
	case RCODENameFailure:
		log.Errorf("DNS packet from %s name failure", conn.RemoteAddr().String())
		return nil, nil, ErrNxDomain
	case RCODEServerFailure:
		log.Errorf("DNS packet from %s server failure", conn.RemoteAddr().String())
		return nil, nil, ErrDNSServerFailure
	}

	if msg.Answers != nil && len(msg.Answers) > 0 {
		answers = msg.Answers
	}

	if msg.Authorities != nil && len(msg.Authorities) > 0 {
		authorities = msg.Authorities
	}

	return
}
