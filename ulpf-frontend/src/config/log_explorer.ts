import type { TableColumn } from "../components/table";
import type { FilterConfig } from "../components/search_and_filter";
import type { LogEvent } from "../types/log";

export const logExplorerColumns: TableColumn<LogEvent>[] = [
  {
    key: "time",
    label: "Time",
  },
  {
    key: "event_uid",
    label: "Event ID",
  },
  {
    key: "category_name",
    label: "Category",
  },
  {
    key: "severity",
    label: "Severity",
  },
  {
    key: "activity_name",
    label: "Activity",
  },
  {
    key: "class_name",
    label: "Class",
  },
  {
    key: "status",
    label: "Status",
  },
  {
    key: "type_name",
    label: "Type",
  },
];

export const logExplorerFilters: FilterConfig[] = [
  {
    key: "time",
    label: "Time",
    options: [
      { label: "Last 15 minutes", value: "15m" },
      { label: "Last 1 hour", value: "1h" },
      { label: "Last 24 hours", value: "24h" },
    ],
  },

  {
    key: "severity",
    label: "Severity",
    options: [
      { label: "Informational", value: "Informational" },
      { label: "Low", value: "Low" },
      { label: "Medium", value: "Medium" },
      { label: "High", value: "High" },
      { label: "Critical", value: "Critical" },
      { label: "Fatal", value: "Fatal" },
    ],
  },

  {
    key: "activity",
    label: "Activity",
    options: [
      { label: "Logon", value: "Logon" },
      { label: "Launch", value: "Launch" },
      { label: "Terminate", value: "Terminate" },
      { label: "Create", value: "Create" },
      { label: "Delete", value: "Delete" },
      { label: "Open", value: "Open" },
      { label: "Close", value: "Close" },
      { label: "Update", value: "Update" },
    ],
  },

  {
    key: "className",
    label: "Class",
    options: [
      { label: "Authentication", value: "Authentication" },
      { label: "System Activity", value: "System Activity" },
      { label: "Process Activity", value: "Process Activity" },
      { label: "Network Activity", value: "Network Activity" },
      { label: "File Activity", value: "File Activity" },
    ],
  },

  {
    key: "status",
    label: "Status",
    options: [
      { label: "Success", value: "Success" },
      { label: "Failure", value: "Failure" },
    ],
  },
];