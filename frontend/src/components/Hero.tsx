import { Button } from "antd";
import { ArrowRight02Icon } from "hugeicons-react";
import LedgerTabStack from "./LedgerTabStack";

export default function Hero({ onRegister }) {
  return (
    <section
      id="top"
      className="max-w-6xl mx-auto px-6 pt-20 pb-24 md:pt-28 md:pb-32 grid md:grid-cols-2 gap-16 items-center"
    >
      <div>
        <div className="inline-flex items-center gap-2 mb-6">
          <span
            className="w-6 h-[2px]"
            style={{ backgroundColor: "#b8863a" }}
          />
          <span
            className="text-[13px] font-semibold tracking-[0.14em] uppercase"
            style={{ color: "#b8863a" }}
          >
            OneDesk ERP
          </span>
        </div>

        <h1
          className="text-[40px] sm:text-[52px] leading-[1.05] font-extrabold tracking-tight mb-6"
          style={{ color: "#16191c" }}
        >
          Every department,
          <br />
          one ledger.
        </h1>

        <p
          className="text-[17px] leading-relaxed max-w-md mb-9"
          style={{ color: "#52585d" }}
        >
          Inventory, sales, invoicing, and procurement — running off the same
          record, updating each other automatically. Place an order and watch
          stock, invoices, and reorders fall into line behind it.
        </p>

        <div className="flex flex-wrap items-center gap-4">
          <Button
            type="primary"
            size="large"
            onClick={onRegister}
            className="!h-[50px] !px-7 !text-[15px]"
          >
            Start free{" "}
            <ArrowRight02Icon size={18} className="inline ml-1 -mb-0.5" />
          </Button>
          <a
            href="#modules"
            className="text-[15px] font-semibold underline underline-offset-4 decoration-1"
            style={{ color: "#16191c" }}
          >
            See how it works
          </a>
        </div>

        <div
          className="flex items-center gap-6 mt-12 pt-8 border-t"
          style={{ borderColor: "#e1ded6" }}
        >
          <div>
            <div
              className="font-mono text-2xl font-semibold"
              style={{ color: "#16191c" }}
            >
              4
            </div>
            <div className="text-[13px]" style={{ color: "#52585d" }}>
              connected modules
            </div>
          </div>
          <div className="w-px h-8" style={{ backgroundColor: "#e1ded6" }} />
          <div>
            <div
              className="font-mono text-2xl font-semibold"
              style={{ color: "#16191c" }}
            >
              0
            </div>
            <div className="text-[13px]" style={{ color: "#52585d" }}>
              spreadsheets required
            </div>
          </div>
        </div>
      </div>

      <LedgerTabStack />
    </section>
  );
}
