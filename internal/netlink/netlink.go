//go:build linux
// +build linux

package netlink

import (
	"fmt"
	"net"

	"github.com/mdlayher/netlink"
	"github.com/mdlayher/netlink/nlenc"
	log "github.com/sirupsen/logrus"
)

const (
	RTM_GETLINK   = 18
	RTM_NEWLINK   = 16
	RTM_GETROUTE  = 26
	RTM_NEWROUTE  = 24
	RTM_NEWADDR   = 20
	RTM_GETADDR   = 22
	NLM_F_REQUEST = 0x1
	NLM_F_DUMP    = 0x300
	NLM_F_ACK     = 0x4
	NLM_F_EXCL    = 0x200
	NLM_F_CREATE  = 0x400

	IFF_UP = 0x1

	IFLA_IFNAME = 3

	IFA_ADDRESS = 1
	IFA_LOCAL   = 2

	RTA_DST      = 1
	RTA_OIF      = 4
	RTA_GATEWAY  = 5
	RTA_PRIORITY = 6
	RTA_PREFSRC  = 7

	RT_SCOPE_UNIVERSE = 0
	RT_SCOPE_LINK     = 253

	NETLINK_ROUTE = 0
)

var (
	scopeNames = map[uint8]string{
		0:   "global",
		200: "site",
		253: "link",
		254: "host",
	}

	protoNames = map[uint8]string{
		0:  "unspec",
		2:  "kernel",
		3:  "boot",
		4:  "static",
		16: "dhcp",
	}
)

type InterfaceAttributes struct {
	Index     uint32
	Name      string
	Flags     uint32
	Type      uint16
	MTU       uint32
	Address   net.HardwareAddr
	Broadcast net.HardwareAddr
	State     string // "up", "down"
	Operstate string // operational state: "unknown", "notpresent", "down", "lowerlayerdown", "testing", "dormant", "up"
}

type InterfaceAddress struct {
	Family    uint8
	PrefixLen uint8
	Scope     uint8
	Index     uint32
	Address   string
	Local     string
	Broadcast string
}

type NetlinkServiceOpts struct {
	WANInterface string
	LANInterface string
}

type NetlinkService struct {
	wanInterface string
	lanInterface string
	netlinkConn  *netlink.Conn
}

func NewNetlinkService(opts NetlinkServiceOpts) *NetlinkService {
	return &NetlinkService{
		wanInterface: opts.WANInterface,
		lanInterface: opts.LANInterface,
	}
}

func (z *NetlinkService) Start() error {
	log.Info("Starting netlink service")
	var err error
	if z.netlinkConn, err = netlink.Dial(NETLINK_ROUTE, nil); err != nil {
		return err
	}

	// check if interfaces are up
	ifaces, err := z.getInterfaces()
	if err != nil {
		return err
	}

	for _, iface := range ifaces {
		if iface.Name == z.wanInterface {
			if iface.State != "up" {
				log.Debugf("interface %s is down, setting up", z.wanInterface)
				if err := z.setInterfaceUp(z.wanInterface); err != nil {
					return err
				}
			}
			addresses, err := z.getInterfaceAddressesByIndex(iface.Index)
			if err != nil {
				return err
			}

			if len(addresses) == 0 {
				log.Debugf("no addresses found for interface %s, adding default", z.wanInterface)
				if err := z.setAddrForInterface(z.wanInterface, "10.0.0.1/24"); err != nil {
					return err
				}
			}

		} else if iface.Name == z.lanInterface {
		}
	}

	return nil
}

func (z *NetlinkService) Stop() error {
	return z.netlinkConn.Close()
}

func (z *NetlinkService) getInterfaces() ([]InterfaceAttributes, error) {
	var interfaces []InterfaceAttributes

	// Build request
	req := netlink.Message{
		Header: netlink.Header{
			Type:  RTM_GETLINK,
			Flags: NLM_F_REQUEST | NLM_F_DUMP,
		},
		Data: make([]byte, 16), // IfInfomsg
	}

	msgs, err := z.netlinkConn.Execute(req)
	if err != nil {
		return nil, err
	}

	for _, msg := range msgs {
		if msg.Header.Type != RTM_NEWLINK {
			continue
		}

		if len(msg.Data) < 16 {
			continue
		}

		ifInfo := InterfaceAttributes{
			Index: nlenc.Uint32(msg.Data[4:8]),
			Flags: nlenc.Uint32(msg.Data[8:12]),
			Type:  nlenc.Uint16(msg.Data[0:2]),
		}

		// Determine state from flags
		if ifInfo.Flags&IFF_UP != 0 {
			ifInfo.State = "up"
		} else {
			ifInfo.State = "down"
		}

		// Parse attributes starting after IfInfomsg (16 bytes)
		parseInterfaceAttributes(msg.Data[16:], &ifInfo)

		interfaces = append(interfaces, ifInfo)
	}

	return interfaces, nil
}

