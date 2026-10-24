package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thatjames/gatekeeper-go/internal/config"
)

func TestFeaturesWithoutAuth(t *testing.T) {
	r := setupRouter(t, &config.Web{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/features", nil))
	if w.Code != http.StatusOK || w.Body.String() != "[]" {
		t.Errorf("expected 200 [], got %d %s", w.Code, w.Body.String())
	}
}
