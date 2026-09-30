# Log Ingestion & Preprocessing Platform — Implementation Specification

## 1. Project Backstory

This project is being developed for a Smart India Hackathon (SIH) problem around **log ingestion, preprocessing, format detection, and analysis** in a corporate environment.

The goal is to build a **lightweight, air-gapped, deterministic log-processing platform** rather than trying to reproduce the full feature set of tools such as Logstash.

In a typical organization, many different applications generate logs in different formats and through different transport mechanisms. For example:

- A payment service may send JSON logs.
- A web server may generate Apache or Nginx access logs.
- Infrastructure software may emit Syslog.
- Custom applications may send plain text logs over TCP, HTTP, or UDP.

The system should provide a common ingestion layer that can receive logs from these applications, identify and authenticate the source application, preserve the original log message, place the event into Kafka, and then deterministically identify known log formats.

The longer-term pipeline is:

```text
LOG SOURCES
   │
   ├── TCP :5000
   ├── HTTP :5001
   ├── UDP :5002
   └── Syslog :5003
             │
             ▼
       Go Ingestion
             │
     authenticate +
    identify application
             │
             ▼
        Kafka raw-logs
             │
             ▼
   Deterministic Format
         Detection
             │
      ┌──────┼──────┐
      ▼      ▼      ▼
    JSON   Syslog  Apache/Nginx
      │      │      │
      └──────┴──────┘
             │
             ▼
           Parser
             │
             ▼
      OCSF / Normalizer
             │
             ▼
       Storage / API
             │
             ▼
    Existing Web Application
             │
             ▼
         Dashboard
```

The **production web dashboard is being developed by another team member**. However, a small **temporary test dashboard must be implemented in this repository** so the complete pipeline can be demonstrated and tested locally.

This test dashboard is NOT the final product UI. It exists only to:
- Verify that processed events reach a frontend.
- Demonstrate the end-to-end pipeline during development/SIH demos.
- Inspect received logs and detected formats.
- Make backend integration easier to test before the other dashboard is ready.

Keep the test dashboard deliberately simple and isolated. Do not design the backend around this temporary UI. The eventual dashboard must be able to replace it without requiring changes to ingestion, Kafka, detection, parsing, or normalization.

---

## 2. Current Implementation Scope

Implement only:

1. Multi-protocol ingestion
2. Application identification and authentication
3. Common internal event envelope
4. Kafka publishing
5. Deterministic known-format detection
6. Automated tests
7. Docker Compose for local development
8. Documentation

Do **not** implement yet:

- Full production log parsing
- Full OCSF normalization
- ML/LLM detection
- Unknown-format processing pipeline
- Production/final dashboard
- Cloud deployment
- Kubernetes
- Complex persistent queues
- Many extra protocols
- Full registration UI/database

The architecture must make these easy to add later.

---

## 3. Design Goals

Prioritize:

- Lightweight deployment
- Air-gapped operation
- Deterministic behavior
- Predictable performance
- Explainability
- Modular architecture
- Easy testing
- Easy extension
- Raw-log preservation
- Clear ingestion/processing separation
- Compatibility with an independently developed web app

Do not introduce ML, LLMs, external cloud APIs, or unnecessary infrastructure.

For known-format detection, use deterministic structural detection.

---

## 4. Technology Stack

Use:

- **Go** for ingestion/processing
- **Kafka** as the ingestion/processing boundary
- **Docker Compose** for local infrastructure
- Go standard networking libraries where practical
- Go's standard testing framework
- JSON for internal event serialization
- YAML or JSON for the initial application registry

Avoid unnecessary frameworks.

---

## 5. Proposed Repository Structure

