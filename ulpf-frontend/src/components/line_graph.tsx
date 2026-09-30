import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";

interface LineItem {
  label: string;
  value: number;
}

interface LineGraphProps {
  title?: string;
  data: LineItem[];
}

function formatTime(label: unknown): string {
  if (typeof label !== "string") {
    return "";
  }

  const date = new Date(label.replace(" ", "T"));

  if (Number.isNaN(date.getTime())) {
    return label;
  }

  return date.toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  });
}

function LineGraph({
  title,
  data,
}: LineGraphProps) {
  return (
    <div className="line-graph">
      {title && (
        <h3 className="line-graph-title">
          {title}
        </h3>
      )}

      <div className="line-graph-chart">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={data}>

            <CartesianGrid
              className="line-graph-grid"
              vertical={false}
            />

            <XAxis
              dataKey="label"
              tickFormatter={formatTime}
              tick={{
                fill: "var(--text-secondary)",
                fontSize: 12,
              }}
            />

            <YAxis
              allowDecimals={false}
              tick={{
                fill: "var(--text-secondary)",
                fontSize: 12,
              }}
            />

            <Tooltip
              labelFormatter={formatTime}
            />

            <Line
              type="monotone"
              dataKey="value"
              stroke="var(--accent)"
              strokeWidth={2}
              dot={{ r: 3 }}
              activeDot={{ r: 5 }}
            />

          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

export default LineGraph;