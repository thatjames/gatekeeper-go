package v1

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
)

// captureLog sends log output to a buffer for the rest of the test.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	out, level := log.StandardLogger().Out, log.GetLevel()
	log.SetOutput(&buf)
	log.SetLevel(log.TraceLevel)
	t.Cleanup(func() {
		log.SetOutput(out)
		log.SetLevel(level)
	})
	return &buf
}

// setupRouter serves the v1 API with the given web config and a test secret.
func setupRouter(t *testing.T, web *config.Web) *gin.Engine {
	t.Helper()
	oldConfig, oldSecret := config.Config, jwtSecret
	config.Config = &config.ConfigInstance{Web: web}
	jwtSecret = []byte("test-secret")
	t.Cleanup(func() {
		config.Config, jwtSecret = oldConfig, oldSecret
	})
	r := gin.New()
	SetupV1Endpoints(r.Group("/api"))
	return r
}

func postLogin(r *gin.Engine, username, password string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(UserLoginRequest{Username: username, Password: password})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func writeHTPasswd(t *testing.T, username, password string) string {
	t.Helper()
	sum := sha1.Sum([]byte(password))
	path := filepath.Join(t.TempDir(), "htpasswd")
	line := username + ":{SHA}" + base64.StdEncoding.EncodeToString(sum[:]) + "\n"
	if err := os.WriteFile(path, []byte(line), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoginNeverLogsPassword(t *testing.T) {
	const good, bad = "correct-horse-battery", "wrong-staple-guess"
	buf := captureLog(t)
	r := setupRouter(t, &config.Web{HTPasswdFile: writeHTPasswd(t, "admin", good)})

	if w := postLogin(r, "admin", bad); w.Code != http.StatusUnauthorized {
		t.Errorf("wrong password: expected 401, got %d", w.Code)
	}
	if w := postLogin(r, "admin", good); w.Code != http.StatusOK {
		t.Errorf("right password: expected 200, got %d", w.Code)
	}

	logged := buf.String()
	if strings.Contains(logged, good) || strings.Contains(logged, bad) {
		t.Errorf("password in log output:\n%s", logged)
	}
	if !strings.Contains(logged, "admin") {
		t.Errorf("expected the username in log output:\n%s", logged)
	}
}

func TestFallbackLoginWorksTwice(t *testing.T) {
	buf := captureLog(t)
	r := setupRouter(t, &config.Web{HTPasswdFile: filepath.Join(t.TempDir(), "missing")})

	password := fallbackPassword
	if len(password) != 16 {
		t.Fatalf("expected a 16-character fallback password, got %q", password)
	}
	for i := range 2 {
		if w := postLogin(r, "admin", password); w.Code != http.StatusOK {
			t.Errorf("login %d: expected 200, got %d", i+1, w.Code)
		}
	}
	if w := postLogin(r, "admin", password+"x"); w.Code != http.StatusUnauthorized {
		t.Errorf("wrong password: expected 401, got %d", w.Code)
	}
	if w := postLogin(r, "root", password); w.Code != http.StatusUnauthorized {
		t.Errorf("wrong user: expected 401, got %d", w.Code)
	}
	if n := strings.Count(buf.String(), password); n != 1 {
		t.Errorf("expected the fallback password logged once, got %d times", n)
	}
}

func TestLoadJWTSecret(t *testing.T) {
	oldSecret := jwtSecret
	t.Cleanup(func() { jwtSecret = oldSecret })

	t.Setenv(jwtSecretVar, "")
	err := LoadJWTSecret()
	if !errors.Is(err, ErrNoJWTSecret) || !strings.Contains(err.Error(), "GATEKEEPER_JWT_SECRET") {
		t.Errorf("expected an error naming GATEKEEPER_JWT_SECRET, got %v", err)
	}

	t.Setenv(jwtSecretVar, "test-secret")
	if err := LoadJWTSecret(); err != nil {
		t.Errorf("expected no error with the secret set, got %v", err)
	}
}

func TestTokenWithoutSecretErrors(t *testing.T) {
	oldSecret := jwtSecret
	jwtSecret = nil
	t.Cleanup(func() { jwtSecret = oldSecret })

	if _, err := CreateAuthToken("admin"); !errors.Is(err, ErrNoJWTSecret) {
		t.Errorf("create: expected ErrNoJWTSecret, got %v", err)
	}
	if _, err := ParseAuthToken("x.y.z"); !errors.Is(err, ErrNoJWTSecret) {
		t.Errorf("parse: expected ErrNoJWTSecret, got %v", err)
	}
}

func TestTokenCookieFlags(t *testing.T) {
	oldConfig := config.Config
	t.Cleanup(func() { config.Config = oldConfig })

	for _, tc := range []struct {
		webURL string
		secure bool
	}{
		{"https://gatekeeper.example.com", true},
		{"http://localhost:5173", false},
	} {
		config.Config = &config.ConfigInstance{Web: &config.Web{WebURL: tc.webURL}}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		setTokenCookie(c, "token", int(tokenLife.Seconds()))

		cookies := w.Result().Cookies()
		if len(cookies) != 1 {
			t.Fatalf("%s: expected one cookie, got %d", tc.webURL, len(cookies))
		}
		cookie := cookies[0]
		if cookie.Name != "oauth_token" || !cookie.HttpOnly || cookie.Secure != tc.secure ||
			cookie.SameSite != http.SameSiteLaxMode || cookie.MaxAge != 7*24*60*60 {
			t.Errorf("%s: unexpected cookie %+v", tc.webURL, cookie)
		}
	}
}

func TestMeReturnsUser(t *testing.T) {
	r := setupRouter(t, &config.Web{})
	token, err := CreateAuthToken("alice")
	if err != nil {
		t.Fatal(err)
	}

	for name, set := range map[string]func(*http.Request){
		"bearer": func(req *http.Request) { req.Header.Set("Authorization", "Bearer "+token) },
		"cookie": func(req *http.Request) { req.AddCookie(&http.Cookie{Name: "oauth_token", Value: token}) },
	} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
		set(req)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var user User
		if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &user) != nil ||
			user.Username != "alice" || user.Role != UserRole {
			t.Errorf("%s: expected 200 with alice, got %d %s", name, w.Code, w.Body.String())
		}
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/me", nil))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("no token: expected 401, got %d", w.Code)
	}
}
