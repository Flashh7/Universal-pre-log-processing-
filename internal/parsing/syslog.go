package parsing

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// SyslogParser supports RFC 3164 BSD-style syslog messages.
//
// Format: <PRI>timestamp hostname app[pid]: message
// or:     <PRI>timestamp hostname app: message
type SyslogParser struct{}

// RFC 3164 pattern with optional PID
var syslogPattern = regexp.MustCompile(
	`^<(?P<priority>\d+)>(?P<timestamp>\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+(?P<hostname>\S+)\s+(?P<app_name>\S+?)(?:\[(?P<pid>\d+)\])?:\s+(?P<message>.+)$`,
)

func (p *SyslogParser) Name() string {
	return "SYSLOG"
}

func (p *SyslogParser) Parse(raw string) (ParsedEvent, error) {
	matches := syslogPattern.FindStringSubmatch(raw)
	if matches == nil {
		return ParsedEvent{}, fmt.Errorf("UNSUPPORTED_LAYOUT: raw does not match supported RFC 3164 syslog layout")
	}

	fields := make(map[string]any)
	names := syslogPattern.SubexpNames()
	for i, name := range names {
		if i == 0 || name == "" {
			continue
		}
		if matches[i] != "" {
			fields[name] = matches[i]
		}
	}

	// Convert priority to int
	if pri, ok := fields["priority"]; ok {
		if v, err := strconv.Atoi(pri.(string)); err == nil {
			fields["priority"] = v
		}
	}

	// Convert PID to int if present
	if pid, ok := fields["pid"]; ok {
		if v, err := strconv.Atoi(pid.(string)); err == nil {
			fields["pid"] = v
		}
	}

	// Parse timestamp (RFC 3164 uses "Jan  2 15:04:05" — no year)
	var ts *time.Time
	if tsStr, ok := fields["timestamp"]; ok {
		// Prepend current year since RFC 3164 omits it
		tsWithYear := fmt.Sprintf("%d %s", time.Now().Year(), tsStr.(string))
		if parsed, err := time.Parse("2006 Jan  2 15:04:05", tsWithYear); err == nil {
			ts = &parsed
		} else if parsed, err := time.Parse("2006 Jan 2 15:04:05", tsWithYear); err == nil {
			ts = &parsed
		}
	}

	return ParsedEvent{
		Timestamp: ts,
		EventType: "", // Classification is separate
		Fields:    fields,
	}, nil
}
