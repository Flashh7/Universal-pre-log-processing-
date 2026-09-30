import type { Source } from "../types/source";

export const mockSources: Source[] = [
  {
    source_uid: "SRC-001",
    source_name: "Firewall-01",
    vendor_name: "Palo Alto",
    source_type: "Firewall",
    input_format: "CEF",
    status: "Active",
    event_count: 482391,
  },
  {
    source_uid: "SRC-002",
    source_name: "Web-Server-01",
    vendor_name: "Apache",
    source_type: "Web Server",
    input_format: "JSON",
    status: "Active",
    event_count: 328742,
  },
  {
    source_uid: "SRC-003",
    source_name: "Router-02",
    vendor_name: "Cisco",
    source_type: "Router",
    input_format: "Syslog",
    status: "Error",
    event_count: 174928,
  },
  {
    source_uid: "SRC-004",
    source_name: "Database-01",
    vendor_name: "PostgreSQL",
    source_type: "Database",
    input_format: "JSON",
    status: "Inactive",
    event_count: 92341,
  },
];