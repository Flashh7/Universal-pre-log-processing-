import React from "react";
import type { SearchFilters } from "../types/log";

export interface FilterOption {
  label: string;
  value: string;
}

export interface FilterConfig {
  key: keyof Omit<SearchFilters, "search">;
  label: string;
  options: FilterOption[];
}

interface SearchFilterBarProps {
  search: string;
  onSearchChange: (value: string) => void;

  filters: FilterConfig[];
  selectedFilters: SearchFilters;
  onFilterChange: (
    key: keyof Omit<SearchFilters, "search">,
    value: string
  ) => void;

  onClear: () => void;
}

const SearchFilterBar: React.FC<SearchFilterBarProps> = ({
  search,
  onSearchChange,
  filters,
  selectedFilters,
  onFilterChange,
  onClear,
}) => {
  const hasFilters =
    search !== "" ||
    Object.entries(selectedFilters).some(
      ([key, value]) =>
        key !== "search" && value !== ""
    );

  return (
    <div className="search-filter-bar">

      <div className="search-box">
        <input
          type="text"
          placeholder="Search logs..."
          value={search}
          onChange={(e) =>
            onSearchChange(e.target.value)
          }
        />
      </div>

      <div className="filter-group">
        {filters.map((filter) => (
          <select
            key={filter.key}
            value={selectedFilters[filter.key]}
            onChange={(e) =>
              onFilterChange(
                filter.key,
                e.target.value
              )
            }
          >
            <option value="">
              {filter.label}
            </option>

            {filter.options.map((option) => (
              <option
                key={option.value}
                value={option.value}
              >
                {option.label}
              </option>
            ))}
          </select>
        ))}

        {hasFilters && (
          <button
            type="button"
            className="clear-filters"
            onClick={onClear}
          >
            Clear
          </button>
        )}
      </div>

    </div>
  );
};

export default SearchFilterBar;