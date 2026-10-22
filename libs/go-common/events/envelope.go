// Package events defines the envelope every Kafka message uses
// (libs/events/schemas/envelope.v1.json).
package events

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var eventTypeRe = regexp.MustCompile(`^[a-z_]+(\.[a-z_]+)+\.v[0-9]+$`)

// Envelope wraps every event payload.
type Envelope struct {
	EventID       uuid.UUID         `json:"event_id"`
	EventType     string            `json:"event_type"`
	SchemaVersion int               `json:"schema_version"`
	OccurredAt    time.Time         `json:"occurred_at"`
	UserID        uuid.UUID         `json:"user_id"`
	Producer      string            `json:"producer"`
	TraceContext  map[string]string `json:"trace_context,omitempty"`
	Payload       json.RawMessage   `json:"payload"`
}

// New builds an envelope with a time-ordered UUIDv7 id and the current UTC time.
func New(eventType string, schemaVersion int, userID uuid.UUID, producer string, payload any) (Envelope, error) {
	if !eventTypeRe.MatchString(eventType) {
		return Envelope{}, fmt.Errorf("event type %q is not of the form domain.action.vN", eventType)
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, fmt.Errorf("marshal payload: %w", err)
	}
	id, err := uuid.NewV7()
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{
		EventID:       id,
		EventType:     eventType,
		SchemaVersion: schemaVersion,
		OccurredAt:    time.Now().UTC(),
		UserID:        userID,
		Producer:      producer,
		Payload:       raw,
	}, nil
}
