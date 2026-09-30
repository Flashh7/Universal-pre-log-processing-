package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ulpf-backend/internal/database"
)

type Source struct {
	SourceUID      string `json:"source_uid"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	VendorName     string `json:"vendor_name"`
	ProductName    string `json:"product_name"`
	OriginalFormat string `json:"original_format"`
	Status         string `json:"status"`
}

// GET /api/sources
func GetSources(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT
			source_uid,
			name,
			type,
			vendor_name,
			product_name,
			original_format,
			status
		FROM sources
		ORDER BY source_uid
		FORMAT JSONEachRow
	`

	result, err := database.Query(query)
	if err != nil {
		http.Error(
			w,
			"Failed to fetch sources: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	var sources []Source

	for _, line := range strings.Split(strings.TrimSpace(string(result)), "\n") {
		if line == "" {
			continue
		}

		var source Source

		if err := json.Unmarshal([]byte(line), &source); err != nil {
			http.Error(
				w,
				"Failed to parse source data: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		sources = append(sources, source)
	}

	if sources == nil {
		sources = []Source{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sources)
}

// POST /api/sources
func CreateSource(w http.ResponseWriter, r *http.Request) {
	var source Source

	err := json.NewDecoder(r.Body).Decode(&source)
	if err != nil {
		http.Error(
			w,
			"Invalid request body: "+err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	if source.Name == "" ||
		source.Type == "" ||
		source.VendorName == "" ||
		source.ProductName == "" ||
		source.OriginalFormat == "" ||
		source.Status == "" {

		http.Error(
			w,
			"Missing required source fields",
			http.StatusBadRequest,
		)
		return
	}

	// Generate source UID on the backend.
	source.SourceUID = fmt.Sprintf(
		"src-%d",
		time.Now().UnixNano(),
	)

	query := fmt.Sprintf(`
		INSERT INTO sources
		(
			source_uid,
			name,
			type,
			vendor_name,
			product_name,
			original_format,
			status
		)
		VALUES
		(
			'%s',
			'%s',
			'%s',
			'%s',
			'%s',
			'%s',
			'%s'
		)
	`,
		escapeSQL(source.SourceUID),
		escapeSQL(source.Name),
		escapeSQL(source.Type),
		escapeSQL(source.VendorName),
		escapeSQL(source.ProductName),
		escapeSQL(source.OriginalFormat),
		escapeSQL(source.Status),
	)

	_, err = database.Query(query)
	if err != nil {
		http.Error(
			w,
			"Failed to create source: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(source)
}

