import { NavLink } from "react-router-dom";
import {
  DashboardSquare01Icon,
  Package01Icon,
  ShoppingCart01Icon,
  Invoice03Icon,
  TruckDeliveryIcon,
  Wallet02Icon,
  Settings02Icon,
  UserCircleIcon,
  Cancel01Icon,
} from "hugeicons-react";

const navItems = [
  {
    to: "/dashboard",
    label: "Overview",
    Icon: DashboardSquare01Icon,
    accent: "#e1ded6",
    enabled: true,
    end: true,
  },
  {
    to: "/dashboard/inventory",
    label: "Inventory",
    Icon: Package01Icon,
    accent: "#24344a",
    enabled: true,
  },
  {
    to: "/dashboard/sales",
    label: "Sales",
    Icon: ShoppingCart01Icon,
    accent: "#b8863a",
    enabled: true,
  },
  {
    to: "/dashboard/finance",
    label: "Finance",
    Icon: Invoice03Icon,
    accent: "#5b7a63",
    enabled: true,
  },
  {
    to: "/dashboard/procurement",
    label: "Procurement",
    Icon: TruckDeliveryIcon,
    accent: "#16191c",
    enabled: true,
  },
  {
    to: "/dashboard/billing",
    label: "Billing",
    Icon: Wallet02Icon,
    accent: "#e1ded6",
    enabled: true,
  },
];

export default function Sidebar({ open, onClose }) {
  return (
    <>
      {open && (
        <div
          className="fixed inset-0 z-40 md:hidden"
          style={{ backgroundColor: "rgba(22,25,28,0.5)" }}
          onClick={onClose}
          aria-hidden="true"
        />
      )}

      <aside
        className={`w-[240px] shrink-0 h-screen fixed md:sticky top-0 left-0 z-50 flex flex-col transition-transform duration-200 ease-in-out md:translate-x-0 ${
          open ? "translate-x-0" : "-translate-x-full"
        }`}
        style={{ backgroundColor: "#24344a" }}
      >
        <div
          className="px-6 h-[64px] flex items-center justify-between border-b shrink-0"
          style={{ borderColor: "rgba(246,245,241,0.12)" }}
        >
          <div className="flex items-center gap-2">
            <span
              className="w-6 h-6 flex items-center justify-center text-xs font-bold"
              style={{ backgroundColor: "#b8863a", color: "#16191c" }}
            >
              L
            </span>
            <span
              className="font-semibold text-[15px] tracking-tight"
              style={{ color: "#f6f5f1" }}
            >
              LedgerOS
            </span>
          </div>
          <button
            onClick={onClose}
            className="md:hidden p-1"
            aria-label="Close menu"
          >
            <Cancel01Icon size={20} color="#f6f5f1" />
          </button>
        </div>

        <nav className="flex-1 px-3 py-5 flex flex-col gap-1 overflow-y-auto">
          {navItems.map((item) => (
            <NavItem key={item.to} {...item} onNavigate={onClose} />
          ))}
        </nav>

        <div className="px-3 pb-3 shrink-0">
          <NavLink
            to="/dashboard/settings"
            className="flex items-center gap-3 px-3 py-2.5 text-[14px] font-medium rounded-sm opacity-60 cursor-not-allowed"
            style={{ color: "#f6f5f1" }}
            onClick={(e) => e.preventDefault()}
          >
            <Settings02Icon size={18} />
            Settings
          </NavLink>
        </div>

        <div
          className="px-3 pb-5 pt-3 border-t shrink-0"
          style={{ borderColor: "rgba(246,245,241,0.12)" }}
        >
          <div className="flex items-center gap-2.5 px-3 py-2">
            <UserCircleIcon size={28} color="#f6f5f1" />
            <div className="min-w-0">
              <div
                className="text-[13px] font-semibold truncate"
                style={{ color: "#f6f5f1" }}
              >
                Acme Co
              </div>
              <div
                className="text-[12px] truncate"
                style={{ color: "#9aa3ab" }}
              >
                Free plan
              </div>
            </div>
          </div>
        </div>
      </aside>
    </>
  );
}

type NavItemProps = {
  to: string;
  label: string;
  Icon: React.ComponentType<{ size?: number }>;
  accent: string;
  enabled: boolean;
  end?: boolean;
  onNavigate: () => void;
};

function NavItem({
  to,
  label,
  Icon,
  accent,
  enabled,
  end,
  onNavigate,
}: NavItemProps) {
  if (!enabled) {
    return (
      <div
        className="flex items-center justify-between px-3 py-2.5 text-[14px] font-medium rounded-sm opacity-40 cursor-not-allowed"
        style={{ color: "#f6f5f1" }}
        title="Not built yet"
      >
        <span className="flex items-center gap-3">
          <Icon size={18} />
          {label}
        </span>
        <span
          className="text-[10px] font-semibold tracking-wide uppercase px-1.5 py-0.5 rounded-sm"
          style={{ backgroundColor: "rgba(246,245,241,0.12)" }}
        >
          Soon
        </span>
      </div>
    );
  }

  return (
    <NavLink
      to={to}
      end={end}
      onClick={onNavigate}
      className={({ isActive }) =>
        `flex items-center gap-3 px-3 py-2.5 text-[14px] font-medium rounded-sm transition-colors ${
          isActive ? "" : "hover:bg-white/5"
        }`
      }
      style={({ isActive }) => ({
        color: "#f6f5f1",
        backgroundColor: isActive ? "rgba(246,245,241,0.1)" : "transparent",
        borderLeft: isActive
          ? `2px solid ${accent === "#e1ded6" ? "#b8863a" : accent}`
          : "2px solid transparent",
      })}
    >
      <Icon size={18} />
      {label}
    </NavLink>
  );
}
