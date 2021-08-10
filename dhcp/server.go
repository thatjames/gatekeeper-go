package dhcp

import (
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type DHCPServer struct {
	opts         *DHCPServerOpts
	issuedLeases leaseDB
	requestChan  chan Message
	responseChan chan Message
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
		requestChan:  make(chan Message, 100),
		responseChan: make(chan Message, 100),
	}
}

func (z *DHCPServer) Start() error {
	for i := 0; i < 10; i++ {
		go requestWorker(z.requestChan, z.responseChan, z.issuedLeases)
	}
	return z.listen()
}

func (z *DHCPServer) listen() error {
	packetConn, err := net.ListenPacket("udp4", z.opts.ListenAddress)
	if err != nil {
		return err
	}
	log.Debug("Listen on ", z.opts.ListenAddress)
	go func() {
		buff := make([]byte, 1500)
		for {
			log.Debug("waiting on packet")
			n, _, err := packetConn.ReadFrom(buff)
			if err != nil {
				log.Error("unable to read datastream: ", err.Error())
				continue
			}

			if msg := Message(buff[:n]); n >= 240 && msg.OpCode() == OpRequest {
				z.requestChan <- Message(buff[:n])
			}
		}
	}()
	return nil
}

func requestWorker(reqChan <-chan Message, respChan chan<- Message, leases leaseDB) {
	for req := range reqChan {
		opts := ParseOptions(req)
		log.Debugf("Transaction: %x %d from %s", req.XId(), opts[OptionDHCPMessageType], req.CHAddr().String())

		resp := Message(make([]byte, 0))
		switch int(opts[OptionDHCPMessageType][0]) {
		case DHCPDiscover:
			log.Debug("discovery packet, responding with offer")

			//respond(offer)

		case DHCPRequest:
			log.Debug("client requests address ", net.IP(opts[OptionRequestedIPAddress]))

			//check and respond with ack/nack

		case DHCPRelease:
			log.Debug("client releasing lease")
		}

		respChan <- resp
	}
}
