package parsing

import "time"

// ParsedEvent is the internal intermediate representation produced by parsers.
// It is NOT OCSF — it is a format-specific structured extraction.
type ParsedEvent struct {
	Timestamp *time.Time     `json:"timestamp,omitempty"`
	EventType string         `json:"event_type"`
	Fields    map[string]any `json:"fields"`
}
