package syslog

import (
	"bufio"
	"context"
	"log"
	"net"
	"strings"
	"time"

	"github.com/sih/log-platform/internal/event"
)

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
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer l.Close()
	log.Printf("Syslog TCP Server listening on %s", s.addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Syslog Accept error: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		rawLog := strings.TrimSpace(scanner.Text())
		if rawLog == "" {
			continue
		}

		ev := event.Event{
			ApplicationID: "syslog-source",
			Protocol:      "syslog",
			ReceivedAt:    time.Now().UTC(),
			Raw:           rawLog,
		}

		if err := s.publisher.Publish(context.Background(), ev); err != nil {
			log.Printf("Syslog Publish error: %v", err)
		}
	}
}
