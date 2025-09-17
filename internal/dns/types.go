package dns

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

var (
	ErrDNSPacketTooShort = errors.New("DNS packet too short for header")
)

type DNSType uint16

const (
	DNSTypeA     DNSType = 1
	DNSTypeAAAA  DNSType = 28
	DNSTypeCNAME DNSType = 5
	DNSTypeMX    DNSType = 15
	DNSTypeNS    DNSType = 2
	DNSTypeTXT   DNSType = 16
	DNSTypeOPT   DNSType = 41
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
	case DNSTypeOPT:
		return "OPT"
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

func (m *DNSMessage) String() string {
	return fmt.Sprintf("DNS Message: ID=%d, QR=%d, Opcode=%s, AA=%d, TC=%d, RD=%d, RA=%d, Z=%d, RCODE=%d", m.Header.ID, m.Header.QR, m.Header.Opcode, m.Header.AA, m.Header.TC, m.Header.RD, m.Header.RA, m.Header.Z, m.Header.RCODE)
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
	*DNSMessage
	ResponseAddr net.Addr
}

func ParseDNSMessage(data []byte) (*DNSMessage, error) {
	if len(data) < 12 {
		return nil, ErrDNSPacketTooShort
	}

	var dnsMessage DNSMessage
	offset := 0

	// Parse header (12 bytes total)
	dnsMessage.Header.ID = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	// Parse flags (2 bytes combined)
	flags := binary.BigEndian.Uint16(data[offset : offset+2])
	dnsMessage.Header.QR = uint8((flags >> 15) & 0x1)
	dnsMessage.Header.Opcode = DNSOpCode((flags >> 11) & 0xF)
	dnsMessage.Header.AA = uint8((flags >> 10) & 0x1)
	dnsMessage.Header.TC = uint8((flags >> 9) & 0x1)
	dnsMessage.Header.RD = uint8((flags >> 8) & 0x1)
	dnsMessage.Header.RA = uint8((flags >> 7) & 0x1)
	dnsMessage.Header.Z = uint8((flags >> 4) & 0x7)
	dnsMessage.Header.RCODE = uint8(flags & 0xF)
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
	dnsMessage.Questions = make([]DNSQuestion, 0, qdCount)
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

		dnsMessage.Questions = append(dnsMessage.Questions, question)
	}

	// Parse answers
	dnsMessage.Answers = make([]DNSRecord, 0, anCount)
	for i := 0; i < int(anCount); i++ {
		record, newOffset, err := parseResourceRecord(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing answer record: %v", err)
		}
		dnsMessage.Answers = append(dnsMessage.Answers, record)
		offset = newOffset
	}

	// Parse authorities
	dnsMessage.Authorities = make([]DNSRecord, 0, nsCount)
	for i := 0; i < int(nsCount); i++ {
		record, newOffset, err := parseResourceRecord(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing authority record: %v", err)
		}
		dnsMessage.Authorities = append(dnsMessage.Authorities, record)
		offset = newOffset
	}

	// Parse additionals
	dnsMessage.Additionals = make([]DNSRecord, 0, arCount)
	for i := 0; i < int(arCount); i++ {
		record, newOffset, err := parseResourceRecord(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing additional record: %v", err)
		}
		dnsMessage.Additionals = append(dnsMessage.Additionals, record)
		offset = newOffset
	}

	return &dnsMessage, nil
}

// parseDNSName parses a DNS name from the packet, handling compression
func parseDNSName(data []byte, offset int) (string, int, error) {
	var name string
	originalOffset := offset
	jumped := false
	jumps := 0
	maxJumps := 5                 // Prevent infinite loops
	visited := make(map[int]bool) // Track visited positions to prevent loops

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

			// Validate pointer is within packet bounds
			if pointer >= len(data) {
				return "", offset, fmt.Errorf("compression pointer %d exceeds packet length %d", pointer, len(data))
			}

			// Prevent pointing to locations that would cause loops
			if visited[pointer] {
				return "", offset, errors.New("compression pointer loop detected")
			}
			visited[pointer] = true

			// Ensure we're not pointing to another compression pointer immediately
			// (this would be valid but we need to be careful about infinite loops)
			if pointer == offset {
				return "", offset, errors.New("compression pointer points to itself")
			}

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

		// Validate label length is reasonable (DNS labels max 63 bytes)
		if length > 63 {
			return "", offset, fmt.Errorf("invalid label length %d (max 63)", length)
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

		// Prevent excessively long domain names
		if len(name) > 253 { // DNS name max length
			return "", offset, errors.New("domain name too long")
		}
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
	startOffset := offset

	// Parse name
	record.Name, offset, err = parseDNSName(data, offset)
	if err != nil {
		return record, offset, fmt.Errorf("error parsing RR name at offset %d: %v", startOffset, err)
	}

	// Check bounds for type, class, TTL, and data length
	if offset+10 > len(data) {
		return record, offset, fmt.Errorf("insufficient data for resource record at offset %d (need %d, have %d)", offset, offset+10, len(data))
	}

	// Parse type
	record.Type = DNSType(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// Special handling for EDNS OPT records (RFC 6891)
	if record.Type == DNSType(41) { // OPT record
		// For OPT records, the "class" field is the requestor's UDP payload size
		udpPayloadSize := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2

		// The TTL field contains extended RCODE, version, and flags
		extendedInfo := binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4

		// Store the UDP payload size in the Class field for OPT records
		record.Class = udpPayloadSize
		record.TTL = extendedInfo

		// Parse data length and data normally
		dataLength := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2

		if offset+int(dataLength) > len(data) {
			return record, offset, fmt.Errorf("OPT record data extends beyond packet at offset %d (need %d, have %d, rdlength %d)", offset, offset+int(dataLength), len(data), dataLength)
		}

		record.RData = make([]byte, dataLength)
		copy(record.RData, data[offset:offset+int(dataLength)])
		offset += int(dataLength)

		return record, offset, nil
	}

	// Regular resource record parsing
	record.Class = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	record.TTL = binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse data length and data
	dataLength := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	if offset+int(dataLength) > len(data) {
		return record, offset, fmt.Errorf("resource record data extends beyond packet at offset %d (need %d, have %d, rdlength %d)", offset, offset+int(dataLength), len(data), dataLength)
	}

	record.RData = make([]byte, dataLength)
	copy(record.RData, data[offset:offset+int(dataLength)])
	offset += int(dataLength)

	return record, offset, nil
}
