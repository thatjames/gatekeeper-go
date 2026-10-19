package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dhcp"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dns"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/service"
)

const testConfig = `DNS:
  LocalDomains:
    nas.example.lan: 10.0.0.2
  BlockLists:
    - https://lists.example.com/a.txt
  BlockedDomains:
    - a.example.com
    - b.example.com
    - c.example.com
DHCP:
  StartAddr: 10.0.0.100
  EndAddr: 10.0.0.200
  ReservedAddresses:
    02:00:00:00:00:01: 10.0.0.150
`

// fakeResolver records what the handlers change in the running resolver.
type fakeResolver struct {
	lock         sync.Mutex
	localDomains map[string]net.IP
	blocked      map[string]bool
}

func (r *fakeResolver) Resolve(string, dns.DNSType) ([]*dns.DNSRecord, []*dns.DNSRecord, error) {
	return nil, nil, nil
}

func (r *fakeResolver) AddLocalDomain(domain string, ip net.IP) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if ip.Equal(net.IPv4zero) {
		return errors.New("invalid IP address")
	}
	r.localDomains[domain] = ip
	return nil
}

func (r *fakeResolver) DeleteLocalDomain(domain string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.localDomains, domain)
}

func (r *fakeResolver) AddBlocklistEntries(entries []string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for _, entry := range entries {
		r.blocked[entry] = true
	}
}

func (r *fakeResolver) DeleteBlocklistEntry(domain string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.blocked, domain)
}

func (r *fakeResolver) FlushBlocklist() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.blocked = make(map[string]bool)
}

// countingFetcher serves one entry per URL and counts the fetches; URLs ending in /missing fail.
type countingFetcher struct {
	lock    sync.Mutex
	fetches map[string]int
}

func (f *countingFetcher) Fetch(url string) ([]string, error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.fetches[url]++
	if filepath.Base(url) == "missing" {
		return nil, errors.New("404 Not Found")
	}
	return []string{"from-" + filepath.Base(url)}, nil
}

type handlerTest struct {
	path     string
	resolver *fakeResolver
	fetcher  *countingFetcher
	router   *gin.Engine
}

func newHandlerTest(t *testing.T) *handlerTest {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(testConfig), 0o644); err != nil {
		t.Fatal(err)
	}
	oldConfig := config.Config
	t.Cleanup(func() { config.Config = oldConfig })
	config.Config = nil
	if err := config.LoadConfig(path); err != nil {
		t.Fatal(err)
	}

	ht := &handlerTest{
		path:     path,
		resolver: &fakeResolver{localDomains: make(map[string]net.IP), blocked: make(map[string]bool)},
		fetcher:  &countingFetcher{fetches: make(map[string]int)},
		router:   gin.New(),
	}
	service.Register(dns.NewDNSServerWithOpts(dns.DNSServerOpts{}, ht.resolver, ht.fetcher), service.DNS)
	service.Register(dhcp.NewDHCPServerFromConfig(config.Config.DHCP), service.DHCP)

	r := ht.router
	r.GET("/dns/config", getDNSConfig)
	r.GET("/dns/local-domains", getLocalDomains)
	r.POST("/dns/local-domains", addLocalDomain)
	r.PUT("/dns/local-domains/:domain", updateLocalDomain)
	r.DELETE("/dns/local-domains/:domain", deleteLocalDomain)
	r.POST("/dns/blocklist", addBlocklist)
	r.DELETE("/dns/blocklist/:id", deleteBlocklist)
	r.PUT("/dns/blockeddomains", addBlockedDomain)
	r.DELETE("/dns/blockeddomains/:id", deleteBlockedDomain)
	r.DELETE("/dhcp/leases/:clientId", deleteLease)
	r.POST("/dhcp/leases/reserve", reserveLease)
	r.PUT("/dhcp/leases", updateLease)
	return ht
}

// do sends the request and checks that exactly one JSON value came back.
func (ht *handlerTest) do(t *testing.T, method, path string, body any) (int, map[string]any) {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reader = bytes.NewReader(data)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ht.router.ServeHTTP(w, req)

	dec := json.NewDecoder(w.Body)
	var answer map[string]any
	// Errorf, not Fatalf: TestHandlersConcurrent calls this from other goroutines.
	if err := dec.Decode(&answer); err != nil {
		t.Errorf("%s %s: answer is not one JSON object: %v", method, path, err)
	}
	if dec.More() {
		t.Errorf("%s %s: answered more than once", method, path)
	}
	return w.Code, answer
}

