package dns

type DNSServer struct {
	resolver *DNSResolver
}

func NewDNSServer(resolver *DNSResolver) *DNSServer {
	return &DNSServer{
		resolver: resolver,
	}
}

func (d *DNSServer) Start() error {
	panic("not implemented") // TODO: Implement
}

func (d *DNSServer) Stop() error {
	panic("not implemented") // TODO: Implement
}
