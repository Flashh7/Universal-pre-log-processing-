package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"ulpf-backend/internal/database"
)

func GetLogs(w http.ResponseWriter, r *http.Request) {

	// Read filters from the URL
	search := r.URL.Query().Get("search")
	severity := r.URL.Query().Get("severity")
	activity := r.URL.Query().Get("activity")
	className := r.URL.Query().Get("class_name")
	status := r.URL.Query().Get("status")

	// Base query
	query := `
		SELECT
			event_uid,
			time,
			category_name,
			category_uid,
			class_name,
			class_uid,
			activity_name,
			activity_id,
			type_name,
			type_uid,
			severity,
			severity_id,
			status,
			status_id,
			raw_event
		FROM events
		WHERE 1 = 1
	`

	// Search across useful normalized fields
	if search != "" {

		value := escapeSQL(search)

		query += fmt.Sprintf(`
			AND (
				event_uid LIKE '%%%s%%'
				OR category_name LIKE '%%%s%%'
				OR class_name LIKE '%%%s%%'
				OR activity_name LIKE '%%%s%%'
				OR type_name LIKE '%%%s%%'
				OR severity LIKE '%%%s%%'
				OR status LIKE '%%%s%%'
			)
		`,
			value,
			value,
			value,
			value,
			value,
			value,
			value,
		)
	}

	// Severity filter
	if severity != "" {
		query += fmt.Sprintf(
			" AND severity = '%s'\n",
			escapeSQL(severity),
		)
	}

	// Activity filter
	if activity != "" {
		query += fmt.Sprintf(
			" AND activity_name = '%s'\n",
			escapeSQL(activity),
		)
	}

	// Class filter
	if className != "" {
		query += fmt.Sprintf(
			" AND class_name = '%s'\n",
			escapeSQL(className),
		)
	}

	// Status filter
	if status != "" {
		query += fmt.Sprintf(
			" AND status = '%s'\n",
			escapeSQL(status),
		)
	}

	// Ordering and result limit
	query += `
		ORDER BY time DESC
		LIMIT 100
		FORMAT JSONEachRow
	`

	// Execute query
	result, err := database.Query(query)

	if err != nil {
		http.Error(
			w,
			"Failed to fetch logs: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	// Convert ClickHouse JSONEachRow into a JSON array
	lines := strings.Split(
		strings.TrimSpace(string(result)),
		"\n",
	)

	logs := make([]json.RawMessage, 0)

	for _, line := range lines {

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		logs = append(
			logs,
			json.RawMessage(line),
		)
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(logs)
}

func GetLogByID(w http.ResponseWriter, r *http.Request) {

	// Extract event UID from:
	// /api/logs/{event_uid}

	eventUID := strings.TrimPrefix(r.URL.Path, "/api/logs/")

	if eventUID == "" {
		http.Error(
			w,
			"Event UID is required",
			http.StatusBadRequest,
		)
		return
	}

	query := fmt.Sprintf(`
		SELECT
			event_uid,
			time,
			category_name,
			category_uid,
			class_name,
			class_uid,
			activity_name,
			activity_id,
			type_name,
			type_uid,
			severity,
			severity_id,
			status,
			status_id,
			raw_event
		FROM events
		WHERE event_uid = '%s'
		LIMIT 1
		FORMAT JSONEachRow
	`, escapeSQL(eventUID))

	result, err := database.Query(query)

	if err != nil {
		http.Error(
			w,
			"Failed to fetch event: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	result = []byte(strings.TrimSpace(string(result)))

	// No event found
	if len(result) == 0 {
		http.Error(
			w,
			"Event not found",
			http.StatusNotFound,
		)
		return
	}

	// Return the single event directly
	w.Header().Set("Content-Type", "application/json")

	w.Write(result)
}

// escapeSQL escapes characters that could break the SQL query.
func escapeSQL(value string) string {

	value = strings.ReplaceAll(
		value,
		`\\`,
		`\\\\`,
	)

	value = strings.ReplaceAll(
		value,
		`'`,
		`\'`,
	)

	return value
}