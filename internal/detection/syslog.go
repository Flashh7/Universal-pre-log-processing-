package detection

import "regexp"

type SyslogDetector struct{}

// A simplistic syslog pattern (RFC 3164 or RFC 5424 basics)
// e.g. <34>Oct 11 22:14:15 mymachine su: 'su root' failed for lonvick on /dev/pts/8
var syslogPattern = regexp.MustCompile(`^<\d+>.*`)

func (d *SyslogDetector) Name() string {
	return "SYSLOG"
}

func (d *SyslogDetector) Detect(raw string) DetectionResult {
	if syslogPattern.MatchString(raw) {
		return DetectionResult{
			Format: "SYSLOG",
			Status: StatusMatch,
			Reason: "Starts with syslog priority <PRI>",
		}
	}
	
	return DetectionResult{
		Format: "SYSLOG",
		Status: StatusNoMatch,
		Reason: "Does not match syslog structure",
	}
}
