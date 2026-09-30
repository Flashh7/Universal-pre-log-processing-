import { useState } from "react";
import type { Page } from "./layout";

type SidebarProps = {
  activePage: Page;
  setActivePage: (page: Page) => void;
};

function Sidebar({
  activePage,
  setActivePage,
}: SidebarProps) {
  const [collapsed, setCollapsed] = useState(false);

  return (
    <aside className={`sidebar ${collapsed ? "collapsed" : ""}`}>

      <div
        className="sidebar-menu"
        onClick={() => setCollapsed(!collapsed)}
      >
        <span className="menu-icon">☰</span>

        {!collapsed && (
          <span className="menu-label">Menu</span>
        )}
      </div>

      <nav className="sidebar-nav">
        
        <button
          className={`nav-item ${
            activePage === "logs" ? "active" : ""
          }`}
          onClick={() => setActivePage("logs")}
        >
          {!collapsed && "Log Explorer"}
        </button>

        <button
          className={`nav-item ${
            activePage === "sources" ? "active" : ""
          }`}
          onClick={() => setActivePage("sources")}
        >
          {!collapsed && "Sources"}
        </button>

        <button
          className={`nav-item ${
            activePage === "overview" ? "active" : ""
          }`}
          onClick={() => setActivePage("overview")}
        >
          {!collapsed && "Overview"}
        </button>       

      </nav>

    </aside>
  );
}

export default Sidebar;