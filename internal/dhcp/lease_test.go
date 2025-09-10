package dhcp

import (
	"encoding/binary"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/common"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/datasource"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/datasource/mocks"
)

var (
	start = net.IP{10, 0, 0, 1}
	end   = net.IP{10, 0, 0, 95}
)

// Helper function to set up mock and cleanup
func setupMockDataSource(t *testing.T) (*mocks.MockDHCPDataSource, func()) {
	ctrl := gomock.NewController(t)
	mockDS := mocks.NewMockDHCPDataSource(ctrl)

	originalDS := datasource.DataSource
	datasource.DataSource = mockDS

	cleanup := func() {
		ctrl.Finish()
		datasource.DataSource = originalDS
	}

	return mockDS, cleanup
}

func Test_ReservedLease_Pass(t *testing.T) {
	// This test doesn't use DataSource, so no mock needed
	db := NewLeasePool(net.ParseIP("10.0.0.1"), net.ParseIP("10.0.0.10"))
	db.ReserveLease("test", net.ParseIP("10.0.0.100"))
	lease := db.GetLease("test")
	if lease == nil {
		t.Fatal("expected reserved lease, none returned")
	}

	assert.Equal(t, net.IP{10, 0, 0, 100}, lease.IP.To4(), "expected address 10.0.0.100 but got", lease.IP.To4())
	assert.Equal(t, common.LeaseReserved, lease.State, "expected LeaseReserved but got", lease.State)
	assert.Equal(t, "test", lease.ClientId, "expected test but got", lease.ClientId)
}

func Test_InitSuccess(t *testing.T) {
	// This test doesn't use DataSource, so no mock needed
	db := basicDB()
	assert.Equal(t, int(binary.BigEndian.Uint32(end)-binary.BigEndian.Uint32(start)), len(db.leases), "should have prepopulated range of addresses")
	for i, lease := range db.leases {
		targetIP := binary.BigEndian.Uint32(start) + uint32(i)
		switch {
		case lease.State != common.LeaseAvailable:
			t.Fatalf("lease should be available but is %s", lease.State)
		case binary.BigEndian.Uint32(lease.IP) != targetIP:
			t.Fatalf("lease IP mismatch on %v", lease.IP)
		case !lease.Expiry.IsZero():
			t.Fatalf("lease %v should not have expiry date", lease)
		}
	}

	for i := 0; i < len(db.leases); i++ {
		lease := db.NextAvailableLease(fmt.Sprintf("%d", i))
		assert.NotNil(t, lease)
		assert.Equal(t, common.LeaseOffered, lease.State)
		targetIP := binary.BigEndian.Uint32(start) + uint32(i)
		assert.Equal(t, targetIP, binary.BigEndian.Uint32(lease.IP))
		db.AcceptLease(lease, time.Minute*5)
		assert.Equal(t, common.LeaseActive, db.leases[i].State)
	}
}

func Test_OfferAndAcceptLeaseOnce_Pass(t *testing.T) {
	// This test doesn't use DataSource, so no mock needed
	db := basicDB()
	lease := db.NextAvailableLease("test")
	assert.True(t, lease.IP.Equal(net.ParseIP("10.0.0.1")), "incorrect IP offered")
	assert.False(t, lease.Expiry.IsZero(), "should not have 0 expiry")
	assert.Equal(t, common.LeaseOffered, lease.State, "lease state should be LeaseOffered but is", lease.State)
	assert.Equal(t, "test", lease.ClientId, "lease clientId should be test but is", lease.ClientId)

	db.AcceptLease(lease, time.Hour)

	resultLease := db.GetLease(lease.ClientId)
	assert.NotNil(t, resultLease, "should have correspoding lease for client, but got nil")
	assert.True(t, resultLease.IP.Equal(lease.IP), "returned lease has IP mismatch, expected", lease.IP, "but got", resultLease.IP)
	assert.Equal(t, common.LeaseActive, resultLease.State, "state should be active, but got", resultLease.State)
}

