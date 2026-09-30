package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"ulpf-backend/internal/database"
)

func GetOverviewMetrics(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT
			count() AS total_events,
			if(
				max(time) = min(time),
				0,
				count() / ((max(time) - min(time)) / 1000.0)
			) AS events_per_second,
			countDistinct(category_name) AS active_categories,
			countIf(status = 'Failure') AS failed_events
		FROM events
		FORMAT JSONEachRow
	`

	result, err := database.Query(query)
	if err != nil {
		http.Error(
			w,
			"Failed to fetch overview metrics: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	writeJSONEachRowAsObject(w, result)
}

func GetOverviewVolume(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT
			toStartOfMinute(fromUnixTimestamp64Milli(time)) AS timestamp,
			count() AS event_count
		FROM events
		GROUP BY timestamp
		ORDER BY timestamp
		FORMAT JSONEachRow
	`

	result, err := database.Query(query)
	if err != nil {
		http.Error(
			w,
			"Failed to fetch overview volume: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	writeJSONEachRowAsArray(w, result)
}

func GetOverviewCategories(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT
			category_name,
			count() AS event_count
		FROM events
		GROUP BY category_name
		ORDER BY event_count DESC
		FORMAT JSONEachRow
	`

	result, err := database.Query(query)
	if err != nil {
		http.Error(
			w,
			"Failed to fetch overview categories: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	writeJSONEachRowAsArray(w, result)
}

func GetOverviewSeverity(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT
			severity,
			count() AS event_count
		FROM events
		GROUP BY severity
		ORDER BY event_count DESC
		FORMAT JSONEachRow
	`

	result, err := database.Query(query)
	if err != nil {
		http.Error(
			w,
			"Failed to fetch overview severity: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	writeJSONEachRowAsArray(w, result)
}

func writeJSONEachRowAsObject(w http.ResponseWriter, result []byte) {
	result = []byte(strings.TrimSpace(string(result)))

	if len(result) == 0 {
		http.Error(
			w,
			"No overview metrics returned",
			http.StatusInternalServerError,
		)
		return
	}

	var object json.RawMessage

	if err := json.Unmarshal(result, &object); err != nil {
		http.Error(
			w,
			"Failed to parse overview metrics: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(object)
}

func writeJSONEachRowAsArray(w http.ResponseWriter, result []byte) {
	lines := strings.Split(
		strings.TrimSpace(string(result)),
		"\n",
	)

	items := make([]json.RawMessage, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		var item json.RawMessage

		if err := json.Unmarshal([]byte(line), &item); err != nil {
			http.Error(
				w,
				"Failed to parse overview result: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
