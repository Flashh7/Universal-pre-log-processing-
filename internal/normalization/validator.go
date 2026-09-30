package normalization

import "fmt"

// ValidationResult describes whether an OCSF event passes schema validation.
type ValidationResult struct {
	Valid  bool
	Errors []string
}

// Validate checks a NormalizedEvent against OCSF 1.7.0 requirements.
func Validate(ev NormalizedEvent) ValidationResult {
	var errs []string

	// Required fields
	if ev.Metadata.Version == "" {
		errs = append(errs, "metadata.version is required")
	}
	if ev.EventUID == "" {
		errs = append(errs, "event_uid is required")
	}
	if ev.Time == 0 {
		errs = append(errs, "time is required")
	}
	if ev.ClassName == "" {
		errs = append(errs, "class_name is required")
	}
	if ev.CategoryName == "" {
		errs = append(errs, "category_name is required")
	}
	if ev.ActivityName == "" {
		errs = append(errs, "activity_name is required")
	}
	if ev.TypeName == "" {
		errs = append(errs, "type_name is required")
	}
	if ev.Severity == "" {
		errs = append(errs, "severity is required")
	}
	if ev.Status == "" {
		errs = append(errs, "status is required")
	}
	if ev.RawData == "" {
		errs = append(errs, "raw_data is required")
	}

	// Validate type_uid derivation: type_uid = class_uid * 100 + activity_id
	expectedTypeUID := ev.ClassUID*100 + ev.ActivityID
	if ev.TypeUID != expectedTypeUID {
		errs = append(errs, fmt.Sprintf("type_uid %d does not match expected %d (class_uid * 100 + activity_id)", ev.TypeUID, expectedTypeUID))
	}

	// Validate severity_id is in OCSF range: 0=Unknown, 1=Informational, 2=Low, 3=Medium, 4=High, 5=Critical, 99=Other
	validSeverity := map[int]bool{0: true, 1: true, 2: true, 3: true, 4: true, 5: true, 99: true}
	if !validSeverity[ev.SeverityID] {
		errs = append(errs, fmt.Sprintf("severity_id %d is not a valid OCSF 1.7.0 value", ev.SeverityID))
	}

	// Validate status_id is in OCSF range: 0=Unknown, 1=Success, 2=Failure, 99=Other
	validStatus := map[int]bool{0: true, 1: true, 2: true, 99: true}
	if !validStatus[ev.StatusID] {
		errs = append(errs, fmt.Sprintf("status_id %d is not a valid OCSF 1.7.0 value", ev.StatusID))
	}

	// Validate known class_uid values
	validClassUIDs := map[int]bool{0: true, 3002: true, 4002: true}
	if !validClassUIDs[ev.ClassUID] {
		errs = append(errs, fmt.Sprintf("class_uid %d is not a recognized OCSF class", ev.ClassUID))
	}

	// Validate category_uid matches class
	switch ev.ClassUID {
	case 0: // Base Event
		if ev.CategoryUID != 0 {
			errs = append(errs, fmt.Sprintf("category_uid %d invalid for Base Event (expected 0)", ev.CategoryUID))
		}
	case 3002: // Authentication
		if ev.CategoryUID != 3 {
			errs = append(errs, fmt.Sprintf("category_uid %d invalid for Authentication (expected 3)", ev.CategoryUID))
		}
	case 4002: // HTTP Activity
		if ev.CategoryUID != 4 {
			errs = append(errs, fmt.Sprintf("category_uid %d invalid for HTTP Activity (expected 4)", ev.CategoryUID))
		}
	}

	// Validate HTTP Activity activity_id: 0-9, 99
	if ev.ClassUID == 4002 {
		validHTTPActivity := map[int]bool{0: true, 1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 8: true, 9: true, 99: true}
		if !validHTTPActivity[ev.ActivityID] {
			errs = append(errs, fmt.Sprintf("activity_id %d is not valid for HTTP Activity", ev.ActivityID))
		}
	}

	// Validate Authentication activity_id: 0-4, 99
	if ev.ClassUID == 3002 {
		validAuthActivity := map[int]bool{0: true, 1: true, 2: true, 3: true, 4: true, 99: true}
		if !validAuthActivity[ev.ActivityID] {
			errs = append(errs, fmt.Sprintf("activity_id %d is not valid for Authentication", ev.ActivityID))
		}
	}

	return ValidationResult{
		Valid:  len(errs) == 0,
		Errors: errs,
	}
}
