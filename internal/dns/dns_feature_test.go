package dns

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

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
	DNSErrorContextKey  = "dnsError"
)

type DNSFeatureTestSuite struct {
	resolver    *DNSResolver
	packetBytes []byte
}

func (ts *DNSFeatureTestSuite) reset() {
	ts.resolver = NewDNSResolverWithDefaultOpts()
	ts.packetBytes = nil
}

func (ts *DNSFeatureTestSuite) givenTheDNSPacket(ctx context.Context, packetBase64 string) error {
	decoded, err := base64.StdEncoding.DecodeString(packetBase64)
	if err != nil {
		return fmt.Errorf("failed to decode base64 packet: %v", err)
	}
	ts.packetBytes = decoded
	return nil
}

func (ts *DNSFeatureTestSuite) iParseTheDNSPacket(ctx context.Context) (context.Context, error) {
	packet, err := ParseDNSMessage(ts.packetBytes)
	if err != nil {
		return context.WithValue(ctx, DNSErrorContextKey, err), nil
	}
	return context.WithValue(ctx, DNSPacketContextKey, packet), nil
}

func (ts *DNSFeatureTestSuite) iShouldReceiveADNSQueryFor(ctx context.Context, expectedDomain string) error {
	packet := ctx.Value(DNSPacketContextKey).(*DNSMessage)

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

func (ts *DNSFeatureTestSuite) iShouldReceiveADNSPacketWithTheIP(ctx context.Context, expectedIP string) error {
	packet := ctx.Value(DNSPacketContextKey).(*DNSMessage)

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

func (ts *DNSFeatureTestSuite) thatServerHasACacheForWithIP(domain string, ip string) error {
	ts.resolver.cache[domain] = &DNSCacheItem{
		record: &DNSRecord{
			Type:  DNSTypeA,
			Class: 1,
			TTL:   300,
			RData: net.ParseIP(ip).To4(),
		},
		ttl: time.Now().Add(time.Second * 300),
	}
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

func (ts *DNSFeatureTestSuite) iShouldReceiveTheError(ctx context.Context, expectedError string) error {
	err := ctx.Value(DNSErrorContextKey).(error)
	if err == nil {
		return fmt.Errorf("expected error %s, but got nil", expectedError)
	}

	if err.Error() != expectedError {
		return fmt.Errorf("expected error %s, but got %s", expectedError, err.Error())
	}

	return nil
}

func (ts *DNSFeatureTestSuite) thePacketShouldParse(ctx context.Context) error {
	if err, ok := ctx.Value(DNSErrorContextKey).(error); ok {
		if err != nil {
			return fmt.Errorf("expected packet to parse, but got error: %v", err)
		}
	}
	return nil
}

func (ts *DNSFeatureTestSuite) thePacketShouldHaveAdditionalRecord(ctx context.Context, expectedCount int) error {
	packet := ctx.Value(DNSPacketContextKey).(*DNSMessage)

	actualCount := len(packet.Additionals)
	if actualCount != expectedCount {
		return fmt.Errorf("expected %d additional records, got %d", expectedCount, actualCount)
	}

	return nil
}

func (ts *DNSFeatureTestSuite) theAdditionalRecordShouldBeAnEDNSOPTRecord(ctx context.Context) error {
	packet := ctx.Value(DNSPacketContextKey).(*DNSMessage)

	if len(packet.Additionals) == 0 {
		return fmt.Errorf("no additional records found")
	}

	optRecord := packet.Additionals[0]

	// Check if it's an OPT record
	if optRecord.Type != DNSTypeOPT {
		return fmt.Errorf("expected OPT record (type 41), got type %d (%s)", optRecord.Type, optRecord.Type)
	}

	// Check if name is root domain (empty string for OPT records)
	if optRecord.Name != "" {
		return fmt.Errorf("expected empty name for OPT record, got %s", optRecord.Name)
	}

	// Validate that class field contains UDP payload size (should be reasonable)
	if optRecord.Class == 0 || optRecord.Class > 65535 {
		return fmt.Errorf("invalid UDP payload size in OPT record: %d", optRecord.Class)
	}

	return nil
}

func TestFeatures(t *testing.T) {
	flag.Parse()

	if *reportFile != "" {
		file, err := os.OpenFile(*reportFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatalf("Failed to create output file: %v", err)
		}
		defer file.Close()
		opts.Output = file
		opts.Format = "junit"
	}

	opts.TestingT = t

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

	ctx.Given(`^the DNS packet "([^"]*)"$`, ts.givenTheDNSPacket)
	ctx.When(`^I parse the DNS packet$`, ts.iParseTheDNSPacket)
	ctx.Then(`^I should receive a DNS query for "([^"]*)"$`, ts.iShouldReceiveADNSQueryFor)
	ctx.Then(`^I should receive a DNS packet with the IP "([^"]*)"$`, ts.iShouldReceiveADNSPacketWithTheIP)

	ctx.Given(`^the resolver has a cache for "([^"]*)" with IP "([^"]*)"$`, ts.thatServerHasACacheForWithIP)
	ctx.When(`^I send a DNS request A for "([^"]*)"$`, ts.iSendADNSRequestAFor)
	ctx.Then(`^I should receive a valid DNS response with IP "([^"]*)"$`, ts.theDNSServerShouldRespondWith)
	ctx.Then(`^I should receive the error "([^"]*)"$`, ts.iShouldReceiveTheError)
	ctx.Step(`^The packet should parse`, ts.thePacketShouldParse)
	ctx.Then(`^the packet should have (\d+) additional record$`, ts.thePacketShouldHaveAdditionalRecord)
	ctx.Then(`^the additional record should be an EDNS OPT record$`, ts.theAdditionalRecordShouldBeAnEDNSOPTRecord)
}
