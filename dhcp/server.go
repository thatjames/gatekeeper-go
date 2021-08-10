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
	requestChan  chan dhcpLeaseRequest
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

type dhcpLeaseRequest struct {
	Message
	addr net.Addr
}

var defaultOpts = &DHCPServerOpts{
	ListenAddress: "127.0.0.1:67",
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
		requestChan:  make(chan dhcpLeaseRequest, 100),
		responseChan: make(chan Message, 100),
	}
}

func (z *DHCPServer) Start() error {
	return z.listen()
}

func (z *DHCPServer) listen() error {
	packetConn, err := net.ListenPacket("udp4", z.opts.ListenAddress)
	if err != nil {
		return err
	}
	go func() {
		buff := make([]byte, 1500)
		for {
			n, addr, err := packetConn.ReadFrom(buff)
			if err != nil {
				log.Fatal("unable to read datastream: ", err.Error())
			}

			if n < 240 {
				continue
			}

			req := dhcpLeaseRequest{
				Message: Message(buff[:n]),
				addr:    addr,
			}
			if req.HLen() < 16 {
				continue
			}

			z.requestChan <- req
		}
	}()
	return nil
}

func requestWorker(reqChan chan dhcpLeaseRequest, respChan chan Message, leases leaseDB) {
	for req := range reqChan {
		log.Debug("request from ", req.addr.String())

		resp := Message([]byte{0})
		respChan <- resp
	}
}
