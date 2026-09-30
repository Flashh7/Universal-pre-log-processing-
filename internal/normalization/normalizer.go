package normalization

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sih/log-platform/internal/classification"
	"github.com/sih/log-platform/internal/parsing"
)

type Metadata struct {
	Version string `json:"version"`
}

type HTTPRequest struct {
	HTTPMethod string `json:"http_method,omitempty"`
	URL        URL    `json:"url,omitempty"`
}

type URL struct {
	Path string `json:"path,omitempty"`
}

type HTTPResponse struct {
	Code int `json:"code,omitempty"`
	Size int `json:"size,omitempty"`
}

type Endpoint struct {
	IP string `json:"ip,omitempty"`
}

type User struct {
	Name string `json:"name,omitempty"`
}

type NormalizedEvent struct {
	Metadata Metadata `json:"metadata"`
	EventUID string   `json:"event_uid"`
	Time     int64    `json:"time"`

	CategoryName string `json:"category_name"`
	CategoryUID  int    `json:"category_uid"`

	ClassName string `json:"class_name"`
	ClassUID  int    `json:"class_uid"`

	ActivityName string `json:"activity_name"`
	ActivityID   int    `json:"activity_id"`

	TypeName string `json:"type_name"`
	TypeUID  int    `json:"type_uid"`

	Severity   string `json:"severity"`
	SeverityID int    `json:"severity_id"`

	Status   string `json:"status"`
	StatusID int    `json:"status_id"`

	RawData string `json:"raw_data"`

	HTTPRequest  *HTTPRequest  `json:"http_request,omitempty"`
	HTTPResponse *HTTPResponse `json:"http_response,omitempty"`
	SrcEndpoint  *Endpoint     `json:"src_endpoint,omitempty"`
	User         *User         `json:"user,omitempty"`

	Unmapped map[string]any `json:"unmapped,omitempty"`
}

type Normalizer interface {
	Normalize(pe parsing.ParsedEvent, cr classification.ClassificationResult, rawEvent string) (NormalizedEvent, error)
}

type OCSFNormalizer struct{}

func NewNormalizer() *OCSFNormalizer {
	return &OCSFNormalizer{}
}

func (n *OCSFNormalizer) Normalize(pe parsing.ParsedEvent, cr classification.ClassificationResult, rawEvent string) (NormalizedEvent, error) {
	eventTime := time.Now().UTC().UnixMilli()
	if pe.Timestamp != nil {
		eventTime = pe.Timestamp.UTC().UnixMilli()
	}

	var ev NormalizedEvent
	ev.Metadata = Metadata{Version: "1.7.0"}
	ev.EventUID = uuid.New().String()
	ev.Time = eventTime
	ev.RawData = rawEvent

	switch cr.EventType {
	case "HTTP_ACTIVITY":
		ev = n.normalizeHTTP(ev, pe)
	case "AUTHENTICATION":
		ev = n.normalizeAuthentication(ev, pe)
	default:
		ev = n.normalizeBaseEvent(ev, pe)
	}

	return ev, nil
}