func (z *NetlinkService) getInterfaceAddressesByIndex(ifIndex uint32) ([]InterfaceAddress, error) {
	var addresses []InterfaceAddress

	// Build request
	req := netlink.Message{
		Header: netlink.Header{
			Type:  RTM_GETADDR,
			Flags: NLM_F_REQUEST | NLM_F_DUMP,
		},
		Data: make([]byte, 8), // IfAddrmsg
	}

	msgs, err := z.netlinkConn.Execute(req)
	if err != nil {
		return nil, err
	}

	for _, msg := range msgs {
		if msg.Header.Type != RTM_NEWADDR {
			continue
		}

		if len(msg.Data) < 8 {
			continue
		}

		// Parse IfAddrmsg
		msgIfIndex := nlenc.Uint32(msg.Data[4:8])

		// Skip if not the interface we're looking for
		if msgIfIndex != ifIndex {
			continue
		}

		addr := InterfaceAddress{
			Family:    msg.Data[0],
			PrefixLen: msg.Data[1],
			Scope:     msg.Data[3],
			Index:     msgIfIndex,
		}

		// Parse attributes starting after IfAddrmsg (8 bytes)
		attrs, err := parseRouteAttributes(msg.Data[8:])
		if err != nil {
			continue
		}

		// IFA_ADDRESS = 1
		if address, ok := attrs[IFA_ADDRESS]; ok {
			addr.Address = net.IP(address).String()
		}

		// IFA_LOCAL = 2
		if local, ok := attrs[IFA_LOCAL]; ok {
			addr.Local = net.IP(local).String()
		}

		// IFA_BROADCAST = 4
		if broadcast, ok := attrs[4]; ok {
			addr.Broadcast = net.IP(broadcast).String()
		}

		addresses = append(addresses, addr)
	}

	return addresses, nil
}

func (z *NetlinkService) getRoutes(interfaces map[uint32]string) ([]Route, error) {
	var routes []Route

	req := netlink.Message{
		Header: netlink.Header{
			Type:  RTM_GETROUTE,
			Flags: NLM_F_REQUEST | NLM_F_DUMP,
		},
		Data: make([]byte, 12), // RtMsg
	}

	msgs, err := z.netlinkConn.Execute(req)
	if err != nil {
		return nil, err
	}

	for _, msg := range msgs {
		if msg.Header.Type != RTM_NEWROUTE {
			continue
		}

		if len(msg.Data) < 12 {
			continue
		}

		route := Route{
			Family:   msg.Data[0],
			DstLen:   msg.Data[1],
			Protocol: msg.Data[4],
			Scope:    msg.Data[5],
		}

		attrs, err := parseRouteAttributes(msg.Data[12:])
		if err != nil {
			continue
		}

		if dst, ok := attrs[RTA_DST]; ok {
			route.Dst = net.IP(dst).String()
		}

		if gateway, ok := attrs[RTA_GATEWAY]; ok {
			route.Gateway = net.IP(gateway).String()
		}

		if oifBytes, ok := attrs[RTA_OIF]; ok && len(oifBytes) >= 4 {
			oif := nlenc.Uint32(oifBytes)
			if ifname, ok := interfaces[oif]; ok {
				route.Dev = ifname
			} else {
				route.Dev = fmt.Sprintf("if%d", oif)
			}
		}

		if prefsrc, ok := attrs[RTA_PREFSRC]; ok {
			route.Src = net.IP(prefsrc).String()
		}

		if metricBytes, ok := attrs[RTA_PRIORITY]; ok && len(metricBytes) >= 4 {
			route.Metric = nlenc.Uint32(metricBytes)
		}

		routes = append(routes, route)
	}

	return routes, nil
}

func (z *NetlinkService) setInterfaceUp(ifname string) error {
	return z.setInterfaceState(ifname, true)
}

func (z *NetlinkService) setInterfaceDown(ifname string) error {
	return z.setInterfaceState(ifname, false)
}

func (z *NetlinkService) setInterfaceState(ifname string, up bool) error {
	ifIndex, err := z.getInterfaceIndex(ifname)
	if err != nil {
		return fmt.Errorf("failed to get interface index: %w", err)
	}

	ifInfo := make([]byte, 16)
	ifInfo[0] = 0 // Family (AF_UNSPEC)
	// bytes 1-3 are padding
	nlenc.PutUint32(ifInfo[4:8], ifIndex)

	if up {
		nlenc.PutUint32(ifInfo[8:12], IFF_UP)  // Flags
		nlenc.PutUint32(ifInfo[12:16], IFF_UP) // Change mask
	} else {
		nlenc.PutUint32(ifInfo[8:12], 0)       // Flags (clear UP)
		nlenc.PutUint32(ifInfo[12:16], IFF_UP) // Change mask
	}

	req := netlink.Message{
		Header: netlink.Header{
			Type:  RTM_NEWLINK,
			Flags: NLM_F_REQUEST | NLM_F_ACK,
		},
		Data: ifInfo,
	}

	_, err = z.netlinkConn.Execute(req)
	if err != nil {
		return fmt.Errorf("failed to set interface state: %w", err)
	}

	return nil
}

