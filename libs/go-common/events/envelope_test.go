package events

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const schemaDir = "../../events/schemas"

func compile(t *testing.T, name string) *jsonschema.Schema {
	t.Helper()
	c := jsonschema.NewCompiler()
	c.AssertFormat()
	s, err := c.Compile(filepath.Join(schemaDir, name+".json"))
	if err != nil {
		t.Fatalf("compile %s: %v", name, err)
	}
	return s
}

func asInstance(t *testing.T, v any) any {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	inst, err := jsonschema.UnmarshalJSON(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return inst
}

// The Go envelope must produce messages the shared JSON Schema accepts, so Go
// producers and Python consumers agree on the wire format.
func TestEnvelopeMatchesSchema(t *testing.T) {
	payload := map[string]any{
		"account_id":      uuid.NewString(),
		"broker_trade_id": "T-1",
		"symbol":          "AAPL",
		"side":            "buy",
		"quantity":        "10",
		"price":           "187.42",
		"executed_at":     "2026-10-22T14:30:00Z",
	}
	env, err := New("trade.received.v1", 1, uuid.New(), "brokerage-sync", payload)
	if err != nil {
		t.Fatal(err)
	}
	if env.EventID.Version() != 7 {
		t.Errorf("event id version = %d, want 7", env.EventID.Version())
	}
	if err := compile(t, "envelope.v1").Validate(asInstance(t, env)); err != nil {
		t.Fatalf("envelope: %v", err)
	}
	if err := compile(t, "trade.received.v1").Validate(asInstance(t, payload)); err != nil {
		t.Fatalf("payload: %v", err)
	}
}

func TestAllSchemasCompile(t *testing.T) {
	entries, err := os.ReadDir(schemaDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		compile(t, strings.TrimSuffix(e.Name(), ".json"))
	}
}

func TestRejectsUnversionedType(t *testing.T) {
	if _, err := New("trade.received", 1, uuid.New(), "x", map[string]any{}); err == nil {
		t.Fatal("expected error for unversioned event type")
	}
}
