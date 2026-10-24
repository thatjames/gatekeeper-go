package v1

import (
	"net"
	"net/http"
	"testing"

	"github.com/thatjames/gatekeeper-go/internal/dhcp"
	"github.com/thatjames/gatekeeper-go/internal/service"
)

func TestUpdateLeaseOntoAnotherReservationRefused(t *testing.T) {
	ht := newHandlerTest(t)
	pool := service.GetService[*dhcp.DHCPServer](service.DHCP).LeaseDB()
	if err := pool.ReserveLease("02:00:00:00:00:01", net.ParseIP("10.0.0.150").To4()); err != nil {
		t.Fatal(err)
	}
	if err := pool.ReserveLease("02:00:00:00:00:02", net.ParseIP("10.0.0.151").To4()); err != nil {
		t.Fatal(err)
	}
	before := ht.file(t)

	code, answer := ht.do(t, http.MethodPut, "/dhcp/leases", DhcpLeaseRequest{ClientId: "02:00:00:00:00:02", IP: "10.0.0.150"})
	if code != http.StatusBadRequest || answer["error"] == nil {
		t.Fatalf("expected 400 with an error, got %d %v", code, answer)
	}
	if ip := pool.GetLease("02:00:00:00:00:02").IP; !ip.Equal(net.ParseIP("10.0.0.151")) {
		t.Errorf("the reservation moved to %s", ip)
	}
	if ht.file(t) != before {
		t.Error("the config file changed")
	}

	if code, _ := ht.do(t, http.MethodPut, "/dhcp/leases", DhcpLeaseRequest{ClientId: "02:00:00:00:00:02", IP: "10.0.0.152"}); code != http.StatusOK {
		t.Errorf("a free IP: expected 200, got %d", code)
	}
}
