package dhcp

import (
	"bytes"
	"fmt"
	"net"

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
	fmt.Println(b)
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

	OptionEndField OptionTag = 255
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
const (
	DHCPDiscover = iota + 1
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

//DHCP Message
type Message []byte

func (m Message) OpCode() OpCode {
	return OpCode(m[0])
}

func (m Message) HType() byte {
	return m[1]
}

func (m Message) HLen() byte {
	if r := m[2]; r > 16 {
		return 16
	}
	return m[2]
}

func (m Message) Hops() byte {
	return m[3]
}

func (m Message) XId() []byte {
	return m[4:8]
}

func (m Message) Secs() []byte {
	return m[8:10]
}

func (m Message) Flags() []byte {
	return m[10:12]
}

func (m Message) CIAddr() net.IP {
	return net.IP(m[12:16])
}

func (m Message) YIAddr() net.IP {
	return net.IP(m[16:20])
}

func (m Message) SIAddr() net.IP {
	return net.IP(m[20:24])
}

func (m Message) GIAddr() net.IP {
	return net.IP(m[24:28])
}

func (m Message) CHAddr() net.HardwareAddr {
	hlen := m.HLen()
	return net.HardwareAddr(m[28 : 28+hlen])
}

func (m Message) Sname() string {
	return string(m[44:108])
}

func (m Message) File() []byte {
	return m[108:236]
}

func (m Message) Options() []byte {
	return m[240:]
}
