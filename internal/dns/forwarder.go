package dns

import (
	"fmt"
	"net"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// ConditionalForwarder sends the queries for Zone, and for every name under
// it, to Server only.
type ConditionalForwarder struct {
	Zone   string
	Server string // IPv4, optionally with :port
}

type conditionalForwarder struct {
	zone string // lower-case, no trailing dot
	addr string // ip:port
}

// parseForwarders normalises the configured forwarders, longest zone first,
// and skips the ones that cannot be used.
func parseForwarders(forwarders []ConditionalForwarder) []conditionalForwarder {
	parsed := make([]conditionalForwarder, 0, len(forwarders))
	for _, forwarder := range forwarders {
		zone := normaliseName(forwarder.Zone)
		if zone == "" {
			log.Warnf("skipping conditional forwarder to %s: empty zone", forwarder.Server)
			continue
		}
		addr, err := forwarderAddr(forwarder.Server)
		if err != nil {
			log.Warnf("skipping conditional forwarder for %s: %v", zone, err)
			continue
		}
		parsed = append(parsed, conditionalForwarder{zone: zone, addr: addr})
	}
	sort.SliceStable(parsed, func(i, j int) bool {
		return len(parsed[i].zone) > len(parsed[j].zone)
	})
	return parsed
}

// forwarderAddr turns "ip" or "ip:port" into "ip:port", port 53 by default.
func forwarderAddr(server string) (string, error) {
	host, port := server, "53"
	if h, p, err := net.SplitHostPort(server); err == nil {
		host, port = h, p
	}
	ip := net.ParseIP(host).To4()
	if ip == nil {
		return "", fmt.Errorf("server %q is not an IPv4 address", server)
	}
	if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 {
		return "", fmt.Errorf("server %q has an invalid port", server)
	}
	return net.JoinHostPort(ip.String(), port), nil
}

func normaliseName(name string) string {
	return strings.ToLower(strings.TrimSuffix(name, "."))
}

// forwarderFor is the forwarder with the longest zone that domain is in, or nil.
func (r *DNSResolver) forwarderFor(domain string) *conditionalForwarder {
	name := normaliseName(domain)
	for i := range r.forwarders {
		zone := r.forwarders[i].zone
		if name == zone || strings.HasSuffix(name, "."+zone) {
			return &r.forwarders[i]
		}
	}
	return nil
}

// forward resolves domain at its zone's forwarder and never at the upstreams.
func (r *DNSResolver) forward(domain string, dnsType DNSType, forwarder *conditionalForwarder, cacheKey string) (answers, authorities []*DNSRecord, err error) {
	if dnsType == DNSTypePTR {
		if answers := r.localPTR(domain); answers != nil {
			queryCounter.With(prometheus.Labels{"domain": domain, "upstream": "local-domain", "result": "success"}).Inc()
			return answers, nil, nil
		}
	}

	log.Debugf("forwarding %s to %s", domain, forwarder.addr)
	answers, authorities, err = r.exchange(domain, dnsType, forwarder.addr)
	if err == ErrNxDomain {
		nxSOA, nxTTL := soaMinimum(authorities)
		if nxSOA == nil {
			return nil, nil, ErrNxDomain
		}
		authorities = []*DNSRecord{nxSOA}
		r.cacheResult(cacheKey, &DNSCacheItem{authorities: authorities, nxDomain: true}, nxTTL)
		return nil, authorities, ErrNxDomain
	}
	if err != nil {
		log.Errorf("conditional forwarder %s for %s failed: %v", forwarder.addr, forwarder.zone, err)
		queryCounter.With(prometheus.Labels{"domain": domain, "upstream": forwarder.addr, "result": "failed"}).Inc()
		return nil, nil, ErrDNSServerFailure
	}
	queryCounter.With(prometheus.Labels{"domain": domain, "upstream": forwarder.addr, "result": "success"}).Inc()
	r.cacheAnswer(cacheKey, answers, authorities)
	return answers, authorities, nil
}

// reverseIPv4 is the IPv4 address of a four-label reverse name
// (d.c.b.a.in-addr.arpa, with or without the trailing dot, any case), or nil.
func reverseIPv4(domain string) net.IP {
	reverse, ok := strings.CutSuffix(normaliseName(domain), ".in-addr.arpa")
	if !ok {
		return nil
	}
	octets := strings.Split(reverse, ".")
	if len(octets) != 4 {
		return nil
	}
	return net.ParseIP(octets[3] + "." + octets[2] + "." + octets[1] + "." + octets[0]).To4()
}

// localPTR answers a reverse lookup from the local domains when one of them
// has the IPv4 of the reverse name, or returns nil. When several share the
// address, the alphabetically first name answers, so the answer is stable.
func (r *DNSResolver) localPTR(domain string) []*DNSRecord {
	ip := reverseIPv4(domain)
	if ip == nil {
		return nil
	}
	hosts := make([]string, 0)
	r.domainLock.RLock()
	for host, localIP := range r.localDomains {
		if localIP.Equal(ip) {
			hosts = append(hosts, host)
		}
	}
	r.domainLock.RUnlock()
	if len(hosts) == 0 {
		return nil
	}
	host := slices.Min(hosts)
	return []*DNSRecord{{
		Name:       compressedDomainVal,
		Type:       DNSTypePTR,
		Class:      DNSClassIN,
		TTL:        uint32(r.cacheTTL.Seconds()),
		ParsedName: host,
		RData:      stringToDNSWireFormat(host),
	}}
}
