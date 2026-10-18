package dns

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	ErrNxDomain               = errors.New("domain unavailable/blocked")
	ErrDNSFormatError         = errors.New("DNS packet format error")
	ErrDNSServerFailure       = errors.New("DNS packet server failure")
	ErrInvalidBlocklistFormat = errors.New("invalid blocklist format")
)

var (
	compressedDomainVal = []byte{0xc0, 0x0c}
)

type Resolver interface {
	Resolve(domain string, dnsType DNSType) (answers, authorities []*DNSRecord, err error)
	AddLocalDomain(domain string, ip net.IP) error
	DeleteLocalDomain(domain string)
	AddBlocklistEntries(entries []string)
	DeleteBlocklistEntry(domain string)
	FlushBlocklist()
}

type DNSResolver struct {
	cache           *dnsCache
	upstream        []net.IP
	blacklist       map[string]struct{}
	localDomains    map[string]net.IP
	forwarders      []conditionalForwarder
	domainLock      *sync.RWMutex
	cacheTTL        time.Duration
	upstreamTimeout time.Duration
	dialTimeout     time.Duration
	readTimeout     time.Duration
	upstreamPort    int
	stopSweep       chan struct{}
	closeOnce       sync.Once
}

type ResolverOpts struct {
	Upstreams             []string
	LocalDomains          map[string]net.IP
	ConditionalForwarders []ConditionalForwarder
	CacheTTL              time.Duration
	UpstreamTimeout       time.Duration
	DialTimeout           time.Duration
	ReadTimeout           time.Duration
}

var defaultResolverOpts = ResolverOpts{
	Upstreams:       []string{"1.1.1.1", "9.9.9.9"},
	LocalDomains:    make(map[string]net.IP),
	CacheTTL:        time.Minute * 5,
	UpstreamTimeout: time.Second * 5,
	DialTimeout:     time.Second * 2,
	ReadTimeout:     time.Second * 2,
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

	cacheTTL := options.CacheTTL
	if cacheTTL == 0 {
		cacheTTL = time.Minute * 5
	}
	upstreamTimeout := options.UpstreamTimeout
	if upstreamTimeout == 0 {
		upstreamTimeout = time.Second * 5
	}
	dialTimeout := options.DialTimeout
	if dialTimeout == 0 {
		dialTimeout = time.Second * 2
	}
	readTimeout := options.ReadTimeout
	if readTimeout == 0 {
		readTimeout = time.Second * 2
	}

	r := &DNSResolver{
		cache:           newDNSCache(defaultCacheLimit),
		upstream:        upstreamAddrs,
		localDomains:    options.LocalDomains,
		forwarders:      parseForwarders(options.ConditionalForwarders),
		domainLock:      new(sync.RWMutex),
		blacklist:       make(map[string]struct{}),
		cacheTTL:        cacheTTL,
		upstreamTimeout: upstreamTimeout,
		dialTimeout:     dialTimeout,
		readTimeout:     readTimeout,
		upstreamPort:    53,
		stopSweep:       make(chan struct{}),
	}
	go r.sweepCache(cacheSweepInterval, r.stopSweep)
	return r
}

// Close stops the cache sweep.
func (r *DNSResolver) Close() {
	r.closeOnce.Do(func() { close(r.stopSweep) })
}

func (r *DNSResolver) sweepCache(interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case now := <-ticker.C:
			r.cache.sweep(now)
		}
	}
}

// cacheResult caches item for ttl from now; a zero TTL is not cached.
func (r *DNSResolver) cacheResult(key string, item *DNSCacheItem, ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	item.expires = time.Now().Add(ttl)
	r.cache.set(key, item)
}

func (r *DNSResolver) Resolve(domain string, dnsType DNSType) (answers, authorities []*DNSRecord, err error) {
	answers, authorities = make([]*DNSRecord, 0), make([]*DNSRecord, 0)
	log.Debugf("resolving %s", domain)

	r.domainLock.RLock()
	_, blocked := r.blacklist[domain]
	responseIP, isLocal := r.localDomains[domain]
	r.domainLock.RUnlock()

	if blocked {
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
			TTL:        uint32(r.cacheTTL.Seconds()),
			ParsedName: domain,
			RData:      result,
		})
		return answers, nil, nil
	}

	// Generate cache key
	cacheKey := domain + "|" + dnsType.String()

	// Check cache
	if cacheItem, ok := r.cache.get(cacheKey, time.Now()); ok {
		if cacheItem.nxDomain {
			return nil, cacheItem.authorities, ErrNxDomain
		}
		queryCounter.With(prometheus.Labels{"domain": domain, "upstream": "cache", "result": "success"}).Inc()
		return cacheItem.answers, cacheItem.authorities, nil
	}

	if isLocal {
		log.Debugf("found %s in local domains", domain)
		if dnsType != DNSTypeA {
			return nil, nil, nil
		}
		queryCounter.With(prometheus.Labels{"domain": domain, "upstream": "local-domain", "result": "success"}).Inc()
		answers = append(answers, &DNSRecord{
			Name:       compressedDomainVal,
			Type:       dnsType,
			Class:      DNSClassIN,
			TTL:        uint32(r.cacheTTL.Seconds()),
			ParsedName: domain,
			RData:      responseIP.To4(),
		})
		return answers, nil, nil
	}

	if forwarder := r.forwarderFor(domain); forwarder != nil {
		return r.forward(domain, dnsType, forwarder, cacheKey)
	}

	type result struct {
		answers     []*DNSRecord
		authorities []*DNSRecord
		upstream    net.IP
		err         error
	}

	// One deadline for the whole fan-out, not a fresh one per reply.
	ctx, cancel := context.WithTimeout(context.Background(), r.upstreamTimeout)
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
	var nxSOA *DNSRecord
	var nxTTL time.Duration

