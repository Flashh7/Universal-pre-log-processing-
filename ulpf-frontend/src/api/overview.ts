import type {
  OverviewMetrics,
  OverviewVolumeItem,
  OverviewCategoryItem,
  OverviewSeverityItem,
} from "../types/overview";

const API_BASE_URL = "http://localhost:8000";

export async function getOverviewMetrics(): Promise<OverviewMetrics> {
  const response = await fetch(
    `${API_BASE_URL}/api/overview/metrics`
  );

  if (!response.ok) {
    throw new Error("Failed to fetch overview metrics");
  }

  return response.json();
}

export async function getOverviewVolume(): Promise<OverviewVolumeItem[]> {
  const response = await fetch(
    `${API_BASE_URL}/api/overview/volume`
  );

  if (!response.ok) {
    throw new Error("Failed to fetch overview volume");
  }

  return response.json();
}

export async function getOverviewCategories(): Promise<OverviewCategoryItem[]> {
  const response = await fetch(
    `${API_BASE_URL}/api/overview/categories`
  );

  if (!response.ok) {
    throw new Error("Failed to fetch overview categories");
  }

  return response.json();
}

export async function getOverviewSeverity(): Promise<OverviewSeverityItem[]> {
  const response = await fetch(
    `${API_BASE_URL}/api/overview/severity`
  );

  if (!response.ok) {
    throw new Error("Failed to fetch overview severity");
  }

  return response.json();
}