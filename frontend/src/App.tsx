import { BrowserRouter, Routes, Route } from "react-router-dom";
import LandingPage from "./pages/LandingPage";
import DashboardShell from "./dashboard/layout/DashboardShell";
import OverviewPage from "./dashboard/modules/overview/OverviewPage";
import InventoryPage from "./dashboard/modules/inventory/InventoryPage";
import ComingSoon from "./dashboard/shared/ComingSoon";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<LandingPage />} />

        <Route path="/dashboard" element={<DashboardShell />}>
          <Route index element={<OverviewPage />} />
          <Route path="inventory" element={<InventoryPage />} />
          <Route path="sales" element={<ComingSoon moduleName="Sales" />} />
          <Route path="finance" element={<ComingSoon moduleName="Finance" />} />
          <Route
            path="procurement"
            element={<ComingSoon moduleName="Procurement" />}
          />
          <Route path="billing" element={<ComingSoon moduleName="Billing" />} />
          <Route
            path="settings"
            element={<ComingSoon moduleName="Settings" />}
          />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