wait:
	for i := 0; i < len(r.upstream); i++ {
		var res result
		select {
		case res = <-results:
		case <-ctx.Done():
			log.Warn("timeout waiting for an upstream response")
			break wait
		}

		if res.err != nil {
			log.Error("unable to lookup: ", res.err.Error())
			lastErr = res.err

			if res.err == ErrNxDomain && nxSOA == nil {
				nxSOA, nxTTL = soaMinimum(res.authorities)
			}
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

			r.cacheAnswer(cacheKey, res.answers, res.authorities)
			log.Debugf("Processing response from upstream %v", res.upstream)

			return res.answers, res.authorities, nil
		}
	}

	// All upstreams failed
	if lastErr != nil && lastErr != ErrNxDomain {
		return nil, nil, lastErr
	}
	if nxSOA == nil {
		return nil, nil, ErrNxDomain
	}
	authorities = []*DNSRecord{nxSOA}
	r.cacheResult(cacheKey, &DNSCacheItem{
		authorities: authorities,
		nxDomain:    true,
	}, nxTTL)
	return nil, authorities, ErrNxDomain
}

// cacheAnswer caches answers and authorities for the TTL of the first
// answer, or of the first authority when there are no answers.
func (r *DNSResolver) cacheAnswer(cacheKey string, answers, authorities []*DNSRecord) {
	var ttl uint32
	if answers != nil {
		ttl = answers[0].TTL
	} else if authorities != nil {
		ttl = authorities[0].TTL
		log.Debugf("TTL: %v", ttl)
	} else {
		return
	}
	r.cacheResult(cacheKey, &DNSCacheItem{
		answers:     answers,
		authorities: authorities,
	}, time.Duration(ttl)*time.Second)
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

func (r *DNSResolver) AddBlocklistEntries(entries []string) {
	defer r.domainLock.Unlock()
	r.domainLock.Lock()
	for _, entry := range entries {
		r.blacklist[entry] = struct{}{}
	}
}

func (r *DNSResolver) DeleteBlocklistEntry(domain string) {
	defer r.domainLock.Unlock()
	r.domainLock.Lock()
	delete(r.blacklist, domain)
}

func (r *DNSResolver) FlushBlocklist() {
	defer r.domainLock.Unlock()
	r.domainLock.Lock()
	r.blacklist = make(map[string]struct{})
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
			r.domainLock.RLock()
			for host, localIp := range r.localDomains {
				log.Tracef("checking %s against %v", host, localIp)
				if localIp.String() == ip {
					log.Tracef("found %s in local domains", host)
					answers = append(answers, &DNSRecord{
						Name:       compressedDomainVal,
						Type:       dnsType,
						Class:      DNSClassIN,
						TTL:        uint32(r.cacheTTL.Seconds()),
						ParsedName: host,
						RData:      stringToDNSWireFormat(host),
					})
					r.domainLock.RUnlock()
					return answers, nil, nil
				}
				r.domainLock.RUnlock()
				return nil, nil, ErrNxDomain // we don't have this local domain defined, so we treat it as a bad name
			}
			r.domainLock.RUnlock()
		}
	}
	return r.exchange(domain, dnsType, fmt.Sprintf("%s:%d", upstream, r.upstreamPort))
}

// exchange asks the DNS server at addr (ip:port) for domain and dnsType.
func (r *DNSResolver) exchange(domain string, dnsType DNSType, addr string) (answers, authorities []*DNSRecord, err error) {
	message := NewDnsMessage()
	message.Header.ID = uint16(rand.Intn(65535))
	message.Header.SetRD(true)

	question := new(DNSQuestion)
	question.Name = stringToDNSWireFormat(domain)
	question.Type = dnsType
	question.Class = DNSClassIN //probably going to regret hardcoding this one day
	message.Questions = append(message.Questions, question)
	message.Additionals = append(message.Additionals, NewOPTRecord(maxUDPSize))

	dat, err := MarshalDNSMessage(message)
	if err != nil {
		return nil, nil, err
	}

	dialer := net.Dialer{
		Timeout: r.dialTimeout,
	}

	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, nil, err
	}

	conn, err := dialer.Dial("udp", raddr.String())
	if err != nil {
		return nil, nil, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(r.readTimeout))

	_, err = conn.Write(dat)
	if err != nil {
		return nil, nil, err
	}

	buff := make([]byte, 65535)
	n, err := conn.Read(buff)
	if err != nil {
		return nil, nil, err
	}

	msg, err := ParseDNSMessage(buff[:n])
	if err != nil {
		return nil, nil, err
	}
	if msg.Header.TC() {
		log.Debugf("truncated reply from %s, asking again over TCP", addr)
		if msg, err = r.exchangeTCP(dat, addr); err != nil {
			return nil, nil, err
		}
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
		return nil, msg.Authorities, ErrNxDomain
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

// exchangeTCP sends the marshaled query to addr over TCP and returns the reply.
func (r *DNSResolver) exchangeTCP(query []byte, addr string) (*DNSMessage, error) {
	dialer := net.Dialer{
		Timeout: r.dialTimeout,
	}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(r.readTimeout))

	if _, err := conn.Write(append(binary.BigEndian.AppendUint16(nil, uint16(len(query))), query...)); err != nil {
		return nil, err
	}
	lengthPrefix := make([]byte, 2)
	if _, err := io.ReadFull(conn, lengthPrefix); err != nil {
		return nil, err
	}
	reply := make([]byte, binary.BigEndian.Uint16(lengthPrefix))
	if _, err := io.ReadFull(conn, reply); err != nil {
		return nil, err
	}
	return ParseDNSMessage(reply)
}
