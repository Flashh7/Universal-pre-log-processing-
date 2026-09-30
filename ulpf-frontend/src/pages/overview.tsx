import { useEffect, useState } from "react";
import Card from "../components/card";
import LineGraph from "../components/line_graph";
import HorizontalBarGraph from "../components/horizontal_bar_graph";

import {
  getOverviewMetrics,
  getOverviewVolume,
  getOverviewCategories,
  getOverviewSeverity,
} from "../api/overview";

import type {
  OverviewMetrics,
  OverviewVolumeItem,
  OverviewCategoryItem,
  OverviewSeverityItem,
} from "../types/overview";

function Overview() {
  const [metrics, setMetrics] =
    useState<OverviewMetrics | null>(null);

  const [volume, setVolume] =
    useState<OverviewVolumeItem[]>([]);

  const [categories, setCategories] =
    useState<OverviewCategoryItem[]>([]);

  const [severity, setSeverity] =
    useState<OverviewSeverityItem[]>([]);

  useEffect(() => {
    const loadOverview = async () => {
      const [
        metricsData,
        volumeData,
        categoriesData,
        severityData,
      ] = await Promise.all([
        getOverviewMetrics(),
        getOverviewVolume(),
        getOverviewCategories(),
        getOverviewSeverity(),
      ]);

      setMetrics(metricsData);
      setVolume(volumeData);
      setCategories(categoriesData);
      setSeverity(severityData);
    };

    loadOverview();
  }, []);

  if (!metrics) {
    return null;
  }

  return (
    <div className="overview-page">

      <div className="overview-metrics">

        <Card title="Total Events">
          <div className="metric-value">
            {metrics.total_events.toLocaleString()}
          </div>
        </Card>

        <Card title="Events/sec">
          <div className="metric-value">
            {metrics.events_per_second.toLocaleString()}
          </div>
        </Card>

        <Card title="Active Categories">
          <div className="metric-value">
            {metrics.active_categories.toLocaleString()}
          </div>
        </Card>

        <Card title="Failed Events">
          <div className="metric-value">
            {metrics.failed_events.toLocaleString()}
          </div>
        </Card>

      </div>

      <div className="overview-section">

        <Card title="Event Volume Over Time">
          <LineGraph
            data={volume.map((item) => ({
              label: item.timestamp,
              value: item.event_count,
            }))}
          />
        </Card>

      </div>

      <div className="overview-charts">

        <Card title="Events by Category">
          <HorizontalBarGraph
            data={categories.map((item) => ({
              label: item.category_name,
              value: item.event_count,
            }))}
          />
        </Card>

        <Card title="Events by Severity">
          <HorizontalBarGraph
            data={severity.map((item) => ({
              label: item.severity,
              value: item.event_count,
            }))}
          />
        </Card>

      </div>

    </div>
  );
}

export default Overview;