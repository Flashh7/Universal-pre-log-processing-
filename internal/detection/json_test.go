package detection

import "testing"

func TestJSONDetector(t *testing.T) {
	detector := &JSONDetector{}

	tests := []struct {
		name     string
		input    string
		expected DetectionStatus
	}{
		{"Valid JSON Object", `{"message": "error", "code": 500}`, StatusMatch},
		{"Valid JSON Array", `[1, 2, 3]`, StatusMatch},
		{"Invalid JSON", `{message: "missing quotes"}`, StatusNoMatch},
		{"Plain Text", "2026-09-28 ERROR database failed", StatusNoMatch},
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
