package dns

import (
	"encoding/binary"
	"io"
	"net"
	"sync/atomic"
	"testing"
)

// fakeDNSServer answers on one loopback port over UDP and TCP with whatever
// its handler makes of the query; tcp tells the handler which one asked.
type fakeDNSServer struct {
	addr    string
	queries atomic.Int32
	handler func(query *DNSMessage, tcp bool) *DNSMessage
	udp     net.PacketConn
	tcp     net.Listener
}

func newFakeDNSServer(t testing.TB, handler func(query *DNSMessage, tcp bool) *DNSMessage) *fakeDNSServer {
	t.Helper()
	udp, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	tcp, err := net.Listen("tcp4", udp.LocalAddr().String())
	if err != nil {
		udp.Close()
		t.Fatal(err)
	}
	f := &fakeDNSServer{addr: udp.LocalAddr().String(), handler: handler, udp: udp, tcp: tcp}
	t.Cleanup(f.close)
	go f.serveUDP()
	go f.serveTCP()
	return f
}

func (f *fakeDNSServer) close() {
	f.udp.Close()
	f.tcp.Close()
}

func (f *fakeDNSServer) reply(data []byte, tcp bool) []byte {
	query, err := ParseDNSMessage(data)
	if err != nil {
		return nil
	}
	f.queries.Add(1)
	reply := f.handler(query, tcp)
	if reply == nil {
		return nil
	}
	reply.Header.ID = query.Header.ID
	reply.Header.SetQR(true)
	out, err := MarshalDNSMessage(reply)
	if err != nil {
		return nil
	}
	return out
}

func (f *fakeDNSServer) serveUDP() {
	buff := make([]byte, 65535)
	for {
		n, addr, err := f.udp.ReadFrom(buff)
		if err != nil {
			return
		}
		if out := f.reply(buff[:n], false); out != nil {
			f.udp.WriteTo(out, addr)
		}
	}
}

func (f *fakeDNSServer) serveTCP() {
	for {
		conn, err := f.tcp.Accept()
		if err != nil {
			return
		}
		go func() {
			defer conn.Close()
			prefix := make([]byte, 2)
			for {
				if _, err := io.ReadFull(conn, prefix); err != nil {
					return
				}
				data := make([]byte, binary.BigEndian.Uint16(prefix))
				if _, err := io.ReadFull(conn, data); err != nil {
					return
				}
				out := f.reply(data, true)
				if out == nil {
					return
				}
				conn.Write(append(binary.BigEndian.AppendUint16(nil, uint16(len(out))), out...))
			}
		}()
	}
}

// replyWith answers every query with rcode and the given answers.
func replyWith(rcode RCODE, answers ...*DNSRecord) func(*DNSMessage, bool) *DNSMessage {
	return func(query *DNSMessage, _ bool) *DNSMessage {
		reply := &DNSMessage{Header: &DNSHeader{}, Questions: query.Questions, Answers: answers}
		reply.Header.SetRCODE(rcode)
		return reply
	}
}

func aRecord(ip string, ttl uint32) *DNSRecord {
	return &DNSRecord{Name: compressedDomainVal, Type: DNSTypeA, Class: DNSClassIN, TTL: ttl, RData: net.ParseIP(ip).To4()}
}

func ptrRecord(host string, ttl uint32) *DNSRecord {
	return &DNSRecord{Name: compressedDomainVal, Type: DNSTypePTR, Class: DNSClassIN, TTL: ttl, RData: stringToDNSWireFormat(host)}
}
