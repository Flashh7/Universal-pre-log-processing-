# Parser + OCSF Normalization Implementation Specification

## Goal

Extend the existing Go log ingestion/preprocessing platform with:

Raw Event -> Format Detection -> Parser -> ParsedEvent -> Deterministic Event Classification -> OCSF Normalization

Keep the existing collector, authentication, Kafka, and detection architecture. Do not rewrite working code.

The project must remain **air-gapped**: no runtime internet access, cloud LLMs/APIs, online parser services, or runtime schema downloads.

Do NOT implement the future unknown-format ML/regex-generation system yet.
Do NOT implement the dashboard/frontend.

## 1. Responsibilities

### Format detection
Answers: "What format is this?"
Examples: Apache, Nginx, JSON, Syslog, UNKNOWN, AMBIGUOUS.

### Parsing
Answers: "Given this known format/layout, what structured fields are present?"

### Event classification
Answers: "What does this parsed event represent?"
Examples: HTTP Activity, Authentication, Network Activity, File Activity, Process Activity.

### OCSF normalization
Answers: "How should this event be represented using OCSF?"

Keep these layers separate.

## 2. Parser strategy

Implement deterministic, format-specific parsers.

For Apache, Nginx, and Syslog, Go's standard `regexp` package is acceptable and preferred initially.

For JSON, use `encoding/json`; do NOT use regex.

Do not implement Grok or a Grok compatibility layer.

Conceptually:

Apache raw --regex--> ParsedEvent
Nginx raw --regex--> ParsedEvent
Syslog raw --regex--> ParsedEvent
JSON raw --JSON--> ParsedEvent

## 3. Format/layout limitation

Apache and Nginx access-log layouts are configurable. "Apache" does not mean every Apache layout, and "Nginx" does not mean every Nginx layout.

Explicitly support a documented set of layouts, such as Apache Common/Combined, the selected Nginx layout, supported Syslog, and JSON.

If a detected Apache/Nginx log does not match a supported layout, return an explicit unsupported-layout/parse error. Do not guess.

If the exact supported layouts cannot be determined safely from the repository/project, ASK ME before implementing.

## 4. ParsedEvent

Use/adapt:

```go
type ParsedEvent struct {
    Timestamp *time.Time
    EventType string
    Fields    map[string]any
}
```

This is an internal intermediate representation, NOT OCSF.

Example:

```json
{
  "timestamp": "2026-09-30T14:32:10+05:30",
  "event_type": "http_activity",
  "fields": {
    "client_ip": "192.168.1.20",
    "method": "GET",
    "path": "/login",
    "protocol": "HTTP/1.1",
    "status": 401,
    "bytes": 532
  }
}
```

## 5. Parser interface

Use/adapt:

```go
type Parser interface {
    Name() string
    Parse(raw string) (ParsedEvent, error)
}
```

Each parser validates its supported layout, extracts fields, converts useful types, and returns explicit errors on failure.

## 6. Apache parser

Support only documented Apache layout(s). Depending on layout, extract applicable:

- client IP
- identity
- user
- timestamp
- HTTP method
- request target/path
- HTTP version
- status
- response bytes
- referrer
- user agent

Do not assume every Apache server has the same fields.

## 7. Nginx parser

Support only documented Nginx layout(s). Depending on layout, extract applicable:

- remote address
- timestamp
- request method
- request URI
- HTTP protocol
- status
- bytes sent
- referrer
- user agent

Do not guess unsupported layouts.

## 8. JSON parser

Use `encoding/json`. Preserve arbitrary source fields. Do not infer semantic event type merely from field names; classification is separate.

## 9. Syslog parser

Extract applicable Syslog envelope fields such as:

- timestamp
- hostname
- process/application
- PID where available
- message

Do not assume the Syslog message has a specific semantic type.

## 10. Deterministic event classification

Create a separate classifier, e.g.:

