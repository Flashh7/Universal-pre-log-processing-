import type { Source } from "../types/source";

const API_BASE_URL = "http://localhost:8000";

export async function getSources(): Promise<Source[]> {
  const response = await fetch(`${API_BASE_URL}/api/sources`);

  if (!response.ok) {
    throw new Error("Failed to fetch sources");
  }

  return response.json();
}

export async function createSource(
  source: Omit<Source, "source_uid">
): Promise<Source> {
  const response = await fetch(`${API_BASE_URL}/api/sources`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(source),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(errorText || "Failed to create source");
  }

  return response.json();
}