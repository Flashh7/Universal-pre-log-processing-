package models

type LogEvent struct {
	EventUID     string `json:"event_uid"`
	Time         int64  `json:"time"`
	CategoryName string `json:"category_name"`
	CategoryUID  int32  `json:"category_uid"`
	ClassName    string `json:"class_name"`
	ClassUID     int32  `json:"class_uid"`
	ActivityName string `json:"activity_name"`
	ActivityID   int32  `json:"activity_id"`
	TypeName     string `json:"type_name"`
	TypeUID      int32  `json:"type_uid"`
	Severity     string `json:"severity"`
	SeverityID   int32  `json:"severity_id"`
	Status       string `json:"status"`
	StatusID     int32  `json:"status_id"`
	RawEvent     string `json:"raw_event"`
}