```text
log-platform/
│
├── cmd/
│   └── collector/
│       └── main.go
│
├── internal/
│   ├── ingestion/
│   │   ├── tcp/
│   │   │   └── server.go
│   │   ├── http/
│   │   │   └── server.go
│   │   ├── udp/
│   │   │   └── server.go
│   │   └── syslog/
│   │       └── server.go
│   │
│   ├── auth/
│   │   └── registry.go
│   │
│   ├── event/
│   │   └── event.go
│   │
│   ├── kafka/
│   │   └── producer.go
│   │
│   └── detection/
│       ├── detector.go
│       ├── manager.go
│       ├── json.go
│       ├── syslog.go
│       ├── apache.go
│       └── nginx.go
│
├── configs/
│   └── applications.yaml
│
├── testdata/
│   ├── json/
│   ├── syslog/
│   ├── apache/
│   └── nginx/
│
├── docker-compose.yml
├── go.mod
└── README.md
```

Keep this modular separation even if exact filenames are adjusted.

---

## 6. Ingestion Layer

Expose:

```text
TCP     :5000
HTTP    :5001
UDP     :5002
Syslog  :5003
```

These are intentionally separate ports.

### TCP

Multiple applications must share the same TCP port:

```text
payment-service ──────┐
auth-service ─────────┤
game-service ─────────┼──> TCP :5000
monitoring-service ───┘
```

A single Go listener can accept many simultaneous connections.

For each connection:

1. Establish connection.
2. Authenticate application.
3. Associate connection with authenticated application.
4. Read log events.
5. Wrap each event in the common envelope.
6. Publish to Kafka.

Use a goroutine per connection.

### TCP framing

TCP is a byte stream. Do **not** assume one `Read()` equals one log.

For the first implementation use **newline-delimited events**.

Handle:

- One log split across multiple reads
- Multiple logs in one read
- Client disconnects
- Empty lines
- Very long lines within a configured maximum
- Partial data at connection close

### HTTP

Expose:

```text
POST http://localhost:5001/logs
```

Support simple input such as:

```http
Content-Type: text/plain

2026-09-28 23:50:01 ERROR database unavailable
```

or:

```json
{
  "message": "2026-09-28 23:50:01 ERROR database unavailable"
}
```

The handler should authenticate, identify, envelope, and publish. It must not parse the log's internal format.

### UDP

Expose:

```text
UDP :5002
```

Treat each datagram as one incoming event initially.

Document that UDP does not guarantee delivery, ordering, or duplicate prevention.

Because UDP is connectionless, authentication needs an event-level mechanism. Keep this isolated so it can be replaced later.

### Syslog

Expose:

```text
Syslog :5003
```

Preserve the original Syslog message in `raw`. Do not fully parse Syslog at ingestion.

---

## 7. Application Registration and Authentication

Distinguish:

- **Application ID** = who the client claims to be
- **Authentication** = proof of identity
- **Authorization** = what the client is allowed to do

For the prototype use a simple registry:

```yaml
applications:
  - id: payment-service
    api_key: abc123

  - id: auth-service
    api_key: xyz456
```

Do not build a registration UI/database yet.

### TCP authentication

1. Client connects.
2. Client sends application ID + API key.
3. Server validates them.
4. Server associates the connection with that application.
5. Subsequent logs on that connection do not repeat the application ID.

Conceptually:

```text
payment-service
      │
      │ application ID + API key
      ▼
Go TCP :5000
      │
      │ authenticate
      ▼
connection = payment-service
      │
      ├── log
      ├── log
      └── log
             │
             ▼
           Kafka
```

Authenticate once per connection. Do not query the registry for every log line.

Reject unknown applications, invalid API keys, and missing authentication.

Never log API keys.

---

## 8. Common Internal Event Envelope

Every ingestion method must produce the same internal event representation.

Example:

```json
{
  "application_id": "payment-service",
  "protocol": "tcp",
  "received_at": "2026-09-28T23:50:01Z",
  "raw": "2026-09-28 23:50:01 ERROR database unavailable"
}
```

Suggested Go type:

```go
type Event struct {
    ApplicationID string    `json:"application_id"`
    Protocol      string    `json:"protocol"`
    ReceivedAt    time.Time `json:"received_at"`
    Raw           string    `json:"raw"`
}
```

The envelope can gain metadata later, but keep it small.

### Raw-log integrity

