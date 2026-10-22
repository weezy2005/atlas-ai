package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestLabelsUseRoutePattern(t *testing.T) {
	m := New("gateway")
	r := chi.NewRouter()
	r.Use(m.Middleware)
	r.Get("/accounts/{id}", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })
	r.Handle("/metrics", m.Handler())

	for _, p := range []string{"/accounts/1", "/accounts/2", "/nope"} {
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", p, nil))
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	body, _ := io.ReadAll(rec.Body)
	out := string(body)

	for _, want := range []string{
		`http_requests_total{method="GET",route="/accounts/{id}",service="gateway",status="418"} 2`,
		`route="unmatched"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s", want)
		}
	}
	if strings.Contains(out, `route="/metrics"`) {
		t.Error("/metrics scrapes should not be recorded")
	}
}
