package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHealthHandler(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	healthHandler(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.Len() == 0 {
		t.Error("expected response body")
	}
}

func TestGetVersion(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/version", nil)

	getVersion(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestOIDCOptions_Defaults(t *testing.T) {
	opts := OIDCOptions{
		IssuerURL:    "https://accounts.google.com",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURL:  "https://example.com/auth/callback",
		Scopes:       []string{"openid", "profile"},
	}

	if opts.IssuerURL != "https://accounts.google.com" {
		t.Errorf("expected IssuerURL, got %s", opts.IssuerURL)
	}
	if opts.ClientID != "test-client" {
		t.Errorf("expected ClientID, got %s", opts.ClientID)
	}
	if len(opts.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(opts.Scopes))
	}
}

func TestOIDCConfig_Struct(t *testing.T) {
	config := OIDCConfig{}
	if config.Provider != nil {
		t.Error("expected nil provider")
	}
	if config.Verifier != nil {
		t.Error("expected nil verifier")
	}
}
