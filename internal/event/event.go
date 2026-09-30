package event

import "time"

// Event represents the common internal event envelope
type Event struct {
	ApplicationID  string    `json:"application_id"`
	Protocol       string    `json:"protocol"`
	ReceivedAt     time.Time `json:"received_at"`
	DetectedFormat string    `json:"detected_format,omitempty"`
	Raw            string    `json:"raw"`
}
