package storage

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sih/log-platform/internal/normalization"
)

// testEvent returns a fully populated NormalizedEvent for testing.
func testEvent() normalization.NormalizedEvent {
	return normalization.NormalizedEvent{
		Metadata:     normalization.Metadata{Version: "1.7.0"},
		EventUID:     "test-uid-123",
		Time:         1791642600000, // epoch milliseconds
		CategoryName: "Network Activity",
		CategoryUID:  4,
		ClassName:    "HTTP Activity",
		ClassUID:     4002,
		ActivityName: "Get",
		ActivityID:   3,
		TypeName:     "HTTP Activity: Get",
		TypeUID:      400203,
		Severity:     "Informational",
		SeverityID:   1,
		Status:       "Success",
		StatusID:     1,
		RawData:      `127.0.0.1 - frank [10/Oct/2026:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`,
	}
}

// --- Test 1: Correct field mapping from OCSF event → ClickHouse row ---

func TestToRow_FieldMapping(t *testing.T) {
	ev := testEvent()
	row := ToRow(ev)

	if row.EventUID != ev.EventUID {
		t.Errorf("EventUID: got %q, want %q", row.EventUID, ev.EventUID)
	}
	if row.Time != ev.Time {
		t.Errorf("Time: got %d, want %d", row.Time, ev.Time)
	}
	if row.CategoryName != ev.CategoryName {
		t.Errorf("CategoryName: got %q, want %q", row.CategoryName, ev.CategoryName)
	}
	if row.ClassName != ev.ClassName {
		t.Errorf("ClassName: got %q, want %q", row.ClassName, ev.ClassName)
	}
	if row.ActivityName != ev.ActivityName {
		t.Errorf("ActivityName: got %q, want %q", row.ActivityName, ev.ActivityName)
	}
	if row.TypeName != ev.TypeName {
		t.Errorf("TypeName: got %q, want %q", row.TypeName, ev.TypeName)
	}
	if row.Severity != ev.Severity {
		t.Errorf("Severity: got %q, want %q", row.Severity, ev.Severity)
	}
	if row.Status != ev.Status {
		t.Errorf("Status: got %q, want %q", row.Status, ev.Status)
	}
}

// --- Test 2: Correct handling of Int32/Int64 fields ---

func TestToRow_IntConversions(t *testing.T) {
	ev := testEvent()
	row := ToRow(ev)

	// Int32 fields
	if row.CategoryUID != int32(ev.CategoryUID) {
		t.Errorf("CategoryUID int32: got %d, want %d", row.CategoryUID, int32(ev.CategoryUID))
	}
	if row.ClassUID != int32(ev.ClassUID) {
		t.Errorf("ClassUID int32: got %d, want %d", row.ClassUID, int32(ev.ClassUID))
	}
	if row.ActivityID != int32(ev.ActivityID) {
		t.Errorf("ActivityID int32: got %d, want %d", row.ActivityID, int32(ev.ActivityID))
	}
	if row.TypeUID != int32(ev.TypeUID) {
		t.Errorf("TypeUID int32: got %d, want %d", row.TypeUID, int32(ev.TypeUID))
	}
	if row.SeverityID != int32(ev.SeverityID) {
		t.Errorf("SeverityID int32: got %d, want %d", row.SeverityID, int32(ev.SeverityID))
	}
	if row.StatusID != int32(ev.StatusID) {
		t.Errorf("StatusID int32: got %d, want %d", row.StatusID, int32(ev.StatusID))
	}

	// Int64 field
	if row.Time != ev.Time {
		t.Errorf("Time int64: got %d, want %d", row.Time, ev.Time)
	}
}

// --- Test 3: raw_event preservation ---

func TestToRow_RawEventPreservation(t *testing.T) {
	ev := testEvent()
	row := ToRow(ev)

	// NormalizedEvent.RawData must map to clickhouseRow.RawEvent
	if row.RawEvent != ev.RawData {
		t.Errorf("RawEvent != RawData.\nGot:  %q\nWant: %q", row.RawEvent, ev.RawData)
	}

	// Verify the raw event is byte-for-byte preserved (no trimming, no case changes)
	originalRaw := `127.0.0.1 - frank [10/Oct/2026:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`
	if row.RawEvent != originalRaw {
		t.Errorf("Raw event was modified during mapping.\nGot:  %q\nWant: %q", row.RawEvent, originalRaw)
	}
}

// --- Test 4: ClickHouse connection failure ---

func TestPing_ConnectionFailure(t *testing.T) {
	// Point at a server that doesn't exist
	writer := NewClickHouseWriterForTest(
		"http://127.0.0.1:1", // unreachable port
		"ulpf", "events", "test", "test",
	)

	err := writer.Ping(context.Background())
	if err == nil {
		t.Fatal("Expected error on connection failure, got nil")
	}
	if !strings.Contains(err.Error(), "clickhouse connection failed") {
		t.Errorf("Expected 'clickhouse connection failed' in error, got: %v", err)
	}
}

// --- Test 5: ClickHouse insert failure (server returns error) ---

