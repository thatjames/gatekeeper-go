package dns

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

var (
	ErrDNSPacketTooShort   = errors.New("DNS packet too short for header")
	ErrDNSTooManyQuestions = errors.New("DNS packet contains more than 1 question")
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
	Name  []byte
	Type  DNSType
	Class uint16
	TTL   uint32
	RData []byte
}

type DNSMessage struct {
	Header      *DNSHeader
	Questions   []*DNSQuestion
	Answers     []*DNSRecord
	Authorities []*DNSRecord
	Additionals []*DNSRecord
}

type DNSHeader struct {
	ID    uint16
	Flags uint16
	/*
			Bit: 15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
		         QR |   Opcode  |AA|TC|RD|RA|   Z    |   RCODE   |
	*/
}

func (h DNSHeader) SetQR(qr bool) { //where true is a response
	if qr {
		h.Flags |= 0x8000
	} else {
		h.Flags &= 0x7FFF
	}
}

func (h DNSHeader) SetOpcode(opcode DNSOpCode) {
	h.Flags = (h.Flags & 0x87FF) | (uint16(opcode) << 11)
}

func (h DNSHeader) SetAA(aa bool) {
	if aa {
		h.Flags |= 0x0400
	} else {
		h.Flags &= 0xFBFF
	}
}

func (h DNSHeader) SetTC(tc bool) {
	if tc {
		h.Flags |= 0x0200
	} else {
		h.Flags &= 0xFDFF
	}
}

func (h DNSHeader) SetRD(rd bool) {
	if rd {
		h.Flags |= 0x0100
	} else {
		h.Flags &= 0xFEFF
	}
}

func (h DNSHeader) SetRA(ra bool) {
	if ra {
		h.Flags |= 0x0080
	} else {
		h.Flags &= 0xFF7F
	}
}

func (h DNSHeader) SetZ(z uint8) {
	h.Flags = (h.Flags & 0xFF8F) | (uint16(z) << 4)
}

func (h DNSHeader) SetRCODE(rcode RCODE) {
	h.Flags = (h.Flags & 0xFFF0) | uint16(rcode)
}

type RCODE uint8

const (
	RCODESuccess       RCODE = 0
	RCODEFormatError   RCODE = 1
	RCODEServerFailure RCODE = 2
	RCODENameFailure   RCODE = 3
)

type DNSQuestion struct {
	Name  []byte
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
	dnsMessage := &DNSMessage{
		Header:      &DNSHeader{},
		Questions:   make([]*DNSQuestion, 0),
		Answers:     make([]*DNSRecord, 0),
		Authorities: make([]*DNSRecord, 0),
		Additionals: make([]*DNSRecord, 0),
	}
	offset := 0

	// Parse header (12 bytes total)
	dnsMessage.Header.ID = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	// Parse flags (2 bytes combined) - store the raw flags
	dnsMessage.Header.Flags = binary.BigEndian.Uint16(data[offset : offset+2])
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

	// 99% of real world traffic only uses 1 question, so we will FORMERR if we see more than 1
	if qdCount > 1 {
		return nil, ErrDNSTooManyQuestions
	}

	// Parse questions
	dnsMessage.Questions = make([]*DNSQuestion, 0, qdCount)
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
		dnsMessage.Questions = append(dnsMessage.Questions, &question)
	}

	// Parse answers
	dnsMessage.Answers = make([]*DNSRecord, 0, anCount)
	for i := 0; i < int(anCount); i++ {
		record, newOffset, err := parseResourceRecord(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing answer record: %v", err)
		}
		dnsMessage.Answers = append(dnsMessage.Answers, &record)
		offset = newOffset
	}

	// Parse authorities
	dnsMessage.Authorities = make([]*DNSRecord, 0, nsCount)
	for i := 0; i < int(nsCount); i++ {
		record, newOffset, err := parseResourceRecord(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing authority record: %v", err)
		}
		dnsMessage.Authorities = append(dnsMessage.Authorities, &record)
		offset = newOffset
	}

	// Parse additionals
	dnsMessage.Additionals = make([]*DNSRecord, 0, arCount)
	for i := 0; i < int(arCount); i++ {
		record, newOffset, err := parseResourceRecord(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing additional record: %v", err)
		}
		dnsMessage.Additionals = append(dnsMessage.Additionals, &record)
		offset = newOffset
	}

	return dnsMessage, nil
}

