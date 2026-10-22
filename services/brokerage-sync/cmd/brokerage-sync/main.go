// Command brokerage-sync connects broker accounts over read-only OAuth and syncs
// positions and trades into Kafka. Phase 1 serves health endpoints only.
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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/weezy2005/atlas-ai/libs/go-common/config"
	"github.com/weezy2005/atlas-ai/libs/go-common/health"
	"github.com/weezy2005/atlas-ai/libs/go-common/logging"
	"github.com/weezy2005/atlas-ai/libs/go-common/metrics"
	"github.com/weezy2005/atlas-ai/libs/go-common/server"
	"github.com/weezy2005/atlas-ai/libs/go-common/tracing"
)

const service = "brokerage-sync"

func main() {
	probe := flag.Bool("healthcheck", false, "probe the local /health endpoint and exit")
	flag.Parse()
	port := config.String("PORT", "8000")
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
	ctx := context.Background()
	shutdown, err := tracing.Setup(ctx, service, config.String("OTEL_EXPORTER_OTLP_ENDPOINT", ""))
	if err != nil {
		return err
	}
	defer func() { _ = shutdown(ctx) }()

	dbURL, err := config.Required("DATABASE_URL")
	if err != nil {
		return err
	}
	redisURL, err := config.Required("REDIS_URL")
	if err != nil {
		return err
	}
	kafka, err := config.Required("KAFKA_BOOTSTRAP_SERVERS")
	if err != nil {
		return err
	}

	// pgxpool connects lazily, so the service starts (and reports not-ready)
	// even while Postgres is still coming up.
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("DATABASE_URL: %w", err)
	}
	defer pool.Close()
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("REDIS_URL: %w", err)
	}
	rdb := redis.NewClient(opts)
	defer func() { _ = rdb.Close() }()

	checks := []health.Check{
		{Name: "postgres", Fn: pool.Ping},
		{Name: "redis", Fn: func(ctx context.Context) error { return rdb.Ping(ctx).Err() }},
		health.TCPCheck("kafka", kafka),
	}
	return server.Run(log, ":"+port, newRouter(metrics.New(service), checks), 30*time.Second)
}

func newRouter(m *metrics.Metrics, checks []health.Check) http.Handler {
	r := chi.NewRouter()
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
