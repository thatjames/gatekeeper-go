package dns

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

var (
	reqDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "dns_req_time",
		Help:    "dns request time in ms buckets",
		Buckets: []float64{1, 10, 100, 250, 500, 1000, 2500, 5000, 10000},
	})

	queryCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "dns_query_counter",
		Help: "Count by domain",
	}, []string{"domain", "upstream", "result"})

	cacheHitCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "dns_cache_hit_counter",
		Help: "count of cache hits by domain",
	}, []string{"domain"})

	blockedDomainCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "dns_blocked_domain_counter",
		Help: "count of blocked domains",
	}, []string{"domain"})

	queryByIPCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "dns_query_by_ip_counter",
		Help: "count of queries by IP",
	}, []string{"ip", "result"})
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

// This is not a type because of OPT records
const (
	DNSClassIN   uint16 = 1
	DNSClassCH   uint16 = 3
	DNSClassHS   uint16 = 4
	DNSClassNone uint16 = 254
	DNSClassANY  uint16 = 255
)

type DNSRecord struct {
	ParsedName  string
	Name        []byte
	Type        DNSType
	Class       uint16
	TTL         uint32
	RData       []byte
	ParsedRData string
}

func (r *DNSRecord) String() string {
	return fmt.Sprintf("%s %s %d %d", r.ParsedName, r.Type, r.Class, r.TTL)
}

type DNSMessage struct {
	Header      *DNSHeader
	Questions   []*DNSQuestion
	Answers     []*DNSRecord
	Authorities []*DNSRecord
	Additionals []*DNSRecord
}

func NewDnsMessage() *DNSMessage {
	return &DNSMessage{
		Header:      &DNSHeader{},
		Questions:   make([]*DNSQuestion, 0),
		Answers:     make([]*DNSRecord, 0),
		Authorities: make([]*DNSRecord, 0),
		Additionals: make([]*DNSRecord, 0),
	}
}

type DNSHeader struct {
	ID    uint16
	Flags uint16
	/*
			Bit: 15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
		         QR |   Opcode  |AA|TC|RD|RA|   Z    |   RCODE   |
	*/
}

func (h *DNSHeader) SetQR(qr bool) { //where true is a response
	if qr {
		h.Flags |= 0x8000
	} else {
		h.Flags &= 0x7FFF
	}
}

func (h *DNSHeader) SetOpcode(opcode DNSOpCode) {
	h.Flags = (h.Flags & 0x87FF) | (uint16(opcode) << 11)
}

func (h *DNSHeader) SetAA(aa bool) {
	if aa {
		h.Flags |= 0x0400
	} else {
		h.Flags &= 0xFBFF
	}
}

func (h *DNSHeader) SetTC(tc bool) {
	if tc {
		h.Flags |= 0x0200
	} else {
		h.Flags &= 0xFDFF
	}
}

func (h *DNSHeader) SetRD(rd bool) {
	if rd {
		h.Flags |= 0x0100
	} else {
		h.Flags &= 0xFEFF
	}
}

func (h *DNSHeader) SetRA(ra bool) {
	if ra {
		h.Flags |= 0x0080
	} else {
		h.Flags &= 0xFF7F
	}
}

func (h *DNSHeader) SetZ(z uint8) {
	h.Flags = (h.Flags & 0xFF8F) | (uint16(z) << 4)
}

func (h *DNSHeader) SetRCODE(rcode RCODE) {
	h.Flags = (h.Flags & 0xFFF0) | uint16(rcode)
}

func (h *DNSHeader) RCODE() RCODE {
	return RCODE(h.Flags & 0x000F)
}

type RCODE uint8

const (
	RCODESuccess       RCODE = 0
	RCODEFormatError   RCODE = 1
	RCODEServerFailure RCODE = 2
	RCODENameFailure   RCODE = 3
)

type DNSQuestion struct {
	Name       []byte
	ParsedName string
	Type       DNSType
	Class      uint16
}

type DNSPacket struct {
	*DNSMessage
	ResponseAddr net.Addr
}

func stringToDNSWireFormat(domain string) []byte {
	if domain == "" || domain == "." {
		return []byte{0}
	}

	var result []byte
	if strings.HasSuffix(domain, ".") {
		domain = domain[:len(domain)-1]
	}

	labels := strings.Split(domain, ".")

	for _, label := range labels {
		if len(label) == 0 {
			continue
		}
		if len(label) > 63 {
			label = label[:63]
		}
		result = append(result, byte(len(label)))
		result = append(result, []byte(label)...)
	}
	result = append(result, 0)
	return result
}

