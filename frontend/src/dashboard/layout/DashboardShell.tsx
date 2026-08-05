import { Outlet } from "react-router-dom";
import Sidebar from "./Sidebar";

export default function DashboardShell() {
  return (
    <div className="flex min-h-screen" style={{ backgroundColor: "#f6f5f1" }}>
      <Sidebar />
      <div className="flex-1 min-w-0">
        <Outlet />
      </div>
    </div>
  );
}
