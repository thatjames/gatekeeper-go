package dns

import (
	"net"

	log "github.com/sirupsen/logrus"
)

type DNSResolver struct {
	cache map[string]net.IP
}

func NewDNSResolver() *DNSResolver {
	return &DNSResolver{
		cache: make(map[string]net.IP),
	}
}

func (r *DNSResolver) Resolve(domain string) (*DNSPacket, error) {
	log.Debugf("resolving %s", domain)
	if ip, ok := r.cache[domain]; ok {
		log.Debugf("found %s in cache", ip.String())
		return &DNSPacket{
			Name:  domain,
			Type:  DNSTypeA,
			Class: 1,
			TTL:   300,
			RData: ip,
		}, nil
	}
	return nil, nil
}
