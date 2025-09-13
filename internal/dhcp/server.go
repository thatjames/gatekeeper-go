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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/common"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
)

const (
	dhcpServerPort = 67
)

var (
	reqDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "dhcp_req_time",
		Help:    "dhcp request time buckets",
		Buckets: []float64{1, 10, 100, 250, 500, 1000, 2500, 5000, 10000},
	})

	opCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "dhcp_op_counter",
		Help: "Count by type of operations",
	}, []string{"op", "client", "hostname"})

	activeLeaseGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "active_lease_count",
		Help: "count of active leases",
	})
)

type DHCPPacket struct {
	Message      Message
	ResponseAddr net.Addr
}

type DHCPServer struct {
	opts          *DHCPServerOpts
	issuedLeases  *LeasePool
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
		LeaseFile:      config.LeaseFile,
	}

	return NewDHCPServerWithOpts(options)
}

func NewDHCPServerWithOpts(opts *DHCPServerOpts) *DHCPServer {
	return &DHCPServer{
		opts:         opts,
		issuedLeases: NewLeasePool(opts.StartFrom, opts.EndAt),
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
			if lease != nil && lease.State == common.LeaseActive {
				counter++
			}
		}
	}

	for clientID, lease := range z.opts.ReservedLeases {
		z.issuedLeases.ReserveLease(clientID, net.ParseIP(lease).To4())
		activeLeaseGauge.Inc()
		log.Infof("reserving %s for %s", lease, clientID)
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

	return z.issuedLeases.PeristLeases(z.opts.LeaseFile)
}

func (z *DHCPServer) LeaseDB() *LeasePool {
	return z.issuedLeases
}

func (z *DHCPServer) Options() *DHCPServerOpts {
	return z.opts
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
		tsStart := time.Now()
		resp := z.handleRequest(req)
		opts := ParseOptions(req.Message)
		host := req.Message.CHAddr().String()
		if opts[OptionHostname] != nil {
			host = string(opts[OptionHostname])
		}
		opCounter.With(prometheus.Labels{"op": DHCPMessageType(opts[OptionDHCPMessageType][0]).String(), "client": req.Message.CHAddr().String(), "hostname": host}).Inc()
		if resp != nil {
			z.responseChan <- &DHCPPacket{
				Message:      resp,
				ResponseAddr: req.ResponseAddr,
			}
			opCounter.With(prometheus.Labels{"op": DHCPMessageType(ParseOptions(resp)[OptionDHCPMessageType][0]).String(), "client": req.Message.CHAddr().String(), "hostname": host}).Inc()
		}
		tsEnd := time.Since(tsStart).Round(time.Millisecond)
		reqDuration.Observe(float64(tsEnd.Milliseconds()))
	}
}

func (z *DHCPServer) handleRequest(req *DHCPPacket) Message {
	opts := ParseOptions(req.Message)
	id := req.Message.CHAddr().String()
	if opts[OptionHostname] != nil {
		id = string(opts[OptionHostname])
	}
	log.Debugf("%s starting transaction %x: %v", id, req.Message.XId(), opts[OptionDHCPMessageType])

	var resp Message
	switch DHCPMessageType(opts[OptionDHCPMessageType][0]) {
	case DHCPDiscover:
		var offeredIp net.IP
		if existingLease := z.issuedLeases.GetLease(req.Message.CHAddr().String()); existingLease != nil {
			log.Debugf("found existing lease %s for %s", existingLease.IP.To4().String(), id)
			offeredIp = existingLease.IP
		} else {
			if nextLease := z.issuedLeases.NextAvailableLease(req.Message.CHAddr().String()); nextLease == nil {
				log.Warn("unable to issue lease: no more remaining in pool")
				return nil
			} else {
				log.Infof("reserving available lease %s for %s", nextLease.IP.To4().String(), id)
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
			copy(nameServers[i*4:(i+1)*4], []byte(nameServer.To4()))
		}
		responseOptions[OptionDomainNameServer] = nameServers
		responseOptions[OptionDomainName] = []byte(z.opts.DomainName)
		resp = MakeReply(req.Message, DHCPOffer, z.interfaceAddr, offeredIp, time.Second*time.Duration(z.opts.LeaseTTL), responseOptions)
		log.Infof("offering address %s to %s", resp.YIAddr().String(), id)

	case DHCPRequest:
		if val, ok := opts[OptionServerIdentifier]; ok && !net.IP.Equal(z.interfaceAddr, net.IP(val)) {
			log.Debug("client requesting address from another server")
			return nil
		}
		var requestedAddr net.IP
		if requestedAddr = opts[OptionRequestedIPAddress]; requestedAddr == nil {
			requestedAddr = req.Message.CIAddr()
		}
		log.Debugf("%s requests address %s", id, requestedAddr)
		var msgType DHCPMessageType
		var setIP net.IP
		if lease := z.issuedLeases.GetLease(req.Message.CHAddr().String()); lease != nil {
			if hostname, ok := opts[OptionHostname]; ok {
				lease.Hostname = string(hostname)
			}
			if lease.IP.Equal(requestedAddr) {
				msgType = DHCPAck
				setIP = requestedAddr
				switch lease.State {
				case common.LeaseOffered:
					log.Infof("confirm address %s for %s", requestedAddr, id)
					z.issuedLeases.AcceptLease(lease, time.Second*time.Duration(z.opts.LeaseTTL))
					activeLeaseGauge.Inc()
				case common.LeaseReserved:
					z.issuedLeases.AcceptLease(lease, time.Second*time.Duration(z.opts.LeaseTTL))
					log.Infof("send ack for reserved address %s to %s", requestedAddr, id)
				case common.LeaseActive:
					z.issuedLeases.AcceptLease(lease, time.Second*time.Duration(z.opts.LeaseTTL))
					lease.Expiry = time.Now().Add(time.Second * time.Duration(z.opts.LeaseTTL))
					log.Infof("send ack for active address %s to %s", lease.IP.To4().String(), id)
				default:
					log.Infof("lease is invalid, resetting and nacking")
					activeLeaseGauge.Dec()
					msgType = DHCPNack
					z.issuedLeases.ReleaseLease(lease)
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
			copy(nameServers[i*4:(i+1)*4], []byte(nameServer.To4()))
		}
		responseOptions[OptionDomainNameServer] = nameServers
		responseOptions[OptionDomainName] = []byte(z.opts.DomainName)
		resp = MakeReply(req.Message, msgType, z.interfaceAddr, setIP, time.Second*time.Duration(z.opts.LeaseTTL), responseOptions)

	case DHCPRelease:
		activeLeaseGauge.Dec()
		lease := z.issuedLeases.GetLease(req.Message.CHAddr().String())
		if lease == nil {
			return nil
		}
		log.Info(lease.Hostname, " releasing lease")
		z.issuedLeases.ReleaseLease(lease)
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
