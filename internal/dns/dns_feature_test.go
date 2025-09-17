package dns

import (
	"context"
	"testing"

	"github.com/cucumber/godog"
)

var (
	outputFile string
)

type DNSFeatureTestSuite struct {
}

func (ts *DNSFeatureTestSuite) reset() {

}

func (ts *DNSFeatureTestSuite) iHaveADNSServer() error {
	return nil
}

func (ts *DNSFeatureTestSuite) iSendADNSRequestAFor(domain string) error {
	return nil
}

func (ts *DNSFeatureTestSuite) theDNSServerShouldRespondWith(expectedResponse string) error {
	return nil
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.FailNow()
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

	ctx.Given(`^I have a DNS server$`, ts.iHaveADNSServer)
	ctx.When(`^I send a DNS request A for "([^"]*)"$`, ts.iSendADNSRequestAFor)
	ctx.Then(`^I should receive a valid DNS response with IP "([^"]*)"$`, ts.theDNSServerShouldRespondWith)
}
