import type { LogEvent, SearchFilters } from "../types/log";

const API_BASE_URL = "http://localhost:8000";

export async function getLogs(
  filters: SearchFilters
): Promise<LogEvent[]> {
  const params = new URLSearchParams();

  if (filters.search) {
    params.set("search", filters.search);
  }

  if (filters.severity) {
    params.set("severity", filters.severity);
  }

  if (filters.activity) {
    params.set("activity", filters.activity);
  }

  if (filters.className) {
    params.set("class_name", filters.className);
  }

  if (filters.status) {
    params.set("status", filters.status);
  }

  const queryString = params.toString();

  const url = queryString
    ? `${API_BASE_URL}/api/logs?${queryString}`
    : `${API_BASE_URL}/api/logs`;

  const response = await fetch(url);

  if (!response.ok) {
    throw new Error(`Failed to fetch logs: ${response.status}`);
  }

  return response.json();
}

export async function getLogById(
  eventUID: string
): Promise<LogEvent> {
  const response = await fetch(
    `${API_BASE_URL}/api/logs/${encodeURIComponent(eventUID)}`
  );

  if (!response.ok) {
    if (response.status === 404) {
      throw new Error("Event not found");
    }

    throw new Error(
      `Failed to fetch event: ${response.status}`
    );
  }

  return response.json();
}