```go
type ClassificationStatus string

const (
    ClassificationConfident ClassificationStatus = "CONFIDENT"
    ClassificationAmbiguous ClassificationStatus = "AMBIGUOUS"
    ClassificationUnknown   ClassificationStatus = "UNKNOWN"
)

type ClassificationResult struct {
    Status     ClassificationStatus
    EventType  string
    Candidates []string
    Reason     string
}
```

Operate on ParsedEvent, not raw strings.

Do not use ML initially.

Use evidence-based deterministic rules rather than one giant rigid regex.

Examples:

HTTP: HTTP method + request target/path + HTTP version.

Authentication: authentication/login-related action + account/user identity + authentication result/status.

Network: source endpoint + destination endpoint + network protocol/connection evidence.

File: file/path object + file operation/action.

Do not classify solely from source format:
Apache != automatically HTTP; JSON != automatically Authentication; Syslog != automatically System Activity.

If multiple strong semantic interpretations exist, return AMBIGUOUS. Do not arbitrarily choose one.

## 11. OCSF version

Target **OCSF 1.7.0** and pin/document it.

Use the actual OCSF 1.7.0 schema as the source of truth. Do not invent IDs, field names, enum values, or classes.

All required schema information must be locally available for the air-gapped deployment.

## 12. Minimal OCSF output for now

Keep the first implementation deliberately small. The normalized output should expose these requested OCSF base fields:

- event_uid
- time

- category_name
- category_uid

- class_name
- class_uid

- activity_name
- activity_id

- type_name
- type_uid

- severity
- severity_id

- status
- status_id

- raw_event

Do NOT implement the entire OCSF hierarchy yet. Later we can add fields such as src_endpoint, dst_endpoint, actor/user, device, process, file, http_request, resources, cloud, etc.

The design must make those additions easy later.

## 13. OCSF fallback

OCSF has a concrete **Base Event** for generic events that do not belong to another defined event category.

For successfully parsed events whose semantic classification is UNKNOWN or AMBIGUOUS, use the actual OCSF Base Event representation as the fallback.

Do NOT invent GenericEvent, UnknownEvent, UniversalEvent, OCSFUnknown, or similar fake classes.

Use the actual OCSF 1.7 values for Base Event category/class/activity/type/severity/status.

## 14. Unknown format vs unknown event

Keep these separate.

### Unknown/ambiguous format

The system cannot reliably parse the raw event.

```text
raw log -> UNKNOWN_FORMAT / AMBIGUOUS_FORMAT
```

Stop before OCSF normalization.

### Unknown/ambiguous event

The system successfully parsed the log but cannot confidently determine its semantic class.

```text
raw -> known format -> parser -> ParsedEvent -> UNKNOWN/AMBIGUOUS EVENT -> OCSF Base Event
```

This distinction is mandatory.

## 15. OCSF normalizer

Create a separate normalizer, e.g.:

```go
type Normalizer interface {
    Normalize(event ParsedEvent, classification ClassificationResult) (NormalizedEvent, error)
}
```

It should select the appropriate OCSF class, map fields, populate required base attributes, preserve raw data, and validate the result.

Do NOT make normalization format-specific. Prefer semantic mappings such as HTTPActivity, Authentication, NetworkActivity, FileActivity.

This allows Apache and Nginx to map to the same semantic OCSF class.

## 16. OCSF field mapping

Map fields explicitly and verify semantic compatibility against OCSF 1.7.

For the minimal output:

Parsed timestamp -> OCSF time.

Classification -> category_name/category_uid, class_name/class_uid, activity_name/activity_id, type_name/type_uid.

Source severity -> severity/severity_id.

Source outcome -> status/status_id.

Original source log -> raw_event in the platform's output; use OCSF's actual raw-data semantics where applicable.

Do not map merely because names look similar.

OCSF documentation states that `_uid` fields identify schema values and `_id` fields are enum values. `type_uid` is derived as:

type_uid = class_uid * 100 + activity_id

Do not independently invent type_uid.

## 17. Severity and status

Use actual OCSF 1.7 enums.

If severity is unknown, use the official OCSF Unknown value rather than inventing one.

