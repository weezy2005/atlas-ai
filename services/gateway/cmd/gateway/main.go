// Command gateway is the single entry point for client traffic: it authenticates,
// rate-limits and routes to backend services. Phase 1 serves health endpoints only.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/weezy2005/atlas-ai/libs/go-common/config"
	"github.com/weezy2005/atlas-ai/libs/go-common/health"
	"github.com/weezy2005/atlas-ai/libs/go-common/logging"
	"github.com/weezy2005/atlas-ai/libs/go-common/metrics"
	"github.com/weezy2005/atlas-ai/libs/go-common/server"
	"github.com/weezy2005/atlas-ai/libs/go-common/tracing"
)

const service = "gateway"

func main() {
	probe := flag.Bool("healthcheck", false, "probe the local /health endpoint and exit")
	flag.Parse()
	port := config.String("PORT", "8080")
	if *probe {
		if err := health.Probe("http://127.0.0.1:" + port + "/health"); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	log := logging.New(os.Stdout, service, config.String("LOG_LEVEL", "info"))
	if err := run(log, port); err != nil {
		log.Error("exit", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(log *slog.Logger, port string) error {
	shutdown, err := tracing.Setup(context.Background(), service, config.String("OTEL_EXPORTER_OTLP_ENDPOINT", ""))
	if err != nil {
		return err
	}
	defer func() { _ = shutdown(context.Background()) }()

	redisURL, err := config.Required("REDIS_URL")
	if err != nil {
		return err
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("REDIS_URL: %w", err)
	}
	rdb := redis.NewClient(opts)
	defer func() { _ = rdb.Close() }()

	checks := []health.Check{
		{Name: "redis", Fn: func(ctx context.Context) error { return rdb.Ping(ctx).Err() }},
	}
	return server.Run(log, ":"+port, newRouter(metrics.New(service), checks), 15*time.Second)
}

func newRouter(m *metrics.Metrics, checks []health.Check) http.Handler {
	r := chi.NewRouter()
	// No middleware.RealIP: it trusts client-supplied X-Forwarded-For (IP spoofing).
	// Client IP for rate limiting will come only from the trusted ingress (Phase 2).
	r.Use(middleware.RequestID, middleware.Recoverer, m.Middleware)
	r.Get("/health", health.LivenessHandler(service))
	r.Get("/ready", health.ReadinessHandler(checks, 2*time.Second))
	r.Handle("/metrics", m.Handler())
	return otelhttp.NewHandler(r, service, otelhttp.WithFilter(func(r *http.Request) bool {
		switch r.URL.Path {
		case "/health", "/ready", "/metrics":
			return false
		}
		return true
	}))
}
