import { Link } from "react-router-dom";
import { Package01Icon, ArrowRight02Icon } from "hugeicons-react";
import PageHeader from "../../shared/PageHeader";

export default function OverviewPage() {
  return (
    <div>
      <PageHeader title="Overview" description="OneDesk · Free plan" />
      <div className="px-8">
        <Link
          to="/dashboard/inventory"
          className="flex items-center justify-between p-6 max-w-md border transition-colors hover:bg-black/[0.02]"
          style={{ borderColor: "#e1ded6", backgroundColor: "#fbfaf8" }}
        >
          <span className="flex items-center gap-3">
            <span
              className="w-10 h-10 flex items-center justify-center"
              style={{ backgroundColor: "#24344a", color: "#f6f5f1" }}
            >
              <Package01Icon size={20} />
            </span>
            <span>
              <span
                className="block font-semibold text-[15px]"
                style={{ color: "#16191c" }}
              >
                Inventory
              </span>
              <span className="block text-[13px]" style={{ color: "#52585d" }}>
                Products, warehouses, stock
              </span>
            </span>
          </span>
          <ArrowRight02Icon size={18} color="#52585d" />
        </Link>
      </div>
    </div>
  );
}