If status is unknown, use the official OCSF Unknown value. If a source status does not map cleanly, use the official Other mechanism and preserve the source-specific label as required by the schema.

Do not invent security severity judgments without an explicit project rule.

## 18. Unmapped fields

Do not silently discard parsed fields that are not represented in the current minimal output.

OCSF provides an `unmapped` attribute for source fields not mapped to the schema. Use the actual OCSF mechanism where appropriate.

Keep a clear distinction between OCSF-standard fields and source-specific/unmapped fields.

Do not invent fake core OCSF attributes.

## 19. Raw event

Preserve the original raw log exactly as received.

Do not alter whitespace, capitalization, timestamps, escaping, or content.

If our external field is called `raw_event` while OCSF 1.7 calls the corresponding core attribute `raw_data`, clearly document the distinction rather than falsely claiming `raw_event` is a core OCSF attribute.

## 20. Air-gapped requirement

The implementation must work without internet access at runtime.

No:
- cloud APIs
- cloud LLMs
- OpenAI/Gemini/Claude APIs
- online parser services
- runtime OCSF downloads
- runtime Grok-pattern downloads
- online validation services

Vendor/package the required OCSF schema information locally or use another fully local build-time strategy.

Document the pinned OCSF version and local schema dependency.

## 21. Future unknown-format engine — not now

Later we may implement:

UNKNOWN FORMAT -> local pattern/ML discovery -> candidate regex/parser -> validation -> new supported layout.

Do not implement this now.

For now:
- UNKNOWN FORMAT -> stop/report
- AMBIGUOUS FORMAT -> stop/report

## 22. Suggested repository structure

Adapt to existing code; do not blindly duplicate packages:

```text
internal/
├── detection/
├── parsing/
│   ├── parser.go
│   ├── event.go
│   ├── apache.go
│   ├── nginx.go
│   ├── json.go
│   └── syslog.go
├── classification/
│   ├── classifier.go
│   └── rules.go
└── normalization/
    ├── normalizer.go
    ├── ocsf.go
    └── mappings/
        ├── http.go
        ├── authentication.go
        ├── network.go
        └── ...
```

## 23. Tests

Parser tests:
- valid Apache
- invalid Apache
- supported Apache variations
- unsupported Apache layout
- valid Nginx
- invalid Nginx
- valid JSON
- malformed JSON
- valid Syslog
- malformed Syslog
- empty input
- unexpected fields
- malformed timestamps

Classification tests:
- clear HTTP -> CONFIDENT
- clear Authentication -> CONFIDENT
- clear Network -> CONFIDENT
- insufficient evidence -> UNKNOWN
- multiple strong interpretations -> AMBIGUOUS

Normalization tests:
- HTTP -> expected minimal OCSF fields
- Authentication -> expected minimal OCSF fields
- Network -> expected minimal OCSF fields
- UNKNOWN semantic event -> valid OCSF Base Event
- AMBIGUOUS semantic event -> valid OCSF Base Event

End-to-end:
- raw Apache -> detect -> parse -> classify -> normalize -> validate
- unknown format -> UNKNOWN_FORMAT and does not enter OCSF normalization

## 24. Error states

Keep these distinct:

DETECTION_UNKNOWN
DETECTION_AMBIGUOUS
PARSE_ERROR
UNSUPPORTED_LAYOUT
CLASSIFICATION_UNKNOWN
CLASSIFICATION_AMBIGUOUS
NORMALIZATION_ERROR
OCSF_VALIDATION_ERROR

Do not silently convert one into another.

## 25. Processor integration

The processor should eventually do:

```text
Kafka raw-logs
  ↓
decode envelope
  ↓
format detection
  ↓
if known:
    parser
    ↓
    event classification
    ↓
    OCSF normalization
else:
    UNKNOWN/AMBIGUOUS FORMAT
```

Do not move parsing into the collector.

Do not break the existing Kafka topic or event envelope.

## 26. Observability

Make these processing outcomes observable:

- detected format
- parse success/failure
- classification result
- normalization success/failure
- OCSF validation failure

Never log API keys or authentication secrets.