`raw` must remain untouched.

For example:

```text
2026-09-28 23:50:01 ERROR   database failed   [retry=3]
```

must survive exactly.

Do not:

- Normalize whitespace
- Change capitalization
- Rewrite timestamps
- Remove characters
- Reformat the message

Metadata may be added around the raw message, but the raw message itself must not be changed.

---

## 9. Kafka

Kafka is the boundary between ingestion and processing.

Initial topic:

```text
raw-logs
```

Architecture:

```text
Applications
     │
     ▼
Go ingestion
     │
     ▼
Kafka: raw-logs
     │
     ▼
Format detection / processing
```

The Go collector publishes serialized event envelopes.

Kafka should allow ingestion and processing to be independently restarted and developed.

### Kafka failure behavior

Do not silently discard logs if Kafka is unavailable.

Use:

- Appropriate acknowledgements
- Producer retries
- Timeouts
- Explicit error handling

A local disk spool may be added later, but do not implement a complicated persistent queue now.

---

## 10. Deterministic Format Detection

Initially detect:

```text
JSON
Syslog
Apache Access Log
Nginx Access Log
```

Do **not** use:

- Machine learning
- LLMs
- External APIs
- Cloud inference
- Probabilistic classification

The environment is intended to be air-gapped and lightweight.

---

## 11. Detector Architecture

Use independent detectors:

```text
detection/
    detector.go
    manager.go
    json.go
    syslog.go
    apache.go
    nginx.go
```

Common interface:

```go
type Detector interface {
    Name() string
    Detect(raw string) DetectionResult
}
```

A result can contain:

```go
type DetectionResult struct {
    Format string
    Status DetectionStatus
    Reason string
}
```

Detector-level states:

```text
MATCH
NO_MATCH
```

Manager-level outcomes:

```text
KNOWN FORMAT
UNKNOWN
AMBIGUOUS
```

---

## 12. Detection Algorithm

Do not initially use confidence scores.

1. Run deterministic detectors.
2. Collect matches.
3. Exactly one match → return that format.
4. No matches → `UNKNOWN`.
5. Multiple matches → `AMBIGUOUS`.

Example:

```text
raw log
   │
   ├── JSON detector ────── NO_MATCH
   ├── Syslog detector ──── NO_MATCH
   ├── Apache detector ──── MATCH
   └── Nginx detector ───── NO_MATCH
              │
              ▼
          format=APACHE
```

If multiple detectors match, do not arbitrarily choose one. Return `AMBIGUOUS`.

Only add deterministic precedence rules or stronger signatures if real ambiguity requires them. Do not introduce scoring prematurely.

---

## 13. JSON Detector

Actually attempt JSON parsing.

Do not just check:

```text
starts with {
```

Return `MATCH` only for valid JSON according to the supported implementation.

Detection may trim input for the purpose of checking, but the event's original `raw` must remain unchanged.

---

## 14. Syslog Detector

Use deterministic Syslog structure.

Typical evidence:

```text
<PRI>timestamp hostname application: message
```

Possible checks:

- Priority structure
- Recognizable timestamp
- Host/application/message structure

Regex is acceptable.

Do not extract all semantic fields at detection time.

---

## 15. Apache Detector

Use structural evidence such as:

- Client IP
- Bracketed timestamp
- Quoted HTTP request
- HTTP method
- HTTP protocol
- Three-digit response status
- Response size

Example:

```text
127.0.0.1 - - [28/Sep/2026:23:50:01 +0530] "GET /login HTTP/1.1" 200 1234
```

Detection should only answer:

> Does this look like Apache?

Do not output semantic fields such as `client_ip`, `timestamp`, `method`, `path`, or `status` yet.

---

## 16. Nginx Detector

Use deterministic structural characteristics of common Nginx access logs.

Apache and Nginx may look similar.

If a log legitimately matches both:

```text
AMBIGUOUS
```

Do not pretend certainty.

If this becomes common, strengthen signatures or use a documented deterministic precedence rule.

---

