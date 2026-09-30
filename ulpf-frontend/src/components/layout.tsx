import { useState } from "react";

import Sidebar from "./sidebar";
import Header from "./header";

import Overview from "../pages/overview";
import LogExplorer from "../pages/log_explorer";
import Sources from "../pages/sources";
import EventDetails from "../pages/event_details";
import AddSource from "../pages/add_source";

import type { LogEvent } from "../types/log";

export type Page = | "logs" | "overview" | "sources" | "event-details" | "add-source";

function Layout() {
  const [activePage, setActivePage] = useState<Page>("logs");

  const [selectedEvent, setSelectedEvent] = useState<LogEvent | null>(null);

  const handleEventSelect = (event: LogEvent) => {setSelectedEvent(event);setActivePage("event-details")};

  return (
    <div className="layout">
      <Sidebar
        activePage={activePage}
        setActivePage={setActivePage}
      />
      <div className="main">
        <Header />
        <main className="content">
          {activePage === "overview" && (<Overview />)}

          {activePage === "logs" && (<LogExplorer onEventSelect={handleEventSelect}/>)}

          {activePage === "sources" && (<Sources setActivePage={setActivePage} />)}

          {activePage === "add-source" && ( <AddSource setActivePage={setActivePage} /> )}

          {activePage === "event-details" && selectedEvent && (<EventDetails event={selectedEvent}/>)}
        </main>
      </div>
    </div>
  );
}

export default Layout;