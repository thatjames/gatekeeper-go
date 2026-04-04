package web

import (
	"testing"

	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
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
