package parsing

import (
	"encoding/json"
	"fmt"
	"time"
)

// JSONParser uses encoding/json to parse arbitrary JSON log events.
// It preserves all source fields. Classification is separate.
type JSONParser struct{}

func (p *JSONParser) Name() string {
	return "JSON"
}

func (p *JSONParser) Parse(raw string) (ParsedEvent, error) {
	var fields map[string]any
	if err := json.Unmarshal([]byte(raw), &fields); err != nil {
		return ParsedEvent{}, fmt.Errorf("PARSE_ERROR: invalid JSON: %w", err)
	}

	// Attempt to extract a timestamp from common field names
	var ts *time.Time
	for _, key := range []string{"timestamp", "time", "@timestamp", "datetime", "ts"} {
		if v, ok := fields[key]; ok {
			if s, ok := v.(string); ok {
				for _, layout := range []string{
					time.RFC3339,
					time.RFC3339Nano,
					"2006-01-02T15:04:05-0700",
					"2006-01-02 15:04:05",
					"2006-01-02T15:04:05Z",
				} {
					if parsed, err := time.Parse(layout, s); err == nil {
						ts = &parsed
						break
					}
				}
			}
			break
		}
	}

	return ParsedEvent{
		Timestamp: ts,
		EventType: "", // Classification is separate
		Fields:    fields,
	}, nil
}
