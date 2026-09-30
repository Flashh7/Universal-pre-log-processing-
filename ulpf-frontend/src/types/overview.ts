export interface OverviewMetrics {
  total_events: number;
  events_per_second: number;
  active_categories: number;
  failed_events: number;
}

export interface OverviewVolumeItem {
  timestamp: string;
  event_count: number;
}

export interface OverviewCategoryItem {
  category_name: string;
  event_count: number;
}

export interface OverviewSeverityItem {
  severity: string;
  event_count: number;
}