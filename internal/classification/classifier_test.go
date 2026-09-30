package classification

import (
	"testing"
	"time"

	"github.com/sih/log-platform/internal/parsing"
)

func TestClassifier_ClearHTTP(t *testing.T) {
	c := NewClassifier()
	ts := time.Now()
	pe := parsing.ParsedEvent{
		Timestamp: &ts,
		Fields: map[string]any{
			"method":   "GET",
			"path":     "/index.html",
			"protocol": "HTTP/1.1",
			"status":   200,
		},
	}
	result := c.Classify(pe)
	if result.Status != ClassificationConfident {
		t.Errorf("Expected CONFIDENT, got %s", result.Status)
	}
	if result.EventType != "HTTP_ACTIVITY" {
		t.Errorf("Expected HTTP_ACTIVITY, got %s", result.EventType)
	}
}

func TestClassifier_ClearAuthentication(t *testing.T) {
	c := NewClassifier()
	pe := parsing.ParsedEvent{
		Fields: map[string]any{
			"action": "login",
			"user":   "admin",
			"result": "success",
		},
	}
	result := c.Classify(pe)
	if result.Status != ClassificationConfident {
		t.Errorf("Expected CONFIDENT, got %s", result.Status)
	}
	if result.EventType != "AUTHENTICATION" {
		t.Errorf("Expected AUTHENTICATION, got %s", result.EventType)
	}
}

func TestClassifier_Unknown(t *testing.T) {
	c := NewClassifier()
	pe := parsing.ParsedEvent{
		Fields: map[string]any{
			"some_field": "some_value",
			"another":    42,
		},
	}
	result := c.Classify(pe)
	if result.Status != ClassificationUnknown {
		t.Errorf("Expected UNKNOWN, got %s", result.Status)
	}
}

func TestClassifier_InsufficientEvidence(t *testing.T) {
	c := NewClassifier()
	// Has method but no path or protocol — insufficient HTTP evidence
	pe := parsing.ParsedEvent{
		Fields: map[string]any{
			"method": "GET",
		},
	}
	result := c.Classify(pe)
	if result.Status != ClassificationUnknown {
		t.Errorf("Expected UNKNOWN for insufficient evidence, got %s", result.Status)
	}
}
