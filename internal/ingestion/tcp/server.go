package tcp

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
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
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer l.Close()
	log.Printf("TCP Server listening on %s", s.addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("TCP Accept error: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	// Authentication phase
	// Expected: APP_ID API_KEY
	if !scanner.Scan() {
		return
	}
	authLine := strings.TrimSpace(scanner.Text())
	parts := strings.Split(authLine, " ")
	if len(parts) != 2 {
		fmt.Fprintln(conn, "ERROR: Invalid authentication format")
		return
	}

	appID, apiKey := parts[0], parts[1]
	if err := s.registry.Authenticate(appID, apiKey); err != nil {
		fmt.Fprintf(conn, "ERROR: %v\n", err)
		return
	}
	fmt.Fprintln(conn, "OK")

	// Read logs
	for scanner.Scan() {
		rawLog := strings.TrimSpace(scanner.Text())
		if rawLog == "" {
			continue
		}

		ev := event.Event{
			ApplicationID: appID,
			Protocol:      "tcp",
			ReceivedAt:    time.Now().UTC(),
			Raw:           rawLog,
		}

		if err := s.publisher.Publish(context.Background(), ev); err != nil {
			log.Printf("TCP Publish error: %v", err)
		}
	}
}