## 17. Detection Output

A future enriched event can look like:

```json
{
  "application_id": "payment-service",
  "protocol": "tcp",
  "detected_format": "apache",
  "raw": "127.0.0.1 - - [28/Sep/2026:23:50:01 +0530] "GET /login HTTP/1.1" 200 1234"
}
```

At detection stage, `detected_format` is the main new information.

Do not extract all fields yet.

---

## 18. Explainability

Detectors may provide reasons such as:

```text
format: apache

matched:
- IPv4 structure
- bracketed timestamp
- quoted HTTP request
- three-digit HTTP status
- response byte count
```

Keep this modular so it can later be exposed to an API or omitted from production event payloads.

---

## 19. Regex and Grok

Regex is a deterministic pattern-matching mechanism.

Grok is a higher-level pattern language built from reusable regex-style patterns and named captures.

Do not reproduce Logstash's full Grok system.

Use normal Go regex/parsing unless Grok compatibility becomes an explicit requirement.

Keep this distinction:

```text
Detection:
"Which format is this?"

Parsing:
"What does each part mean?"
```

Example:

```text
Raw Apache log
      │
      ▼
Apache detector
      │
      ▼
"APACHE"
      │
      ▼
Apache parser
      │
      ▼
client_ip, timestamp, method, path, status, bytes
      │
      ▼
OCSF normalizer
```

---

## 20. Future Parser / OCSF Boundary

Full parsing is out of scope now, but design for:

```text
Kafka raw-logs
      │
      ▼
Format Detector
      │
      ├── JSON
      ├── SYSLOG
      ├── APACHE
      └── NGINX
             │
             ▼
          Parser
             │
             ▼
      Structured Event
             │
             ▼
       OCSF Normalizer
             │
             ▼
      Storage / API
```

Detector code must not be coupled to the dashboard.

---

## 21. Future Web Application Integration

A different team member is developing the production web application/dashboard.

**Do not implement any dashboard or frontend in this repository.**

The current implementation should only provide clean backend interfaces that make future integration possible.

The processing system should eventually expose structured data through a clean, documented interface such as a REST API or another agreed backend interface.

Keep the output/storage boundary behind an abstraction:

```text
Processing pipeline
        │
        ▼
ProcessedEvent
        │
        ▼
Output / Storage Interface
        │
        ├── future REST API
        ├── future database
        └── future message stream
```

Do not hard-code assumptions about:

- React/Vue/plain HTML
- PostgreSQL/SQLite/etc.
- frontend routes
- frontend components
- final dashboard implementation

The other developer should be able to integrate without depending on internal Go package structure.

Do not expose internal Kafka messages as the permanent frontend API contract.

Consider API versioning once the external contract is established.

The **application registry is also part of the backend infrastructure**, not the dashboard. For this implementation phase, use the simple registry configuration described in the authentication section. A future administrative interface may manage the registry, but that is outside the current scope.

## 22. Future Unknown-Format Pipeline

Unknown formats are handled by a separate future pipeline.

The detector must still return:

```text
UNKNOWN
```

Future routing:

```text
                    Format Detection
                           │
             ┌─────────────┴─────────────┐
             │                           │
         Known format                UNKNOWN
             │                           │
             ▼                           ▼
          Parser                 Unknown-format
             │                     pipeline
             ▼
       Normalization
```

Do not implement the unknown pipeline now.

---

## 23. Testing

### Detector unit tests

For every detector, include:

- Positive cases
- Negative cases
- Malformed input
- Empty input
- Whitespace
- Very long input
- Ambiguous cases

### TCP tests

Test:

1. Multiple applications simultaneously
2. Multiple logs in one read
3. One log split across reads
4. Client disconnect
5. Empty input
6. Very long input
7. Invalid authentication
8. Unknown application ID
9. Correct application identity preserved

Concurrent test clients:

```text
payment-service
auth-service
game-service
monitoring-service
```

All use:

```text
TCP :5000
```

Verify there are no identity-mixing errors.

---

## 24. Integration Tests

