package dns

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

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

type DNSOpCode uint8

const (
	DNSOpCodeQuery  DNSOpCode = 0 // standard query
	DNSOpCodeIQuery DNSOpCode = 1 // inverse query
	DNSOpCodeStatus DNSOpCode = 2 // server status request
)

func (o DNSOpCode) String() string {
	switch o {
	case DNSOpCodeQuery:
		return "Query"
	case DNSOpCodeIQuery:
		return "IQuery"
	case DNSOpCodeStatus:
		return "Status"
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

type DNSMessage struct {
	Header      DNSHeader
	Questions   []DNSQuestion
	Answers     []DNSRecord
	Authorities []DNSRecord
	Additionals []DNSRecord
}

type DNSHeader struct {
	ID     uint16    // Identification: Identifier for the message
	QR     uint8     // Query/Response: 0 for a query, 1 for a response
	Opcode DNSOpCode // Operation Code: See DnsOpCode for more
	AA     uint8     // Authoritative Answer: This server is an authority for the domain and can return answers
	TC     uint8     // Truncated: Message was truncated
	RD     uint8     // Recursion Desired: This message is a query with the recursion desired bit set
	RA     uint8     // Recursion Available: Message is a response to a query with the recursion available bit set
	Z      uint8     // Reserved: Reserved for future use
	RCODE  uint8     // Response Code: Response code
}

type DNSQuestion struct {
	Name  string
	Type  DNSType
	Class uint16
}

type DNSPacket struct {
	DNSMessage
	ResponseAddr net.IP
}

func ParseDNSPacket(data []byte) (*DNSPacket, error) {
	if len(data) < 12 {
		return nil, errors.New("DNS packet too short for header")
	}

	var dnsPacket DNSPacket
	offset := 0

	// Parse header (12 bytes total)
	dnsPacket.Header.ID = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	// Parse flags (2 bytes combined)
	flags := binary.BigEndian.Uint16(data[offset : offset+2])
	dnsPacket.Header.QR = uint8((flags >> 15) & 0x1)
	dnsPacket.Header.Opcode = DNSOpCode((flags >> 11) & 0xF)
	dnsPacket.Header.AA = uint8((flags >> 10) & 0x1)
	dnsPacket.Header.TC = uint8((flags >> 9) & 0x1)
	dnsPacket.Header.RD = uint8((flags >> 8) & 0x1)
	dnsPacket.Header.RA = uint8((flags >> 7) & 0x1)
	dnsPacket.Header.Z = uint8((flags >> 4) & 0x7)
	dnsPacket.Header.RCODE = uint8(flags & 0xF)
	offset += 2

	// Parse counts
	qdCount := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	anCount := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	nsCount := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	arCount := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	// Parse questions
	dnsPacket.Questions = make([]DNSQuestion, 0, qdCount)
	for i := 0; i < int(qdCount); i++ {
		var question DNSQuestion
		var err error

		question.Name, offset, err = parseDNSName(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing question name: %v", err)
		}

		if offset+4 > len(data) {
			return nil, errors.New("insufficient data for question type and class")
		}

		question.Type = DNSType(binary.BigEndian.Uint16(data[offset : offset+2]))
		offset += 2
		question.Class = binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2

		dnsPacket.Questions = append(dnsPacket.Questions, question)
	}

	// Parse answers
	dnsPacket.Answers = make([]DNSRecord, 0, anCount)
	for i := 0; i < int(anCount); i++ {
		record, newOffset, err := parseResourceRecord(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing answer record: %v", err)
		}
		dnsPacket.Answers = append(dnsPacket.Answers, record)
		offset = newOffset
	}

	// Parse authorities
	dnsPacket.Authorities = make([]DNSRecord, 0, nsCount)
	for i := 0; i < int(nsCount); i++ {
		record, newOffset, err := parseResourceRecord(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing authority record: %v", err)
		}
		dnsPacket.Authorities = append(dnsPacket.Authorities, record)
		offset = newOffset
	}

	// Parse additionals
	dnsPacket.Additionals = make([]DNSRecord, 0, arCount)
	for i := 0; i < int(arCount); i++ {
		record, newOffset, err := parseResourceRecord(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing additional record: %v", err)
		}
		dnsPacket.Additionals = append(dnsPacket.Additionals, record)
		offset = newOffset
	}

	return &dnsPacket, nil
}

// parseDNSName parses a DNS name from the packet, handling compression
func parseDNSName(data []byte, offset int) (string, int, error) {
	var name string
	originalOffset := offset
	jumped := false
	jumps := 0
	maxJumps := 5 // Prevent infinite loops

	for {
		if offset >= len(data) {
			return "", offset, errors.New("unexpected end of data while parsing name")
		}

		length := data[offset]

		// Check for compression pointer (top 2 bits set)
		if (length & 0xC0) == 0xC0 {
			if offset+1 >= len(data) {
				return "", offset, errors.New("incomplete compression pointer")
			}
			// Extract 14-bit offset
			pointer := int(binary.BigEndian.Uint16(data[offset:offset+2]) & 0x3FFF)
			if !jumped {
				originalOffset = offset + 2 // Save position after pointer
			}
			offset = pointer
			jumped = true
			jumps++
			if jumps > maxJumps {
				return "", offset, errors.New("too many compression jumps")
			}
			continue
		}

		offset++

		// End of name
		if length == 0 {
			break
		}

		// Check bounds for label
		if offset+int(length) > len(data) {
			return "", offset, errors.New("label extends beyond packet")
		}

		// Add dot separator if not first label
		if name != "" {
			name += "."
		}

		// Extract label
		name += string(data[offset : offset+int(length)])
		offset += int(length)
	}

	// If we jumped, return the saved offset
	if jumped {
		return name, originalOffset, nil
	}

	return name, offset, nil
}

// parseResourceRecord parses a resource record from the packet
func parseResourceRecord(data []byte, offset int) (DNSRecord, int, error) {
	var record DNSRecord
	var err error

	// Parse name
	record.Name, offset, err = parseDNSName(data, offset)
	if err != nil {
		return record, offset, err
	}

	// Check bounds for type, class, TTL, and data length
	if offset+10 > len(data) {
		return record, offset, errors.New("insufficient data for resource record")
	}

	// Parse type, class, TTL
	record.Type = DNSType(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2
	record.Class = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	record.TTL = binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse data length and data
	dataLength := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	if offset+int(dataLength) > len(data) {
		return record, offset, errors.New("resource record data extends beyond packet")
	}

	record.RData = make([]byte, dataLength)
	copy(record.RData, data[offset:offset+int(dataLength)])
	offset += int(dataLength)

	return record, offset, nil
}
