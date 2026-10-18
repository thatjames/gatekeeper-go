package dns

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/cucumber/godog"
)

// parityTestSuite runs a real DNS server on 127.0.0.1 with fake forwarders
// and no upstreams, and talks to it as a client would.
type parityTestSuite struct {
	localDomains map[string]net.IP
	forwarders   []ConditionalForwarder
	fakes        map[string]*fakeDNSServer // by zone
	server       *DNSServer
	reply        *DNSMessage
}

func (ps *parityTestSuite) reset() {
	if ps.server != nil {
		ps.server.Stop()
	}
	for _, fake := range ps.fakes {
		fake.close()
	}
	*ps = parityTestSuite{localDomains: map[string]net.IP{}, fakes: map[string]*fakeDNSServer{}}
}

func (ps *parityTestSuite) theLocalDomainWithIP(domain, ip string) error {
	ps.localDomains[domain] = net.ParseIP(ip).To4()
	return nil
}

func (ps *parityTestSuite) forwardZone(zone string, handler func(*DNSMessage, bool) *DNSMessage) error {
	fake, err := startFakeDNSServer(handler)
	if err != nil {
		return err
	}
	ps.fakes[zone] = fake
	ps.forwarders = append(ps.forwarders, ConditionalForwarder{Zone: zone, Server: fake.addr})
	return nil
}

func (ps *parityTestSuite) zoneForwardedWithRecordsOverTCPOnly(zone string, count int) error {
	answers := make([]*DNSRecord, 0, count)
	for i := 0; i < count; i++ {
		answers = append(answers, aRecord(net.IPv4(10, 0, byte(i>>8), byte(i)).String(), 60))
	}
	return ps.forwardZone(zone, func(query *DNSMessage, tcp bool) *DNSMessage {
		reply := &DNSMessage{Header: &DNSHeader{}, Questions: query.Questions}
		if tcp {
			reply.Answers = answers
		} else {
			reply.Header.SetTC(true)
		}
		return reply
	})
}

func (ps *parityTestSuite) zoneForwardedAnsweringPTR(zone, host string) error {
	return ps.forwardZone(zone, replyWith(RCODESuccess, ptrRecord(host, 60)))
}

func (ps *parityTestSuite) zoneForwardedToSilentServer(zone string) error {
	return ps.forwardZone(zone, func(*DNSMessage, bool) *DNSMessage { return nil })
}

func (ps *parityTestSuite) startServer() error {
	if ps.server != nil {
		return nil
	}
	resolver := NewDNSResolverWithOpts(ResolverOpts{
		LocalDomains:          ps.localDomains,
		ConditionalForwarders: ps.forwarders,
		ReadTimeout:           time.Millisecond * 500,
	})
	ps.server = NewDNSServerWithOpts(DNSServerOpts{Interface: "lo", Port: 0, ReadDeadline: time.Millisecond * 100}, resolver, nil)
	return ps.server.Start()
}

func (ps *parityTestSuite) iAskFor(recordType, name, transport, edns string) error {
	if err := ps.startServer(); err != nil {
		return err
	}
	query := NewDnsMessage()
	query.Header.ID = uint16(rand.Intn(65535))
	query.Header.SetRD(true)
	types := map[string]DNSType{"A": DNSTypeA, "PTR": DNSTypePTR}
	query.Questions = append(query.Questions, &DNSQuestion{Name: stringToDNSWireFormat(name), Type: types[recordType], Class: DNSClassIN})
	if edns != "" {
		query.Additionals = append(query.Additionals, NewOPTRecord(maxUDPSize))
	}
	data, err := MarshalDNSMessage(query)
	if err != nil {
		return err
	}

	addr := ps.server.packetConn.LocalAddr().String()
	var reply []byte
	if transport == "TCP" {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err != nil {
			return err
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(time.Second * 3))
		if _, err := conn.Write(append(binary.BigEndian.AppendUint16(nil, uint16(len(data))), data...)); err != nil {
			return err
		}
		prefix := make([]byte, 2)
		if _, err := io.ReadFull(conn, prefix); err != nil {
			return err
		}
		reply = make([]byte, binary.BigEndian.Uint16(prefix))
		if _, err := io.ReadFull(conn, reply); err != nil {
			return err
		}
	} else {
		conn, err := net.Dial("udp", addr)
		if err != nil {
			return err
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(time.Second * 3))
		if _, err := conn.Write(data); err != nil {
			return err
		}
		buff := make([]byte, 65535)
		n, err := conn.Read(buff)
		if err != nil {
			return err
		}
		reply = buff[:n]
		if limit := udpSizeLimit(query); n > limit {
			return fmt.Errorf("UDP reply of %d bytes is over the %d byte limit", n, limit)
		}
	}
	if ps.reply, err = ParseDNSMessage(reply); err != nil {
		return err
	}
	if ps.reply.Header.ID != query.Header.ID {
		return fmt.Errorf("reply ID %d does not match query ID %d", ps.reply.Header.ID, query.Header.ID)
	}
	return nil
}

