package dns

import (
	"fmt"
	"net"
	"time"

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
	receiverChan chan *dnsWorkItem
	responseChan chan *dnsWorkItem
	exitChan     chan struct{}
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
		resolver:     NewDNSResolverWithDefaultOpts(),
		opts:         &opts,
		packetConn:   nil,
		receiverChan: make(chan *dnsWorkItem, 100),
		responseChan: make(chan *dnsWorkItem, 100),
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
	go d.responseWorker()
	return nil
}

func (d *DNSServer) Stop() error {
	log.Info("stopping DNS server")
	close(d.exitChan)
	return nil
}

func (d *DNSServer) listen() {
	buff := make([]byte, 1500)
	defer d.packetConn.Close()
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
			workItem := &dnsWorkItem{
				startTime: time.Now(),
			}
			buff = buff[:n]
			log.Tracef("received %d bytes from %s", n, d.packetConn.LocalAddr().String())
			msg, err := ParseDNSMessage(buff)
			if err != nil {
				log.Error("unable to parse DNS message: ", err.Error())
				workItem.err = err
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
		if packet.err != nil { //almost always because of a malformed packet
			packet.DNSMessage.Header.SetRCODE(RCODEFormatError)
		} else {
			log.Tracef("received DNS packet from %s", packet.ResponseAddr.String())
			response, err := d.resolver.Resolve(packet.DNSMessage.Questions[0].ParsedName, packet.DNSMessage.Questions[0].Type)
			if err != nil {
				if err == ErrNxDomain {
					packet.DNSMessage.Header.SetRCODE(RCODENameFailure)
				} else {
					packet.DNSMessage.Header.SetRCODE(RCODEServerFailure)
				}
			} else {
				packet.DNSMessage.Header.SetRCODE(RCODESuccess)
				packet.DNSMessage.Header.SetQR(true)
				packet.DNSMessage.Answers = append(make([]*DNSRecord, 0), response)
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
