import {
  Package01Icon,
  ShoppingCart01Icon,
  Invoice03Icon,
  TruckDeliveryIcon,
} from "hugeicons-react";

const modules = [
  {
    Icon: Package01Icon,
    name: "Inventory",
    accent: "#24344a",
    copy: "Stock tracked per warehouse, adjusted the instant an order or delivery touches it — never two requests fighting over the same count.",
  },
  {
    Icon: ShoppingCart01Icon,
    name: "Sales",
    accent: "#b8863a",
    copy: "Orders that don't just get recorded — placing one sets everything downstream in motion automatically.",
  },
  {
    Icon: Invoice03Icon,
    name: "Finance",
    accent: "#5b7a63",
    copy: "An invoice appears the moment an order does. No end-of-day reconciliation, no separate step to remember.",
  },
  {
    Icon: TruckDeliveryIcon,
    name: "Procurement",
    accent: "#16191c",
    copy: "Falling stock suggests its own reorder before you're out — a purchase order waiting on a vendor, not a stockout waiting to happen.",
  },
];

export default function Modules() {
  return (
    <section
      id="modules"
      className="border-t"
      style={{ borderColor: "#e1ded6", backgroundColor: "#fbfaf8" }}
    >
      <div className="max-w-6xl mx-auto px-6 py-24">
        <div className="max-w-lg mb-16">
          <div className="inline-flex items-center gap-2 mb-5">
            <span
              className="w-6 h-[2px]"
              style={{ backgroundColor: "#b8863a" }}
            />
            <span
              className="text-[13px] font-semibold tracking-[0.14em] uppercase"
              style={{ color: "#b8863a" }}
            >
              Four departments, one record
            </span>
          </div>
          <h2
            className="text-[32px] sm:text-[38px] font-extrabold tracking-tight leading-tight"
            style={{ color: "#16191c" }}
          >
            They don't just coexist. They talk to each other.
          </h2>
        </div>

        <div
          className="grid sm:grid-cols-2 gap-px"
          style={{ backgroundColor: "#e1ded6" }}
        >
          {modules.map((m) => (
            <div
              key={m.name}
              className="p-8 md:p-10"
              style={{ backgroundColor: "#fbfaf8" }}
            >
              <div
                className="w-11 h-11 flex items-center justify-center mb-6"
                style={{ backgroundColor: m.accent, color: "#f6f5f1" }}
              >
                <m.Icon size={22} strokeWidth={1.75} />
              </div>
              <h3
                className="text-[19px] font-bold tracking-tight mb-2.5"
                style={{ color: "#16191c" }}
              >
                {m.name}
              </h3>
              <p
                className="text-[15px] leading-relaxed"
                style={{ color: "#52585d" }}
              >
                {m.copy}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
