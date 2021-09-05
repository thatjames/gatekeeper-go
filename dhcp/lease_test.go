package dhcp

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	start = net.IP{10, 0, 0, 1}
	end   = net.IP{10, 0, 0, 95}
)

func Test_ReservedLease_Pass(t *testing.T) {
	db := NewLeaseDB(net.ParseIP("10.0.0.1"), net.ParseIP("10.0.0.10"))
	db.ReserveLease("test", net.ParseIP("10.0.0.100"))
	lease := db.GetLease("test")
	if lease == nil {
		t.Fatal("expected reserved lease, none returned")
	}

	assert.Equal(t, net.IP{10, 0, 0, 100}, lease.IP.To4(), "expected address 10.0.0.100 but got", lease.IP.To4())
	assert.Equal(t, LeaseReserved, lease.State, "expected LeaseReserved but got", lease.State)
	assert.Equal(t, "test", lease.ClientId, "expected test but got", lease.ClientId)

}

func Test_InitSuccess(t *testing.T) {
	db := basicDB()
	assert.Equal(t, int(binary.BigEndian.Uint32(end)-binary.BigEndian.Uint32(start)), len(db.leases), "should have prepopulated range of addresses")
	for i, lease := range db.leases {
		targetIP := binary.BigEndian.Uint32(start) + uint32(i)
		switch {
		case lease.State != LeaseAvailable:
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
		assert.Equal(t, LeaseOffered, lease.State)
		targetIP := binary.BigEndian.Uint32(start) + uint32(i)
		assert.Equal(t, targetIP, binary.BigEndian.Uint32(lease.IP))
		db.AcceptLease(lease, time.Minute*5)
		assert.Equal(t, LeaseActive, db.leases[i].State)
	}
}

func Test_OfferAndAcceptLeaseOnce_Pass(t *testing.T) {
	db := basicDB()
	lease := db.NextAvailableLease("test")
	assert.True(t, lease.IP.Equal(net.ParseIP("10.0.0.1")), "incorrect IP offered")
	assert.False(t, lease.Expiry.IsZero(), "should not have 0 expiry")
	assert.Equal(t, LeaseOffered, lease.State, "lease state should be LeaseOffered but is", lease.State)
	assert.Equal(t, "test", lease.ClientId, "lease clientId should be test but is", lease.ClientId)

	db.AcceptLease(lease, time.Hour)

	resultLease := db.GetLease(lease.ClientId)
	assert.NotNil(t, resultLease, "should have correspoding lease for client, but got nil")
	assert.True(t, resultLease.IP.Equal(lease.IP), "returned lease has IP mismatch, expected", lease.IP, "but got", resultLease.IP)
	assert.Equal(t, LeaseActive, resultLease.State, "state should be active, but got", resultLease.State)
}

func Test_OfferLeaseTwiceAndAcceptLeaseOnce_Pass(t *testing.T) {
	db := basicDB()
	db.NextAvailableLease("test")
	lease := db.NextAvailableLease("test")
	assert.True(t, lease.IP.Equal(net.ParseIP("10.0.0.1")), "incorrect IP offered")
	assert.False(t, lease.Expiry.IsZero(), "should not have 0 expiry")
	assert.Equal(t, LeaseOffered, lease.State, "lease state should be LeaseOffered but is", lease.State)
	assert.Equal(t, "test", lease.ClientId, "lease clientId should be test but is", lease.ClientId)

	db.AcceptLease(lease, time.Hour)

	resultLease := db.GetLease(lease.ClientId)
	assert.NotNil(t, resultLease, "should have correspoding lease for client, but got nil")
	assert.True(t, resultLease.IP.Equal(lease.IP), "returned lease has IP mismatch, expected", lease.IP, "but got", resultLease.IP)
	assert.Equal(t, LeaseActive, resultLease.State, "state should be active, but got", resultLease.State)
}

func Test_OfferLeaseTwiceAndAcceptLeaseAfterExpiryForNewClient_Pass(t *testing.T) {
	db := basicDB()
	db.NextAvailableLease("asd")
	db.leases[0].Expiry = time.Now().Add(time.Second * -1)
	lease := db.NextAvailableLease("test")
	assert.True(t, lease.IP.Equal(net.ParseIP("10.0.0.1")), "incorrect IP offered")
	assert.False(t, lease.Expiry.IsZero(), "should not have 0 expiry")
	assert.Equal(t, LeaseOffered, lease.State, "lease state should be LeaseOffered but is", lease.State)
	assert.Equal(t, "test", lease.ClientId, "lease clientId should be test but is", lease.ClientId)

	db.AcceptLease(lease, time.Hour)

	resultLease := db.GetLease(lease.ClientId)
	assert.NotNil(t, resultLease, "should have correspoding lease for client, but got nil")
	assert.True(t, resultLease.IP.Equal(lease.IP), "returned lease has IP mismatch, expected", lease.IP, "but got", resultLease.IP)
	assert.Equal(t, LeaseActive, resultLease.State, "state should be active, but got", resultLease.State)
}

func Test_GetExipiredLease(t *testing.T) {
	db := basicDB()
	db.NextAvailableLease("test")
	db.leases[0].Expiry = time.Now().Add(time.Second * -1)
	assert.Nil(t, db.GetLease("test"), "lease has expired and a nil should be returned")
	assert.Equal(t, new(Lease), db.leases[0], "lease should be zeroed")
}

func Test_SaveAndReadLeases(t *testing.T) {
	db := basicDB()
	for i := range db.leases {
		db.AcceptLease(db.NextAvailableLease(fmt.Sprintf("%d", i)), time.Hour)
	}

	fileName := path.Join(os.TempDir(), "test-leases")
	if err := db.PeristLeases(fileName); err != nil {
		t.Fatal("unexpected error while saving leases:", err.Error())
	}

	if err := db.LoadLeases(fileName, time.Hour); err != nil {
		t.Fatal("unexpected error while loading leases:", err.Error())
	}

	for i, lease := range db.leases {
		targetIP := binary.BigEndian.Uint32(start) + uint32(i)
		assert.Equal(t, targetIP, binary.BigEndian.Uint32(lease.IP), "incorrect IP assigned")
		assert.Equal(t, fmt.Sprintf("%d", i), lease.ClientId, "mismatched client ID")
		assert.Equal(t, LeaseActive, lease.State, "incorrect lease state")
	}
}

func basicDB() *LeaseDB {
	return NewLeaseDB(start, end)
}
