// Package health serves /health (liveness) and /ready (dependency readiness).
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Check verifies one dependency is reachable.
type Check struct {
	Name string
	Fn   func(ctx context.Context) error
}

// TCPCheck dials the first address in a comma-separated list. Used for Kafka
// until services hold a real client (Phase 3).
func TCPCheck(name, addrs string) Check {
	addr := strings.TrimSpace(strings.Split(addrs, ",")[0])
	return Check{Name: name, Fn: func(ctx context.Context) error {
		var d net.Dialer
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err != nil {
			return err
		}
		return conn.Close()
	}}
}

// Run executes all checks concurrently, each bounded by timeout.
func Run(ctx context.Context, checks []Check, timeout time.Duration) map[string]string {
	results := make(map[string]string, len(checks))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, c := range checks {
		wg.Add(1)
		go func(c Check) {
			defer wg.Done()
			cctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			status := "ok"
			if err := c.Fn(cctx); err != nil {
				status = fmt.Sprintf("error: %v", err)
			}
			mu.Lock()
			results[c.Name] = status
			mu.Unlock()
		}(c)
	}
	wg.Wait()
	return results
}

// LivenessHandler always reports ok while the process can serve HTTP.
func LivenessHandler(service string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": service})
	}
}

// ReadinessHandler returns 200 only when every check passes, else 503 naming the failures.
func ReadinessHandler(checks []Check, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := Run(r.Context(), checks, timeout)
		status, code := "ready", http.StatusOK
		for _, v := range results {
			if v != "ok" {
				status, code = "not_ready", http.StatusServiceUnavailable
				break
			}
		}
		writeJSON(w, code, map[string]any{"status": status, "checks": results})
	}
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

// Probe performs a GET against url and fails unless it returns 200. It backs the
// binary's -healthcheck flag, since distroless images have no curl.
func Probe(url string) error {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url) //nolint:noctx // one-shot CLI probe
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