func Test_OfferLeaseTwiceAndAcceptLeaseOnce_Pass(t *testing.T) {
	// This test doesn't use DataSource, so no mock needed
	db := basicDB()
	db.NextAvailableLease("test")
	lease := db.NextAvailableLease("test")
	assert.True(t, lease.IP.Equal(net.ParseIP("10.0.0.1")), "incorrect IP offered")
	assert.False(t, lease.Expiry.IsZero(), "should not have 0 expiry")
	assert.Equal(t, common.LeaseOffered, lease.State, "lease state should be LeaseOffered but is", lease.State)
	assert.Equal(t, "test", lease.ClientId, "lease clientId should be test but is", lease.ClientId)

	db.AcceptLease(lease, time.Hour)

	resultLease := db.GetLease(lease.ClientId)
	assert.NotNil(t, resultLease, "should have correspoding lease for client, but got nil")
	assert.True(t, resultLease.IP.Equal(lease.IP), "returned lease has IP mismatch, expected", lease.IP, "but got", resultLease.IP)
	assert.Equal(t, common.LeaseActive, resultLease.State, "state should be active, but got", resultLease.State)
}

func Test_OfferLeaseTwiceAndAcceptLeaseAfterExpiryForNewClient_Pass(t *testing.T) {
	// This test doesn't use DataSource, so no mock needed
	db := basicDB()
	db.NextAvailableLease("asd")
	db.leases[0].Expiry = time.Now().Add(time.Second * -1)
	lease := db.NextAvailableLease("test")
	assert.True(t, lease.IP.Equal(net.ParseIP("10.0.0.1")), "incorrect IP offered")
	assert.False(t, lease.Expiry.IsZero(), "should not have 0 expiry")
	assert.Equal(t, common.LeaseOffered, lease.State, "lease state should be LeaseOffered but is", lease.State)
	assert.Equal(t, "test", lease.ClientId, "lease clientId should be test but is", lease.ClientId)

	db.AcceptLease(lease, time.Hour)

	resultLease := db.GetLease(lease.ClientId)
	assert.NotNil(t, resultLease, "should have correspoding lease for client, but got nil")
	assert.True(t, resultLease.IP.Equal(lease.IP), "returned lease has IP mismatch, expected", lease.IP, "but got", resultLease.IP)
	assert.Equal(t, common.LeaseActive, resultLease.State, "state should be active, but got", resultLease.State)
}

func Test_GetExipiredLease(t *testing.T) {
	// This test doesn't use DataSource, so no mock needed
	db := basicDB()
	db.NextAvailableLease("test")
	db.leases[0].Expiry = time.Now().Add(time.Second * -1)
	assert.Nil(t, db.GetLease("test"), "lease has expired and a nil should be returned")
	assert.Empty(t, db.leases[1].ClientId, "client id should be zeroed")
	assert.Empty(t, db.leases[1].Hostname, "hostname should be zeroed")
	assert.Empty(t, db.leases[1].Expiry, "expiry should be zeroed")
}

func Test_SaveAndReadLeases(t *testing.T) {
	// Set up mock
	mockDS, cleanup := setupMockDataSource(t)
	defer cleanup()

	db := basicDB()

	// Prepare test data - fill up all leases
	var expectedLeases []common.Lease
	for i := range db.leases {
		lease := db.NextAvailableLease(fmt.Sprintf("%d", i))
		db.AcceptLease(lease, time.Hour)

		// Add to expected leases (including reserved leases from ReservedLeases())
		expectedLeases = append(expectedLeases, common.Lease{
			ClientId: lease.ClientId,
			IP:       lease.IP,
			State:    common.LeaseActive,
			Hostname: lease.Hostname,
			Expiry:   lease.Expiry,
		})
	}

	// Set expectation for PersistLeases call
	mockDS.EXPECT().
		PersistLeases(gomock.Any()).
		Return(nil).
		Times(1)

	// Test PersistLeases
	fileName := "test-leases" // We don't actually use this since it's mocked
	err := db.PeristLeases(fileName)
	assert.NoError(t, err, "unexpected error while saving leases")

	// Set expectation for ListLeases call - return the leases we expect to load
	mockDS.EXPECT().
		ListLeases().
		Return(expectedLeases, nil).
		Times(1)

	// Test LoadLeases
	err = db.LoadLeases(fileName, time.Hour)
	assert.NoError(t, err, "unexpected error while loading leases")

	// Verify that leases were loaded correctly
	for i, lease := range db.leases {
		targetIP := binary.BigEndian.Uint32(start) + uint32(i)
		assert.Equal(t, targetIP, binary.BigEndian.Uint32(lease.IP), "incorrect IP assigned")
		assert.Equal(t, fmt.Sprintf("%d", i), lease.ClientId, "mismatched client ID")
		assert.Equal(t, common.LeaseActive, lease.State, "incorrect lease state")
	}
}

