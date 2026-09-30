import Card from "../components/card";
import type { Source } from "../types/source";
import { useEffect, useState } from "react";
import { getSources } from "../api/sources";


function Sources() {  
  const [sources, setSources] = useState<Source[]>([]);

  useEffect(() => {
    const loadSources = async () => {
      const data = await getSources();
      setSources(data);
    };

    loadSources();
  }, []);

  return (
    <div className="sources-page">

      <div className="sources-actions">
        <button className="add-source-button">
          + Add Source
        </button>
      </div>

      <div className="sources-grid">
        {sources.map((source) => (
          <Card
            key={source.source_uid}
            title={source.source_name}
            subtitle={`${source.vendor_name} · ${source.source_type}`}
            className="source-card"
          >
            <div className="source-details">

              <div className="source-detail">
                <span className="source-label">
                  Source ID
                </span>
                <span>
                  {source.source_uid}
                </span>
              </div>

              <div className="source-detail">
                <span className="source-label">
                  Format
                </span>
                <span>
                  {source.input_format}
                </span>
              </div>

              <div className="source-detail">
                <span className="source-label">
                  Events
                </span>
                <span>
                  {source.event_count.toLocaleString()}
                </span>
              </div>

              <div className="source-detail">
                <span className="source-label">
                  Status
                </span>
                <span
                  className={`source-status ${source.status.toLowerCase()}`}
                >
                  {source.status}
                </span>
              </div>

            </div>
          </Card>
        ))}
      </div>

    </div>
  );
}

export default Sources;