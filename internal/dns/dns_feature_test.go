package dns

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
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

var (
	dnsTypeMappings = map[string]DNSType{
		"A":     DNSTypeA,
		"AAAA":  DNSTypeAAAA,
		"CNAME": DNSTypeCNAME,
		"MX":    DNSTypeMX,
		"NS":    DNSTypeNS,
		"TXT":   DNSTypeTXT,
		"OPT":   DNSTypeOPT,
	}
)

var (
	countMapping = map[string]int{
		"first":  0,
		"second": 1,
		"third":  2,
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

	if len(packet.Questions) != 1 {
		return fmt.Errorf("expected 1 question, got %d", len(packet.Questions))
	}

	question := packet.Questions[0]
	if question.ParsedName != expectedDomain {
		return fmt.Errorf("expected domain %s, got %s", expectedDomain, question.ParsedName)
	}

	if len(packet.Answers) != 0 {
		return fmt.Errorf("expected 0 answers in query, got %d", len(packet.Answers))
	}

	return nil
}

func (ts *DNSFeatureTestSuite) iShouldReceiveADNSPacketWithTheIP(ctx context.Context, expectedIP string) error {
	packet := ctx.Value(DNSPacketContextKey).(*DNSMessage)

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
	keyBuff := bytes.NewBufferString(domain)
	binary.Write(keyBuff, binary.BigEndian, DNSTypeA)
	cacheKey := fmt.Sprintf("%x", keyBuff.Bytes())
	records := make([]*DNSRecord, 0)
	records = append(records, &DNSRecord{
		Type:  DNSTypeA,
		Class: 1,
		TTL:   300,
		RData: net.ParseIP(ip).To4(),
	})
	ts.resolver.cache[cacheKey] = &DNSCacheItem{
		ttl:     time.Now().Add(time.Second * 300),
		records: records,
	}
	return nil
}

func (ts *DNSFeatureTestSuite) iSendADNSRequestAFor(ctx context.Context, domain string) (context.Context, error) {
	dnsRecord, _, err := ts.resolver.Resolve(domain, DNSTypeA)
	if err != nil {
		return context.Background(), err
	}

	return context.WithValue(ctx, DNSQueryContextKey, dnsRecord), nil
}

func (ts *DNSFeatureTestSuite) theDNSServerShouldRespondWithASingleRecord(ctx context.Context, expectedResponse string) error {
	dnsRecord := ctx.Value(DNSQueryContextKey).([]*DNSRecord)[0]
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
	if optRecord.Name != nil && len(optRecord.Name) > 0 {
		return fmt.Errorf("expected empty name for OPT record, got %v", optRecord.Name)
	}

	// Validate that class field contains UDP payload size (should be reasonable)
	if optRecord.Class == 0 || optRecord.Class > 65535 {
		return fmt.Errorf("invalid UDP payload size in OPT record: %d", optRecord.Class)
	}

	return nil
}

func (ts *DNSFeatureTestSuite) thenIShouldReceiveADNSPacketWithAnAuthorityCountOf(ctx context.Context, expectedCount int) error {
	packet := ctx.Value(DNSPacketContextKey).(*DNSMessage)

	actualCount := len(packet.Authorities)
	if actualCount != expectedCount {
		return fmt.Errorf("expected %d authority records, got %d", expectedCount, actualCount)
	}

	return nil
}

func (ts *DNSFeatureTestSuite) andThePacketShouldHaveAnswers(ctx context.Context, expectedCount int) error {
	packet := ctx.Value(DNSPacketContextKey).(*DNSMessage)

	actualCount := len(packet.Answers)
	if actualCount != expectedCount {
		return fmt.Errorf("expected %d answers, got %d", expectedCount, actualCount)
	}

	return nil
}

func (ts *DNSFeatureTestSuite) andThePacketShouldHaveAuthority(ctx context.Context, expectedCount int) error {
	packet := ctx.Value(DNSPacketContextKey).(*DNSMessage)

	actualCount := len(packet.Authorities)
	if actualCount != expectedCount {
		return fmt.Errorf("expected %d authority records, got %d", expectedCount, actualCount)
	}

	return nil
}

func (ts *DNSFeatureTestSuite) andTheRecordShouldBeARecordPointingTo(ctx context.Context, positionType, recordType, expectedValue string) error {
	packet := ctx.Value(DNSPacketContextKey).(*DNSMessage)
	var record *DNSRecord
	switch positionType {
	case "first":
		if len(packet.Answers) == 0 {
			return fmt.Errorf("expected at least 1 answer, got %d", len(packet.Answers))
		}
		record = packet.Answers[0]
	case "second":
		if len(packet.Answers) < 2 {
			return fmt.Errorf("expected at least 2 answers, got %d", len(packet.Answers))
		}
		record = packet.Answers[1]
	case "third":
		if len(packet.Answers) < 3 {
			return fmt.Errorf("expected at least 3 answers, got %d", len(packet.Answers))
		}
		record = packet.Answers[2]
	}

	if record.Type != dnsTypeMappings[recordType] {
		return fmt.Errorf("expected record type %s, got %s", recordType, record.Type)
	}

	if record.RData == nil {
		return fmt.Errorf("expected record RData to be non-nil")
	}

	rdata := record.RData
	switch recordType {
	case "A":
		if len(rdata) != 4 {
			return fmt.Errorf("expected 4 bytes for A record, got %d", len(rdata))
		}
		actualIP := fmt.Sprintf("%d.%d.%d.%d", rdata[0], rdata[1], rdata[2], rdata[3])
		if actualIP != expectedValue {
			return fmt.Errorf("expected IP %s, got %s", expectedValue, actualIP)
		}
	case "CNAME":
		if len(rdata) < 1 {
			return fmt.Errorf("expected CNAME record to have at least 1 byte of data, got %d", len(rdata))
		}
		if record.ParsedRData != expectedValue {
			return fmt.Errorf("expected CNAME %s, got %s", expectedValue, record.ParsedRData)
		}
	default:
		return fmt.Errorf("record type %s not supported", recordType)
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
	ctx.Then(`^I should receive a single valid DNS response with IP "([^"]*)"$`, ts.theDNSServerShouldRespondWithASingleRecord)
	ctx.Then(`^I should receive the error "([^"]*)"$`, ts.iShouldReceiveTheError)
	ctx.Step(`^The packet should parse`, ts.thePacketShouldParse)
	ctx.Then(`^the packet should have (\d+) additional record$`, ts.thePacketShouldHaveAdditionalRecord)
	ctx.Then(`^the additional record should be an EDNS OPT record$`, ts.theAdditionalRecordShouldBeAnEDNSOPTRecord)
	ctx.Then(`^I should receive a DNS packet with an authority count of (\d+)$`, ts.thenIShouldReceiveADNSPacketWithAnAuthorityCountOf)
	ctx.Step(`^the packet should have (\d+) (?:answer|answers)$`, ts.andThePacketShouldHaveAnswers)
	ctx.Step(`^the packet should have (\d+) authority$`, ts.andThePacketShouldHaveAuthority)

	// And the first record should be a CNAME record pointing to "home.slimjim.xyz"
	// And the second record should be an A record pointing to "84.82.5.244"
	ctx.Step(`^the (first|second|third) record should be (?:a|an) ([a-z|A-Z]+) record pointing to "([^"]*)"$`, ts.andTheRecordShouldBeARecordPointingTo)

}
