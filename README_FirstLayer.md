# Log Ingestion & Preprocessing Platform

A lightweight, air-gapped, deterministic log-processing platform built for the Smart India Hackathon (SIH). Receives logs across multiple protocols, authenticates sources, detects formats, parses structured fields, classifies events, and normalizes to OCSF 1.7.0.

**The web dashboard is developed separately by another team and is NOT included in this repository.**

## Architecture

```text
LOG SOURCES
   │
   ├── TCP :5050
   ├── HTTP :5001
   ├── UDP :5002
   └── Syslog :5003
             │
             ▼
       Go Collector (cmd/collector)
             │
     authenticate + identify app
             │
             ▼
        Kafka: raw-logs
             │
             ▼
       Go Processor (cmd/processor)
             │
     ┌───────┴────────┐
     │                │
  Format Detection    │
     │                │
  KNOWN FORMAT    UNKNOWN/AMBIGUOUS
     │            → stop/report
     ▼
   Parser
     │
     ▼
  Classification
     │
     ▼
  OCSF 1.7.0 Normalization
     │
     ▼
  OCSF Validation
     │
     ▼
  Terminal output (development)
```

## Supported Protocols

| Protocol | Port | Authentication |
|----------|------|----------------|
| TCP      | 5050 | First-line: `APP_ID API_KEY` |
| HTTP     | 5001 | Headers: `X-App-ID`, `X-API-Key` |
| UDP      | 5002 | N/A (connectionless) |
| Syslog   | 5003 | N/A |

## Authentication

Application identities and API keys are managed in `configs/applications.yaml`.

## Processing Pipeline

| Stage | Responsibility |
|-------|---------------|
| **Detection** | "What format is this?" → JSON, SYSLOG, APACHE, NGINX, UNKNOWN, AMBIGUOUS |
| **Parsing** | "What structured fields are present?" → ParsedEvent |
| **Classification** | "What does this event represent?" → HTTP_ACTIVITY, AUTHENTICATION, UNKNOWN |
| **Normalization** | "How should this be represented in OCSF 1.7.0?" → NormalizedEvent |
| **Validation** | "Does the OCSF output pass schema validation?" → PASS/FAIL |

### OCSF 1.7.0 Classes Supported

| Event Type | OCSF Class | class_uid | category_uid |
|------------|-----------|-----------|-------------|
| HTTP Activity | HTTP Activity | 4002 | 4 (Network Activity) |
| Authentication | Authentication | 3002 | 3 (Identity & Access Management) |
| Unknown/Ambiguous | Base Event | 0 | 0 (Uncategorized) |

## Running Locally

### Prerequisites
- Go 1.21+
- Docker (for Kafka)

### Terminal 1: Start Kafka
```bash
docker-compose up -d
```

### Terminal 2: Start the Collector
```bash
go run cmd/collector/main.go
```

### Terminal 3: Start the Processor
```bash
go run cmd/processor/main.go
```

### Terminal 4: Send test logs

**Apache HTTP event:**
```bash
(echo "payment-service abc123"; echo '127.0.0.1 - frank [10/Oct/2026:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326') | nc localhost 5050 
```

**JSON authentication event:**
```bash
curl -X POST http://localhost:5001/logs \
  -H "X-App-ID: payment-service" \
  -H "X-API-Key: abc123" \
  -H "Content-Type: text/plain" \
  -d '{"timestamp": "2026-10-10T14:30:00Z", "action": "login", "user": "admin", "result": "success"}'
```

**Syslog authentication event:**
```bash
echo '<34>Oct 11 22:14:15 mymachine sshd[12345]: Failed password for user admin from 10.0.0.1 port 22 ssh2' | nc localhost 5003
```

**Unknown format (will stop at detection):**
```bash
(echo "payment-service abc123"; echo 'This is a completely custom log format') | nc localhost 5050
```

**Ambiguous format (Apache/Nginx overlap):**
```bash
(echo "payment-service abc123"; echo '192.168.1.1 - - [10/Oct/2026:13:55:36 -0700] "GET /index.html HTTP/1.1" 200 612 "-" "Mozilla/5.0"') | nc localhost 5050
```

## Running Tests
```bash
go test ./... -v
```

## Future Architecture
```text
Detection → Parsing → OCSF Normalization → Storage/API → Web Dashboard
```

The dashboard is intentionally external to this repository. The processing pipeline exposes clean backend interfaces for future integration.