func TestWrite_InsertFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Code: 60. DB::Exception: Table ulpf.events doesn't exist"))
	}))
	defer server.Close()

	writer := NewClickHouseWriterForTest(
		server.URL, "ulpf", "events", "test", "test",
	)

	err := writer.Write(context.Background(), testEvent())
	if err == nil {
		t.Fatal("Expected error on insert failure, got nil")
	}
	if !strings.Contains(err.Error(), "clickhouse insert failed") {
		t.Errorf("Expected 'clickhouse insert failed' in error, got: %v", err)
	}
}

// --- Test 6: Successful insert ---

func TestWrite_Success(t *testing.T) {
	var receivedBody []byte
	var receivedQuery string
	var receivedUser string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.Query().Get("query")
		receivedUser = r.Header.Get("X-ClickHouse-User")
		var err error
		receivedBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	writer := NewClickHouseWriterForTest(
		server.URL, "ulpf", "events", "testuser", "testpass",
	)

	ev := testEvent()
	err := writer.Write(context.Background(), ev)
	if err != nil {
		t.Fatalf("Expected successful write, got: %v", err)
	}

	// Verify the INSERT query targets the correct table
	if !strings.Contains(receivedQuery, "INSERT INTO ulpf.events FORMAT JSONEachRow") {
		t.Errorf("Unexpected query: %q", receivedQuery)
	}

	// Verify authentication header was sent
	if receivedUser != "testuser" {
		t.Errorf("Expected X-ClickHouse-User 'testuser', got %q", receivedUser)
	}

	// Verify the JSON body contains all 15 fields with correct values
	var row clickhouseRow
	if err := json.Unmarshal(receivedBody, &row); err != nil {
		t.Fatalf("Failed to unmarshal body: %v", err)
	}

	if row.EventUID != ev.EventUID {
		t.Errorf("Body EventUID: got %q, want %q", row.EventUID, ev.EventUID)
	}
	if row.Time != ev.Time {
		t.Errorf("Body Time: got %d, want %d", row.Time, ev.Time)
	}
	if row.ClassUID != int32(ev.ClassUID) {
		t.Errorf("Body ClassUID: got %d, want %d", row.ClassUID, int32(ev.ClassUID))
	}
	if row.RawEvent != ev.RawData {
		t.Errorf("Body RawEvent: got %q, want %q", row.RawEvent, ev.RawData)
	}
}

// --- Test 7: Verify all 15 JSON field names match ClickHouse columns ---

func TestToRow_JSONFieldNames(t *testing.T) {
	ev := testEvent()
	row := ToRow(ev)

	data, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("Failed to marshal row: %v", err)
	}

	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatalf("Failed to unmarshal row: %v", err)
	}

	// These 15 field names must exactly match the ClickHouse column names
	expectedColumns := []string{
		"event_uid", "time", "category_name", "category_uid",
		"class_name", "class_uid", "activity_name", "activity_id",
		"type_name", "type_uid", "severity", "severity_id",
		"status", "status_id", "raw_event",
	}

	if len(fields) != len(expectedColumns) {
		t.Errorf("Expected %d fields, got %d", len(expectedColumns), len(fields))
	}

	for _, col := range expectedColumns {
		if _, ok := fields[col]; !ok {
			t.Errorf("Missing expected ClickHouse column %q in JSON output", col)
		}
	}
}

// --- Test 8: Ping success ---

func TestPing_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("1\n"))
	}))
	defer server.Close()

	writer := NewClickHouseWriterForTest(
		server.URL, "ulpf", "events", "test", "test",
	)

	err := writer.Ping(context.Background())
	if err != nil {
		t.Fatalf("Expected successful ping, got: %v", err)
	}
}

// --- Test 9: Ping auth failure ---

func TestPing_AuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Code: 516. Authentication failed"))
	}))
	defer server.Close()

	writer := NewClickHouseWriterForTest(
		server.URL, "ulpf", "events", "wrong", "wrong",
	)

	err := writer.Ping(context.Background())
	if err == nil {
		t.Fatal("Expected error on auth failure, got nil")
	}
	if !strings.Contains(err.Error(), "clickhouse ping failed") {
		t.Errorf("Expected 'clickhouse ping failed' in error, got: %v", err)
	}
}

// --- Test 10: NewClickHouseWriter config validation ---

func TestNewClickHouseWriter_MissingHost(t *testing.T) {
	_, err := NewClickHouseWriter(ClickHouseConfig{
		Username: "user",
		Password: "pass",
	})
	if err == nil || !strings.Contains(err.Error(), "CLICKHOUSE_HOST") {
		t.Errorf("Expected CLICKHOUSE_HOST error, got: %v", err)
	}
}

func TestNewClickHouseWriter_MissingUsername(t *testing.T) {
	_, err := NewClickHouseWriter(ClickHouseConfig{
		Host:     "localhost",
		Password: "pass",
	})
	if err == nil || !strings.Contains(err.Error(), "CLICKHOUSE_USERNAME") {
		t.Errorf("Expected CLICKHOUSE_USERNAME error, got: %v", err)
	}
}

func TestNewClickHouseWriter_MissingPassword(t *testing.T) {
	_, err := NewClickHouseWriter(ClickHouseConfig{
		Host:     "localhost",
		Username: "user",
	})
	if err == nil || !strings.Contains(err.Error(), "CLICKHOUSE_PASSWORD") {
		t.Errorf("Expected CLICKHOUSE_PASSWORD error, got: %v", err)
	}
}
