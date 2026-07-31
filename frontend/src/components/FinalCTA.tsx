import { Button } from "antd";
import { ArrowRight02Icon } from "hugeicons-react";

export default function FinalCTA({ onRegister }) {
  return (
    <>
      <section
        className="border-t"
        style={{ borderColor: "#e1ded6", backgroundColor: "#fbfaf8" }}
      >
        <div className="max-w-6xl mx-auto px-6 py-24 text-center">
          <h2
            className="text-[32px] sm:text-[42px] font-extrabold tracking-tight leading-tight mb-6 max-w-2xl mx-auto"
            style={{ color: "#16191c" }}
          >
            Stop reconciling four tools by hand.
          </h2>
          <p
            className="text-[16px] leading-relaxed mb-9 max-w-md mx-auto"
            style={{ color: "#52585d" }}
          >
            Free plan, no card required. Upgrade the moment the caps actually
            get in your way, not before.
          </p>
          <Button
            type="primary"
            size="large"
            onClick={onRegister}
            className="!h-[52px] !px-8 !text-[15px]"
          >
            Start free{" "}
            <ArrowRight02Icon size={18} className="inline ml-1 -mb-0.5" />
          </Button>
        </div>
      </section>
    </>
  );
}
