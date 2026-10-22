package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/weezy2005/atlas-ai/libs/go-common/health"
	"github.com/weezy2005/atlas-ai/libs/go-common/metrics"
)

func TestEndpoints(t *testing.T) {
	up := []health.Check{{Name: "redis", Fn: func(context.Context) error { return nil }}}
	down := []health.Check{{Name: "redis", Fn: func(context.Context) error { return errors.New("refused") }}}

	cases := []struct {
		name   string
		checks []health.Check
		path   string
		code   int
		body   string
	}{
		{"health", down, "/health", http.StatusOK, `"service":"gateway"`},
		{"ready", up, "/ready", http.StatusOK, `"status":"ready"`},
		{"not ready", down, "/ready", http.StatusServiceUnavailable, `"redis":"error: refused"`},
		{"metrics", up, "/metrics", http.StatusOK, "go_goroutines"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			newRouter(metrics.New(service), tc.checks).ServeHTTP(rec, httptest.NewRequest("GET", tc.path, nil))
			if rec.Code != tc.code {
				t.Fatalf("code = %d, want %d", rec.Code, tc.code)
			}
			if !strings.Contains(rec.Body.String(), tc.body) {
				t.Fatalf("body %q missing %q", rec.Body.String(), tc.body)
			}
		})
	}
}
