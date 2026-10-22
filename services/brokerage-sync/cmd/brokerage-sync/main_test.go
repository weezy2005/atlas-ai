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

func TestReadinessReportsEachDependency(t *testing.T) {
	checks := []health.Check{
		{Name: "postgres", Fn: func(context.Context) error { return nil }},
		{Name: "redis", Fn: func(context.Context) error { return nil }},
		{Name: "kafka", Fn: func(context.Context) error { return errors.New("dial tcp: refused") }},
	}
	rec := httptest.NewRecorder()
	newRouter(metrics.New(service), checks).ServeHTTP(rec, httptest.NewRequest("GET", "/ready", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("code = %d, want 503", rec.Code)
	}
	for _, want := range []string{`"postgres":"ok"`, `"redis":"ok"`, `"kafka":"error: dial tcp: refused"`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("body %s missing %s", rec.Body.String(), want)
		}
	}
}

func TestHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	newRouter(metrics.New(service), nil).ServeHTTP(rec, httptest.NewRequest("GET", "/health", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"service":"brokerage-sync"`) {
		t.Fatalf("got %d %s", rec.Code, rec.Body.String())
	}
}
