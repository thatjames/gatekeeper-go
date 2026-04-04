package dhcp

import (
	"net"
	"testing"
	"time"
)

func TestLeaseStateString(t *testing.T) {
	tests := []struct {
		state    LeaseState
		expected string
	}{
		{LeaseAvailable, "Available"},
		{LeaseOffered, "Offered"},
		{LeaseReserved, "Reserved"},
		{LeaseActive, "Active"},
		{99, "unknown"},
	}

	for _, tt := range tests {
		result := tt.state.String()
		if result != tt.expected {
			t.Errorf("LeaseState(%d).String() = %s; want %s", tt.state, result, tt.expected)
		}
	}
}

func TestLeaseString(t *testing.T) {
	lease := &Lease{
		Hostname: "test-host",
		IP:       net.ParseIP("192.168.1.100"),
		ClientId: "00:11:22:33:44:55",
		State:    LeaseActive,
		Expiry:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	result := lease.String()
	if result == "" {
		t.Error("expected non-empty string")
	}
}

func TestLeaseClear(t *testing.T) {
	lease := &Lease{
		ClientId: "00:11:22:33:44:55",
		Hostname: "test-host",
		IP:       net.ParseIP("192.168.1.100"),
		State:    LeaseActive,
		Expiry:   time.Now().Add(time.Hour),
	}

	lease.Clear()

	if lease.ClientId != "" {
		t.Errorf("expected ClientId to be empty, got %s", lease.ClientId)
	}
	if lease.Hostname != "" {
		t.Errorf("expected Hostname to be empty, got %s", lease.Hostname)
	}
	if lease.State != LeaseAvailable {
		t.Errorf("expected State to be Available, got %v", lease.State)
	}
}

func TestNewLeasePool(t *testing.T) {
	start := net.ParseIP("192.168.1.100")
	end := net.ParseIP("192.168.1.110")

	pool := NewLeasePool(start, end)

	if pool == nil {
		t.Fatal("expected non-nil LeasePool")
	}
}

func TestLeasePoolGetLease(t *testing.T) {
	start := net.ParseIP("192.168.1.100")
	end := net.ParseIP("192.168.1.110")

	pool := NewLeasePool(start, end)

	lease := pool.GetLease("00:11:22:33:44:55")
	if lease != nil {
		t.Error("expected nil for non-existent lease")
	}
}

func TestLeasePoolReserveLease(t *testing.T) {
	start := net.ParseIP("192.168.1.100")
	end := net.ParseIP("192.168.1.110")

	pool := NewLeasePool(start, end)

	err := pool.ReserveLease("00:11:22:33:44:55", net.ParseIP("192.168.1.105"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lease := pool.GetLease("00:11:22:33:44:55")
	if lease == nil {
		t.Fatal("expected lease to exist")
	}
	if lease.State != LeaseReserved {
		t.Errorf("expected state Reserved, got %v", lease.State)
	}
}