func parseDNSNameWithWireFormat(data []byte, offset int) (string, []byte, int, error) {
	startOffset := offset
	parsedName, newOffset, err := parseDNSName(data, offset)
	if err != nil {
		return "", nil, offset, err
	}

	var wireFormat []byte

	if startOffset < len(data) && data[startOffset] == 0 {
		wireFormat = []byte{0}
		return parsedName, wireFormat, newOffset, nil
	}

	currentPos := startOffset
	for currentPos < len(data) {
		length := data[currentPos]

		if (length & 0xC0) == 0xC0 {
			wireFormat = append(wireFormat, data[currentPos:currentPos+2]...)
			break
		}

		wireFormat = append(wireFormat, length)
		currentPos++

		if length == 0 {
			break
		}

		if currentPos+int(length) <= len(data) {
			wireFormat = append(wireFormat, data[currentPos:currentPos+int(length)]...)
			currentPos += int(length)
		}
	}

	return parsedName, wireFormat, newOffset, nil
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

	dnsMessage.Header.ID = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	dnsMessage.Header.Flags = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	qdCount := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	anCount := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	nsCount := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	arCount := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	log.Tracef("parsing %d questions, %d answers, %d authorities, %d additionals", qdCount, anCount, nsCount, arCount)

	// 99% of real world traffic only uses 1 question, so we will FORMERR if we see more than 1
	if qdCount > 1 {
		return nil, ErrDNSTooManyQuestions
	}

	dnsMessage.Questions = make([]*DNSQuestion, 0, qdCount)
	for i := 0; i < int(qdCount); i++ {
		var question DNSQuestion
		var err error
		question.ParsedName, question.Name, offset, err = parseDNSNameWithWireFormat(data, offset)
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

	dnsMessage.Answers = make([]*DNSRecord, 0, anCount)
	for i := 0; i < int(anCount); i++ {
		record, newOffset, err := parseResourceRecord(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing answer record: %v", err)
		}
		dnsMessage.Answers = append(dnsMessage.Answers, record)
		offset = newOffset
	}

	dnsMessage.Authorities = make([]*DNSRecord, 0, nsCount)
	for i := 0; i < int(nsCount); i++ {
		record, newOffset, err := parseResourceRecord(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing authority record: %v", err)
		}
		dnsMessage.Authorities = append(dnsMessage.Authorities, record)
		offset = newOffset
	}

	dnsMessage.Additionals = make([]*DNSRecord, 0, arCount)
	for i := 0; i < int(arCount); i++ {
		record, newOffset, err := parseResourceRecord(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing additional record: %v", err)
		}
		dnsMessage.Additionals = append(dnsMessage.Additionals, record)
		offset = newOffset
	}

	return dnsMessage, nil
}

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

		if (length & 0xC0) == 0xC0 {
			if offset+1 >= len(data) {
				return "", offset, errors.New("incomplete compression pointer")
			}

			pointer := int(binary.BigEndian.Uint16(data[offset:offset+2]) & 0x3FFF)

			if pointer >= len(data) {
				return "", offset, fmt.Errorf("compression pointer %d exceeds packet length %d", pointer, len(data))
			}

			if visited[pointer] {
				return "", offset, errors.New("compression pointer loop detected")
			}
			visited[pointer] = true

			if pointer == offset {
				return "", offset, errors.New("compression pointer points to itself")
			}

			if !jumped {
				originalOffset = offset + 2
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

		if length == 0 {
			break
		}

		if length > 63 {
			return "", offset, fmt.Errorf("invalid label length %d (max 63)", length)
		}

		if offset+int(length) > len(data) {
			return "", offset, errors.New("label extends beyond packet")
		}

		if len(name) > 0 {
			name += "."
		}

		name += string(data[offset : offset+int(length)])
		offset += int(length)

		if len(name) > 253 {
			return "", offset, errors.New("domain name too long")
		}
	}

	if jumped {
		return name, originalOffset, nil
	}

	return name, offset, nil
}

func parseResourceRecord(data []byte, offset int) (*DNSRecord, int, error) {
	var record DNSRecord
	var err error
	startOffset := offset

	record.ParsedName, record.Name, offset, err = parseDNSNameWithWireFormat(data, offset)
	if err != nil {
		return nil, offset, fmt.Errorf("error parsing RR name at offset %d: %v", startOffset, err)
	}

	if offset+10 > len(data) {
		return nil, offset, fmt.Errorf("insufficient data for resource record at offset %d (need %d, have %d)", offset, offset+10, len(data))
	}

	record.Type = DNSType(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2

	if record.Type == DNSTypeOPT {
		record.ParsedName = ""
		record.Name = nil

		udpPayloadSize := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2

		extendedInfo := binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4

		record.Class = udpPayloadSize
		record.TTL = extendedInfo

		dataLength := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2

		if offset+int(dataLength) > len(data) {
			return nil, offset, fmt.Errorf("OPT record data extends beyond packet at offset %d (need %d, have %d, rdlength %d)", offset, offset+int(dataLength), len(data), dataLength)
		}

		record.RData = make([]byte, dataLength)
		copy(record.RData, data[offset:offset+int(dataLength)])

		offset += int(dataLength)

		return &record, offset, nil
	}

	record.Class = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	record.TTL = binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	dataLength := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	if offset+int(dataLength) > len(data) {
		return nil, offset, fmt.Errorf("resource record data extends beyond packet at offset %d (need %d, have %d, rdlength %d)", offset, offset+int(dataLength), len(data), dataLength)
	}

	record.RData = make([]byte, dataLength)
	if record.Type == DNSTypeCNAME || record.Type == DNSTypeNS {
		parsedTarget, _, err := parseDNSName(data, offset)
		if err != nil {
			log.Debugf("unable to parse domain in RData for %s record: %v", record.Type, err)
		} else {
			record.ParsedRData = parsedTarget
		}
	}
	copy(record.RData, data[offset:offset+int(dataLength)])
	offset += int(dataLength)

	return &record, offset, nil
}

func MarshalDNSMessage(msg *DNSMessage) ([]byte, error) {
	var buf []byte

	header := make([]byte, 12)
	binary.BigEndian.PutUint16(header[0:2], msg.Header.ID)
	binary.BigEndian.PutUint16(header[2:4], msg.Header.Flags)
	binary.BigEndian.PutUint16(header[4:6], uint16(len(msg.Questions)))
	binary.BigEndian.PutUint16(header[6:8], uint16(len(msg.Answers)))
	binary.BigEndian.PutUint16(header[8:10], uint16(len(msg.Authorities)))
	binary.BigEndian.PutUint16(header[10:12], uint16(len(msg.Additionals)))
	buf = append(buf, header...)

	for _, q := range msg.Questions {
		buf = append(buf, q.Name...)

		typeClass := make([]byte, 4)
		binary.BigEndian.PutUint16(typeClass[0:2], uint16(q.Type))
		binary.BigEndian.PutUint16(typeClass[2:4], q.Class)
		buf = append(buf, typeClass...)
	}

	for _, r := range msg.Answers {
		recordBytes, err := marshalResourceRecord(r)
		if err != nil {
			return nil, fmt.Errorf("error marshaling answer record: %v", err)
		}
		buf = append(buf, recordBytes...)
	}

	for _, r := range msg.Authorities {
		log.Tracef("marshaling authority record %v", r)
		recordBytes, err := marshalResourceRecord(r)
		if err != nil {
			return nil, fmt.Errorf("error marshaling authority record: %v", err)
		}
		buf = append(buf, recordBytes...)
	}

	for _, r := range msg.Additionals {
		recordBytes, err := marshalResourceRecord(r)
		if err != nil {
			return nil, fmt.Errorf("error marshaling additional record: %v", err)
		}
		buf = append(buf, recordBytes...)
	}

	return buf, nil
}

func marshalResourceRecord(record *DNSRecord) ([]byte, error) {
	var buf []byte
	if record.Name != nil {
		buf = append(buf, record.Name...)
	}
	rrHeader := make([]byte, 10)
	binary.BigEndian.PutUint16(rrHeader[0:2], uint16(record.Type))
	binary.BigEndian.PutUint16(rrHeader[2:4], record.Class)
	binary.BigEndian.PutUint32(rrHeader[4:8], record.TTL)
	binary.BigEndian.PutUint16(rrHeader[8:10], uint16(len(record.RData)))
	buf = append(buf, rrHeader...)
	buf = append(buf, record.RData...)
	return buf, nil
}