Run Kafka + collector and verify:

```text
payment-service → TCP :5000 → Kafka
payment-service → HTTP :5001 → Kafka
payment-service → UDP :5002 → Kafka
source → Syslog :5003 → Kafka
```

Verify:

- protocol metadata
- application ID
- raw message
- Kafka delivery

---

## 25. Detection Tests

Expected:

```text
JSON example       → JSON
Syslog example     → SYSLOG
Apache example     → APACHE
Nginx example      → NGINX
Unknown example    → UNKNOWN
Ambiguous example  → AMBIGUOUS
```

---

## 26. Raw Integrity Test

Send:

```text
2026-09-28 23:50:01 ERROR   database failed   [retry=3]
```

Verify the exact string survives:

```text
ingestion
    ↓
event envelope
    ↓
Kafka
```

No whitespace normalization is allowed.

---

## 27. Failure Tests

Test:

- Invalid API key
- Unknown application ID
- Kafka unavailable
- Kafka timeout
- Client disconnect
- Malformed HTTP request
- Empty TCP message
- Very large message
- Multiple TCP messages in one read
- One TCP message split across reads

Fail predictably and report useful errors.

---

## 28. Docker Compose

Provide local infrastructure with Kafka and the minimum dependencies required.

README must explain:

1. Start infrastructure
2. Start collector
3. Register sample applications
4. Send sample logs
5. Verify Kafka messages
6. Run tests

No cloud service or external API should be required.

---

## 29. Configuration

Do not hard-code important operational settings.

Configure:

```text
TCP port
HTTP port
UDP port
Syslog port
Kafka broker
Kafka topic
Application registry location
Maximum log size
Authentication settings
```

Use environment variables and/or configuration files with sensible local defaults.

---

## 30. Error Handling

Do not ignore errors.

Handle:

- Listener startup failure
- Authentication failure
- Invalid configuration
- Kafka connection failure
- Kafka publish failure
- Malformed HTTP
- Oversized messages
- Connection read failure

Do not log secrets.

---

## 31. Backpressure and Reliability

Do not assume infinite throughput.

Use bounded buffers where appropriate.

If Kafka becomes slower than incoming traffic, define controlled behavior instead of allowing unbounded memory growth.

Initial acceptable behavior:

```text
Incoming logs
      │
      ▼
bounded processing/publish path
      │
      ├── Kafka accepts → success
      │
      └── Kafka unavailable/overloaded → controlled failure
```

A disk-backed spool can be added later if stronger durability is required.

---

## 32. Observability

Keep the design ready for metrics such as:

```text
logs_received_total
logs_processed_total
logs_failed_total
logs_unknown_total

kafka_produce_success_total
kafka_produce_failure_total

active_connections
bytes_received

processing_latency
kafka_lag
```

A complete metrics stack is optional for this phase.

---

## 33. Scalability

The first version can be a single collector:

```text
Sources → Go Collector → Kafka
```

The architecture should later support:

```text
                ┌── Go Collector 1 ──┐
Sources ────────┼── Go Collector 2 ──┼──> Kafka
                └── Go Collector 3 ──┘
```

Do not introduce Kubernetes or complex service discovery prematurely.

---

## 34. Security

Minimum requirements:

- Authenticate applications.
- Never expose API keys in logs.
- Validate request sizes.
- Reject unknown applications.
- Do not trust client-provided identity without authentication.
- Preserve raw logs.
- Keep authentication mechanisms explicit.
- Keep secrets configurable.

TLS can be added later. Keep the network layer modular enough that secure transport can be introduced without rewriting the system.

---

## 35. Architectural Principles

### Ingestion does not parse

```text
receive
→ authenticate
→ identify
→ wrap
→ Kafka
```

Not:

```text
receive
→ parse
→ normalize
→ modify
```

### Detection does not fully parse

Detection:

```text
"Is this Apache?"
```

Parsing:

```text
"What is the IP?
 What is the timestamp?
 What is the HTTP method?
 What is the status?"
```

### Raw data is preserved

