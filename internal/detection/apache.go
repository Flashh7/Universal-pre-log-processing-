package detection

import "regexp"

type ApacheDetector struct{}

// Apache common/combined access log pattern signature
var apachePattern = regexp.MustCompile(`^\S+ \S+ \S+ \[[^\]]+\] "[A-Z]+ .*? HTTP/\d\.\d" \d{3} \d+`)

func (d *ApacheDetector) Name() string {
	return "APACHE"
}

func (d *ApacheDetector) Detect(raw string) DetectionResult {
	if apachePattern.MatchString(raw) {
		return DetectionResult{
			Format: "APACHE",
			Status: StatusMatch,
			Reason: "Matches Apache access log signature (IP - - [date] \"method path protocol\" status bytes)",
		}
	}

	return DetectionResult{
		Format: "APACHE",
		Status: StatusNoMatch,
		Reason: "Does not match Apache access log signature",
	}
}
