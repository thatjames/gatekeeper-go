package dns

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHTTPBlocklistFetcher_Fetch_HTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("127.0.0.1 localhost\n"))
	}))
	defer server.Close()

	fetcher := NewHTTPBlocklistFetcher()
	hosts, err := fetcher.Fetch(server.URL)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hosts) == 0 {
		t.Error("expected hosts to be returned")
	}
}

func TestHTTPBlocklistFetcher_Fetch_File(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "blocklist.txt")
	os.WriteFile(tmpFile, []byte("127.0.0.1 localhost\n"), 0644)

	fetcher := NewHTTPBlocklistFetcher()
	hosts, err := fetcher.Fetch(tmpFile)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hosts) == 0 {
		t.Error("expected hosts to be returned")
	}
}

func TestHTTPBlocklistFetcher_Fetch_InvalidURL(t *testing.T) {
	fetcher := NewHTTPBlocklistFetcher()
	_, err := fetcher.Fetch("invalid-url")

	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestHTTPBlocklistFetcher_Fetch_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	fetcher := NewHTTPBlocklistFetcher()
	_, err := fetcher.Fetch(server.URL)

	if err == nil {
		t.Error("expected error for HTTP 404")
	}
}

func TestHTTPBlocklistFetcher_Fetch_InvalidFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("singlefield"))
	}))
	defer server.Close()

	fetcher := NewHTTPBlocklistFetcher()
	hosts, err := fetcher.Fetch(server.URL)

	if err != ErrInvalidBlocklistFormat {
		t.Errorf("expected ErrInvalidBlocklistFormat, got %v (hosts: %v)", err, hosts)
	}
}

func TestHTTPBlocklistFetcher_Fetch_NonExistentFile(t *testing.T) {
	fetcher := NewHTTPBlocklistFetcher()
	_, err := fetcher.Fetch("/non/existent/file.txt")

	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestFileBlocklistFetcher_Fetch(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "blocklist.txt")
	os.WriteFile(tmpFile, []byte("127.0.0.1 localhost\n192.168.1.1 router.local\n"), 0644)

	fetcher := &FileBlocklistFetcher{}
	hosts, err := fetcher.Fetch(tmpFile)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hosts) == 0 {
		t.Error("expected hosts to be returned")
	}
}

func TestFileBlocklistFetcher_Fetch_NonExistent(t *testing.T) {
	fetcher := &FileBlocklistFetcher{}
	_, err := fetcher.Fetch("/non/existent/file.txt")

	if err == nil {
		t.Error("expected error for non-existent file")
	}
}
