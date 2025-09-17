package dns

type DNSPacket struct {
	Name  string
	Type  DNSType
	Class uint16
	TTL   uint32
	RData []byte
}

type DNSType uint16

const (
	DNSTypeA     DNSType = 1
	DNSTypeAAAA  DNSType = 28
	DNSTypeCNAME DNSType = 5
	DNSTypeMX    DNSType = 15
	DNSTypeNS    DNSType = 2
	DNSTypeTXT   DNSType = 16
)

func (t DNSType) String() string {
	switch t {
	case DNSTypeA:
		return "A"
	case DNSTypeAAAA:
		return "AAAA"
	case DNSTypeCNAME:
		return "CNAME"
	case DNSTypeMX:
		return "MX"
	case DNSTypeNS:
		return "NS"
	case DNSTypeTXT:
		return "TXT"
	default:
		return "unknown"
	}
}

type DNSRecord struct {
	Name  string
	Type  DNSType
	Class uint16
	TTL   uint32
	RData []byte
}

func ParseDNSPacket(data []byte) (*DNSPacket, error) {
	return nil, nil
}
