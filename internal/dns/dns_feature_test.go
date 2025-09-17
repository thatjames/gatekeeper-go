package dns

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var (
	reportFile = flag.String("report-file", "", "Output cucumber file for JSON report")

	opts = godog.Options{
		Output: colors.Colored(os.Stdout),
		Format: "pretty",
	}
)

const (
	DNSQueryContextKey  = "dnsQuery"
	DNSPacketContextKey = "dnsPacket"
)

type DNSFeatureTestSuite struct {
	resolver    *DNSResolver
	packetBytes []byte
}

func (ts *DNSFeatureTestSuite) reset() {
	ts.resolver = NewDNSResolver()
	ts.packetBytes = nil
}

// === DNS PACKET PARSING FEATURES ===

func (ts *DNSFeatureTestSuite) givenTheDNSPacket(ctx context.Context, packetBase64 string) error {
	decoded, err := base64.StdEncoding.DecodeString(packetBase64)
	if err != nil {
		return fmt.Errorf("failed to decode base64 packet: %v", err)
	}
	ts.packetBytes = decoded
	return nil
}

func (ts *DNSFeatureTestSuite) iParseTheDNSPacket(ctx context.Context) (context.Context, error) {
	packet, err := ParseDNSPacket(ts.packetBytes)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse DNS packet: %v", err)
	}
	return context.WithValue(ctx, DNSPacketContextKey, packet), nil
}

func (ts *DNSFeatureTestSuite) iShouldReceiveADNSQueryFor(ctx context.Context, expectedDomain string) error {
	packet := ctx.Value(DNSPacketContextKey).(*DNSPacket)

	if packet.Header.QR != 0 {
		return fmt.Errorf("expected DNS query (QR=0), got QR=%d", packet.Header.QR)
	}

	if len(packet.Questions) != 1 {
		return fmt.Errorf("expected 1 question, got %d", len(packet.Questions))
	}

	question := packet.Questions[0]
	if question.Name != expectedDomain {
		return fmt.Errorf("expected domain %s, got %s", expectedDomain, question.Name)
	}

	if len(packet.Answers) != 0 {
		return fmt.Errorf("expected 0 answers in query, got %d", len(packet.Answers))
	}

	return nil
}

// This function handles the expectation of IP in a DNS packet
// It should work for RESPONSE packets, not query packets
func (ts *DNSFeatureTestSuite) iShouldReceiveADNSPacketWithTheIP(ctx context.Context, expectedIP string) error {
	packet := ctx.Value(DNSPacketContextKey).(*DNSPacket)

	// Check if this is a response
	if packet.Header.QR != 1 {
		return fmt.Errorf("expected DNS response packet to contain IP, but got query packet (QR=%d)", packet.Header.QR)
	}

	if len(packet.Answers) == 0 {
		return fmt.Errorf("expected at least 1 answer with IP, got %d", len(packet.Answers))
	}

	answer := packet.Answers[0]
	if answer.Type != DNSTypeA {
		return fmt.Errorf("expected A record answer, got %s", answer.Type)
	}

	if len(answer.RData) != 4 {
		return fmt.Errorf("expected 4 bytes for IPv4, got %d", len(answer.RData))
	}

	actualIP := fmt.Sprintf("%d.%d.%d.%d", answer.RData[0], answer.RData[1], answer.RData[2], answer.RData[3])
	if actualIP != expectedIP {
		return fmt.Errorf("expected IP %s, got %s", expectedIP, actualIP)
	}

	return nil
}

// === DNS RESOLVER FEATURES ===

func (ts *DNSFeatureTestSuite) thatServerHasACacheForWithIP(domain string, ip string) error {
	ts.resolver.cache[domain] = net.ParseIP(ip).To4()
	return nil
}

func (ts *DNSFeatureTestSuite) iSendADNSRequestAFor(ctx context.Context, domain string) (context.Context, error) {
	dnsRecord, err := ts.resolver.Resolve(domain)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, DNSQueryContextKey, dnsRecord), nil
}

func (ts *DNSFeatureTestSuite) theDNSServerShouldRespondWith(ctx context.Context, expectedResponse string) error {
	dnsRecord := ctx.Value(DNSQueryContextKey).(*DNSRecord)
	if dnsRecord.Type != DNSTypeA {
		return fmt.Errorf("expected DNS record type A, but got %s", dnsRecord.Type)
	}

	if dnsRecord.Class != 1 {
		return fmt.Errorf("expected class %d, but got %d", 1, dnsRecord.Class)
	}

	if dnsRecord.TTL != 300 {
		return fmt.Errorf("expected TTL %d, but got %d", 300, dnsRecord.TTL)
	}

	netIP := net.ParseIP(expectedResponse).To4()
	if !net.IP(dnsRecord.RData).Equal(netIP) {
		return fmt.Errorf("expected IP %s, but got %s", netIP, dnsRecord.RData)
	}
	return nil
}

func TestFeatures(t *testing.T) {
	opts.TestingT = t
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}

	if suite.Run() != 0 {
		t.FailNow()
	}
}

func TestFeaturesWithOutputFile(t *testing.T) {
	flag.Parse()

	if *reportFile != "" {
		file, err := os.OpenFile(*reportFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatalf("Failed to create output file: %v", err)
		}
		defer file.Close()
		opts.Output = file
	}

	opts.TestingT = t
	opts.Format = "junit"

	status := godog.TestSuite{
		Name:                "DNS Management",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	if status == 2 {
		t.SkipNow()
	}

	if status != 0 {
		t.Fatalf("zero status code expected, %d received", status)
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ts := &DNSFeatureTestSuite{}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		ts.reset()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		ts.reset()
		return ctx, nil
	})

	// DNS Packet Parsing step definitions
	ctx.Given(`^the DNS packet "([^"]*)"$`, ts.givenTheDNSPacket)
	ctx.When(`^I parse the DNS packet$`, ts.iParseTheDNSPacket)
	ctx.Then(`^I should receive a DNS query for "([^"]*)"$`, ts.iShouldReceiveADNSQueryFor)
	ctx.Then(`^I should receive a DNS packet with the IP "([^"]*)"$`, ts.iShouldReceiveADNSPacketWithTheIP)

	// DNS Resolver step definitions
	ctx.Given(`^the resolver has a cache for "([^"]*)" with IP "([^"]*)"$`, ts.thatServerHasACacheForWithIP)
	ctx.When(`^I send a DNS request A for "([^"]*)"$`, ts.iSendADNSRequestAFor)
	ctx.Then(`^I should receive a valid DNS response with IP "([^"]*)"$`, ts.theDNSServerShouldRespondWith)
}