func (n *OCSFNormalizer) normalizeHTTP(ev NormalizedEvent, pe parsing.ParsedEvent) NormalizedEvent {
	ev.CategoryName = "Network Activity"
	ev.CategoryUID = 4
	ev.ClassName = "HTTP Activity"
	ev.ClassUID = 4002

	ev.ActivityID = httpMethodToActivityID(pe.Fields)
	ev.ActivityName = httpActivityName(ev.ActivityID)
	ev.TypeUID = ev.ClassUID*100 + ev.ActivityID
	ev.TypeName = fmt.Sprintf("HTTP Activity: %s", ev.ActivityName)

	ev.SeverityID, ev.Severity = httpStatusToSeverity(pe.Fields)
	ev.StatusID, ev.Status = httpStatusToStatus(pe.Fields)

	req := &HTTPRequest{}
	res := &HTTPResponse{}
	src := &Endpoint{}
	mappedKeys := make(map[string]bool)

	for k, v := range pe.Fields {
		kl := strings.ToLower(k)
		switch kl {
		case "method", "request_method":
			if s, ok := v.(string); ok { req.HTTPMethod = strings.ToUpper(s) }
			mappedKeys[k] = true
		case "path", "uri", "request_uri", "url", "request_target":
			if s, ok := v.(string); ok { req.URL.Path = s }
			mappedKeys[k] = true
		case "status":
			res.Code = toInt(v)
			mappedKeys[k] = true
		case "bytes", "size", "bytes_sent":
			res.Size = toInt(v)
			mappedKeys[k] = true
		case "client_ip", "remote_addr", "ip":
			if s, ok := v.(string); ok { src.IP = s }
			mappedKeys[k] = true
		}
	}

	if req.HTTPMethod != "" || req.URL.Path != "" { ev.HTTPRequest = req }
	if res.Code != 0 || res.Size != 0 { ev.HTTPResponse = res }
	if src.IP != "" { ev.SrcEndpoint = src }

	unmapped := make(map[string]any)
	for k, v := range pe.Fields {
		if !mappedKeys[k] {
			unmapped[k] = v
		}
	}
	if len(unmapped) > 0 {
		ev.Unmapped = unmapped
	}

	return ev
}

func (n *OCSFNormalizer) normalizeAuthentication(ev NormalizedEvent, pe parsing.ParsedEvent) NormalizedEvent {
	ev.CategoryName = "Identity & Access Management"
	ev.CategoryUID = 3
	ev.ClassName = "Authentication"
	ev.ClassUID = 3002

	ev.ActivityID = authActivityID(pe.Fields)
	ev.ActivityName = authActivityName(ev.ActivityID)
	ev.TypeUID = ev.ClassUID*100 + ev.ActivityID
	ev.TypeName = fmt.Sprintf("Authentication: %s", ev.ActivityName)
	ev.SeverityID, ev.Severity = authSeverity(pe.Fields)
	ev.StatusID, ev.Status = authStatus(pe.Fields)

	src := &Endpoint{}
	user := &User{}
	mappedKeys := make(map[string]bool)

	for k, v := range pe.Fields {
		kl := strings.ToLower(k)
		switch kl {
		case "user", "username", "user_name", "account", "uid":
			if s, ok := v.(string); ok { user.Name = s }
			mappedKeys[k] = true
		case "client_ip", "remote_addr", "ip", "src_ip":
			if s, ok := v.(string); ok { src.IP = s }
			mappedKeys[k] = true
		}
	}

	if src.IP != "" { ev.SrcEndpoint = src }
	if user.Name != "" { ev.User = user }

	unmapped := make(map[string]any)
	for k, v := range pe.Fields {
		if !mappedKeys[k] {
			unmapped[k] = v
		}
	}
	if len(unmapped) > 0 {
		ev.Unmapped = unmapped
	}

	return ev
}

func (n *OCSFNormalizer) normalizeBaseEvent(ev NormalizedEvent, pe parsing.ParsedEvent) NormalizedEvent {
	ev.CategoryName = "Uncategorized"
	ev.CategoryUID = 0
	ev.ClassName = "Base Event"
	ev.ClassUID = 0
	ev.ActivityID = 0
	ev.ActivityName = "Unknown"
	ev.TypeUID = ev.ClassUID*100 + ev.ActivityID
	ev.TypeName = "Base Event: Unknown"
	ev.SeverityID = 0
	ev.Severity = "Unknown"
	ev.StatusID = 0
	ev.Status = "Unknown"

	if len(pe.Fields) > 0 {
		ev.Unmapped = make(map[string]any)
		for k, v := range pe.Fields {
			ev.Unmapped[k] = v
		}
	}

	return ev
}

func httpMethodToActivityID(fields map[string]any) int {
	method := ""
	for k, v := range fields {
		kl := strings.ToLower(k)
		if kl == "method" || kl == "request_method" {
			if s, ok := v.(string); ok {
				method = strings.ToUpper(s)
			}
		}
	}

	switch method {
	case "CONNECT": return 1
	case "DELETE": return 2
	case "GET": return 3
	case "HEAD": return 4
	case "OPTIONS": return 5
	case "POST": return 6
	case "PUT": return 7
	case "TRACE": return 8
	case "PATCH": return 9
	default:
		if method != "" { return 99 }
		return 0
	}
}

