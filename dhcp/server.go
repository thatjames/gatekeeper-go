package dhcp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"gitlab.com/thatjames-go/gatekeeper-go/config"
	"gitlab.com/thatjames-go/gatekeeper-go/service"
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
	broadcastAddr net.IP
	requestChan   chan *DHCPPacket
	responseChan  chan *DHCPPacket
}

type DHCPServerOpts struct {
	Interface      string
	StartFrom      net.IP
	EndAt          net.IP
	NameServers    []net.IP
	LeaseTTL       int
	Router         net.IP
	SubnetMask     net.IP
	DomainName     string
	ReservedLeases map[string]string
	LeaseFile      string
}

var defaultOpts = &DHCPServerOpts{
	Interface: "enp34s0",
	StartFrom: net.ParseIP("10.0.0.2").To4(),
	EndAt:     net.ParseIP("10.0.0.99").To4(),
}

func NewDHCPServer() *DHCPServer {
	return NewDHCPServerWithOpts(defaultOpts)
}

func NewDHCPServerFromConfig(config *config.DHCP) *DHCPServer {
	nameServers := make([]net.IP, len(config.NameServers))
	for i := range nameServers {
		nameServers[i] = net.ParseIP(config.NameServers[i])
	}
	options := &DHCPServerOpts{
		Interface:      config.Interface,
		StartFrom:      net.ParseIP(config.StartAddr).To4(),
		EndAt:          net.ParseIP(config.EndAddr).To4(),
		NameServers:    nameServers,
		LeaseTTL:       config.LeaseTTL,
		Router:         net.ParseIP(config.Router).To4(),
		SubnetMask:     net.ParseIP(config.SubnetMask).To4(),
		DomainName:     config.DomainName,
		ReservedLeases: config.ReservedAddresses,
	}

	return NewDHCPServerWithOpts(options)
}

func NewDHCPServerWithOpts(opts *DHCPServerOpts) *DHCPServer {
	return &DHCPServer{
		opts:         opts,
		issuedLeases: NewLeaseDB(opts.StartFrom, opts.EndAt),
		responseChan: make(chan *DHCPPacket, 100),
		requestChan:  make(chan *DHCPPacket, 100),
	}
}

func (z *DHCPServer) Type() service.ServiceKey {
	return service.DHCP
}

