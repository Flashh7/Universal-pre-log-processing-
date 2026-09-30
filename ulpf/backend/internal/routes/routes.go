package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"ulpf-backend/internal/database"
	"ulpf-backend/internal/handlers"
)

func Setup() {
	http.HandleFunc("/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ULPF Backend is running")
	}))

	http.HandleFunc("/api/test-db", withCORS(func(w http.ResponseWriter, r *http.Request) {
		result, err := database.Query("SELECT 1")
		if err != nil {
			http.Error(
				w,
				"Database connection failed: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		response := map[string]string{
			"result": string(result),
		}

		json.NewEncoder(w).Encode(response)
	}))

	// Logs
	http.HandleFunc("/api/logs/", withCORS(handlers.GetLogByID))
	http.HandleFunc("/api/logs", withCORS(handlers.GetLogs))

	// Overview
	http.HandleFunc("/api/overview/metrics", withCORS(handlers.GetOverviewMetrics))
	http.HandleFunc("/api/overview/volume", withCORS(handlers.GetOverviewVolume))
	http.HandleFunc("/api/overview/categories", withCORS(handlers.GetOverviewCategories))
	http.HandleFunc("/api/overview/severity", withCORS(handlers.GetOverviewSeverity))

	// Sources
	http.HandleFunc("/api/sources", withCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetSources(w, r)

		case http.MethodPost:
			handlers.CreateSource(w, r)

		default:
			http.Error(
				w,
				"Method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	}))
}

func withCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(
			"Access-Control-Allow-Origin",
			"http://localhost:5173",
		)

		w.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, OPTIONS",
		)

		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type",
		)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		handler(w, r)
	}
}