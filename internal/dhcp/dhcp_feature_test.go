package dhcp

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/golang/mock/gomock"
)

type TestSuite struct {
	leasePool    *LeasePool
	ctrl         *gomock.Controller
	currentLease *Lease
	error        error
	leases       []Lease
	clientID     string
	filename     string
	leaseFile    string
}

func (ts *TestSuite) reset() {
	if ts.ctrl != nil {
		ts.ctrl.Finish()
	}
	ts.leasePool = nil
	ts.ctrl = nil
	ts.currentLease = nil
	ts.error = nil
	ts.leases = nil
	ts.clientID = ""
	ts.filename = ""
	ts.leaseFile = ""
}

type mockTestingT struct{}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func (m *mockTestingT) Fatalf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func (m *mockTestingT) Helper() {}

func (ts *TestSuite) iHaveALeasePoolWithIPRange(startIP, endIP string) error {
	start := net.ParseIP(startIP)
	end := net.ParseIP(endIP)
	if start == nil || end == nil {
		return fmt.Errorf("invalid IP addresses: %s, %s", startIP, endIP)
	}
	ts.leasePool = NewLeasePool(start.To4(), end.To4())
	return nil
}

func (ts *TestSuite) theLeasePoolShouldHaveNLeases(expectedCount int) error {
	actualCount := len(ts.leasePool.ActiveLeases()) + len(ts.leasePool.ReservedLeases())
	if actualCount != expectedCount {
		return fmt.Errorf("expected %d leases, but got %d", expectedCount, actualCount)
	}
	return nil
}

func (ts *TestSuite) iReserveAnIPAddressForClient(ipAddress, clientID string) error {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s", ipAddress)
	}
	ts.leasePool.ReserveLease(clientID, ip)
	ts.clientID = clientID
	return nil
}

func (ts *TestSuite) iRequestALeaseForClient(clientID string) error {
	ts.clientID = clientID
	ts.currentLease = ts.leasePool.GetLease(clientID)
	return nil
}

func (ts *TestSuite) iShouldReceiveALeaseWithIP(expectedIP string) error {
	if ts.currentLease == nil {
		return fmt.Errorf("expected lease with IP %s, but got nil", expectedIP)
	}
	expected := net.ParseIP(expectedIP)
	if !ts.currentLease.IP.Equal(expected) {
		return fmt.Errorf("expected IP %s, but got %s", expectedIP, ts.currentLease.IP.String())
	}
	return nil
}

func (ts *TestSuite) theLeaseStateShouldBe(expectedState string) error {
	if ts.currentLease == nil {
		return fmt.Errorf("no current lease to check state")
	}

	var expected LeaseState
	switch strings.ToLower(expectedState) {
	case "reserved":
		expected = LeaseReserved
	case "active":
		expected = LeaseActive
	case "offered":
		expected = LeaseOffered
	case "available":
		expected = LeaseAvailable
	default:
		return fmt.Errorf("unknown lease state: %s", expectedState)
	}

	if ts.currentLease.State != expected {
		return fmt.Errorf("expected state %s, but got %s", expectedState, ts.currentLease.State)
	}
	return nil
}

func (ts *TestSuite) iRequestTheNextAvailableLeaseForClient(clientID string) error {
	ts.clientID = clientID
	ts.currentLease = ts.leasePool.NextAvailableLease(clientID)
	return nil
}

func (ts *TestSuite) iShouldReceiveALease() error {
	if ts.currentLease == nil {
		return fmt.Errorf("expected to receive a lease, but got nil")
	}
	return nil
}

func (ts *TestSuite) iAcceptTheLeaseForDuration(duration string) error {
	if ts.currentLease == nil {
		return fmt.Errorf("no lease to accept")
	}

	ttl, err := time.ParseDuration(duration)
	if err != nil {
		return fmt.Errorf("invalid duration %s: %v", duration, err)
	}

	ts.leasePool.AcceptLease(ts.currentLease, ttl)
	return nil
}

func (ts *TestSuite) theLeaseClientIDShouldBe(expectedClientID string) error {
	if ts.currentLease == nil {
		return fmt.Errorf("no current lease to check client ID")
	}
	if ts.currentLease.ClientId != expectedClientID {
		return fmt.Errorf("expected client ID %s, but got %s", expectedClientID, ts.currentLease.ClientId)
	}
	return nil
}

