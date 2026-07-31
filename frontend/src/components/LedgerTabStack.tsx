import {
  Package01Icon,
  ShoppingCart01Icon,
  Invoice03Icon,
  TruckDeliveryIcon,
  Wallet02Icon,
} from "hugeicons-react";

const tabs = [
  {
    label: "Billing",
    Icon: Wallet02Icon,
    bg: "#e1ded6",
    ink: "#16191c",
    rotate: -10,
    z: 1,
  },
  {
    label: "Procurement",
    Icon: TruckDeliveryIcon,
    bg: "#5b7a63",
    ink: "#f6f5f1",
    rotate: -5,
    z: 2,
  },
  {
    label: "Finance",
    Icon: Invoice03Icon,
    bg: "#24344a",
    ink: "#f6f5f1",
    rotate: 0,
    z: 3,
  },
  {
    label: "Sales",
    Icon: ShoppingCart01Icon,
    bg: "#b8863a",
    ink: "#16191c",
    rotate: 5,
    z: 4,
  },
  {
    label: "Inventory",
    Icon: Package01Icon,
    bg: "#fbfaf8",
    ink: "#16191c",
    rotate: 10,
    z: 5,
    border: true,
  },
];

export default function LedgerTabStack() {
  return (
    <div
      className="relative h-[320px] w-full max-w-[420px] mx-auto select-none"
      aria-hidden="true"
    >
      {tabs.map((tab, i) => (
        <div
          key={tab.label}
          className="absolute inset-x-0 top-1/2 flex items-center gap-3 px-6 py-5 w-[280px] left-1/2"
          style={{
            backgroundColor: tab.bg,
            color: tab.ink,
            transform: `translate(-50%, -50%) rotate(${tab.rotate}deg) translateY(${i * 6}px)`,
            zIndex: tab.z,
            border: tab.border ? "1px solid #e1ded6" : "none",
            boxShadow: tab.border ? "3px 3px 0 0 #e1ded6" : "none",
          }}
        >
          <tab.Icon size={22} strokeWidth={1.75} />
          <span className="font-semibold text-[15px] tracking-tight">
            {tab.label}
          </span>
        </div>
      ))}
    </div>
  );
}