#
# 30. End-to-End Manual Testing

The implementation must be testable by the developer from log injection all the way through OCSF normalization.

The existing log injection mechanisms are already implemented and must remain usable.

The developer must be able to run the entire pipeline locally in an air-gapped environment:

```text
Log injection
    ↓
Go Collector
    ↓
Kafka raw-logs
    ↓
Processor
    ↓
Format Detection
    ↓
Parser
    ↓
Event Classification
    ↓
OCSF Normalization
    ↓
Terminal/file/API output for inspection
```

Do not require the dashboard/frontend for testing.

## 30.1 Required local test output

Provide a simple way to observe the processor's final normalized output.

For the initial implementation, printing the normalized event to stdout is sufficient.

For example:

```text
FORMAT: Apache
PARSE: SUCCESS
CLASSIFICATION: HTTP_ACTIVITY
NORMALIZATION: SUCCESS
OCSF VALIDATION: PASS
```

followed by the normalized event:

```json
{
  "event_uid": "...",
  "time": "...",
  "category_name": "...",
  "category_uid": "...",
  "class_name": "...",
  "class_uid": "...",
  "activity_name": "...",
  "activity_id": "...",
  "type_name": "...",
  "type_uid": "...",
  "severity": "...",
  "severity_id": "...",
  "status": "...",
  "status_id": "...",
  "raw_event": "..."
}
```

Use the repository's existing logging/output mechanism if one already exists.

## 30.2 Manual test cases

Document commands/examples for injecting at least:

### Test 1 — Apache HTTP event

Inject a supported Apache access log through the existing collector.

Expected:

```text
Detection → Apache
Parsing → SUCCESS
Classification → appropriate HTTP event
Normalization → SUCCESS
OCSF validation → PASS
```

### Test 2 — Nginx HTTP event

Inject a supported Nginx log.

Expected:

```text
Detection → Nginx
Parsing → SUCCESS
Classification → appropriate HTTP event
Normalization → SUCCESS
OCSF validation → PASS
```

### Test 3 — JSON authentication event

Inject a JSON event representing an authentication attempt.

Expected:

```text
Detection → JSON
Parsing → SUCCESS
Classification → Authentication
Normalization → SUCCESS
OCSF validation → PASS
```

### Test 4 — Syslog authentication event

Inject a Syslog event containing a recognizable authentication event.

Expected:

```text
Detection → Syslog
Parsing → SUCCESS
Classification → Authentication
Normalization → SUCCESS
OCSF validation → PASS
```

### Test 5 — Unknown format

Inject a deliberately unsupported log format.

Expected:

```text
Detection → UNKNOWN
No parser
No OCSF normalization
```

The processor must clearly report that processing stopped because the format was unknown.

### Test 6 — Ambiguous format

Inject a log that causes the existing detector to return AMBIGUOUS, if such a case can be constructed with the current detectors.

Expected:

```text
Detection → AMBIGUOUS
No parser
No OCSF normalization
```

### Test 7 — Known format but unsupported layout

Inject a log that is classified as Apache/Nginx but intentionally does not match one of the supported layouts.

Expected:

```text
Detection → Apache/Nginx
Parsing → UNSUPPORTED_LAYOUT / PARSE_ERROR
No OCSF normalization
```

### Test 8 — Parsed but semantically ambiguous/unknown event

Provide a known-format log whose parsed fields do not provide enough evidence to confidently choose a semantic event type.

Expected:

```text
Detection → known format
Parsing → SUCCESS
Classification → UNKNOWN or AMBIGUOUS
Normalization → OCSF Base Event
Normalization → SUCCESS
OCSF validation → PASS
```

This verifies that semantic uncertainty does not prevent standardized output.

## 30.3 Raw-event preservation test

For every successful normalization test:

1. Save the exact injected raw log.
2. Compare it with `raw_event` in the final normalized output.
3. They must be identical.

No whitespace, capitalization, timestamp, or escaping changes are allowed.

## 30.4 OCSF validation test

