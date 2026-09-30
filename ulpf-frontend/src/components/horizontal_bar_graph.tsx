interface BarItem {
  label: string;
  value: number;
}

interface HorizontalBarGraphProps {
  title?: string;
  data: BarItem[];
}

function HorizontalBarGraph({
  title,
  data,
}: HorizontalBarGraphProps) {
  const maxValue = Math.max(...data.map((item) => item.value), 0);

  return (
    <div className="horizontal-bar-graph">
      {title && (
        <h3 className="horizontal-bar-title">
          {title}
        </h3>
      )}

      <div className="horizontal-bar-list">
        {data.map((item) => {
          const width =
            maxValue === 0
              ? 0
              : (item.value / maxValue) * 100;

          return (
            <div
              className="horizontal-bar-item"
              key={item.label}
            >
              <div className="horizontal-bar-header">
                <span>{item.label}</span>
                <span>{item.value.toLocaleString()}</span>
              </div>

              <div className="horizontal-bar-track">
                <div
                  className="horizontal-bar-fill"
                  style={{ width: `${width}%` }}
                />
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

export default HorizontalBarGraph;
