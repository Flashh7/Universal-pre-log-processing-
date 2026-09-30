# ULPF Frontend

Frontend for the **Universal Log Pre-processing Framework (ULPF)** — a platform designed to ingest, normalize, standardize, and explore heterogeneous logs and events from different hardware and software systems.

The frontend provides a clean interface for exploring normalized events, inspecting individual events, monitoring connected sources, and viewing system-wide analytics.

## Tech Stack

* React
* TypeScript
* Vite
* Recharts
* CSS
* PWA-oriented architecture

## Current Features

### Log Explorer

* Search logs
* Filter logs by:

  * Time
  * Device
  * Severity
  * Activity
  * Class
* Display normalized events in a reusable table
* Open individual events for detailed inspection

### Event Details

Designed to display:

* Normalized event information
* Original/raw event
* Source and vendor information
* Event identifiers
* Raw-to-normalized field mapping

### Sources / Devices

Designed to display:

* Connected log sources
* Vendor
* Source type
* Input format
* Ingestion status
* Event counts

### Overview / Analytics

Designed to display:

* Total events
* Events/sec
* Active sources
* Parsing errors
* Event volume over time
* Events by source
* Events by severity

## Project Structure

```text
src/
├── components/
│   ├── Card.tsx
│   ├── HorizontalBarGraph.tsx
│   ├── LineGraph.tsx
│   ├── SearchFilterBar.tsx
│   └── Table.tsx
│
├── pages/
│   ├── Overview.tsx
│   ├── LogExplorer.tsx
│   ├── EventDetails.tsx
│   └── Sources.tsx
│
├── services/
│   └── api.ts
│
├── types/
│   ├── log.ts
│   ├── source.ts
│   └── overview.ts
│
├── data/
│   └── mockLogs.ts
│
├── App.tsx
├── main.tsx
└── index.css
```

## Architecture

The frontend is being designed to remain independent of the backend implementation.

The intended data flow is:

```text
Backend
   ↓
API / Service Layer
   ↓
Frontend Data Models
   ↓
React Pages
   ↓
Reusable Components
```

The frontend should not depend directly on backend implementation details such as:

* Programming language
* Database
* Message queue
* Storage engine
* Internal processing architecture

Backend communication is isolated inside the `services/` layer.

For example:

```text
LogExplorer
    ↓
searchLogs()
    ↓
services/api.ts
    ↓
Backend
```

During development, `api.ts` can use mock data. When the real backend becomes available, the service implementation can be replaced without requiring changes to the UI components.

## TypeScript Data Contracts

Frontend data structures are defined separately from UI components.

Important models currently include:

* `LogEvent`
* `SearchFilters`
* `Source`
* `Overview`

These types act as the stable contracts between the service layer and the frontend.

## Design Principles

### Reusable Components

Common UI elements are implemented as reusable components rather than being recreated inside individual pages.

Examples:

* Cards
* Tables
* Search/filter bars
* Horizontal bar graphs
* Line graphs

### Backend Independence

Pages and components should consume frontend models rather than directly depending on backend responses.

Backend-specific transformations should happen inside the service/API layer.

### Minimal UI

The frontend intentionally follows a minimal interface focused on:

* Readability
* Information density
* Fast log exploration
* Clear data hierarchy
* Reusable components

### Consistent Visual System

Current design tokens:

```css
--background: #F7F6F2;
--surface: #FFFFFF;
--accent: #B5A642;
--text: #222222;
--text-secondary: #6B6B66;
--border: #E4E1D8;
```

## Development

Install dependencies:

```bash
npm install
```

Start the development server:

```bash
npm run dev
```

The Vite development server will provide the local development URL.

## Current Development Status

The frontend is currently being developed with mock data.

The next architectural step is to complete the service/API boundary so that mock data can later be replaced by the actual ULPF backend without requiring changes to the React UI.

## ULPF Pipeline Context

The frontend is one part of the larger ULPF architecture:

```text
Collectors
    ↓
Message Queue
    ↓
Ingestion Workers
    ↓
Parsing / Normalization
    ↓
Unified Event Schema
    ↓
Storage / Search
    ↓
ULPF Frontend
```

The frontend is primarily responsible for **exploration, inspection, monitoring, and visualization** of the normalized event data.