func (ht *handlerTest) file(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(ht.path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// breakSave makes every later UpdateConfig fail: the config path becomes a directory.
func (ht *handlerTest) breakSave(t *testing.T) {
	t.Helper()
	if err := os.Remove(ht.path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(ht.path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestAddLocalDomainLiveErrorChangesNothing(t *testing.T) {
	ht := newHandlerTest(t)
	before := ht.file(t)

	code, answer := ht.do(t, http.MethodPost, "/dns/local-domains", LocalDomainRequest{Domain: "bad.example.lan", IP: "0.0.0.0"})
	if code != http.StatusBadRequest || answer["error"] == nil {
		t.Fatalf("expected 400 with an error, got %d %v", code, answer)
	}
	if _, ok := config.Config.DNS.LocalDomains["bad.example.lan"]; ok {
		t.Error("the domain reached the config")
	}
	if ht.file(t) != before {
		t.Error("the config file changed")
	}
}

func TestAddLocalDomain(t *testing.T) {
	ht := newHandlerTest(t)
	code, answer := ht.do(t, http.MethodPost, "/dns/local-domains", LocalDomainRequest{Domain: "pi.example.lan", IP: "10.0.0.3"})
	if code != http.StatusOK || answer["pi.example.lan"] != "10.0.0.3" || answer["nas.example.lan"] != "10.0.0.2" {
		t.Fatalf("expected 200 with both domains, got %d %v", code, answer)
	}
	if ht.resolver.localDomains["pi.example.lan"] == nil {
		t.Error("the domain did not reach the resolver")
	}
}

func TestLocalDomainSaveFailureRollsBack(t *testing.T) {
	ht := newHandlerTest(t)
	ht.resolver.localDomains["nas.example.lan"] = net.ParseIP("10.0.0.2").To4()
	ht.breakSave(t)

	if code, _ := ht.do(t, http.MethodPost, "/dns/local-domains", LocalDomainRequest{Domain: "pi.example.lan", IP: "10.0.0.3"}); code != http.StatusInternalServerError {
		t.Fatalf("add: expected 500, got %d", code)
	}
	if code, _ := ht.do(t, http.MethodPut, "/dns/local-domains/nas.example.lan", LocalDomainRequest{Domain: "files.example.lan", IP: "10.0.0.4"}); code != http.StatusInternalServerError {
		t.Fatalf("update: expected 500, got %d", code)
	}
	if code, _ := ht.do(t, http.MethodDelete, "/dns/local-domains/nas.example.lan", nil); code != http.StatusInternalServerError {
		t.Fatalf("delete: expected 500, got %d", code)
	}

	want := map[string]string{"nas.example.lan": "10.0.0.2"}
	if fmt.Sprint(config.Config.DNS.LocalDomains) != fmt.Sprint(want) {
		t.Errorf("config: expected %v, got %v", want, config.Config.DNS.LocalDomains)
	}
	if len(ht.resolver.localDomains) != 1 || !ht.resolver.localDomains["nas.example.lan"].Equal(net.ParseIP("10.0.0.2")) {
		t.Errorf("resolver: expected only nas.example.lan -> 10.0.0.2, got %v", ht.resolver.localDomains)
	}
}

func TestAddBlocklistFetchesOnceAndAnswers(t *testing.T) {
	ht := newHandlerTest(t)
	code, answer := ht.do(t, http.MethodPost, "/dns/blocklist", BlocklistRequest{Url: "https://lists.example.com/b.txt"})
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d %v", code, answer)
	}
	if fmt.Sprint(answer["blocklist"]) != "[https://lists.example.com/a.txt https://lists.example.com/b.txt]" {
		t.Errorf("expected both lists in the answer, got %v", answer)
	}
	if ht.fetcher.fetches["https://lists.example.com/b.txt"] != 1 || ht.fetcher.fetches["https://lists.example.com/a.txt"] != 0 {
		t.Errorf("expected one fetch of the new list and none of the old, got %v", ht.fetcher.fetches)
	}
	if !ht.resolver.blocked["from-b.txt"] {
		t.Error("the new list's entries are not blocked")
	}
}

func TestAddBlocklistFetchErrorChangesNothing(t *testing.T) {
	ht := newHandlerTest(t)
	before := ht.file(t)
	if code, _ := ht.do(t, http.MethodPost, "/dns/blocklist", BlocklistRequest{Url: "https://lists.example.com/missing"}); code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", code)
	}
	if len(config.Config.DNS.BlockLists) != 1 || ht.file(t) != before {
		t.Errorf("the config changed: %v", config.Config.DNS.BlockLists)
	}
}

func TestDeleteBlocklistAnswers(t *testing.T) {
	ht := newHandlerTest(t)
	ht.resolver.blocked["from-a.txt"] = true
	code, answer := ht.do(t, http.MethodDelete, "/dns/blocklist/0", nil)
	if code != http.StatusOK || answer["blockedDomains"] == nil {
		t.Fatalf("expected 200 with the DNS config, got %d %v", code, answer)
	}
	if len(config.Config.DNS.BlockLists) != 0 || ht.resolver.blocked["from-a.txt"] {
		t.Errorf("the list is still configured or blocked: %v %v", config.Config.DNS.BlockLists, ht.resolver.blocked)
	}
}

func TestBlockedDomainSaveFailureRollsBack(t *testing.T) {
	ht := newHandlerTest(t)
	ht.resolver.blocked["a.example.com"] = true
	// Spare capacity lets an in-place append or delete reach the old slice.
	config.Config.DNS.BlockedDomains = append(make([]string, 0, 8), "a.example.com", "b.example.com", "c.example.com")
	config.Config.DNS.BlockLists = append(make([]string, 0, 8), "https://lists.example.com/a.txt", "https://lists.example.com/b.txt")
	ht.breakSave(t)

	if code, _ := ht.do(t, http.MethodDelete, "/dns/blockeddomains/0", nil); code != http.StatusInternalServerError {
		t.Fatalf("delete: expected 500, got %d", code)
	}
	if code, _ := ht.do(t, http.MethodPut, "/dns/blockeddomains", BlocklistRequest{Url: "d.example.com"}); code != http.StatusInternalServerError {
		t.Fatalf("add: expected 500, got %d", code)
	}
	if code, _ := ht.do(t, http.MethodDelete, "/dns/blocklist/0", nil); code != http.StatusInternalServerError {
		t.Fatalf("delete list: expected 500, got %d", code)
	}

	if want := []string{"a.example.com", "b.example.com", "c.example.com"}; !slices.Equal(config.Config.DNS.BlockedDomains, want) {
		t.Errorf("blocked domains: expected %v, got %v", want, config.Config.DNS.BlockedDomains)
	}
	if want := []string{"https://lists.example.com/a.txt", "https://lists.example.com/b.txt"}; !slices.Equal(config.Config.DNS.BlockLists, want) {
		t.Errorf("blocklists: expected %v, got %v", want, config.Config.DNS.BlockLists)
	}
	if !ht.resolver.blocked["a.example.com"] || ht.resolver.blocked["d.example.com"] {
		t.Errorf("resolver changed: %v", ht.resolver.blocked)
	}
}

// TestHandlersConcurrent is for go test -race: handlers that read and write the config run at once.
func TestHandlersConcurrent(t *testing.T) {
	ht := newHandlerTest(t)
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			host := fmt.Sprintf("host%d.example.lan", i)
			client := fmt.Sprintf("02:00:00:00:01:%02x", i)
			ht.do(t, http.MethodPost, "/dns/local-domains", LocalDomainRequest{Domain: host, IP: fmt.Sprintf("10.0.0.%d", 10+i)})
			ht.do(t, http.MethodGet, "/dns/local-domains", nil)
			ht.do(t, http.MethodPut, "/dns/blockeddomains", BlocklistRequest{Url: host})
			ht.do(t, http.MethodGet, "/dns/config", nil)
			ht.do(t, http.MethodPost, "/dns/blocklist", BlocklistRequest{Url: "https://lists.example.com/" + host})
			ht.do(t, http.MethodPost, "/dhcp/leases/reserve", DhcpLeaseRequest{ClientId: client, IP: fmt.Sprintf("10.0.0.%d", 160+i)})
			ht.do(t, http.MethodDelete, "/dhcp/leases/"+client, nil)
			ht.do(t, http.MethodDelete, "/dns/local-domains/"+host, nil)
		}()
	}
	wg.Wait()

	if len(config.Config.DNS.BlockedDomains) != 3+8 || len(config.Config.DNS.BlockLists) != 1+8 {
		t.Errorf("lost an update: %d blocked domains, %d blocklists", len(config.Config.DNS.BlockedDomains), len(config.Config.DNS.BlockLists))
	}
	if len(config.Config.DNS.LocalDomains) != 1 || len(config.Config.DHCP.ReservedAddresses) != 1 {
		t.Errorf("lost an update: local domains %v, reservations %v", config.Config.DNS.LocalDomains, config.Config.DHCP.ReservedAddresses)
	}
}
