package dhcp

import (
	"net"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type DHCPPacket struct {
	Message      Message
	ResponseAddr net.Addr
}

type DHCPServer struct {
	opts         *DHCPServerOpts
	issuedLeases leaseDB
	packetConn   net.PacketConn
	requestChan  chan *DHCPPacket
	responseChan chan *DHCPPacket
}

type DHCPServerOpts struct {
	ListenAddress string
	StartFrom     net.IP
	NumLeases     int
}

type lease struct {
	ClientId string
	Expiry   time.Time
}
type leaseDB struct {
	leases map[string]*lease
	lock   *sync.Mutex
}

func (l *leaseDB) get(id string) (*lease, bool) {
	l.lock.Lock()
	defer l.lock.Unlock()
	ls, ok := l.leases[id]
	return ls, ok
}

func (l *leaseDB) set(id string, ls *lease) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.leases[id] = ls
}

var defaultOpts = &DHCPServerOpts{
	ListenAddress: "0.0.0.0:67",
	StartFrom:     net.ParseIP("10.0.0.1"),
	NumLeases:     254,
}

func NewDHCPServer() *DHCPServer {
	return NewDHCPServerWithOpts(defaultOpts)
}

func NewDHCPServerWithOpts(opts *DHCPServerOpts) *DHCPServer {
	return &DHCPServer{
		opts: opts,
		issuedLeases: leaseDB{
			leases: make(map[string]*lease),
			lock:   new(sync.Mutex),
		},
		responseChan: make(chan *DHCPPacket, 100),
		requestChan:  make(chan *DHCPPacket, 100),
	}
}

func (z *DHCPServer) Start() error {
	packetConn, err := net.ListenPacket("udp4", z.opts.ListenAddress)
	if err != nil {
		return err
	}
	z.packetConn = packetConn
	log.Debug("Listen on ", z.opts.ListenAddress)
	go z.listen()
	for i := 0; i < 10; i++ {
		go z.receivePacketWorker()
		go z.responsePacketWorker()
	}
	return nil
}

func (z *DHCPServer) listen() error {
	go func() {
		buff := make([]byte, 1500)
		for {
			log.Debug("waiting on packet")
			n, addr, err := z.packetConn.ReadFrom(buff)
			if err != nil {
				log.Error("unable to read datastream: ", err.Error())
				continue
			}

			if msg := Message(buff[:n]); n >= 240 && msg.OpCode() == OpRequest {
				z.requestChan <- &DHCPPacket{
					Message:      msg,
					ResponseAddr: addr,
				}

			}
		}
	}()
	return nil
}

func (z *DHCPServer) receivePacketWorker() {
	for req := range z.requestChan {
		opts := ParseOptions(req.Message)
		log.Debugf("Transaction: %x %d from %s", req.Message.XId(), opts[OptionDHCPMessageType], req.Message.CHAddr().String())
		log.Debug("Request options: ", opts)

		var resp Message
		switch DHCPMessageType(opts[OptionDHCPMessageType][0]) {
		case DHCPDiscover:
			log.Debug("discovery packet, responding with offer")

			opts[OptionDHCPMessageType] = []byte{byte(DHCPOffer)}
			resp = DHCPReply(req.Message, []byte{10, 0, 0, 1}, []byte{10, 0, 0, 100}, time.Second*86400, opts)
			log.Debugf("offering address %s", resp.CIAddr().String())

		case DHCPRequest:
			log.Debug("client requests address ", net.IP(opts[OptionRequestedIPAddress]))
			opts = make(Options)
			opts[OptionDHCPMessageType] = []byte{byte(DHCPAck)}
			opts[OptionDomainNameServer] = []byte{8, 8, 8, 8, 1, 1, 1, 1, 9, 9, 9, 9}
			opts[OptionDomainName] = []byte("international-space-station")
			// opts[OptionNetbiosNameServer] = []byte{10, 0, 0, 1}
			resp = DHCPReply(req.Message, []byte{10, 0, 0, 1}, []byte{10, 0, 0, 100}, time.Second*86400, opts)
			log.Debugf("acking address %s", net.IP(opts[OptionRequestedIPAddress]).String())

			//check and respond with ack/nack

		case DHCPRelease:
			log.Debug("client releasing lease")
		}

		if resp != nil {
			z.responseChan <- &DHCPPacket{
				Message:      resp,
				ResponseAddr: req.ResponseAddr,
			}
		}
	}
}

func (z *DHCPServer) responsePacketWorker() {
	for resp := range z.responseChan {
		log.Debugf("Responding to transaction %x", resp.Message.XId())
		addr := resp.ResponseAddr
		ip, port, err := net.SplitHostPort(addr.String())
		if err != nil {
			log.Error(err.Error())
			continue
		}

		if net.ParseIP(ip).Equal(net.IPv4zero) {
			p, _ := strconv.Atoi(port)
			addr = &net.UDPAddr{
				IP:   net.IPv4bcast,
				Port: p,
			}
		}
		if _, err := z.packetConn.WriteTo(resp.Message, addr); err != nil {
			log.Error("Unable to respond to client: ", err.Error())
		}
	}
}