func (ts *TestSuite) theLeaseForClientExpiresInThePast(clientID string) error {
	for _, lease := range ts.leasePool.leases {
		if lease.ClientId == clientID {
			lease.Expiry = time.Now().Add(-time.Hour)
			break
		}
	}
	return nil
}

func (ts *TestSuite) iShouldReceiveNoLease() error {
	if ts.currentLease != nil {
		return fmt.Errorf("expected no lease, but got lease with IP %s", ts.currentLease.IP.String())
	}
	return nil
}

func (ts *TestSuite) iPersistLeasesToFile(filename string) error {
	ts.filename = filename
	ts.error = ts.leasePool.PeristLeases(ts.leaseFile)
	return nil
}

func (ts *TestSuite) iLoadNLeasesFromThePersistenceLayer(n int) error {
	return nil
}

func (ts *TestSuite) theOperationShouldSucceed() error {
	if ts.error != nil {
		return fmt.Errorf("expected operation to succeed, but got error: %v", ts.error)
	}
	return nil
}

func (ts *TestSuite) theLeasePoolShouldContainNActiveLeases(expectedCount int) error {
	activeLeases := ts.leasePool.ActiveLeases()
	actualCount := len(activeLeases)
	if actualCount != expectedCount {
		return fmt.Errorf("expected %d active leases, but got %d", expectedCount, actualCount)
	}
	return nil
}

func (ts *TestSuite) allLeasesShouldInitiallyBeAvailable() error {
	return nil
}

func (ts *TestSuite) iRequestNLeasesForDifferentClients(count int) error {
	for i := 0; i < count; i++ {
		clientID := fmt.Sprintf("client%d", i)
		lease := ts.leasePool.NextAvailableLease(clientID)
		if lease == nil {
			activeLeases := ts.leasePool.ActiveLeases()
			reservedLeases := ts.leasePool.ReservedLeases()
			return fmt.Errorf("failed to get lease for client %s (iteration %d/%d, active: %d, reserved: %d)",
				clientID, i+1, count, len(activeLeases), len(reservedLeases))
		}

		if lease.IP.String() == "0.0.0.0" || lease.IP.String() == "<nil>" {
			return fmt.Errorf("client %s got lease with invalid IP: %s", clientID, lease.IP.String())
		}

		ts.leasePool.AcceptLease(lease, time.Hour)
	}
	return nil
}

func (ts *TestSuite) iUpdateClientIP(clientID, ipAddress string) error {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s", ipAddress)
	}
	ts.leasePool.UpdateLease(clientID, ip)
	return nil
}

func (ts *TestSuite) iHaveALeasePoolWithClientAndReservedIP(clientID, ipAddress string) error {
	ts.clientID = clientID
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s", ipAddress)
	}
	ts.leasePool = NewLeasePool(net.ParseIP("10.0.0.1").To4(), net.ParseIP("10.0.0.2").To4())
	ts.leasePool.ReserveLease(clientID, ip)
	return nil
}

func (ts *TestSuite) iHaveALeasePoolWithClientAndIP(clientID, ipAddress string) error {
	ts.clientID = clientID
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s", ipAddress)
	}
	ts.leasePool = NewLeasePool(net.ParseIP("10.0.0.1").To4(), net.ParseIP("10.0.0.2").To4())
	for _, lease := range ts.leasePool.leases {
		if lease.IP.Equal(net.ParseIP(ipAddress).To4()) {
			lease.ClientId = clientID
			lease.State = LeaseActive
			lease.Expiry = time.Now().Add(time.Hour)
			break
		}
	}
	return nil
}

func (ts *TestSuite) thereIsAnExpiredLeaseForClientWithIP(clientID, ipAddress string) error {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s", ipAddress)
	}
	ts.leasePool.leases[0] = &Lease{
		ClientId: clientID,
		IP:       ip,
		State:    LeaseActive,
		Expiry:   time.Now().Add(-time.Hour),
	}
	return nil
}

func (ts *TestSuite) iRequestTheFirstExpiredLeaseForClient(clientID string) error {
	lease := ts.leasePool.NextAvailableLease(clientID)
	if lease == nil {
		return fmt.Errorf("no lease found for client %s", clientID)
	}
	ts.currentLease = lease
	return nil
}

