package dns

import (
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type DNSServerOpts struct {
	Options      *DNSServerOpts
	Interface    string
	Port         int
	Upstream     []string
	ResolverOpts *ResolverOpts
}

var defaultDNSServerOpts = DNSServerOpts{
	Interface: "eth0",
	Port:      53,
}

type DNSServer struct {
	opts         *DNSServerOpts
	resolver     *DNSResolver
	packetConn   net.PacketConn
	receiverChan chan *dnsWorkItem
	responseChan chan *dnsWorkItem
	exitChan     chan struct{}
	lock         sync.Mutex
}

type dnsWorkItem struct {
	*DNSPacket
	err       error
	startTime time.Time
}

func NewDNSServer() *DNSServer {
	return NewDNSServerWithOpts(defaultDNSServerOpts)
}

func NewDNSServerWithOpts(opts DNSServerOpts) *DNSServer {
	return &DNSServer{
		resolver:     NewDNSResolverWithOpts(*opts.ResolverOpts),
		opts:         &opts,
		packetConn:   nil,
		receiverChan: make(chan *dnsWorkItem, 100),
		responseChan: make(chan *dnsWorkItem, 100),
		lock:         sync.Mutex{},
	}
}

func (d *DNSServer) Start() error {
	log.Info("starting DNS server")
	log.Tracef("starting DNS server on port %d", d.opts.Port)
	var err error
	if d.packetConn, err = net.ListenPacket("udp4", fmt.Sprintf(":%d", d.opts.Port)); err != nil {
		log.Error("unable to start DNS server: ", err.Error())
		return err
	}
	d.exitChan = make(chan struct{})
	go d.listen()
	go d.receiverWorker()
	go d.responseWorker()
	log.Info("DNS server started")
	return nil
}

func (d *DNSServer) Stop() error {
	close(d.exitChan)
	log.Info("stopping DNS server")
	d.lock.Lock() //wait for the goroutines to finish
	defer d.lock.Unlock()
	return nil
}

func (d *DNSServer) Options() *DNSServerOpts {
	return d.opts
}

func (d *DNSServer) UpdateOptions(opts *DNSServerOpts) {
	d.opts = opts
}

func (d *DNSServer) AddLocalDomain(domain string, ip string) error {
	return d.resolver.AddLocalDomain(domain, net.ParseIP(ip).To4())
}

func (d *DNSServer) DeleteLocalDomain(domain string) {
	d.resolver.DeleteLocalDomain(domain)
}

func (d *DNSServer) listen() {
	d.lock.Lock()
	defer func() {
		log.Tracef("closing DNS server")
		d.packetConn.Close()
		d.lock.Unlock()
	}()
	for {
		select {
		case <-d.exitChan:
			return
		default:
			buff := make([]byte, 1500)
			d.packetConn.SetReadDeadline(time.Now().Add(time.Second * 2))
			n, addr, err := d.packetConn.ReadFrom(buff)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// This is a timeout error, don't log it
					continue
				}
				log.Error("unable to read datastream: ", err.Error())
				continue
			}
			workItem := &dnsWorkItem{
				startTime: time.Now(),
			}
			buff = buff[:n]
			log.Tracef("received %d bytes from %s", n, addr.String())
			msg, err := ParseDNSMessage(buff)
			if err != nil {
				log.Error("unable to parse DNS message: ", err.Error())
				workItem.err = err
			} else if msg.Header == nil {
				log.Error("unable to parse DNS message: no header")
			} else {
				workItem.DNSPacket = &DNSPacket{
					DNSMessage:   msg,
					ResponseAddr: addr,
				}
			}
			d.receiverChan <- workItem
		}
	}
}

func (d *DNSServer) receiverWorker() {
	for packet := range d.receiverChan {
		if packet.err != nil {
			log.Errorf("skipping malformed packet: %v", packet.err)
			continue
		}

		if packet.DNSMessage == nil || packet.DNSMessage.Header == nil {
			log.Error("received packet with nil DNSMessage or Header")
			continue
		}

		log.Tracef("received DNS packet from %s", packet.ResponseAddr.String())
		response, authority, err := d.resolver.Resolve(packet.DNSMessage.Questions[0].ParsedName, packet.DNSMessage.Questions[0].Type)

		packet.DNSMessage.Header.SetQR(true)

		if err != nil {
			if err == ErrNxDomain {
				packet.DNSMessage.Header.SetRCODE(RCODENameFailure)
			} else {
				packet.DNSMessage.Header.SetRCODE(RCODEServerFailure)
			}
		} else {
			packet.DNSMessage.Header.SetRCODE(RCODESuccess)
			if response != nil {
				packet.DNSMessage.Answers = append(packet.DNSMessage.Answers, response)
				log.Trace("adding answer %s", response)
			}
			if authority != nil {
				packet.DNSMessage.Authorities = append(packet.DNSMessage.Authorities, authority)
				log.Tracef("adding authority %s", authority)
			}
		}

		d.responseChan <- packet
	}
}

func (d *DNSServer) responseWorker() {
	for packet := range d.responseChan {
		log.Tracef("sending DNS response packet to %s", packet.ResponseAddr.String())
		packet.Header.SetQR(true)
		data, err := MarshalDNSMessage(packet.DNSMessage)
		if err != nil {
			log.Error("unable to marshal DNS packet: ", err.Error())
			continue
		}
		_, err = d.packetConn.WriteTo(data, packet.ResponseAddr)
		timeElapsed := time.Since(packet.startTime).Round(time.Millisecond)
		reqDuration.Observe(float64(timeElapsed.Milliseconds()))
		if err != nil {
			log.Error("unable to send DNS packet: ", err.Error())
		}
	}
}
