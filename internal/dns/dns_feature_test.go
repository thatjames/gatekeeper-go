package dns

import (
	"context"
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
	DNSQueryContextKey = "dnsQuery"
)

type DNSFeatureTestSuite struct {
	resolver *DNSResolver
}

func (ts *DNSFeatureTestSuite) reset() {
	ts.resolver = NewDNSResolver()
}
func (ts *DNSFeatureTestSuite) iSendADNSRequestAFor(ctx context.Context, domain string) (context.Context, error) {
	dnsPacket, err := ts.resolver.Resolve(domain)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, DNSQueryContextKey, dnsPacket), nil
}

func (ts *DNSFeatureTestSuite) theDNSServerShouldRespondWith(ctx context.Context, expectedResponse string) error {
	dnsQuery := ctx.Value(DNSQueryContextKey).(*DNSPacket)
	if dnsQuery.Type != DNSTypeA {
		return fmt.Errorf("expected DNS query type A, but got %s", dnsQuery.Type)
	}

	if dnsQuery.Class != 1 {
		return fmt.Errorf("expected class %d, but got %d", 1, dnsQuery.Class)
	}

	if dnsQuery.TTL != 300 {
		return fmt.Errorf("expected TTL %d, but got %d", 300, dnsQuery.TTL)
	}

	netIP := net.ParseIP(expectedResponse).To4()
	if !net.IP(dnsQuery.RData).Equal(netIP) {
		return fmt.Errorf("expected IP %s, but got %s", netIP, dnsQuery.RData)
	}
	return nil
}

func (ts *DNSFeatureTestSuite) thatServerHasACacheForWithIP(domain string, ip string) error {
	ts.resolver.cache[domain] = net.ParseIP(ip).To4()
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
		// Create output file for JSON
		file, err := os.Create(*reportFile)
		if err != nil {
			t.Fatalf("Failed to create output file: %v", err)
		}
		defer file.Close()
		opts.Output = file
	}

	opts.TestingT = t
	opts.Format = "junit"

	status := godog.TestSuite{
		Name:                "DHCP Lease Management",
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

	ctx.Step(`^the resolver has a cache for "([^"]*)" with IP "([^"]*)"$`, ts.thatServerHasACacheForWithIP)
	ctx.When(`^I send a DNS request A for "([^"]*)"$`, ts.iSendADNSRequestAFor)
	ctx.Then(`^I should receive a valid DNS response with IP "([^"]*)"$`, ts.theDNSServerShouldRespondWith)
}