func (ps *parityTestSuite) theReplyHasRCODEAndAnswers(rcode string, count int) error {
	rcodes := map[string]RCODE{"NOERROR": RCODESuccess, "SERVFAIL": RCODEServerFailure, "NXDOMAIN": RCODENameFailure}
	if got := ps.reply.Header.RCODE(); got != rcodes[rcode] {
		return fmt.Errorf("expected RCODE %s, got %d", rcode, got)
	}
	if len(ps.reply.Answers) != count {
		return fmt.Errorf("expected %d answers, got %d", count, len(ps.reply.Answers))
	}
	return nil
}

func (ps *parityTestSuite) theReplyIsTruncatedWithAnswers(count int) error {
	if !ps.reply.Header.TC() {
		return fmt.Errorf("expected TC set")
	}
	if len(ps.reply.Answers) != count {
		return fmt.Errorf("expected %d answers, got %d", count, len(ps.reply.Answers))
	}
	return nil
}

func (ps *parityTestSuite) theReplyIsNotTruncated() error {
	if ps.reply.Header.TC() {
		return fmt.Errorf("expected TC clear")
	}
	return nil
}

func (ps *parityTestSuite) theReplyHasAnOPTRecord() error {
	if opt := queryOPT(ps.reply); opt == nil || opt.Class != maxUDPSize {
		return fmt.Errorf("expected an OPT record with UDP size %d, got %v", maxUDPSize, ps.reply.Additionals)
	}
	return nil
}

func (ps *parityTestSuite) theReplyHasNoOPTRecord() error {
	if len(ps.reply.Additionals) != 0 {
		return fmt.Errorf("expected no additional records, got %v", ps.reply.Additionals)
	}
	return nil
}

func (ps *parityTestSuite) theFirstAnswerIsPTR(host string) error {
	if len(ps.reply.Answers) == 0 || ps.reply.Answers[0].Type != DNSTypePTR || !bytes.Equal(ps.reply.Answers[0].RData, stringToDNSWireFormat(host)) {
		return fmt.Errorf("expected PTR %s, got %v", host, ps.reply.Answers)
	}
	return nil
}

func (ps *parityTestSuite) theForwarderWasAsked(zone string, count int) error {
	if got := ps.fakes[zone].queries.Load(); int(got) != count {
		return fmt.Errorf("expected the forwarder for %s asked %d times, got %d", zone, count, got)
	}
	return nil
}

func initializeParityScenario(ctx *godog.ScenarioContext) {
	ps := &parityTestSuite{}
	ps.reset()
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		ps.reset()
		return ctx, nil
	})

	ctx.Given(`^the local domain "([^"]*)" with IP "([^"]*)"$`, ps.theLocalDomainWithIP)
	ctx.Given(`^the zone "([^"]*)" is forwarded to a server with (\d+) A records over TCP only$`, ps.zoneForwardedWithRecordsOverTCPOnly)
	ctx.Given(`^the zone "([^"]*)" is forwarded to a server answering PTR "([^"]*)"$`, ps.zoneForwardedAnsweringPTR)
	ctx.Given(`^the zone "([^"]*)" is forwarded to a server that does not answer$`, ps.zoneForwardedToSilentServer)
	ctx.When(`^I ask for (A|PTR) "([^"]*)" over (UDP|TCP)( with EDNS)?$`, ps.iAskFor)
	ctx.Then(`^the reply has RCODE (NOERROR|NXDOMAIN|SERVFAIL) and (\d+) answers?$`, ps.theReplyHasRCODEAndAnswers)
	ctx.Then(`^the reply is truncated with (\d+) answers?$`, ps.theReplyIsTruncatedWithAnswers)
	ctx.Then(`^the reply is not truncated$`, ps.theReplyIsNotTruncated)
	ctx.Then(`^the reply has an OPT record$`, ps.theReplyHasAnOPTRecord)
	ctx.Then(`^the reply has no OPT record$`, ps.theReplyHasNoOPTRecord)
	ctx.Then(`^the first answer is PTR "([^"]*)"$`, ps.theFirstAnswerIsPTR)
	ctx.Then(`^the forwarder for "([^"]*)" was asked (\d+) times?$`, ps.theForwarderWasAsked)
}
