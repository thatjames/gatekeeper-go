package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/thatjames/gatekeeper-go/internal/config"
	v1 "github.com/thatjames/gatekeeper-go/internal/web/v1"
)

func TestInitDefaultAddress(t *testing.T) {
	cfg := &config.Web{
		Address: "",
		WebURL:  "http://localhost:8085",
	}

	if cfg.Address == "" {
		cfg.Address = ":8085"
	}

	if cfg.Address != ":8085" {
		t.Errorf("expected ':8085', got '%s'", cfg.Address)
	}
}

func TestInitDefaultWebURL(t *testing.T) {
	cfg := &config.Web{
		Address: ":8085",
		WebURL:  "",
	}

	if cfg.WebURL == "" {
		cfg.WebURL = "http://localhost:5173"
	}

	if cfg.WebURL != "http://localhost:5173" {
		t.Errorf("expected 'http://localhost:5173', got '%s'", cfg.WebURL)
	}
}

func TestSpaMiddlewareExists(t *testing.T) {
	middleware := spaMiddleware()
	if middleware == nil {
		t.Error("expected spaMiddleware to not be nil")
	}
}

func TestMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("GATEKEEPER_JWT_SECRET", "test-secret")
	if err := v1.LoadJWTSecret(); err != nil {
		t.Fatal(err)
	}
	oldConfig := config.Config
	t.Cleanup(func() { config.Config = oldConfig })

	get := func(r *gin.Engine, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}
	token, err := v1.CreateAuthToken("admin", v1.AdminRole)
	if err != nil {
		t.Fatal(err)
	}

	on := &config.Web{WebURL: "http://localhost:5173", Prometheus: true}
	config.Config = &config.ConfigInstance{Web: on}
	r := newRouter(on)
	if w := get(r, ""); w.Code != http.StatusUnauthorized {
		t.Errorf("no token: expected 401, got %d", w.Code)
	}
	if w := get(r, token); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "# HELP") {
		t.Errorf("with token: expected 200 with metrics, got %d", w.Code)
	}

	off := &config.Web{WebURL: "http://localhost:5173"}
	config.Config = &config.ConfigInstance{Web: off}
	if w := get(newRouter(off), token); w.Code != http.StatusNotFound {
		t.Errorf("Prometheus off: expected 404, got %d", w.Code)
	}
}

func TestCORSOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Web{WebURL: "https://gatekeeper.example.com", AllowedOrigins: []string{"http://localhost:5173"}}
	oldConfig := config.Config
	t.Cleanup(func() { config.Config = oldConfig })
	config.Config = &config.ConfigInstance{Web: cfg}
	r := newRouter(cfg)

	for origin, allowed := range map[string]bool{
		"https://gatekeeper.example.com":  true,
		"http://localhost:5173":           true,
		"https://gatekeeper.thatjames.nl": false,
		"https://evil.example.com":        false,
	} {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		req.Header.Set("Origin", origin)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		got := w.Header().Get("Access-Control-Allow-Origin")
		if allowed && got != origin {
			t.Errorf("%s: expected Access-Control-Allow-Origin %q, got %q", origin, origin, got)
		}
		if !allowed && got != "" {
			t.Errorf("%s: expected no CORS headers, got Access-Control-Allow-Origin %q", origin, got)
		}
	}
}
