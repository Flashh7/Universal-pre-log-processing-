package detection

import "testing"

func TestManager(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name           string
		input          string
		expectedResult string
		expectedFormat string
	}{
		{
			name:           "JSON Input",
			input:          `{"key":"value"}`,
			expectedResult: OutcomeKnownFormat,
			expectedFormat: "JSON",
		},
		{
			name:           "Syslog Input",
			input:          `<34>Oct 11 22:14:15 mymachine app: test log`,
			expectedResult: OutcomeKnownFormat,
			expectedFormat: "SYSLOG",
		},
		{
			name:           "Apache Common (no referer/UA) → KNOWN APACHE",
			input:          `127.0.0.1 - - [28/Sep/2026:23:50:01 +0530] "GET /login HTTP/1.1" 200 1234`,
			expectedResult: OutcomeKnownFormat,
			expectedFormat: "APACHE",
		},
		{
			name:           "Ambiguous Apache Combined / Nginx Combined",
			input:          `192.168.1.1 - - [10/Oct/2026:13:55:36 -0700] "GET /index.html HTTP/1.1" 200 612 "-" "Mozilla/5.0"`,
			expectedResult: OutcomeAmbiguous,
			expectedFormat: "",
		},
		{
			name:           "Unknown Format",
			input:          `2026-09-28 ERROR database failed`,
			expectedResult: OutcomeUnknown,
			expectedFormat: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := manager.Identify(tt.input)
			if res.Outcome != tt.expectedResult {
				t.Errorf("Expected outcome %s, got %s for %q", tt.expectedResult, res.Outcome, tt.input)
			}
			if res.Format != tt.expectedFormat {
				t.Errorf("Expected format %s, got %s for %q", tt.expectedFormat, res.Format, tt.input)
			}
		})
	}
}
