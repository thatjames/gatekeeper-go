package dhcp

import (
	"bytes"
	"encoding/binary"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

type OpCode byte

const (
	OpUnkown OpCode = iota
	OpRequest
	OpResponse
)

func (o OpCode) String() string {
	switch o {
	case OpUnkown:
		return "unknown"

	case OpRequest:
		return "Request"

	case OpResponse:
		return "Response"

	default:
		return "illegal value"
	}
}

type OptionTag byte

type Options map[OptionTag][]byte

func ParseOptions(m Message) Options {
	opts := make(map[OptionTag][]byte)
	b := m.Options()
	buff := bytes.NewBuffer(b)
	for optTag := OptionTag(buff.Next(1)[0]); buff.Len() > 0; {
		switch optTag {
		case OptionPad: //Do nothing
		case OptionEndField:
			buff.Reset() //drain
		case OptionSubnetMask, OptionTimeOffset: //these 2 are always 4 bytes in size, but their length is included in the packet
			opts[optTag] = buff.Next(5)[1:] //read 5 bytes and cut the first 1, which is always 0x04
		default:
			size := int(buff.Next(1)[0])
			opts[optTag] = buff.Next(size)
			log.Debug("set opt tag ", optTag)
		}
		if buff.Len() > 0 {
			optTag = OptionTag(buff.Next(1)[0])
		}
	}

	return opts
}

const (
	OptionPad                       OptionTag = iota //0
	OptionSubnetMask                                 //1
	OptionTimeOffset                                 //2
	OptionRouter                                     //3
	OptionTimeServer                                 //4
	OptionNameServer                                 //5
	OptionDomainNameServer                           //6
	OptionLogServer                                  //7
	OptionQuoteServer                                //8
	OptionLPRServer                                  //9
	OptionImpressServer                              //10
	OptionRLPServer                                  //11
	OptionHostname                                   //12
	OptionBootFileSize                               //13
	OptionMeritDumpFile                              //14
	OptionDomainName                                 //15
	OptionSwapServer                                 //16
	OptionRootPath                                   //17
	OptionExtensionsPath                             //18
	OptionIPForwarding                               //19
	OptionNonLocalSourceRouting                      //20
	OptionPolicyFilter                               //21
	OptionMaxDatagramReassemblySize                  //22
	OptionIPTTL                                      //23

	OptionNetbiosNameServer           = 44
	OptionEndField          OptionTag = 255
)

//DHCP Extensions
const (
	OptionRequestedIPAddress   OptionTag = iota + 50 //50
	OptionIPLeaseTime                                //51
	OptionOverload                                   //52
	OptionDHCPMessageType                            //53
	OptionServerIdentifier                           //54
	OptionParameterRequestList                       //55
	OptionMessage                                    //56
	OptionMaxDHCPMessageSize                         //57
	OptionRenewalTimeValue                           //58
	OptionRebindingTimeValue                         //59
	OptionClassIdentifier                            //60
	OptionClientIdentifier                           //61
)

//DHCP Message Type

type DHCPMessageType int

const (
	DHCPDiscover DHCPMessageType = iota + 1
	DHCPOffer
	DHCPRequest
	DHCPDecline
	DHCPAck
	DHCPNack
	DHCPRelease
)

func (o OptionTag) String() string {
	switch o {
	case OptionSubnetMask:
		return "SubnetMask"
	case OptionTimeOffset:
		return "TimeOffset"
	case OptionRouter:
		return "Router"
	case OptionTimeServer:
		return "TimeServer"
	case OptionNameServer:
		return "NameServer"
	case OptionDomainNameServer:
		return "DomainNameServer"
	case OptionLogServer:
		return "LogServer"
	case OptionQuoteServer:
		return "QuoteServer"
	case OptionLPRServer:
		return "LPRServer"
	case OptionImpressServer:
		return "ImpressServer"
	case OptionRLPServer:
		return "RLPServer"
	case OptionHostname:
		return "Hostname"
	case OptionBootFileSize:
		return "BootFileSize"
	case OptionMeritDumpFile:
		return "MeritDumpFile"
	case OptionDomainName:
		return "DomainName"
	case OptionSwapServer:
		return "SwapServer"
	case OptionRootPath:
		return "RootPath"
	case OptionExtensionsPath:
		return "ExtensionsPath"
	case OptionIPForwarding:
		return "IPForwarding"
	case OptionNonLocalSourceRouting:
		return "NonLocalSourceRouting"
	case OptionPolicyFilter:
		return "PolicyFilter"
	case OptionMaxDatagramReassemblySize:
		return "MaxDatagramReassemblySize"
	case OptionIPTTL:
		return "IPTTL"
	case OptionEndField:
		return "EndField"
	case OptionRequestedIPAddress:
		return "RequestedIPAddress"
	case OptionIPLeaseTime:
		return "IPLeaseTime"
	case OptionOverload:
		return "Overload"
	case OptionDHCPMessageType:
		return "DHCPMessageType"
	case OptionServerIdentifier:
		return "ServerIdentifier"
	case OptionParameterRequestList:
		return "ParameterRequestList"
	case OptionMessage:
		return "Message"
	case OptionMaxDHCPMessageSize:
		return "MaxDHCPMessageSize"
	case OptionRenewalTimeValue:
		return "RenewalTimeValue"
	case OptionRebindingTimeValue:
		return "RebindingTimeValue"
	case OptionClassIdentifier:
		return "ClassIdentifier"
	case OptionClientIdentifier:
		return "ClientIdentifier"
	default:
		return "unknown"
	}
}

var DHCPCookie = []byte{99, 130, 83, 99}

// The entire packet needs to be 300 bytes when sent over UDP
var zeroes [300]byte

//DHCP Message
type Message []byte

func NewMessage(opCode OpCode) Message {
	m := make(Message, 300)
	m.SetOpCode(opCode)
	m.SetHType(1) //ethernet
	m.SetCookie(DHCPCookie)
	return m
}

func (m Message) OpCode() OpCode {
	return OpCode(m[0])
}

func (m Message) SetOpCode(opCode OpCode) {
	m[0] = byte(opCode)
}

func (m Message) HType() byte {
	return m[1]
}

func (m Message) SetHType(htype byte) {
	m[1] = htype
}

func (m Message) HLen() byte {
	if r := m[2]; r > 16 {
		return 16
	}
	return m[2]
}

func (m Message) SetHLen(size int) {
	if size > 16 {
		size = 16
	}
	m[2] = byte(size)
}

func (m Message) Hops() byte {
	return m[3]
}

func (m Message) XId() []byte {
	return m[4:8]
}

func (m Message) SetXId(b []byte) {
	copy(m[4:8], b)
}

func (m Message) Secs() []byte {
	return m[8:10]
}

func (m Message) Flags() []byte {
	return m[10:12]
}

func (m Message) SetFlags(b []byte) {
	copy(m[8:10], b)
}

func (m Message) CIAddr() net.IP {
	return net.IP(m[12:16])
}

func (m Message) YIAddr() net.IP {
	return net.IP(m[16:20])
}

func (m Message) SetYIAddr(a net.IP) {
	copy(m[16:20], a)
}

func (m Message) SIAddr() net.IP {
	return net.IP(m[20:24])
}

func (m Message) GIAddr() net.IP {
	return net.IP(m[24:28])
}

func (m Message) SetGIAddr(a net.IP) {
	copy(m[24:28], a)
}

func (m Message) CHAddr() net.HardwareAddr {
	hlen := m.HLen()
	return net.HardwareAddr(m[28 : 28+hlen])
}

func (m Message) SetCHAddr(a net.HardwareAddr) {
	copy(m[28:28+len(a)], a)
	m.SetHLen(len([]byte(a)))
}

func (m Message) Sname() string {
	return string(m[44:108])
}

func (m Message) File() []byte {
	return m[108:236]
}

func (m Message) Cookie() []byte {
	return m[236:240]
}

func (m Message) SetCookie(cookie []byte) {
	copy(m[236:240], cookie)
}

func (m Message) Options() []byte {
	return m[240:]
}

func (m Message) SetOptions(options Options) {
	b := make([]byte, 0)
	for tag, val := range options {
		b = append(b, byte(tag))
		b = append(b, byte(len(val)))
		b = append(b, val...)
	}
	b = append(b, byte(OptionEndField))
}

func (m *Message) FillPadding() {
	if l := len(*m); l < 272 {
		*m = append(*m, zeroes[:272-l]...)
	}
}

func DHCPReply(request Message, serverAddr, yiaddr net.IP, leaseDuration time.Duration, options Options) Message {
	m := NewMessage(OpResponse)
	m.SetXId(request.XId())
	m.SetFlags(request.Flags())
	m.SetYIAddr(yiaddr)
	m.SetGIAddr(request.GIAddr())
	m.SetCHAddr(request.CHAddr())
	if leaseDuration > 0 {
		options[OptionIPLeaseTime] = make([]byte, 4)
		binary.BigEndian.PutUint32(options[OptionIPLeaseTime], uint32(leaseDuration/time.Second))
	}
	m.SetOptions(options)
	b := make([]byte, 0)
	for tag, val := range options {
		b = append(b, byte(tag))
		b = append(b, byte(len(val)))
		b = append(b, val...)
	}
	b = append(b, byte(OptionEndField))

	copy(m[240:], b)
	return m
}