func (z *DHCPServer) Start() error {
	log.Debug("looking for interface ", z.opts.Interface)
	iface, err := net.InterfaceByName(z.opts.Interface)
	if err != nil {
		return err
	}

	if iface.Flags&net.FlagBroadcast == 0 {
		return errors.New("interface does not support broadcast")
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		if a, ok := addr.(*net.IPNet); ok {
			z.interfaceAddr = a.IP
			z.broadcastAddr = make(net.IP, 4)

			//calculate broadcast for the given network mask by OR'ing the complement of the mask with the current address
			//example: given address 10.0.0.1 and mask 255.255.255.0:
			// 1. take complement of the mask: ^mask = [00 00 00 ff]
			// 2. ^mask | addr: [10 00 00 ff] (10.0.0.255)
			binary.BigEndian.PutUint32(z.broadcastAddr, binary.BigEndian.Uint32(a.IP.To4())|^binary.BigEndian.Uint32(net.IP(a.Mask).To4()))
			log.Debug("set broadcast address ", z.broadcastAddr, " from mask ", net.IP(a.Mask).String())
			break
		}
	}

	if z.interfaceAddr == nil {
		return fmt.Errorf("could not find IP network address for interface %s", z.opts.Interface)
	}

	log.Debug("load any existing leases")
	leaseFile := z.opts.LeaseFile
	if leaseFile == "" {
		leaseFile = "/var/lib/gatekeeper/leases"
	}
	if err := z.issuedLeases.LoadLeases(leaseFile, time.Second*time.Duration(z.opts.LeaseTTL)); err != nil {
		log.Warn("unable to load leases: ", err.Error())
	} else {
		counter := 0
		for _, lease := range z.issuedLeases.leases {
			if lease != nil && lease.State == LeaseActive {
				counter++
			}
		}
		log.Debug("loaded ", counter, " leases")
	}

	for clientID, lease := range z.opts.ReservedLeases {
		z.issuedLeases.ReserveLease(clientID, net.ParseIP(lease).To4())
		log.Debugf("reserving %s for %s", lease, clientID)
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

func (z *DHCPServer) Stop() error {
	leaseFile := z.opts.LeaseFile
	if leaseFile == "" {
		leaseFile = "/var/lib/gatekeeper/leases"
	}

	if err := os.MkdirAll(path.Dir(leaseFile), os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}

	return z.issuedLeases.PeristLeases(leaseFile)
}

func (z *DHCPServer) listen() {
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
			if nextLease := z.issuedLeases.NextAvailableLease(req.Message.CHAddr().String()); nextLease == nil {
				log.Warn("unable to issue lease: no more remaining in pool")
				return nil
			} else {
				offeredIp = nextLease.IP
			}
		}
		responseOptions := make(Options)
		responseOptions[OptionDHCPMessageType] = []byte{(byte(DHCPOffer))}
		responseOptions[OptionServerIdentifier] = z.interfaceAddr.To4()
		responseOptions[OptionIPLeaseTime] = make([]byte, 4)
		binary.BigEndian.PutUint32(responseOptions[OptionIPLeaseTime], uint32(z.opts.LeaseTTL))
		responseOptions[OptionRouter] = z.opts.Router.To4()
		responseOptions[OptionSubnetMask] = z.opts.SubnetMask.To4()
		nameServers := make([]byte, 4*len(z.opts.NameServers))
		for i, nameServer := range z.opts.NameServers {
			copy(nameServers[i*4:(i+1)*4], nameServer)
		}
		responseOptions[OptionDomainNameServer] = nameServers
		responseOptions[OptionDomainName] = []byte(z.opts.DomainName)
		resp = MakeReply(req.Message, DHCPOffer, z.interfaceAddr, offeredIp, time.Second*time.Duration(z.opts.LeaseTTL), responseOptions)
		log.Infof("offering address %s to %s", resp.YIAddr().String(), req.Message.CHAddr().String())

	case DHCPRequest:
		if val, ok := opts[OptionServerIdentifier]; ok && !net.IP.Equal(z.interfaceAddr, net.IP(val)) {
			log.Debug("client requesting address from another server")
			return nil
		}
		var requestedAddr net.IP
		if requestedAddr = opts[OptionRequestedIPAddress]; requestedAddr == nil {
			requestedAddr = req.Message.CIAddr()
		}
		log.Debug("client requests address ", requestedAddr)
		var msgType DHCPMessageType
		var setIP net.IP
		if lease := z.issuedLeases.GetLease(req.Message.CHAddr().String()); lease != nil {
			if lease.IP.Equal(requestedAddr) {
				log.Infof("send ack for %s", requestedAddr)
				msgType = DHCPAck
				setIP = requestedAddr
				if lease.State == LeaseOffered {
					z.issuedLeases.AcceptLease(lease, time.Second*time.Duration(z.opts.LeaseTTL))
				}
			} else {
				log.Infof("reject requested address from %s", req.Message.CHAddr().String())
				msgType = DHCPNack
			}
		} else {
			//The client needs to ask for a new address
			msgType = DHCPNack
		}
		responseOptions := make(Options)
		responseOptions[OptionDHCPMessageType] = []byte{(byte(msgType))}
		responseOptions[OptionServerIdentifier] = z.interfaceAddr.To4()
		responseOptions[OptionIPLeaseTime] = make([]byte, 4)
		binary.BigEndian.PutUint32(responseOptions[OptionIPLeaseTime], uint32(z.opts.LeaseTTL))
		responseOptions[OptionRouter] = z.opts.Router.To4()
		responseOptions[OptionSubnetMask] = z.opts.SubnetMask.To4()
		nameServers := make([]byte, 4*len(z.opts.NameServers))
		for i, nameServer := range z.opts.NameServers {
			copy(nameServers[i*4:(i+1)*4], nameServer)
		}
		responseOptions[OptionDomainNameServer] = nameServers
		responseOptions[OptionDomainName] = []byte(z.opts.DomainName)
		resp = MakeReply(req.Message, msgType, z.interfaceAddr, setIP, time.Second*time.Duration(z.opts.LeaseTTL), responseOptions)

		//check and respond with ack/nack

	case DHCPRelease:
		log.Debug("client releasing lease")
	}
	return resp
}

func (z *DHCPServer) responsePacketWorker() {
	for resp := range z.responseChan {
		addr := resp.ResponseAddr
		ip, port, err := net.SplitHostPort(addr.String())
		if err != nil {
			log.Error(err.Error())
			continue
		}

		if net.ParseIP(ip).Equal(net.IPv4zero) {
			p, _ := strconv.Atoi(port)
			addr = &net.UDPAddr{
				IP:   z.broadcastAddr,
				Port: p,
			}
		}

		log.Debugf("responding to transaction %x at %s", resp.Message.XId(), addr.String())
		if _, err := z.packetConn.WriteTo(resp.Message, addr); err != nil {
			log.Error("unable to respond to client: ", err.Error())
		}
	}
}

func (z *DHCPServer) LeaseDB() *LeaseDB {
	return z.issuedLeases
}
