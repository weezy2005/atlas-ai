package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func decode(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("not JSON: %q", buf.String())
	}
	return m
}

func TestLineHasServiceTraceAndUser(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, "gateway", "info")

	ctx, span := sdktrace.NewTracerProvider().Tracer("t").Start(context.Background(), "op")
	defer span.End()
	ctx = WithUserID(ctx, "u-1")

	log.InfoContext(ctx, "routed", "route", "/portfolio")
	m := decode(t, &buf)

	want := map[string]any{
		"service":  "gateway",
		"msg":      "routed",
		"level":    "info",
		"user_id":  "u-1",
		"trace_id": span.SpanContext().TraceID().String(),
		"route":    "/portfolio",
	}
	for k, v := range want {
		if m[k] != v {
			t.Errorf("%s = %v, want %v", k, m[k], v)
		}
	}
	if _, ok := m["ts"]; !ok {
		t.Error("missing ts")
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	New(&buf, "x", "warn").Info("hidden")
	if buf.Len() != 0 {
		t.Fatalf("info logged at warn level: %s", buf.String())
	}
}