Every successful normalized output must be validated against the pinned OCSF 1.7 schema.

Make validation failures obvious.

Example:

```text
NORMALIZATION: SUCCESS
OCSF VALIDATION: PASS
```

or:

```text
NORMALIZATION: SUCCESS
OCSF VALIDATION: FAIL
reason: ...
```

## 30.5 Local workflow

Where practical, provide a simple documented workflow such as:

Terminal 1:

```bash
docker compose up -d
```

Terminal 2:

```bash
go run cmd/collector/main.go
```

Terminal 3:

```bash
go run cmd/processor/main.go
```

Terminal 4:

```text
existing log injection command
```

The exact commands, ports, topic names, and paths MUST be taken from the existing repository/configuration. Do not invent or duplicate configuration.

The README should explain what each terminal is running.

## 30.6 Test fixtures

Store representative local fixtures, for example:

```text
testdata/
    apache/
    nginx/
    json/
    syslog/
    unknown/
```

Fixtures must not require internet access.

Where practical, reuse the same fixtures for unit tests and manual end-to-end testing.

## 30.7 Debug visibility

During development, make each processing stage visible:

```text
RAW RECEIVED
    ↓
FORMAT: APACHE
    ↓
PARSE: SUCCESS
    ↓
CLASSIFICATION: HTTP_ACTIVITY
    ↓
NORMALIZATION: SUCCESS
    ↓
OCSF VALIDATION: PASS
```

This is development/debug output and must not expose API keys, authentication secrets, or other sensitive credentials.

## 30.8 End-to-end definition of done

The implementation is not complete until the developer can personally:

```text
inject a log
    ↓
observe it enter Kafka
    ↓
observe format detection
    ↓
observe parsing
    ↓
observe classification
    ↓
observe OCSF normalization
    ↓
inspect the final OCSF event
    ↓
verify OCSF validation
```

All of this must work without the dashboard/frontend and without internet access.

# 27. If Antigravity finds a better implementation

If you discover a materially better implementation approach while inspecting the repository, OCSF schema, or Go ecosystem, **STOP before making that architectural change and ASK ME.**

Explain:
1. current approach
2. proposed alternative
3. why it is better
4. files/components affected
5. air-gapped implications
6. collector/Kafka implications
7. OCSF-output implications

Small implementation improvements that do not change the architecture can be made normally.

Changes that require asking include:
- replacing regex parsers with a substantially different parser architecture
- substantially changing ParsedEvent
- changing OCSF schema strategy
- introducing a new runtime dependency
- changing Kafka boundaries
- changing unknown/ambiguous handling
- introducing an external service
- changing the air-gapped model

## 28. If there are holes

Inspect the repository first.

If an essential detail cannot be determined from the repository or this specification, ASK ME rather than guessing.

Especially ask about:
1. exact Apache layouts
2. exact Nginx layouts
3. existing detection-manager interfaces
4. existing event-envelope structure
5. how detected format reaches the processor
6. OCSF schema packaging if not already present
7. conflicts between existing code and this specification

Do not ask questions whose answers can safely be determined from existing code or the pinned OCSF schema.

## 29. Definition of done

Complete only when:
- existing collector works
- existing Kafka flow works
- existing authentication works
- existing detection works
- Apache parser works for documented layouts
- Nginx parser works for documented layouts
- JSON parser works
- Syslog parser works
- all parsers return ParsedEvent
- parsing/classification/normalization are separate
- classification is deterministic
- CONFIDENT/AMBIGUOUS/UNKNOWN are supported
- known semantic events map to appropriate OCSF 1.7 classes
- unknown/ambiguous semantic events use OCSF Base Event
- minimal requested OCSF fields are produced
- raw event is preserved
- unmapped data is preserved appropriately
- OCSF IDs/names are schema-correct
- type_uid follows OCSF
- OCSF output is validated
- tests pass
- no runtime internet dependency exists
- no ML/LLM unknown-format generator exists
- no dashboard/frontend is added
- documentation is updated
- any materially better architecture discovered by Antigravity was presented to me before adoption
