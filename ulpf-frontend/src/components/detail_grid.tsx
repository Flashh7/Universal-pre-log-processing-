import React from "react";

interface DetailItem {
  label: string;
  value: React.ReactNode;
}

interface DetailGridProps {
  items: DetailItem[];
}

function DetailGrid({ items }: DetailGridProps) {
  return (
    <div className="detail-grid">
      {items.map((item) => (
        <div className="detail-item" key={item.label}>
          <span className="detail-label">
            {item.label}
          </span>

          <span>
            {item.value}
          </span>
        </div>
      ))}
    </div>
  );
}

export default DetailGrid;

