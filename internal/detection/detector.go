package detection

type DetectionStatus string

const (
	StatusMatch   DetectionStatus = "MATCH"
	StatusNoMatch DetectionStatus = "NO_MATCH"
)

const (
	OutcomeKnownFormat = "KNOWN FORMAT"
	OutcomeUnknown     = "UNKNOWN"
	OutcomeAmbiguous   = "AMBIGUOUS"
)

type DetectionResult struct {
	Format string
	Status DetectionStatus
	Reason string
}

type Detector interface {
	Name() string
	Detect(raw string) DetectionResult
}
