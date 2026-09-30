package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sih/log-platform/internal/auth"
	"github.com/sih/log-platform/internal/ingestion/http"
	"github.com/sih/log-platform/internal/ingestion/syslog"
	"github.com/sih/log-platform/internal/ingestion/tcp"
	"github.com/sih/log-platform/internal/ingestion/udp"
	"github.com/sih/log-platform/internal/kafka"
)

func main() {
	registry, err := auth.LoadRegistry("configs/applications.yaml")
	if err != nil {
		log.Fatalf("Failed to load registry: %v", err)
	}

	kafkaBrokers := []string{"localhost:9092"}
	kafkaTopic := "raw-logs"
	producer := kafka.NewProducer(kafkaBrokers, kafkaTopic)
	defer producer.Close()

	tcpServer := tcp.NewServer(":5050", registry, producer)
	httpServer := http.NewServer(":5001", registry, producer)
	udpServer := udp.NewServer(":5002", producer)
	syslogServer := syslog.NewServer(":5003", producer)

	go func() {
		if err := tcpServer.Start(); err != nil {
			log.Fatalf("TCP server failed: %v", err)
		}
	}()

	go func() {
		if err := httpServer.Start(); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	go func() {
		if err := udpServer.Start(); err != nil {
			log.Fatalf("UDP server failed: %v", err)
		}
	}()

	go func() {
		if err := syslogServer.Start(); err != nil {
			log.Fatalf("Syslog server failed: %v", err)
		}
	}()

	log.Println("Log ingestion platform started.")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Shutting down...")
}
