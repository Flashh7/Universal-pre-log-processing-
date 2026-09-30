package udp

import (
	"context"
	"log"
	"net"
	"strings"
	"time"

	"github.com/sih/log-platform/internal/event"
)

// UDP doesn't have connections, so auth is trickier. 
// For now, we will assume a known APP_ID in the payload, or just a default for the sake of the spec.
type Server struct {
	addr      string
	publisher Publisher
}

type Publisher interface {
	Publish(ctx context.Context, ev event.Event) error
}

func NewServer(addr string, publisher Publisher) *Server {
	return &Server{
		addr:      addr,
		publisher: publisher,
	}
}

func (s *Server) Start() error {
	addr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Printf("UDP Server listening on %s", s.addr)

	buf := make([]byte, 65535)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("UDP Read error: %v", err)
			continue
		}

		rawLog := string(buf[:n])
		if strings.TrimSpace(rawLog) == "" {
			continue
		}

		// Hardcoded to unknown-udp for now since UDP is connectionless.
		ev := event.Event{
			ApplicationID: "unknown-udp",
			Protocol:      "udp",
			ReceivedAt:    time.Now().UTC(),
			Raw:           strings.TrimSpace(rawLog),
		}

		if err := s.publisher.Publish(context.Background(), ev); err != nil {
			log.Printf("UDP Publish error: %v", err)
		}
	}
}
