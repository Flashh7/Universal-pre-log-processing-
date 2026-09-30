package parsing

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// ApacheParser supports Apache Common Log Format and Combined Log Format.
//
// Common:  %h %l %u %t "%r" %>s %b
// Combined: %h %l %u %t "%r" %>s %b "%{Referer}i" "%{User-Agent}i"
type ApacheParser struct{}

// Apache Combined Log Format regex with named groups.
// Also matches Common (referrer and user-agent will simply be absent).
var apacheCombinedPattern = regexp.MustCompile(
	`^(?P<client_ip>\S+) (?P<ident>\S+) (?P<user>\S+) \[(?P<timestamp>[^\]]+)\] "(?P<method>[A-Z]+) (?P<path>\S+) (?P<protocol>HTTP/\d\.\d)" (?P<status>\d{3}) (?P<bytes>\d+|-)(?:\s+"(?P<referrer>[^"]*)" "(?P<user_agent>[^"]*)")?`,
)

func (p *ApacheParser) Name() string {
	return "APACHE"
}

func (p *ApacheParser) Parse(raw string) (ParsedEvent, error) {
	matches := apacheCombinedPattern.FindStringSubmatch(raw)
	if matches == nil {
		return ParsedEvent{}, fmt.Errorf("UNSUPPORTED_LAYOUT: raw does not match supported Apache Common/Combined layout")
	}

	fields := make(map[string]any)
	names := apacheCombinedPattern.SubexpNames()
	for i, name := range names {
		if i == 0 || name == "" {
			continue
		}
		if matches[i] != "" {
			fields[name] = matches[i]
		}
	}

	// Convert status to int
	if s, ok := fields["status"]; ok {
		if v, err := strconv.Atoi(s.(string)); err == nil {
			fields["status"] = v
		}
	}

	// Convert bytes to int (handle "-" as 0)
	if b, ok := fields["bytes"]; ok {
		bs := b.(string)
		if bs == "-" {
			fields["bytes"] = 0
		} else if v, err := strconv.Atoi(bs); err == nil {
			fields["bytes"] = v
		}
	}

	// Parse timestamp: 28/Sep/2026:23:50:01 +0530
	var ts *time.Time
	if tsStr, ok := fields["timestamp"]; ok {
		if parsed, err := time.Parse("02/Jan/2006:15:04:05 -0700", tsStr.(string)); err == nil {
			ts = &parsed
		}
	}

	return ParsedEvent{
		Timestamp: ts,
		EventType: "", // Classification is separate
		Fields:    fields,
	}, nil
}
