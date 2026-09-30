export interface LogEvent {
  event_uid: string;
  time: number;

  category_name: string;
  category_uid: number;

  class_name: string;
  class_uid: number;

  activity_name: string;
  activity_id: number;

  type_name: string;
  type_uid: number;

  severity:
    | "Informational"
    | "Low"
    | "Medium"
    | "High"
    | "Critical"
    | "Fatal";

  severity_id: number;

  status: string;
  status_id: number;

  raw_event: string;
}

export interface SearchFilters {
  search: string;
  time: string;
  severity: string;
  activity: string;
  className: string;
  status: string;
}