package http

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sih/log-platform/internal/auth"
	"github.com/sih/log-platform/internal/event"
)

type Server struct {
	addr      string
	registry  *auth.Registry
	publisher Publisher
}

type Publisher interface {
	Publish(ctx context.Context, ev event.Event) error
}

func NewServer(addr string, registry *auth.Registry, publisher Publisher) *Server {
	return &Server{
		addr:      addr,
		registry:  registry,
		publisher: publisher,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/logs", s.handleLogs)
	return http.ListenAndServe(s.addr, mux)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	appID := r.Header.Get("X-App-ID")
	apiKey := r.Header.Get("X-API-Key")

	if err := s.registry.Authenticate(appID, apiKey); err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	rawLog := string(body)

	// If it's a JSON payload like {"message": "..."}, we can optionally unwrap it.
	// But the spec says "It must not parse the log's internal format."
	// We'll just take the raw string.

	ev := event.Event{
		ApplicationID: appID,
		Protocol:      "http",
		ReceivedAt:    time.Now().UTC(),
		Raw:           strings.TrimSpace(rawLog),
	}

	if err := s.publisher.Publish(r.Context(), ev); err != nil {
		http.Error(w, "Failed to publish", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
