package detection

import "encoding/json"

type JSONDetector struct{}

func (d *JSONDetector) Name() string {
	return "JSON"
}

func (d *JSONDetector) Detect(raw string) DetectionResult {
	var js interface{}
	err := json.Unmarshal([]byte(raw), &js)
	if err == nil {
		return DetectionResult{
			Format: "JSON",
			Status: StatusMatch,
			Reason: "Valid JSON syntax parsed",
		}
	}
	
	return DetectionResult{
		Format: "JSON",
		Status: StatusNoMatch,
		Reason: "Invalid JSON syntax",
	}
}
