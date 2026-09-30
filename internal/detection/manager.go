package detection

type Manager struct {
	detectors []Detector
}

func NewManager() *Manager {
	return &Manager{
		detectors: []Detector{
			&JSONDetector{},
			&SyslogDetector{},
			&ApacheDetector{},
			&NginxDetector{},
		},
	}
}

type ManagerResult struct {
	Outcome string
	Format  string
}

func (m *Manager) Identify(raw string) ManagerResult {
	var matches []DetectionResult

	for _, d := range m.detectors {
		res := d.Detect(raw)
		if res.Status == StatusMatch {
			matches = append(matches, res)
		}
	}

	if len(matches) == 0 {
		return ManagerResult{
			Outcome: OutcomeUnknown,
			Format:  "",
		}
	}

	if len(matches) == 1 {
		return ManagerResult{
			Outcome: OutcomeKnownFormat,
			Format:  matches[0].Format,
		}
	}

	return ManagerResult{
		Outcome: OutcomeAmbiguous,
		Format:  "",
	}
}
