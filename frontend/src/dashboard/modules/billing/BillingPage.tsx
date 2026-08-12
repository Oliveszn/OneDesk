import { useState } from "react";
import { Button, Tag, message } from "antd";
import PageHeader from "../../shared/PageHeader";
import UsageBar from "./UsageBar";
import UpgradeModal from "./UpgradeModal";
import { mockUsage as initialUsage, mockPlans } from "../../mockData/billing";
import { formatNaira } from "../../shared/format";

export default function BillingPage() {
  const [usage, setUsage] = useState(initialUsage);
  const [upgradeModalOpen, setUpgradeModalOpen] = useState(false);

  const isPaid = usage.plan_name === "paid";
  const currentPlan = mockPlans.find((p) => p.name === usage.plan_name);

  const handleUpgrade = (values) => {
    console.log("upgrade requested", values);
    setUsage((prev) => ({
      ...prev,
      plan_name: "paid",
      products_cap: null,
      orders_cap: null,
      users_cap: null,
    }));
    message.success(
      "Upgraded to Paid (simulated — no real checkout happened).",
    );
  };

  return (
    <div>
      <PageHeader title="Billing" description="Your plan and usage." />

      <div className="px-8 max-w-lg">
        <div
          className="p-6 mb-8"
          style={{
            backgroundColor: isPaid ? "#24344a" : "#fbfaf8",
            border: isPaid ? "none" : "1px solid #e1ded6",
          }}
        >
          <div className="flex items-center justify-between mb-1">
            <span
              className="text-[13px] font-bold tracking-[0.08em] uppercase"
              style={{ color: "#b8863a" }}
            >
              {usage.plan_name} plan
            </span>
            {isPaid && (
              <Tag color="#5b7a63" style={{ margin: 0, fontWeight: 600 }}>
                Active
              </Tag>
            )}
          </div>
          <div
            className="font-mono text-[28px] font-semibold mt-2"
            style={{ color: isPaid ? "#f6f5f1" : "#16191c" }}
          >
            {currentPlan ? formatNaira(currentPlan.price_amount) : "—"}
            {currentPlan?.billing_interval && (
              <span
                className="text-[15px] font-sans font-normal ml-1"
                style={{ color: isPaid ? "#c7ccd1" : "#52585d" }}
              >
                /month
              </span>
            )}
          </div>

          {!isPaid && (
            <Button
              type="primary"
              className="mt-5"
              onClick={() => setUpgradeModalOpen(true)}
            >
              Upgrade to Paid
            </Button>
          )}
        </div>

        <div
          className="text-[13px] font-semibold uppercase tracking-wide mb-4"
          style={{ color: "#52585d" }}
        >
          Usage this period
        </div>

        <UsageBar
          label="Products"
          used={usage.products_used}
          cap={usage.products_cap}
        />
        <UsageBar
          label="Orders"
          used={usage.orders_used}
          cap={usage.orders_cap}
        />
        <UsageBar
          label="Team members"
          used={usage.users_used}
          cap={usage.users_cap}
        />
      </div>

      <UpgradeModal
        open={upgradeModalOpen}
        onClose={() => setUpgradeModalOpen(false)}
        onUpgrade={handleUpgrade}
      />
    </div>
  );
}
