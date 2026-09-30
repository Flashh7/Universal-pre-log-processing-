import Card from "../components/card";
import type { Source } from "../types/source";
import { useEffect, useState } from "react";
import { getSources } from "../api/sources";
import type { Page } from "../components/layout";

interface SourcesProps {
  setActivePage: (page: Page) => void;
}

function Sources({ setActivePage }: SourcesProps) {
  const [sources, setSources] = useState<Source[]>([]);

  useEffect(() => {
    const loadSources = async () => {
      try {
        const data = await getSources();
        setSources(data);
      } catch (error) {
        console.error("Failed to load sources:", error);
      }
    };

    loadSources();
  }, []);

  return (
    <div className="sources-page">

      <div className="sources-actions">
        <button
          className="add-source-button"
          onClick={() => setActivePage("add-source")}
        >
          + Add Source
        </button>
      </div>

      <div className="sources-grid">
        {sources.map((source) => (
          <Card
            key={source.source_uid}
            title={source.name}
            subtitle={`${source.vendor_name} · ${source.type}`}
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
                  Product
                </span>

                <span>
                  {source.product_name}
                </span>
              </div>

              <div className="source-detail">
                <span className="source-label">
                  Format
                </span>

                <span>
                  {source.original_format}
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