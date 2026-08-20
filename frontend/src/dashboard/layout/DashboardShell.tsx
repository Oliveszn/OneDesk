import { useState } from "react";
import { Outlet } from "react-router-dom";
import { Menu01Icon } from "hugeicons-react";
import Sidebar from "./Sidebar";

export default function DashboardShell() {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <div className="flex min-h-screen" style={{ backgroundColor: "#f6f5f1" }}>
      <Sidebar open={sidebarOpen} onClose={() => setSidebarOpen(false)} />

      <div className="flex-1 min-w-0">
        <div
          className="md:hidden sticky top-0 z-30 flex items-center gap-3 px-4 h-14 border-b"
          style={{ backgroundColor: "#f6f5f1", borderColor: "#e1ded6" }}
        >
          <button
            onClick={() => setSidebarOpen(true)}
            aria-label="Open menu"
            className="p-2 -ml-2"
          >
            <Menu01Icon size={22} color="#16191c" />
          </button>
          <span
            className="font-semibold text-[15px] tracking-tight"
            style={{ color: "#16191c" }}
          >
            LedgerOS
          </span>
        </div>

        <Outlet />
      </div>
    </div>
  );
}
