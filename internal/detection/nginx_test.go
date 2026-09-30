package detection

import "testing"

func TestNginxDetector(t *testing.T) {
	detector := &NginxDetector{}

	tests := []struct {
		name     string
		input    string
		expected DetectionStatus
	}{
		{"Valid Nginx Combined", `192.168.1.1 - - [10/Oct/2026:13:55:36 -0700] "GET /index.html HTTP/1.1" 200 612 "-" "Mozilla/5.0"`, StatusMatch},
		{"Apache Common (no referer/UA)", `127.0.0.1 - - [28/Sep/2026:23:50:01 +0530] "GET /login HTTP/1.1" 200 1234`, StatusNoMatch},
		{"Not Nginx", "<34>Oct 11 22:14:15 mymachine su: 'su root' failed", StatusNoMatch},
		{"Empty Input", "", StatusNoMatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := detector.Detect(tt.input)
			if res.Status != tt.expected {
				t.Errorf("Expected %v, got %v for %q", tt.expected, res.Status, tt.input)
			}
		})
	}
}