func (ts *TestSuite) clientShouldHaveIP(clientID, ipAddress string) error {
	lease := ts.leasePool.GetLease(clientID)
	if lease == nil {
		return fmt.Errorf("client %s does not have a lease", clientID)
	}
	if lease.IP.String() != ipAddress {
		return fmt.Errorf("client %s has lease with IP %s, expected %s", clientID, lease.IP.String(), ipAddress)
	}
	return nil
}

func (ts *TestSuite) theIPAddressShouldBe(expectedIP string) error {
	if ts.currentLease == nil {
		return fmt.Errorf("no current lease to check IP address")
	}
	if ts.currentLease.IP.String() != expectedIP {
		return fmt.Errorf("expected IP %s, but got %s", expectedIP, ts.currentLease.IP.String())
	}
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ts := &TestSuite{}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		ts.reset()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		ts.reset()
		return ctx, nil
	})

	/*
	   And the IP address should be "10.0.0.1"
	*/

	ctx.Step(`^there is an expired lease for client "([^"]*)" with IP "([^"]*)"$`, ts.thereIsAnExpiredLeaseForClientWithIP)
	ctx.When(`^I request the first expired lease for client "([^"]*)"$`, ts.iRequestTheFirstExpiredLeaseForClient)
	ctx.Given(`^I have a lease pool with client "([^"]*)" and reserved IP "([^"]*)"$`, ts.iHaveALeasePoolWithClientAndReservedIP)
	ctx.Given(`^I have a lease pool with client "([^"]*)" and IP "([^"]*)"$`, ts.iHaveALeasePoolWithClientAndIP)
	ctx.Step(`^I have a lease pool with IP range "([^"]*)" to "([^"]*)"$`, ts.iHaveALeasePoolWithIPRange)
	ctx.Step(`^the lease pool should have (\d+) leases$`, ts.theLeasePoolShouldHaveNLeases)
	ctx.Step(`^all leases should initially be available$`, ts.allLeasesShouldInitiallyBeAvailable)
	ctx.Step(`^the IP address should be "([^"]*)"$`, ts.theIPAddressShouldBe)

	ctx.Step(`^I update client "([^"]*)" ip to "([^"]*)"`, ts.iUpdateClientIP)
	ctx.Step(`^I reserve IP address "([^"]*)" for client "([^"]*)"$`, ts.iReserveAnIPAddressForClient)
	ctx.Step(`^I request a lease for client "([^"]*)"$`, ts.iRequestALeaseForClient)
	ctx.Step(`^I should receive a lease with IP "([^"]*)"$`, ts.iShouldReceiveALeaseWithIP)
	ctx.Step(`^the lease state should be "([^"]*)"$`, ts.theLeaseStateShouldBe)
	ctx.Step(`^the lease client ID should be "([^"]*)"$`, ts.theLeaseClientIDShouldBe)

	ctx.Step(`^I request the next available lease for client "([^"]*)"$`, ts.iRequestTheNextAvailableLeaseForClient)
	ctx.Step(`^I should receive a lease$`, ts.iShouldReceiveALease)
	ctx.Step(`^I should receive no lease$`, ts.iShouldReceiveNoLease)
	ctx.Step(`^I accept the lease for "([^"]*)"$`, ts.iAcceptTheLeaseForDuration)

	ctx.Step(`^Client "([^"]*)" should have IP "([^"]*)"$`, ts.clientShouldHaveIP)
	ctx.Step(`^the lease for client "([^"]*)" expires in the past$`, ts.theLeaseForClientExpiresInThePast)

	ctx.Step(`^I request (\d+) leases for different clients$`, ts.iRequestNLeasesForDifferentClients)
	ctx.Step(`^the lease pool should contain (\d+) active leases$`, ts.theLeasePoolShouldContainNActiveLeases)
}

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "pretty",
}

var reportFile = flag.String("report-file", "", "Output cucumber file for JSON report")

func init() {
	godog.BindCommandLineFlags("godog.", &opts)
}

func TestFeatures(t *testing.T) {
	flag.Parse()
	opts.TestingT = t

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

func TestReservedLeases(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "Reserved Leases",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/reserved_leases.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func TestLeaseInitialization(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "Lease Initialization",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/lease_initialization.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func TestLeaseOffering(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "Lease Offering",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/lease_offering.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
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
