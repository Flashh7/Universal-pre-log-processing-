package classification

import (
	"strings"

	"github.com/sih/log-platform/internal/parsing"
)

// ClassificationStatus represents the confidence of a classification.
type ClassificationStatus string

const (
	ClassificationConfident ClassificationStatus = "CONFIDENT"
	ClassificationAmbiguous ClassificationStatus = "AMBIGUOUS"
	ClassificationUnknown   ClassificationStatus = "UNKNOWN"
)

// ClassificationResult contains the outcome of event classification.
type ClassificationResult struct {
	Status     ClassificationStatus
	EventType  string   // e.g. "HTTP_ACTIVITY", "AUTHENTICATION"
	Candidates []string // populated when AMBIGUOUS
	Reason     string
}

// Classifier determines the semantic event type from parsed fields.
type Classifier struct {
	rules []Rule
}

// NewClassifier creates a classifier with the default rule set.
func NewClassifier() *Classifier {
	return &Classifier{
		rules: defaultRules(),
	}
}

// Classify examines parsed event fields and determines the semantic event type.
// It does NOT rely on the source format — Apache does not automatically mean HTTP.
func (c *Classifier) Classify(pe parsing.ParsedEvent) ClassificationResult {
	var matches []string

	for _, r := range c.rules {
		if r.Match(pe) {
			matches = append(matches, r.EventType())
		}
	}

	switch len(matches) {
	case 0:
		return ClassificationResult{
			Status:    ClassificationUnknown,
			EventType: "",
			Reason:    "No classification rule matched the parsed event fields",
		}
	case 1:
		return ClassificationResult{
			Status:    ClassificationConfident,
			EventType: matches[0],
			Reason:    "Single classification rule matched",
		}
	default:
		return ClassificationResult{
			Status:     ClassificationAmbiguous,
			EventType:  "",
			Candidates: matches,
			Reason:     "Multiple classification rules matched",
		}
	}
}

// Rule is a deterministic evidence-based classification rule.
type Rule interface {
	EventType() string
	Match(pe parsing.ParsedEvent) bool
}

func defaultRules() []Rule {
	return []Rule{
		&httpRule{},
		&authRule{},
	}
}

// httpRule matches events with HTTP evidence: method + path/URI + HTTP version.
type httpRule struct{}

func (r *httpRule) EventType() string { return "HTTP_ACTIVITY" }

func (r *httpRule) Match(pe parsing.ParsedEvent) bool {
	hasMethod := false
	hasPath := false
	hasProtocol := false

	for k, v := range pe.Fields {
		kl := strings.ToLower(k)
		if kl == "method" || kl == "request_method" {
			if s, ok := v.(string); ok && s != "" {
				hasMethod = true
			}
		}
		if kl == "path" || kl == "uri" || kl == "request_uri" || kl == "url" || kl == "request_target" {
			if s, ok := v.(string); ok && s != "" {
				hasPath = true
			}
		}
		if kl == "protocol" || kl == "http_version" {
			if s, ok := v.(string); ok && strings.HasPrefix(s, "HTTP/") {
				hasProtocol = true
			}
		}
	}

	return hasMethod && hasPath && hasProtocol
}

// authRule matches events with authentication evidence:
// action (login/auth/logon) + user/account identity + result/status.
type authRule struct{}

func (r *authRule) EventType() string { return "AUTHENTICATION" }

func (r *authRule) Match(pe parsing.ParsedEvent) bool {
	hasAction := false
	hasIdentity := false
	hasResult := false

	authKeywords := []string{"login", "logon", "logout", "logoff", "auth", "authenticate", "authentication", "signin", "sign_in", "signout", "sign_out", "password"}
	resultKeywords := []string{"success", "failure", "failed", "denied", "granted", "accepted", "rejected", "error", "ok"}

	for k, v := range pe.Fields {
		kl := strings.ToLower(k)

		// Check for auth action
		if kl == "action" || kl == "event" || kl == "event_type" || kl == "type" || kl == "operation" {
			if s, ok := v.(string); ok {
				sl := strings.ToLower(s)
				for _, kw := range authKeywords {
					if strings.Contains(sl, kw) {
						hasAction = true
						break
					}
				}
			}
		}

		// Check for user/account identity
		if kl == "user" || kl == "username" || kl == "user_name" || kl == "account" || kl == "email" || kl == "uid" || kl == "user_id" {
			if v != nil && v != "" {
				hasIdentity = true
			}
		}

		// Check for result/status
		if kl == "result" || kl == "status" || kl == "outcome" || kl == "auth_result" {
			if s, ok := v.(string); ok {
				sl := strings.ToLower(s)
				for _, kw := range resultKeywords {
					if strings.Contains(sl, kw) {
						hasResult = true
						break
					}
				}
			}
		}

		// Also check message field in syslog for authentication keywords
		if kl == "message" {
			if s, ok := v.(string); ok {
				sl := strings.ToLower(s)
				for _, kw := range authKeywords {
					if strings.Contains(sl, kw) {
						hasAction = true
						break
					}
				}
				for _, kw := range resultKeywords {
					if strings.Contains(sl, kw) {
						hasResult = true
						break
					}
				}
				// Check for user references in message (e.g. "for bob", "for user bob", "user=alice")
				if strings.Contains(sl, "user") || strings.Contains(sl, "account") ||
					strings.Contains(sl, " for ") {
					hasIdentity = true
				}
			}
		}
	}

	return hasAction && hasIdentity && hasResult
}
