package detection

import "regexp"

type NginxDetector struct{}

// Nginx combined access log requires trailing "referer" "user-agent" fields,
// distinguishing it from Apache Common format which lacks those fields.
var nginxPattern = regexp.MustCompile(`^\S+ - \S+ \[[^\]]+\] "[A-Z]+ .*? HTTP/\d\.\d" \d{3} \d+ "[^"]*" "[^"]*"`)

func (d *NginxDetector) Name() string {
	return "NGINX"
}

func (d *NginxDetector) Detect(raw string) DetectionResult {
	if nginxPattern.MatchString(raw) {
		return DetectionResult{
			Format: "NGINX",
			Status: StatusMatch,
			Reason: "Matches Nginx access log signature",
		}
	}

	return DetectionResult{
		Format: "NGINX",
		Status: StatusNoMatch,
		Reason: "Does not match Nginx access log signature",
	}
}
