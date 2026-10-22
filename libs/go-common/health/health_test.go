package health

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func ok(context.Context) error { return nil }

func TestReadyWhenAllChecksPass(t *testing.T) {
	rec := httptest.NewRecorder()
	ReadinessHandler([]Check{{"postgres", ok}, {"redis", ok}}, time.Second)(rec, httptest.NewRequest("GET", "/ready", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
}

func TestNotReadyNamesFailingDependency(t *testing.T) {
	down := func(context.Context) error { return errors.New("connection refused") }
	rec := httptest.NewRecorder()
	ReadinessHandler([]Check{{"postgres", ok}, {"kafka", down}}, time.Second)(rec, httptest.NewRequest("GET", "/ready", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("code = %d, want 503", rec.Code)
	}
	var body struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "not_ready" || body.Checks["postgres"] != "ok" || body.Checks["kafka"] != "error: connection refused" {
		t.Fatalf("unexpected body %+v", body)
	}
}

func TestChecksAreBoundedByTimeout(t *testing.T) {
	hang := func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }
	start := time.Now()
	res := Run(context.Background(), []Check{{"a", hang}, {"b", hang}}, 50*time.Millisecond)
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("checks ran serially or ignored timeout: %v", elapsed)
	}
	if res["a"] == "ok" || res["b"] == "ok" {
		t.Fatalf("hung checks reported ok: %v", res)
	}
}

func TestTCPCheck(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	if err := TCPCheck("kafka", addr+",other:9092").Fn(context.Background()); err != nil {
		t.Fatalf("open port: %v", err)
	}
	ln.Close()
	if err := TCPCheck("kafka", addr).Fn(context.Background()); err == nil {
		t.Fatal("closed port reported reachable")
	}
}

func TestProbe(t *testing.T) {
	srv := httptest.NewServer(LivenessHandler("gw"))
	defer srv.Close()
	if err := Probe(srv.URL); err != nil {
		t.Fatal(err)
	}
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(503) }))
	defer bad.Close()
	if err := Probe(bad.URL); err == nil {
		t.Fatal("503 accepted")
	}
}