Never modify the original log.

### Kafka decouples stages

Ingestion and processing can be restarted independently.

### Web application is separate

No dashboard is implemented here. The future web application should consume a clean backend contract rather than depend on ingestion internals.

### Known-format detection is deterministic

No ML is required for known formats.

### Unknown is valid

The system must distinguish:

```text
KNOWN FORMAT
UNKNOWN
AMBIGUOUS
```

---

## 36. README Requirements

Document:

### Project purpose

What problem the platform solves.

### Architecture

Include the complete pipeline diagram.

### Supported protocols

```text
TCP :5000
HTTP :5001
UDP :5002
Syslog :5003
```

### Authentication

Application registration and API keys.

### Kafka

The `raw-logs` topic.

### Format detection

Deterministic detection and supported formats.

### Running locally

Exact commands.

### Testing

Exact commands.

### Sending sample logs

Examples for TCP, HTTP, UDP, and Syslog.

### Future architecture

```text
Detection
→ Parsing
→ OCSF normalization
→ Storage/API
→ Existing Web Dashboard
```

Clearly state that the dashboard is intentionally external to this repository.

---

## 37. Definition of Done

- [ ] Go collector builds successfully.
- [ ] TCP server works on port 5000.
- [ ] HTTP server works on port 5001.
- [ ] UDP server works on port 5002.
- [ ] Syslog ingestion works on port 5003.
- [ ] Multiple applications can share the TCP port.
- [ ] Applications are authenticated.
- [ ] Application identity is associated with events.
- [ ] TCP authentication occurs once per connection.
- [ ] Common event envelope exists.
- [ ] Raw logs are preserved exactly.
- [ ] Kafka producer works.
- [ ] `raw-logs` topic is used.
- [ ] Kafka failures are handled explicitly.
- [ ] JSON detection works.
- [ ] Syslog detection works.
- [ ] Apache detection works.
- [ ] Nginx detection works.
- [ ] UNKNOWN is supported.
- [ ] AMBIGUOUS is supported.
- [ ] Detection is deterministic.
- [ ] Detector modules are independently testable.
- [ ] TCP framing edge cases are tested.
- [ ] Concurrent ingestion is tested.
- [ ] Docker Compose runs locally.
- [ ] README contains setup/test instructions.
- [ ] No dashboard or frontend is implemented.
- [ ] Future API/storage boundary is modular and replaceable.
- [ ] No cloud service is required.
- [ ] No ML or LLM is required.

---

## 38. Recommended Implementation Order

Build incrementally:

```text
1. Go project structure
        ↓
2. Configuration
        ↓
3. Application registry/authentication
        ↓
4. Common Event type
        ↓
5. TCP ingestion
        ↓
6. Kafka producer
        ↓
7. TCP → Kafka
        ↓
8. HTTP ingestion
        ↓
9. UDP ingestion
        ↓
10. Syslog ingestion
        ↓
11. Detector interface
        ↓
12. JSON detector
        ↓
13. Syslog detector
        ↓
14. Apache detector
        ↓
15. Nginx detector
        ↓
16. Detection manager
        ↓
17. Unit tests
        ↓
18. Integration tests
        ↓
19. Docker Compose
        ↓
20. Backend output/storage interface
        ↓
21. README
        ↓
22. Clean future interfaces for parser,
    OCSF, storage/API, and web application
```

Do not create one huge monolithic implementation.

Keep each component independently testable.

The critical architectural boundary is:

```text
                    CURRENT PHASE
                         │
                         ▼
Sources → Go Ingestion → Kafka → Deterministic Detection
                                      │
                                      ▼
                                detected format


                    FUTURE PHASE
                                      │
                                      ▼
                                   Parser
                                      │
                                      ▼
                              OCSF Normalizer
                                      │
                                      ▼
                              Storage / API
                                      │
                                      ▼
                          Existing Web Application
                                      │
                                      ▼
                                  Dashboard
```

The final implementation should be easy for another developer to extend without rewriting the ingestion layer.
