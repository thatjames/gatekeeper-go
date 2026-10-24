package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/thatjames/gatekeeper-go/internal/config"
)

func TestRoleFromClaims(t *testing.T) {
	mapped := map[string]string{"gk-admins": "admin", "gk-viewers": "viewer"}
	for _, tc := range []struct {
		name   string
		claims string
		claim  string
		roles  map[string]string
		role   Role
		ok     bool
	}{
		{"default claim, values are roles", `{"groups":["viewer"]}`, "", nil, ViewerRole, true},
		{"admin wins", `{"groups":["viewer","admin"]}`, "", nil, AdminRole, true},
		{"single string", `{"groups":"admin"}`, "", nil, AdminRole, true},
		{"mapped group", `{"groups":["staff","gk-viewers"]}`, "", mapped, ViewerRole, true},
		{"mapped admin", `{"groups":["gk-admins"]}`, "", mapped, AdminRole, true},
		{"map hides plain names", `{"groups":["admin"]}`, "", mapped, "", false},
		{"other claim", `{"roles":["admin"],"groups":["viewer"]}`, "roles", nil, AdminRole, true},
		{"no known value", `{"groups":["staff"]}`, "", nil, "", false},
		{"no claim", `{"email":"a@example.com"}`, "", nil, "", false},
		{"not strings", `{"groups":[1,true]}`, "", nil, "", false},
	} {
		var claims map[string]any
		if err := json.Unmarshal([]byte(tc.claims), &claims); err != nil {
			t.Fatal(err)
		}
		role, ok := RoleFromClaims(claims, tc.claim, tc.roles)
		if role != tc.role || ok != tc.ok {
			t.Errorf("%s: expected %q %v, got %q %v", tc.name, tc.role, tc.ok, role, ok)
		}
	}
}

// roleRouter serves the DHCP and DNS route tables and two plain routes
// behind the auth and role middleware.
func roleRouter(t *testing.T) *gin.Engine {
	t.Helper()
	oldConfig, oldSecret := config.Config, jwtSecret
	config.Config = &config.ConfigInstance{Web: &config.Web{}}
	jwtSecret = []byte("test-secret")
	t.Cleanup(func() { config.Config, jwtSecret = oldConfig, oldSecret })

	r := gin.New()
	protected := r.Group("/api/v1", AuthMiddleware(), RoleMiddleware())
	setupDHCPRoutes(protected)
	setupDNSRoutes(protected)
	protected.GET("/read", func(c *gin.Context) { c.Status(http.StatusOK) })
	protected.POST("/change", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func callAs(t *testing.T, r *gin.Engine, role Role, method, path string) int {
	t.Helper()
	token, err := CreateAuthToken("someone", role)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestViewerRefusedOnEveryChangeRoute(t *testing.T) {
	r := roleRouter(t)
	changes := 0
	for _, route := range r.Routes() {
		if route.Method == http.MethodGet || route.Path == "/api/v1/change" {
			continue
		}
		changes++
		if code := callAs(t, r, ViewerRole, route.Method, route.Path); code != http.StatusForbidden {
			t.Errorf("viewer %s %s: expected 403, got %d", route.Method, route.Path, code)
		}
	}
	if changes < 10 {
		t.Errorf("expected the DHCP and DNS change routes, found %d", changes)
	}
}

func TestRolesOnPlainRoutes(t *testing.T) {
	r := roleRouter(t)
	for _, tc := range []struct {
		role   Role
		method string
		path   string
		code   int
	}{
		{ViewerRole, http.MethodGet, "/api/v1/read", http.StatusOK},
		{ViewerRole, http.MethodPost, "/api/v1/change", http.StatusForbidden},
		{AdminRole, http.MethodGet, "/api/v1/read", http.StatusOK},
		{AdminRole, http.MethodPost, "/api/v1/change", http.StatusOK},
		{"user", http.MethodGet, "/api/v1/read", http.StatusForbidden},
		{"", http.MethodGet, "/api/v1/read", http.StatusForbidden},
	} {
		if code := callAs(t, r, tc.role, tc.method, tc.path); code != tc.code {
			t.Errorf("%q %s %s: expected %d, got %d", tc.role, tc.method, tc.path, tc.code, code)
		}
	}
}

func TestPasswordLoginIsAdmin(t *testing.T) {
	r := setupRouter(t, &config.Web{HTPasswdFile: writeHTPasswd(t, "admin", "pw")})
	w := postLogin(r, "admin", "pw")
	var resp struct{ Token string }
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &resp) != nil {
		t.Fatalf("login: expected 200 with a token, got %d %s", w.Code, w.Body.String())
	}
	claims, err := ParseAuthToken(resp.Token)
	if err != nil || claims.Role != AdminRole {
		t.Errorf("expected an admin token, got %+v %v", claims, err)
	}
}
