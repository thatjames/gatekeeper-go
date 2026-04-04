package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	content := `DHCP:
  Interface: eth0
  StartAddr: 10.0.0.2
  EndAddr: 10.0.0.99
Web:
  Address: :8085
  WebURL: https://gatekeeper.example.com
  Prometheus: true
DNS:
  UpstreamServers:
    - 1.1.1.1
  Interface: eth0
  Port: 53
`

	tmpFile, err := os.CreateTemp("", "config-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	Config = &ConfigInstance{}
	err = LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if Config.Web.Address != ":8085" {
		t.Errorf("expected Web.Address ':8085', got '%s'", Config.Web.Address)
	}
	if Config.Web.WebURL != "https://gatekeeper.example.com" {
		t.Errorf("expected Web.WebURL 'https://gatekeeper.example.com', got '%s'", Config.Web.WebURL)
	}
	if !Config.Web.Prometheus {
		t.Error("expected Web.Prometheus to be true")
	}
	if Config.DNS.Port != 53 {
		t.Errorf("expected DNS.Port 53, got %d", Config.DNS.Port)
	}
	if len(Config.DNS.UpstreamServers) != 1 || Config.DNS.UpstreamServers[0] != "1.1.1.1" {
		t.Errorf("unexpected DNS.UpstreamServers: %v", Config.DNS.UpstreamServers)
	}
}

func TestLoadConfigOIDC(t *testing.T) {
	content := `Web:
  Address: :8085
Auth:
  AuthType: oidc
  IssuerURL: https://accounts.google.com
  ClientID: test-client
  ClientSecretVar: GOOGLE_CLIENT_SECRET
  RedirectURL: https://example.com/auth/callback
  Scopes:
    - openid
    - email
`

	tmpFile, err := os.CreateTemp("", "config-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	Config = &ConfigInstance{}
	err = LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	oidcAuth, ok := Config.Auth.(*OIDCAuth)
	if !ok {
		t.Fatal("expected Auth to be *OIDCAuth")
	}
	if oidcAuth.AuthType != "oidc" {
		t.Errorf("expected AuthType 'oidc', got '%s'", oidcAuth.AuthType)
	}
	if oidcAuth.IssuerURL != "https://accounts.google.com" {
		t.Errorf("expected IssuerURL, got '%s'", oidcAuth.IssuerURL)
	}
	if oidcAuth.ClientID != "test-client" {
		t.Errorf("expected ClientID, got '%s'", oidcAuth.ClientID)
	}
	if len(oidcAuth.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(oidcAuth.Scopes))
	}
}

func TestLoadConfigHTPasswd(t *testing.T) {
	content := `Web:
  Address: :8085
Auth:
  AuthType: htpasswd
  HTPasswdFile: /path/to/.htpasswd
`

	tmpFile, err := os.CreateTemp("", "config-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	Config = &ConfigInstance{}
	err = LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	htAuth, ok := Config.Auth.(*HTTPAuth)
	if !ok {
		t.Fatal("expected Auth to be *HTTPAuth")
	}
	if htAuth.AuthType != "htpasswd" {
		t.Errorf("expected AuthType 'htpasswd', got '%s'", htAuth.AuthType)
	}
	if htAuth.HTPasswdFile != "/path/to/.htpasswd" {
		t.Errorf("expected HTPasswdFile, got '%s'", htAuth.HTPasswdFile)
	}
}

func TestLoadConfigUnknownAuth(t *testing.T) {
	content := `Auth:
  AuthType: unknown
`

	tmpFile, err := os.CreateTemp("", "config-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	Config = &ConfigInstance{}
	err = LoadConfig(tmpFile.Name())
	if err == nil {
		t.Error("expected error for unknown AuthType")
	}
}

func TestBaseAuthType(t *testing.T) {
	auth := BaseAuth{AuthType: "test"}
	if auth.Type() != "test" {
		t.Errorf("expected 'test', got '%s'", auth.Type())
	}
}

func TestConfigInstanceString(t *testing.T) {
	Config = &ConfigInstance{
		Web: &Web{Address: ":8085"},
		DNS: &DNS{Port: 53},
	}
	str := Config.String()
	if str == "" {
		t.Error("expected non-empty string")
	}
	if !contains(str, "Web") {
		t.Error("expected string to contain 'Web'")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
