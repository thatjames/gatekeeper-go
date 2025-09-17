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
	responseChan chan *DNSPacket
	exitChan     chan struct{}
}

func NewDNSServer() *DNSServer {
	return NewDNSServerWithOpts(defaultDNSServerOpts)
}

func NewDNSServerWithOpts(opts DNSServerOpts) *DNSServer {
	return &DNSServer{
		resolver:     NewDNSResolverWithDefaultOpts(),
		opts:         &opts,
		packetConn:   nil,
		receiverChan: make(chan *DNSPacket, 100),
		responseChan: make(chan *DNSPacket, 100),
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
		log.Debugf("received DNS packet from %s", packet.ResponseAddr.String())
		response, err := d.resolver.Resolve(packet.DNSMessage.Questions[0].Name)
		if err != nil {
			if err == ErrNxDomain {
				packet.DNSMessage.Header.SetRCODE(RCODENameFailure)
			} else {
				packet.DNSMessage.Header.SetRCODE(RCODEServerFailure)
			}
			log.Error("unable to resolve: ", err.Error())
			continue
		} else {
			packet.DNSMessage.Header.SetRCODE(RCODESuccess)
			packet.DNSMessage.Answers = append(make([]DNSRecord, 0), *response)
		}
		d.responseChan <- packet
	}
}

func (d *DNSServer) responseWorker() {
	for packet := range d.responseChan {
		log.Debugf("sending DNS packet to %s", packet.ResponseAddr.String())
		packet.Header.SetQR(true)
		data, err := MarshalDNSMessage(packet.DNSMessage)
		if err != nil {
			log.Error("unable to marshal DNS packet: ", err.Error())
			continue
		}
		n, err := d.packetConn.WriteTo(data, packet.ResponseAddr)
		if err != nil {
			log.Error("unable to send DNS packet: ", err.Error())
		} else {
			log.Debugf("sent %d bytes to %s", n, packet.ResponseAddr.String())
		}
	}
}