// New test to specifically test PersistLeases with mock verification
func Test_PersistLeases_CallsDataSource(t *testing.T) {
	mockDS, cleanup := setupMockDataSource(t)
	defer cleanup()

	db := basicDB()

	// Add some leases
	lease1 := db.NextAvailableLease("client1")
	db.AcceptLease(lease1, time.Hour)

	lease2 := db.NextAvailableLease("client2")
	db.AcceptLease(lease2, time.Hour)

	// Add a reserved lease
	db.ReserveLease("reserved-client", net.ParseIP("10.0.0.100"))

	// Set expectation with a custom matcher to verify the leases
	mockDS.EXPECT().
		PersistLeases(gomock.Any()).
		Do(func(leases []common.Lease) {
			// Verify we got the expected number of leases
			// Should include all active leases + reserved leases
			assert.GreaterOrEqual(t, len(leases), 3, "should have at least active + reserved leases")

			// Find and verify our specific leases exist
			foundClient1 := false
			foundClient2 := false
			foundReserved := false

			for _, lease := range leases {
				switch lease.ClientId {
				case "client1":
					foundClient1 = true
					assert.Equal(t, common.LeaseActive, lease.State)
				case "client2":
					foundClient2 = true
					assert.Equal(t, common.LeaseActive, lease.State)
				case "reserved-client":
					foundReserved = true
					assert.Equal(t, common.LeaseReserved, lease.State)
				}
			}

			assert.True(t, foundClient1, "should find client1 lease")
			assert.True(t, foundClient2, "should find client2 lease")
			assert.True(t, foundReserved, "should find reserved lease")
		}).
		Return(nil).
		Times(1)

	// Call PersistLeases
	err := db.PeristLeases("test-file")
	assert.NoError(t, err)
}

// New test for LoadLeases error handling
func Test_LoadLeases_HandlesDataSourceError(t *testing.T) {
	mockDS, cleanup := setupMockDataSource(t)
	defer cleanup()

	db := basicDB()

	// Set expectation to return an error
	mockDS.EXPECT().
		ListLeases().
		Return(nil, fmt.Errorf("database connection failed")).
		Times(1)

	// Test LoadLeases error handling
	err := db.LoadLeases("test-file", time.Hour)
	assert.Error(t, err, "should return error from datasource")
	assert.Contains(t, err.Error(), "database connection failed")
}

// New test for LoadLeases with empty result
func Test_LoadLeases_WithEmptyResult(t *testing.T) {
	mockDS, cleanup := setupMockDataSource(t)
	defer cleanup()

	db := basicDB()

	// Set expectation to return empty slice
	mockDS.EXPECT().
		ListLeases().
		Return([]common.Lease{}, nil).
		Times(1)

	// Test LoadLeases with empty result
	err := db.LoadLeases("test-file", time.Hour)
	assert.NoError(t, err, "should handle empty lease list without error")

	// Verify all leases are still in their initial state
	for _, lease := range db.leases {
		assert.Equal(t, common.LeaseAvailable, lease.State, "lease should remain available")
		assert.Empty(t, lease.ClientId, "client ID should be empty")
	}
}

func basicDB() *LeasePool {
	return NewLeasePool(start, end)
}
