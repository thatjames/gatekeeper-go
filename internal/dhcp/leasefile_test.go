package dhcp

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testMAC(i int) string {
	return fmt.Sprintf("02:00:00:00:%02x:%02x", i>>8, i&0xff)
}

// activePool fills a pool from 10.0.0.1 to 10.0.1.255 with n active leases of synthetic clients.
func activePool(t *testing.T, n int, ttl time.Duration) *LeasePool {
	t.Helper()
	pool := NewLeasePool(net.ParseIP("10.0.0.1").To4(), net.ParseIP("10.0.1.255").To4())
	for i := 0; i < n; i++ {
		lease := pool.NextAvailableLease(testMAC(i))
		if lease == nil {
			t.Fatalf("no lease for client %d", i)
		}
		pool.SetHostname(lease, fmt.Sprintf("host-%d", i))
		pool.AcceptLease(lease, ttl)
	}
	return pool
}

func badFiles(t *testing.T, file string) []string {
	t.Helper()
	matches, err := filepath.Glob(file + ".bad-*")
	if err != nil {
		t.Fatal(err)
	}
	return matches
}

func TestLeaseFile_300LeasesRoundTrip(t *testing.T) {
	file := filepath.Join(t.TempDir(), "sub", "leases")
	saved := activePool(t, 300, 90*time.Minute)
	if err := saved.SaveLeases(file); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(file); err != nil {
		t.Fatal(err)
	} else if info.Mode().Perm() != 0600 {
		t.Errorf("lease file mode %v, want 0600", info.Mode().Perm())
	}

	loaded := NewLeasePool(net.ParseIP("10.0.0.1").To4(), net.ParseIP("10.0.1.255").To4())
	if err := loaded.LoadLeases(file); err != nil {
		t.Fatal(err)
	}
	if got := len(loaded.ActiveLeases()); got != 300 {
		t.Fatalf("loaded %d active leases, want 300", got)
	}
	for i := 0; i < 300; i++ {
		want, got := saved.GetLease(testMAC(i)), loaded.GetLease(testMAC(i))
		if got == nil {
			t.Fatalf("client %d lost its lease", i)
		}
		if !got.IP.Equal(want.IP) || got.Hostname != want.Hostname || got.State != LeaseActive {
			t.Errorf("client %d: got %s, want %s", i, got, want)
		}
		if got.Expiry.Unix() != want.Expiry.Unix() {
			t.Errorf("client %d: expiry %v, want %v", i, got.Expiry, want.Expiry)
		}
	}
}

func TestLeaseFile_MissingOrEmptyLoadsNothing(t *testing.T) {
	dir := t.TempDir()
	pool := NewLeasePool(net.ParseIP("10.0.0.1").To4(), net.ParseIP("10.0.0.10").To4())
	if err := pool.LoadLeases(filepath.Join(dir, "missing")); err != nil {
		t.Errorf("missing file: %v", err)
	}
	empty := filepath.Join(dir, "empty")
	if err := os.WriteFile(empty, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := pool.LoadLeases(empty); err != nil {
		t.Errorf("empty file: %v", err)
	}
	if len(badFiles(t, empty)) != 0 {
		t.Error("empty file was moved aside")
	}
}

func TestLeaseFile_RejectedFilesMovedAside(t *testing.T) {
	good := activePool(t, 3, time.Hour).encodeLeases(time.Now())
	oldFormat := append([]byte{1, 17}, "02:00:00:00:00:01"...)
	oldFormat = append(oldFormat, 4)
	oldFormat = append(oldFormat, "host"...)
	oldFormat = append(oldFormat, 10, 0, 0, 5, byte(LeaseActive))
	badState := append([]byte{}, good...)
	badState[len(badState)-9] = byte(LeaseOffered)

	for name, data := range map[string][]byte{
		"truncated":      good[:len(good)-5],
		"old format":     oldFormat,
		"trailing bytes": append(append([]byte{}, good...), 0),
		"huge count":     append([]byte("GKLF\x02\xff\xff\xff\xff"), good[leaseFileHeader:]...),
		"other version":  append([]byte("GKLF\x03"), good[len(leaseFileMagic)+1:]...),
		"unknown state":  badState,
	} {
		t.Run(name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "leases")
			if err := os.WriteFile(file, data, 0600); err != nil {
				t.Fatal(err)
			}
			pool := NewLeasePool(net.ParseIP("10.0.0.1").To4(), net.ParseIP("10.0.0.10").To4())
			err := pool.LoadLeases(file)
			if !errors.Is(err, errBadLeaseFile) {
				t.Fatalf("got %v, want a bad lease file error", err)
			}
			if _, err := os.Stat(file); !os.IsNotExist(err) {
				t.Error("rejected file still in place")
			}
			if moved := badFiles(t, file); len(moved) != 1 {
				t.Errorf("want one moved-aside file, got %v", moved)
			}
			if got := len(pool.ActiveLeases()); got != 0 {
				t.Errorf("pool has %d active leases after a rejected file", got)
			}
		})
	}
}

