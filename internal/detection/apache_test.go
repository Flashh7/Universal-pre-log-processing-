package detection

import "testing"

func TestApacheDetector(t *testing.T) {
	detector := &ApacheDetector{}

	tests := []struct {
		name     string
		input    string
		expected DetectionStatus
	}{
		{"Valid Apache", `127.0.0.1 - - [28/Sep/2026:23:50:01 +0530] "GET /login HTTP/1.1" 200 1234`, StatusMatch},
		{"Not Apache", "<34>Oct 11 22:14:15 mymachine su: 'su root' failed", StatusNoMatch},
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
