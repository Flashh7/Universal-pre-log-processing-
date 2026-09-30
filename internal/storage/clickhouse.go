package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"time"

	"github.com/sih/log-platform/internal/normalization"
)

// ClickHouseConfig holds connection parameters for ClickHouse.
// All values are loaded from environment variables.
type ClickHouseConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	Table    string
}

// LoadClickHouseConfig reads ClickHouse configuration from environment variables.
// Host, username, and password have no defaults and must be set explicitly.
func LoadClickHouseConfig() ClickHouseConfig {
	return ClickHouseConfig{
		Host:     os.Getenv("CLICKHOUSE_HOST"),
		Port:     getEnv("CLICKHOUSE_PORT", "8123"),
		Database: getEnv("CLICKHOUSE_DATABASE", "ulpf"),
		Username: os.Getenv("CLICKHOUSE_USERNAME"),
		Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		Table:    getEnv("CLICKHOUSE_TABLE", "events"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// clickhouseRow maps exactly to the 15 columns of the ulpf.events table.
// Field names match the ClickHouse column names for JSONEachRow insertion.
type clickhouseRow struct {
	EventUID     string `json:"event_uid"`
	Time         int64  `json:"time"`
	CategoryName string `json:"category_name"`
	CategoryUID  int32  `json:"category_uid"`
	ClassName    string `json:"class_name"`
	ClassUID     int32  `json:"class_uid"`
	ActivityName string `json:"activity_name"`
	ActivityID   int32  `json:"activity_id"`
	TypeName     string `json:"type_name"`
	TypeUID      int32  `json:"type_uid"`
	Severity     string `json:"severity"`
	SeverityID   int32  `json:"severity_id"`
	Status       string `json:"status"`
	StatusID     int32  `json:"status_id"`
	RawEvent     string `json:"raw_event"`
}

// ToRow maps a NormalizedEvent to the ClickHouse row structure.
// NormalizedEvent.RawData is mapped to ClickHouse column raw_event.
// Go int fields are narrowed to int32 for ClickHouse Int32 columns.
// Time remains int64 (epoch milliseconds per OCSF timestamp_t).
func ToRow(ev normalization.NormalizedEvent) clickhouseRow {
	return clickhouseRow{
		EventUID:     ev.EventUID,
		Time:         ev.Time,
		CategoryName: ev.CategoryName,
		CategoryUID:  int32(ev.CategoryUID),
		ClassName:    ev.ClassName,
		ClassUID:     int32(ev.ClassUID),
		ActivityName: ev.ActivityName,
		ActivityID:   int32(ev.ActivityID),
		TypeName:     ev.TypeName,
		TypeUID:      int32(ev.TypeUID),
		Severity:     ev.Severity,
		SeverityID:   int32(ev.SeverityID),
		Status:       ev.Status,
		StatusID:     int32(ev.StatusID),
		RawEvent:     ev.RawData,
	}
}

// ClickHouseWriter implements EventWriter using ClickHouse's HTTP interface (port 8123).
// It sends inserts via POST with FORMAT JSONEachRow — no SQL string concatenation.
type ClickHouseWriter struct {
	baseURL  string
	database string
	table    string
	username string
	password string
	client   *http.Client
}

// NewClickHouseWriter creates a new ClickHouseWriter from the given config.
// Returns an error if required configuration (host, username, password) is missing.
func NewClickHouseWriter(cfg ClickHouseConfig) (*ClickHouseWriter, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("CLICKHOUSE_HOST is required")
	}
	if cfg.Username == "" {
		return nil, fmt.Errorf("CLICKHOUSE_USERNAME is required")
	}
	if cfg.Password == "" {
		return nil, fmt.Errorf("CLICKHOUSE_PASSWORD is required")
	}

	return &ClickHouseWriter{
		baseURL:  fmt.Sprintf("http://%s:%s", cfg.Host, cfg.Port),
		database: cfg.Database,
		table:    cfg.Table,
		username: cfg.Username,
		password: cfg.Password,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// NewClickHouseWriterForTest creates a ClickHouseWriter pointing at a custom base URL.
// This is used by tests with httptest.NewServer.
func NewClickHouseWriterForTest(baseURL, database, table, username, password string) *ClickHouseWriter {
	return &ClickHouseWriter{
		baseURL:  baseURL,
		database: database,
		table:    table,
		username: username,
		password: password,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Ping verifies connectivity to ClickHouse by executing SELECT 1.
func (w *ClickHouseWriter) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "POST", w.baseURL+"/", bytes.NewBufferString("SELECT 1"))
	if err != nil {
		return fmt.Errorf("failed to create ping request: %w", err)
	}
	req.Header.Set("X-ClickHouse-User", w.username)
	req.Header.Set("X-ClickHouse-Key", w.password)
	req.Header.Set("X-ClickHouse-Database", w.database)

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("clickhouse connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("clickhouse ping failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// Write inserts a single validated OCSF event into ClickHouse using FORMAT JSONEachRow.
// The JSON field names match the ClickHouse column names exactly.
func (w *ClickHouseWriter) Write(ctx context.Context, event normalization.NormalizedEvent) error {
	row := ToRow(event)

	body, err := json.Marshal(row)
	if err != nil {
		return fmt.Errorf("failed to marshal event for ClickHouse: %w", err)
	}

	query := fmt.Sprintf("INSERT INTO %s.%s FORMAT JSONEachRow", w.database, w.table)
	endpoint := fmt.Sprintf("%s/?query=%s", w.baseURL, neturl.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create insert request: %w", err)
	}
	req.Header.Set("X-ClickHouse-User", w.username)
	req.Header.Set("X-ClickHouse-Key", w.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("clickhouse insert failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("clickhouse insert failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// Close releases resources. For the HTTP-based client, this is a no-op.
func (w *ClickHouseWriter) Close() error {
	return nil
}
