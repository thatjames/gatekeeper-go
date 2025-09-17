package dns

import (
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

type DNSServerOpts struct {
	Options   *DNSServerOpts
	Interface string
	Port      int
	Upstream  []string
}

var defaultDNSServerOpts = DNSServerOpts{
	Interface: "eth0",
	Port:      53,
}

type DNSServer struct {
	opts         *DNSServerOpts
	resolver     *DNSResolver
	packetConn   net.PacketConn
	receiverChan chan *DNSPacket
}

func NewDNSServer() *DNSServer {
	return &DNSServer{
		resolver: NewDNSResolver(),
		opts:     &defaultDNSServerOpts,
	}
}

func NewDNSServerWithOpts(opts DNSServerOpts) *DNSServer {
	return &DNSServer{
		resolver:   NewDNSResolver(),
		opts:       &opts,
		packetConn: nil,
	}
}

func (d *DNSServer) Start() error {
	log.Info("starting DNS server")
	var err error
	if d.packetConn, err = net.ListenPacket("udp4", fmt.Sprintf(":%d", d.opts.Port)); err != nil {
		return err
	}
	return nil
}

func (d *DNSServer) Stop() error {
	log.Info("stopping DNS server")
	return nil
}

func (d *DNSServer) run() {

}
