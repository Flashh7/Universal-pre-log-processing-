package parsing

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// NginxParser supports the Nginx default combined access log format.
//
// Default: $remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent"
type NginxParser struct{}

var nginxCombinedPattern = regexp.MustCompile(
	`^(?P<remote_addr>\S+) - (?P<remote_user>\S+) \[(?P<timestamp>[^\]]+)\] "(?P<method>[A-Z]+) (?P<uri>\S+) (?P<protocol>HTTP/\d\.\d)" (?P<status>\d{3}) (?P<bytes_sent>\d+|-)(?:\s+"(?P<referrer>[^"]*)" "(?P<user_agent>[^"]*)")?`,
)

func (p *NginxParser) Name() string {
	return "NGINX"
}

func (p *NginxParser) Parse(raw string) (ParsedEvent, error) {
	matches := nginxCombinedPattern.FindStringSubmatch(raw)
	if matches == nil {
		return ParsedEvent{}, fmt.Errorf("UNSUPPORTED_LAYOUT: raw does not match supported Nginx combined layout")
	}

	fields := make(map[string]any)
	names := nginxCombinedPattern.SubexpNames()
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

	// Convert bytes_sent to int
	if b, ok := fields["bytes_sent"]; ok {
		bs := b.(string)
		if bs == "-" {
			fields["bytes_sent"] = 0
		} else if v, err := strconv.Atoi(bs); err == nil {
			fields["bytes_sent"] = v
		}
	}

	// Parse timestamp
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
