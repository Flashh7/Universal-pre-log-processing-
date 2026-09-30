import { useEffect, useState } from "react";
import { getLogs } from "../api/logs";
import Table from "../components/table";
import SearchFilterBar from "../components/search_and_filter";
import type { LogEvent, SearchFilters } from "../types/log";
import {
  logExplorerColumns,
  logExplorerFilters,
} from "../config/log_explorer";

interface LogExplorerProps {
  onEventSelect: (event: LogEvent) => void;
}

function LogExplorer({
  onEventSelect,
}: LogExplorerProps) {
  const [logs, setLogs] = useState<LogEvent[]>([]);

  const [searchFilters, setSearchFilters] =
    useState<SearchFilters>({
      search: "",
      time: "",
      severity: "",
      activity: "",
      className: "",
      status: "",
    });

  useEffect(() => {
    const loadLogs = async () => {
      try {
        const data = await getLogs(searchFilters);
        setLogs(data);
      } catch (error) {
        console.error("Failed to load logs:", error);
        setLogs([]);
      }
    };

    loadLogs();
  }, [searchFilters]);

  const handleFilterChange = (
    key: keyof Omit<SearchFilters, "search">,
    value: string
  ) => {
    setSearchFilters((previous) => ({
      ...previous,
      [key]: value,
    }));
  };

  const handleClear = () => {
    setSearchFilters({
      search: "",
      time: "",
      severity: "",
      activity: "",
      className: "",
      status: "",
    });
  };

  return (
    <div className="log-explorer">

      <SearchFilterBar
        search={searchFilters.search}
        onSearchChange={(value) =>
          setSearchFilters((previous) => ({
            ...previous,
            search: value,
          }))
        }
        filters={logExplorerFilters}
        selectedFilters={searchFilters}
        onFilterChange={handleFilterChange}
        onClear={handleClear}
      />

      <Table
        columns={logExplorerColumns}
        data={logs}
        onRowClick={onEventSelect}
      />

    </div>
  );
}

export default LogExplorer;