// parseDNSName parses a DNS name from the packet, handling compression
func parseDNSName(data []byte, offset int) ([]byte, int, error) {
	var name string
	originalOffset := offset
	jumped := false
	jumps := 0
	maxJumps := 5                 // Prevent infinite loops
	visited := make(map[int]bool) // Track visited positions to prevent loops

	for {
		if offset >= len(data) {
			return nil, offset, errors.New("unexpected end of data while parsing name")
		}

		length := data[offset]

		// Check for compression pointer (top 2 bits set)
		if (length & 0xC0) == 0xC0 {
			if offset+1 >= len(data) {
				return nil, offset, errors.New("incomplete compression pointer")
			}

			// Extract 14-bit offset
			pointer := int(binary.BigEndian.Uint16(data[offset:offset+2]) & 0x3FFF)

			// Validate pointer is within packet bounds
			if pointer >= len(data) {
				return nil, offset, fmt.Errorf("compression pointer %d exceeds packet length %d", pointer, len(data))
			}

			// Prevent pointing to locations that would cause loops
			if visited[pointer] {
				return nil, offset, errors.New("compression pointer loop detected")
			}
			visited[pointer] = true

			// Ensure we're not pointing to another compression pointer immediately
			// (this would be valid but we need to be careful about infinite loops)
			if pointer == offset {
				return nil, offset, errors.New("compression pointer points to itself")
			}

			if !jumped {
				originalOffset = offset + 2 // Save position after pointer
			}
			offset = pointer
			jumped = true
			jumps++

			if jumps > maxJumps {
				return nil, offset, errors.New("too many compression jumps")
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
			return nil, offset, fmt.Errorf("invalid label length %d (max 63)", length)
		}

		// Check bounds for label
		if offset+int(length) > len(data) {
			return nil, offset, errors.New("label extends beyond packet")
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
			return nil, offset, errors.New("domain name too long")
		}
	}

	// If we jumped, return the saved offset
	if jumped {
		return []byte(name), originalOffset, nil
	}

	return []byte(name), offset, nil
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

func MarshalDNSMessage(msg *DNSMessage) ([]byte, error) {
	var buf []byte

	// Header (12 bytes)
	header := make([]byte, 12)
	binary.BigEndian.PutUint16(header[0:2], msg.Header.ID)
	binary.BigEndian.PutUint16(header[2:4], msg.Header.Flags)
	binary.BigEndian.PutUint16(header[4:6], uint16(len(msg.Questions)))
	binary.BigEndian.PutUint16(header[6:8], uint16(len(msg.Answers)))
	binary.BigEndian.PutUint16(header[8:10], uint16(len(msg.Authorities)))
	binary.BigEndian.PutUint16(header[10:12], uint16(len(msg.Additionals)))
	buf = append(buf, header...)

	// Questions
	for _, q := range msg.Questions {
		buf = append(buf, q.Name...)

		// Type and Class (4 bytes)
		typeClass := make([]byte, 4)
		binary.BigEndian.PutUint16(typeClass[0:2], uint16(q.Type))
		binary.BigEndian.PutUint16(typeClass[2:4], q.Class)
		buf = append(buf, typeClass...)
	}

	// Answers
	for _, r := range msg.Answers {
		recordBytes, err := marshalResourceRecord(r)
		if err != nil {
			return nil, fmt.Errorf("error marshaling answer record: %v", err)
		}
		buf = append(buf, recordBytes...)
	}

	// Authorities
	for _, r := range msg.Authorities {
		recordBytes, err := marshalResourceRecord(r)
		if err != nil {
			return nil, fmt.Errorf("error marshaling authority record: %v", err)
		}
		buf = append(buf, recordBytes...)
	}

	// Additionals
	for _, r := range msg.Additionals {
		recordBytes, err := marshalResourceRecord(r)
		if err != nil {
			return nil, fmt.Errorf("error marshaling additional record: %v", err)
		}
		buf = append(buf, recordBytes...)
	}

	return buf, nil
}

// marshalResourceRecord converts a DNSRecord to bytes
func marshalResourceRecord(record *DNSRecord) ([]byte, error) {
	var buf []byte

	// Marshal name
	buf = append(buf, record.Name...)

	// Type, Class, TTL, Data Length (10 bytes)
	rrHeader := make([]byte, 10)
	binary.BigEndian.PutUint16(rrHeader[0:2], uint16(record.Type))
	binary.BigEndian.PutUint16(rrHeader[2:4], record.Class)
	binary.BigEndian.PutUint32(rrHeader[4:8], record.TTL)
	binary.BigEndian.PutUint16(rrHeader[8:10], uint16(len(record.RData)))
	buf = append(buf, rrHeader...)

	// RData
	buf = append(buf, record.RData...)

	return buf, nil
}
