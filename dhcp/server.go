package dhcp

import (
	"fmt"
	"net"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	dhcpServerPort = 67
)

type DHCPPacket struct {
	Message      Message
	ResponseAddr net.Addr
}

type DHCPServer struct {
	opts          *DHCPServerOpts
	issuedLeases  *LeaseDB
	packetConn    net.PacketConn
	interfaceAddr net.IP
	requestChan   chan *DHCPPacket
	responseChan  chan *DHCPPacket
}

type DHCPServerOpts struct {
	Interface   string
	StartFrom   net.IP
	EndAt       net.IP
	NameServers []net.IP
	LeaseTTL    int
}

var defaultOpts = &DHCPServerOpts{
	Interface: "enp34s0",
	StartFrom: net.ParseIP("10.0.0.2").To4(),
	EndAt:     net.ParseIP("10.0.0.99").To4(),
}

func NewDHCPServer() *DHCPServer {
	return NewDHCPServerWithOpts(defaultOpts)
}

func NewDHCPServerWithOpts(opts *DHCPServerOpts) *DHCPServer {
	return &DHCPServer{
		opts:         opts,
		issuedLeases: NewLeaseDB(opts.StartFrom, opts.EndAt),
		responseChan: make(chan *DHCPPacket, 100),
		requestChan:  make(chan *DHCPPacket, 100),
	}
}

func (z *DHCPServer) Start() error {
	log.Debug("looking for interface ", z.opts.Interface)
	iface, err := net.InterfaceByName(z.opts.Interface)
	if err != nil {
		return err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		if a, ok := addr.(*net.IPNet); ok {
			z.interfaceAddr = a.IP
			break
		}
	}

	if z.interfaceAddr == nil {
		return fmt.Errorf("could not find IP network address for interface %s", z.opts.Interface)
	}

	packetConn, err := net.ListenPacket("udp4", fmt.Sprintf(":%d", dhcpServerPort))
	if err != nil {
		return err
	}
	z.packetConn = packetConn
	log.Debug("listen on ", z.interfaceAddr.String())
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
		resp := z.handleRequest(req)
		if resp != nil {
			z.responseChan <- &DHCPPacket{
				Message:      resp,
				ResponseAddr: req.ResponseAddr,
			}
		}
	}
}

func (z *DHCPServer) handleRequest(req *DHCPPacket) Message {
	opts := ParseOptions(req.Message)
	log.Debugf("transaction: %x %d from %s", req.Message.XId(), opts[OptionDHCPMessageType], req.Message.CHAddr().String())

	var resp Message
	switch DHCPMessageType(opts[OptionDHCPMessageType][0]) {
	case DHCPDiscover:
		log.Debug("discovery packet, responding with offer")

		var offeredIp net.IP
		if existingLease := z.issuedLeases.GetLease(req.Message.CHAddr().String()); existingLease != nil {
			log.Debug("found existing lease for ", req.Message.CHAddr().String())
			offeredIp = existingLease.IP
		} else {
			if nextLease := z.issuedLeases.NextAvailableLease(req.Message.CHAddr().String()); offeredIp == nil {
				log.Warn("unable to issue lease: no more remaining in pool")
				return nil
			} else {
				offeredIp = nextLease.IP
			}
		}
		resp = MakeReply(req.Message, DHCPOffer, z.interfaceAddr, offeredIp, time.Second*86400, opts)
		log.Infof("offering address %s to %s", resp.CIAddr().String(), req.Message.CHAddr().String())

	case DHCPRequest:
		if !net.IP.Equal(z.interfaceAddr, opts[OptionServerIdentifier]) {
			log.Debug("client requesting address from another server")
			return nil
		}
		requestedAddr := net.IP(opts[OptionRequestedIPAddress])
		log.Debug("client requests address ", requestedAddr)
		var msgType DHCPMessageType
		var setIP net.IP
		if lease := z.issuedLeases.GetLease(req.Message.CHAddr().String()); lease != nil {
			if net.IP.Equal(lease.IP, requestedAddr) {
				log.Info("confirming address %s for %s", requestedAddr, req.Message.CHAddr().String())
				msgType = DHCPAck
				setIP = requestedAddr
			} else {
				log.Info("reject requested address from %s", req.Message.CHAddr().String())
				msgType = DHCPNack
			}
		} else {
			msgType = DHCPAck
			if nextLease := z.issuedLeases.NextAvailableLease(req.Message.CHAddr().String()); nextLease != nil {
				setIP = nextLease.IP
			} else {
				log.Warn("no available leases")
				return nil
			}
		}
		opts = make(Options)
		nameServers := make([]byte, 0, 4*len(z.opts.NameServers))
		for i, nameServer := range z.opts.NameServers {
			copy(nameServers[i*4:i+1*4], nameServer)
		}
		opts[OptionDomainNameServer] = nameServers
		opts[OptionDomainName] = []byte("international-space-station")
		resp = MakeReply(req.Message, msgType, z.interfaceAddr, setIP, time.Second*time.Duration(z.opts.LeaseTTL), opts)
		log.Info("send ack for address %s", setIP.String())

		//check and respond with ack/nack

	case DHCPRelease:
		log.Debug("client releasing lease")
	}
	return resp
}

func (z *DHCPServer) responsePacketWorker() {
	for resp := range z.responseChan {
		log.Debugf("responding to transaction %x", resp.Message.XId())
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
		// if _, err := z.packetConn.WriteTo(resp.Message, addr); err != nil {
		// 	log.Error("Unable to respond to client: ", err.Error())
		// }
	}
}