func (z *NetlinkService) setAddrForInterface(ifname, cidr string) error {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %w", err)
	}

	ifIndex, err := z.getInterfaceIndex(ifname)
	if err != nil {
		return fmt.Errorf("failed to get interface index: %w", err)
	}

	// Determine address family
	var family uint8
	var ipBytes []byte
	if ip.To4() != nil {
		family = 2 // AF_INET
		ipBytes = ip.To4()
	} else {
		family = 10 // AF_INET6
		ipBytes = ip.To16()
	}

	prefixLen, _ := ipNet.Mask.Size()

	ifAddr := make([]byte, 8)
	ifAddr[0] = family                    // Family
	ifAddr[1] = uint8(prefixLen)          // Prefix length
	ifAddr[2] = 0                         // Flags
	ifAddr[3] = RT_SCOPE_UNIVERSE         // Scope
	nlenc.PutUint32(ifAddr[4:8], ifIndex) // Interface index

	attrs := make([]byte, 0)

	localAttr := buildAttribute(IFA_LOCAL, ipBytes)
	attrs = append(attrs, localAttr...)

	addrAttr := buildAttribute(IFA_ADDRESS, ipBytes)
	attrs = append(attrs, addrAttr...)

	data := append(ifAddr, attrs...)

	req := netlink.Message{
		Header: netlink.Header{
			Type:  RTM_NEWADDR,
			Flags: NLM_F_REQUEST | NLM_F_ACK | NLM_F_CREATE | NLM_F_EXCL,
		},
		Data: data,
	}

	_, err = z.netlinkConn.Execute(req)
	if err != nil {
		return fmt.Errorf("failed to add address: %w", err)
	}

	return nil
}

func (z *NetlinkService) getInterfaceIndex(ifname string) (uint32, error) {
	interfaces, err := z.getInterfaces()
	if err != nil {
		return 0, err
	}

	for _, iface := range interfaces {
		if iface.Name == ifname {
			return iface.Index, nil
		}
	}

	return 0, fmt.Errorf("interface %s not found", ifname)
}

func buildAttribute(attrType uint16, data []byte) []byte {
	length := uint16(4 + len(data))
	attr := make([]byte, (length+3)&^3) // Align to 4 bytes

	nlenc.PutUint16(attr[0:2], length)
	nlenc.PutUint16(attr[2:4], attrType)
	copy(attr[4:], data)

	return attr
}

func parseRouteAttributes(data []byte) (map[uint16][]byte, error) {
	attrs := make(map[uint16][]byte)

	for len(data) >= 4 {
		length := nlenc.Uint16(data[0:2])
		attrType := nlenc.Uint16(data[2:4])

		if length < 4 || int(length) > len(data) {
			break
		}

		value := make([]byte, length-4)
		copy(value, data[4:length])
		attrs[attrType] = value

		aligned := (length + 3) & ^uint16(3)
		data = data[aligned:]
	}

	return attrs, nil
}

func parseInterfaceAttributes(data []byte, ifInfo *InterfaceAttributes) {
	for len(data) >= 4 {
		length := nlenc.Uint16(data[0:2])
		attrType := nlenc.Uint16(data[2:4])

		if length < 4 || int(length) > len(data) {
			break
		}

		value := data[4:length]

		switch attrType {
		case IFLA_IFNAME: // 3 - interface name
			if len(value) > 0 && value[len(value)-1] == 0 {
				value = value[:len(value)-1]
			}
			ifInfo.Name = string(value)
		case 4: // IFLA_MTU
			if len(value) >= 4 {
				ifInfo.MTU = nlenc.Uint32(value)
			}
		case 1: // IFLA_ADDRESS
			ifInfo.Address = make(net.HardwareAddr, len(value))
			copy(ifInfo.Address, value)
		case 2: // IFLA_BROADCAST
			ifInfo.Broadcast = make(net.HardwareAddr, len(value))
			copy(ifInfo.Broadcast, value)
		case 16: // IFLA_OPERSTATE
			if len(value) >= 1 {
				ifInfo.Operstate = operstateToString(value[0])
			}
		}

		aligned := (length + 3) & ^uint16(3)
		data = data[aligned:]
	}
}

func operstateToString(state uint8) string {
	states := map[uint8]string{
		0: "unknown",
		1: "notpresent",
		2: "down",
		3: "lowerlayerdown",
		4: "testing",
		5: "dormant",
		6: "up",
	}
	if s, ok := states[state]; ok {
		return s
	}
	return "unknown"
}
