package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/segmentio/kafka-go"
	"github.com/sih/log-platform/internal/classification"
	"github.com/sih/log-platform/internal/detection"
	"github.com/sih/log-platform/internal/event"
	"github.com/sih/log-platform/internal/normalization"
	"github.com/sih/log-platform/internal/parsing"
	"github.com/sih/log-platform/internal/storage"
)

func main() {
	// Initialize pipeline stages
	detector := detection.NewManager()
	classifier := classification.NewClassifier()
	normalizer := normalization.NewNormalizer()

	// Parser registry keyed by detected format name
	parsers := map[string]parsing.Parser{
		"JSON":   &parsing.JSONParser{},
		"SYSLOG": &parsing.SyslogParser{},
		"APACHE": &parsing.ApacheParser{},
		"NGINX":  &parsing.NginxParser{},
	}

	// Initialize ClickHouse storage (optional — disabled if env vars are not set)
	var writer storage.EventWriter
	chCfg := storage.LoadClickHouseConfig()
	if chCfg.Host != "" {
		chWriter, err := storage.NewClickHouseWriter(chCfg)
		if err != nil {
			log.Fatalf("CLICKHOUSE: CONFIGURATION ERROR — %v", err)
		}
		if err := chWriter.Ping(context.Background()); err != nil {
			log.Fatalf("CLICKHOUSE: CONNECTION FAILED — %v", err)
		}
		log.Printf("CLICKHOUSE: CONNECTED (host: %s:%s, database: %s, table: %s)",
			chCfg.Host, chCfg.Port, chCfg.Database, chCfg.Table)
		writer = chWriter
		defer chWriter.Close()
	} else {
		log.Println("CLICKHOUSE: DISABLED (CLICKHOUSE_HOST not set)")
	}

	// Connect to Kafka
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "raw-logs",
		GroupID:  "format-detector-group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	log.Println("Processor started. Listening for messages on 'raw-logs'...")
	log.Println(strings.Repeat("=", 70))

	// NOTE: ReadMessage auto-commits Kafka offsets upon return, BEFORE processing.
	// This means if processing or ClickHouse insertion fails, the message will not
	// be redelivered. Switching to FetchMessage + CommitMessages (commit after
	// successful storage) would be an architectural change to the Kafka consumer.
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Failed to read message: %v", err)
			break
		}

		// Decode the event envelope from the collector
		var ev event.Event
		if err := json.Unmarshal(m.Value, &ev); err != nil {
			log.Printf("Invalid event envelope: %v", err)
			continue
		}

		processEvent(ev, detector, parsers, classifier, normalizer, writer)
	}

	if err := r.Close(); err != nil {
		log.Fatal("Failed to close reader:", err)
	}
}

func processEvent(
	ev event.Event,
	detector *detection.Manager,
	parsers map[string]parsing.Parser,
	classifier *classification.Classifier,
	normalizer *normalization.OCSFNormalizer,
	writer storage.EventWriter,
) {
	fmt.Println()
	log.Printf("RAW RECEIVED [App: %s] [Protocol: %s]", ev.ApplicationID, ev.Protocol)
	log.Printf("  Raw: %s", ev.Raw)

	// Stage 1: Format Detection
	detResult := detector.Identify(ev.Raw)
	log.Printf("  FORMAT: %s (outcome: %s)", detResult.Format, detResult.Outcome)

	// If UNKNOWN or AMBIGUOUS format, stop before parsing
	if detResult.Outcome == detection.OutcomeUnknown {
		log.Println("  ⛔ DETECTION: UNKNOWN_FORMAT — stopping before parsing/normalization")
		log.Println(strings.Repeat("-", 70))
		return
	}
	if detResult.Outcome == detection.OutcomeAmbiguous {
		log.Println("  ⛔ DETECTION: AMBIGUOUS_FORMAT — stopping before parsing/normalization")
		log.Println(strings.Repeat("-", 70))
		return
	}

	// Stage 2: Parsing
	parser, ok := parsers[detResult.Format]
	if !ok {
		log.Printf("  ⛔ PARSE: No parser registered for format %q", detResult.Format)
		log.Println(strings.Repeat("-", 70))
		return
	}

	parsed, err := parser.Parse(ev.Raw)
	if err != nil {
		log.Printf("  ⛔ PARSE: %v", err)
		log.Println(strings.Repeat("-", 70))
		return
	}
	log.Println("  PARSE: SUCCESS")

	// Stage 3: Classification
	classResult := classifier.Classify(parsed)
	log.Printf("  CLASSIFICATION: %s (status: %s)", classResult.EventType, classResult.Status)
	if classResult.Status == classification.ClassificationAmbiguous {
		log.Printf("  CLASSIFICATION CANDIDATES: %v", classResult.Candidates)
	}

	// Stage 4: OCSF Normalization
	normalized, err := normalizer.Normalize(parsed, classResult, ev.Raw)
	if err != nil {
		log.Printf("  ⛔ NORMALIZATION: FAILED — %v", err)
		log.Println(strings.Repeat("-", 70))
		return
	}
	log.Println("  NORMALIZATION: SUCCESS")

	// Stage 5: OCSF Validation
	valResult := normalization.Validate(normalized)
	if valResult.Valid {
		log.Println("  OCSF VALIDATION: PASS")
	} else {
		log.Printf("  OCSF VALIDATION: FAIL")
		for _, e := range valResult.Errors {
			log.Printf("    reason: %s", e)
		}
		log.Println(strings.Repeat("-", 70))
		return
	}

	// Stage 6: ClickHouse Storage (only for validated events)
	if writer != nil {
		if err := writer.Write(context.Background(), normalized); err != nil {
			log.Printf("  ⛔ CLICKHOUSE: INSERT FAILED — %v", err)
		} else {
			log.Println("  CLICKHOUSE: INSERT SUCCESS")
		}
	}

	// Print the normalized OCSF event
	out, _ := json.MarshalIndent(normalized, "  ", "  ")
	fmt.Printf("  %s\n", string(out))

	// Raw event preservation check
	if normalized.RawData == ev.Raw {
		log.Println("  RAW PRESERVATION: ✅ PASS (raw_data == original raw)")
	} else {
		log.Println("  RAW PRESERVATION: ❌ FAIL (raw_data differs from original)")
	}

	log.Println(strings.Repeat("-", 70))
}
