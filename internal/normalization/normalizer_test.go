package normalization

import (
	"testing"
	"time"

	"github.com/sih/log-platform/internal/classification"
	"github.com/sih/log-platform/internal/parsing"
)

func TestNormalizeHTTP(t *testing.T) {
	n := NewNormalizer()
	ts := time.Now()
	pe := parsing.ParsedEvent{
		Timestamp: &ts,
		Fields: map[string]any{
			"method":    "GET",
			"path":      "/index.html",
			"protocol":  "HTTP/1.1",
			"status":    200,
			"client_ip": "10.0.0.1",
			"bytes":     1024,
			"extra":     "keep me",
		},
	}
	cr := classification.ClassificationResult{
		Status:    classification.ClassificationConfident,
		EventType: "HTTP_ACTIVITY",
	}

	result, err := n.Normalize(pe, cr, "raw log here")
	if err != nil {
		t.Fatalf("Normalization failed: %v", err)
	}

	// Validate OCSF Schema mapping
	if result.HTTPRequest == nil || result.HTTPRequest.HTTPMethod != "GET" || result.HTTPRequest.URL.Path != "/index.html" {
		t.Errorf("HTTPRequest mapping failed")
	}
	if result.HTTPResponse == nil || result.HTTPResponse.Code != 200 || result.HTTPResponse.Size != 1024 {
		t.Errorf("HTTPResponse mapping failed")
	}
	if result.SrcEndpoint == nil || result.SrcEndpoint.IP != "10.0.0.1" {
		t.Errorf("SrcEndpoint mapping failed")
	}
	
	// Unmapped should only contain 'protocol' and 'extra'
	if _, ok := result.Unmapped["extra"]; !ok {
		t.Errorf("Expected 'extra' to be in unmapped")
	}
	if _, ok := result.Unmapped["method"]; ok {
		t.Errorf("Expected 'method' to be removed from unmapped")
	}

	// Validate standard fields
	if result.Metadata.Version != "1.7.0" {
		t.Errorf("Expected metadata.version 1.7.0")
	}
	if result.Time == 0 {
		t.Errorf("Expected non-zero time")
	}

	vr := Validate(result)
	if !vr.Valid {
		t.Errorf("OCSF validation failed: %v", vr.Errors)
	}
}

func TestNormalizeAuthentication(t *testing.T) {
	n := NewNormalizer()
	pe := parsing.ParsedEvent{
		Fields: map[string]any{
			"action":    "login",
			"user":      "admin",
			"result":    "success",
			"client_ip": "192.168.1.1",
			"random":    "data",
		},
	}
	cr := classification.ClassificationResult{
		Status:    classification.ClassificationConfident,
		EventType: "AUTHENTICATION",
	}

	result, err := n.Normalize(pe, cr, "raw auth log")
	if err != nil {
		t.Fatalf("Normalization failed: %v", err)
	}

	// Validate OCSF schema mapping
	if result.User == nil || result.User.Name != "admin" {
		t.Errorf("User mapping failed")
	}
	if result.SrcEndpoint == nil || result.SrcEndpoint.IP != "192.168.1.1" {
		t.Errorf("SrcEndpoint mapping failed")
	}

	// Validate unmapped
	if _, ok := result.Unmapped["random"]; !ok {
		t.Errorf("Expected 'random' in unmapped")
	}
	if _, ok := result.Unmapped["user"]; ok {
		t.Errorf("Expected 'user' to be removed from unmapped")
	}

	vr := Validate(result)
	if !vr.Valid {
		t.Errorf("OCSF validation failed: %v", vr.Errors)
	}
}

func TestRawDataPreservation(t *testing.T) {
	n := NewNormalizer()
	rawEvent := "2026-09-28 23:50:01 ERROR database failed"
	pe := parsing.ParsedEvent{
		Fields: map[string]any{"data": "test"},
	}
	cr := classification.ClassificationResult{
		Status:    classification.ClassificationUnknown,
		EventType: "",
	}

	result, err := n.Normalize(pe, cr, rawEvent)
	if err != nil {
		t.Fatalf("Normalization failed: %v", err)
	}

	if result.RawData != rawEvent {
		t.Errorf("Raw data not preserved.\nExpected: %q\nGot:      %q", rawEvent, result.RawData)
	}
}

func TestValidation_InvalidTypeUID(t *testing.T) {
	ev := NormalizedEvent{
		Metadata:     Metadata{Version: "1.7.0"},
		EventUID:     "test",
		Time:         1234567890,
		CategoryName: "Network Activity",
		CategoryUID:  4,
		ClassName:    "HTTP Activity",
		ClassUID:     4002,
		ActivityName: "Get",
		ActivityID:   3,
		TypeName:     "HTTP Activity: Get",
		TypeUID:      999999, // wrong
		Severity:     "Informational",
		SeverityID:   1,
		Status:       "Success",
		StatusID:     1,
		RawData:      "raw",
	}

	vr := Validate(ev)
	if vr.Valid {
		t.Error("Expected validation failure for wrong type_uid")
	}
}
