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
	exitChan     chan struct{}
}

func NewDNSServer() *DNSServer {
	return NewDNSServerWithOpts(defaultDNSServerOpts)
}

func NewDNSServerWithOpts(opts DNSServerOpts) *DNSServer {
	return &DNSServer{
		resolver:     NewDNSResolver(),
		opts:         &opts,
		packetConn:   nil,
		receiverChan: make(chan *DNSPacket, 10),
		exitChan:     make(chan struct{}),
	}
}

func (d *DNSServer) Start() error {
	log.Info("starting DNS server")
	var err error
	if d.packetConn, err = net.ListenPacket("udp4", fmt.Sprintf(":%d", d.opts.Port)); err != nil {
		return err
	}
	go d.listen()
	go d.receiverWorker()
	return nil
}

func (d *DNSServer) Stop() error {
	log.Info("stopping DNS server")
	return nil
}

func (d *DNSServer) listen() {
	buff := make([]byte, 1500)
	for {
		select {
		case <-d.exitChan:
			return
		default:
			n, addr, err := d.packetConn.ReadFrom(buff)
			if err != nil {
				log.Error("unable to read datastream: ", err.Error())
				continue
			}
			buff = buff[:n]
			log.Debugf("received %d bytes from %s", n, d.packetConn.LocalAddr().String())
			msg, err := ParseDNSMessage(buff)
			if err != nil {
				log.Error("unable to parse DNS message: ", err.Error())
			} else {
				d.receiverChan <- &DNSPacket{
					DNSMessage:   msg,
					ResponseAddr: addr,
				}
			}
		}
	}
}

func (d *DNSServer) receiverWorker() {
	for packet := range d.receiverChan {
		log.Debugf("received DNS packet %s from %s", packet, packet.ResponseAddr.String())
	}
}