func httpActivityName(activityID int) string {
	switch activityID {
	case 1: return "Connect"
	case 2: return "Delete"
	case 3: return "Get"
	case 4: return "Head"
	case 5: return "Options"
	case 6: return "Post"
	case 7: return "Put"
	case 8: return "Trace"
	case 9: return "Patch"
	case 99: return "Other"
	default: return "Unknown"
	}
}

func httpStatusToSeverity(fields map[string]any) (int, string) {
	status := 0
	for k, v := range fields {
		if strings.ToLower(k) == "status" {
			switch sv := v.(type) {
			case int: status = sv
			case float64: status = int(sv)
			}
		}
	}

	switch {
	case status == 0: return 0, "Unknown"
	case status < 400: return 1, "Informational"
	case status < 500: return 2, "Low"
	default: return 3, "Medium"
	}
}

func httpStatusToStatus(fields map[string]any) (int, string) {
	status := 0
	for k, v := range fields {
		if strings.ToLower(k) == "status" {
			switch sv := v.(type) {
			case int: status = sv
			case float64: status = int(sv)
			}
		}
	}

	switch {
	case status == 0: return 0, "Unknown"
	case status < 400: return 1, "Success"
	default: return 2, "Failure"
	}
}

func authActivityID(fields map[string]any) int {
	for k, v := range fields {
		kl := strings.ToLower(k)
		if kl == "action" || kl == "event" || kl == "event_type" || kl == "type" || kl == "operation" || kl == "message" {
			if s, ok := v.(string); ok {
				sl := strings.ToLower(s)
				if strings.Contains(sl, "logout") || strings.Contains(sl, "logoff") || strings.Contains(sl, "signout") || strings.Contains(sl, "sign_out") {
					return 2
				}
				if strings.Contains(sl, "login") || strings.Contains(sl, "logon") || strings.Contains(sl, "signin") || strings.Contains(sl, "sign_in") || strings.Contains(sl, "auth") {
					return 1
				}
			}
		}
	}
	return 0
}

func authActivityName(activityID int) string {
	switch activityID {
	case 1: return "Logon"
	case 2: return "Logoff"
	case 99: return "Other"
	default: return "Unknown"
	}
}

func authSeverity(fields map[string]any) (int, string) {
	for k, v := range fields {
		kl := strings.ToLower(k)
		if kl == "result" || kl == "status" || kl == "outcome" || kl == "auth_result" || kl == "message" {
			if s, ok := v.(string); ok {
				sl := strings.ToLower(s)
				if strings.Contains(sl, "fail") || strings.Contains(sl, "denied") || strings.Contains(sl, "rejected") {
					return 3, "Medium"
				}
				if strings.Contains(sl, "success") || strings.Contains(sl, "granted") || strings.Contains(sl, "accepted") {
					return 1, "Informational"
				}
			}
		}
	}
	return 0, "Unknown"
}

func authStatus(fields map[string]any) (int, string) {
	for k, v := range fields {
		kl := strings.ToLower(k)
		if kl == "result" || kl == "status" || kl == "outcome" || kl == "auth_result" || kl == "message" {
			if s, ok := v.(string); ok {
				sl := strings.ToLower(s)
				if strings.Contains(sl, "success") || strings.Contains(sl, "granted") || strings.Contains(sl, "accepted") {
					return 1, "Success"
				}
				if strings.Contains(sl, "fail") || strings.Contains(sl, "denied") || strings.Contains(sl, "rejected") {
					return 2, "Failure"
				}
			}
		}
	}
	return 0, "Unknown"
}

// toInt safely converts a value to int, handling both int and float64
// (Go's encoding/json unmarshals JSON numbers as float64 into map[string]any).
func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	default:
		return 0
	}
}