func TestLeaseFile_LoadDropsExpiredAndOutOfRange(t *testing.T) {
	file := filepath.Join(t.TempDir(), "leases")
	pool := activePool(t, 2, time.Hour)
	pool.GetLease(testMAC(0)).Expiry = time.Now().Add(-time.Minute)
	if err := os.WriteFile(file, pool.encodeLeases(time.Now().Add(-2*time.Hour)), 0600); err != nil {
		t.Fatal(err)
	}

	loaded := NewLeasePool(net.ParseIP("10.0.0.2").To4(), net.ParseIP("10.0.0.10").To4())
	if err := loaded.LoadLeases(file); err != nil {
		t.Fatal(err)
	}
	// Client 0 (10.0.0.1) is expired and outside the new range; client 1 (10.0.0.2) is kept.
	if got := loaded.ActiveLeases(); len(got) != 1 || got[0].ClientId != testMAC(1) {
		t.Errorf("got %v, want only client 1", got)
	}
}

func TestLeaseFile_Reservations(t *testing.T) {
	const printer, laptop, gone = "02:00:00:00:01:01", "02:00:00:00:01:02", "02:00:00:00:01:03"
	file := filepath.Join(t.TempDir(), "leases")
	saved := NewLeasePool(net.ParseIP("10.0.0.1").To4(), net.ParseIP("10.0.0.10").To4())
	if err := saved.ReserveLease(printer, net.ParseIP("10.0.0.5").To4()); err != nil {
		t.Fatal(err)
	}
	if err := saved.ReserveLease(gone, net.ParseIP("10.0.0.6").To4()); err != nil {
		t.Fatal(err)
	}
	saved.SetHostname(saved.GetLease(printer), "printer")
	saved.SetHostname(saved.GetLease(gone), "gone")
	for _, lease := range saved.leases {
		if lease.IP.Equal(net.ParseIP("10.0.0.7")) {
			lease.ClientId, lease.State, lease.Expiry = laptop, LeaseActive, time.Now().Add(time.Hour)
		}
	}
	if err := saved.SaveLeases(file); err != nil {
		t.Fatal(err)
	}

	// The config now reserves the printer as before and 10.0.0.7, the laptop's IP, for another client.
	loaded := NewLeasePool(net.ParseIP("10.0.0.1").To4(), net.ParseIP("10.0.0.10").To4())
	if err := loaded.ReserveLease(printer, net.ParseIP("10.0.0.5").To4()); err != nil {
		t.Fatal(err)
	}
	if err := loaded.ReserveLease("02:00:00:00:01:04", net.ParseIP("10.0.0.7").To4()); err != nil {
		t.Fatal(err)
	}
	if err := loaded.LoadLeases(file); err != nil {
		t.Fatal(err)
	}
	if got := loaded.GetLease(printer); got == nil || got.Hostname != "printer" || got.State != LeaseReserved {
		t.Errorf("printer reservation: got %v", got)
	}
	if got := loaded.GetLease(gone); got != nil {
		t.Errorf("a reservation the config dropped came back: %v", got)
	}
	if got := loaded.GetLease(laptop); got != nil {
		t.Errorf("an active lease on a reserved IP came back: %v", got)
	}
	if got := len(loaded.ReservedLeases()); got != 2 {
		t.Errorf("got %d reservations, want 2", got)
	}
}

// A lease ACKed while the writer runs is on disk without a Stop: what survives a SIGKILL.
func TestLeaseFile_WrittenOnChange(t *testing.T) {
	const mac = "02:00:00:00:02:01"
	file := filepath.Join(t.TempDir(), "leases")
	z := testServer("10.0.0.1", "10.0.0.10")
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		z.issuedLeases.PersistOnChange(file, 20*time.Millisecond, stop)
	}()
	defer func() { close(stop); <-done }()

	offer := z.handleRequest(testRequest(t, DHCPDiscover, mac, nil))
	if offer == nil {
		t.Fatal("no OFFER")
	}
	ack := z.handleRequest(testRequest(t, DHCPRequest, mac, Options{
		OptionRequestedIPAddress: offer.YIAddr().To4(),
		OptionHostname:           []byte("laptop"),
	}))
	if ack == nil || DHCPMessageType(ParseOptions(ack)[OptionDHCPMessageType][0]) != DHCPAck {
		t.Fatal("no ACK")
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		restarted := NewLeasePool(net.ParseIP("10.0.0.1").To4(), net.ParseIP("10.0.0.10").To4())
		err := restarted.LoadLeases(file)
		if err != nil {
			t.Fatal(err)
		}
		if lease := restarted.GetLease(mac); lease != nil {
			if !lease.IP.Equal(offer.YIAddr()) || lease.Hostname != "laptop" || lease.State != LeaseActive {
				t.Fatalf("restored %s", lease)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("the ACKed lease never reached the lease file")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
