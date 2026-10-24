package v1

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
)

func TestMissingHTPasswdFileIsGenerated(t *testing.T) {
	buf := captureLog(t)
	path := filepath.Join(t.TempDir(), "htpasswd")
	r := setupRouter(t, &config.Web{HTPasswdFile: path})

	if fallbackPassword != "" {
		t.Error("expected the generated file to be used, not the fallback password")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("expected mode 0600, got %o", info.Mode().Perm())
	}
	data, _ := os.ReadFile(path)
	if !strings.HasPrefix(string(data), "admin:$2") {
		t.Errorf("expected an admin line with a bcrypt hash, got %q", data)
	}

	match := regexp.MustCompile(`password: ([A-Za-z0-9]{16})`).FindStringSubmatch(buf.String())
	if match == nil {
		t.Fatalf("expected the password in the log:\n%s", buf.String())
	}
	password := match[1]
	if strings.Contains(string(data), password) {
		t.Error("the password is in the file in plain text")
	}
	if n := strings.Count(buf.String(), password); n != 1 {
		t.Errorf("expected the password logged once, got %d times", n)
	}
	if w := postLogin(r, "admin", password); w.Code != http.StatusOK {
		t.Errorf("login with the generated password: expected 200, got %d", w.Code)
	}
	if w := postLogin(r, "admin", password+"x"); w.Code != http.StatusUnauthorized {
		t.Errorf("wrong password: expected 401, got %d", w.Code)
	}
}

func TestExistingHTPasswdFileIsKept(t *testing.T) {
	path := writeHTPasswd(t, "admin", "kept")
	before, _ := os.ReadFile(path)
	wrote, err := ensureHTPasswdFile(path)
	after, _ := os.ReadFile(path)
	if wrote || err != nil || string(before) != string(after) {
		t.Errorf("expected the file left alone, got wrote=%v err=%v", wrote, err)
	}
	if wrote, err := ensureHTPasswdFile(""); wrote || err != nil {
		t.Errorf("unset path: expected nothing written, got wrote=%v err=%v", wrote, err)
	}
}
