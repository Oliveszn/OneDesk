import { Button } from "antd";
import { CheckmarkCircle02Icon } from "hugeicons-react";

const plans = [
  {
    name: "Free",
    price: "₦0",
    period: "forever",
    tagline:
      "Everything you need to try it for real, not a watered-down trial.",
    features: [
      "Up to 3 team members",
      "Up to 50 products",
      "100 orders a month",
      "All four modules, no feature gating",
    ],
    cta: "Start free",
    featured: false,
  },
  {
    name: "Paid",
    price: "₦5,000",
    period: "/ month",
    tagline: "Once you outgrow the caps, not before.",
    features: [
      "Unlimited team members",
      "Unlimited products",
      "Unlimited orders",
      "Priority support",
    ],
    cta: "Go paid",
    featured: true,
  },
];

export default function Pricing() {
  return (
    <section id="pricing" className="max-w-6xl mx-auto px-6 py-24">
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
            Pricing
          </span>
        </div>
        <h2
          className="text-[32px] sm:text-[38px] font-extrabold tracking-tight leading-tight mb-4"
          style={{ color: "#16191c" }}
        >
          Priced on usage caps, not per-seat games.
        </h2>
        <p className="text-[16px] leading-relaxed" style={{ color: "#52585d" }}>
          One plan is free until you're actually running a business through it.
          The other removes the ceiling.
        </p>
      </div>

      <div className="grid md:grid-cols-2 gap-6">
        {plans.map((plan) => (
          <div
            key={plan.name}
            className="p-9 flex flex-col"
            style={{
              backgroundColor: plan.featured ? "#24344a" : "#fbfaf8",
              border: plan.featured ? "none" : "1px solid #e1ded6",
            }}
          >
            <div className="flex items-baseline justify-between mb-1">
              <h3
                className="text-[15px] font-bold tracking-[0.08em] uppercase"
                style={{ color: plan.featured ? "#b8863a" : "#b8863a" }}
              >
                {plan.name}
              </h3>
            </div>

            <div className="flex items-baseline gap-2 mt-4 mb-3">
              <span
                className="font-mono text-[40px] font-semibold tracking-tight"
                style={{ color: plan.featured ? "#f6f5f1" : "#16191c" }}
              >
                {plan.price}
              </span>
              <span
                className="text-[15px]"
                style={{ color: plan.featured ? "#c7ccd1" : "#52585d" }}
              >
                {plan.period}
              </span>
            </div>

            <p
              className="text-[15px] leading-relaxed mb-8"
              style={{ color: plan.featured ? "#c7ccd1" : "#52585d" }}
            >
              {plan.tagline}
            </p>

            <ul className="flex flex-col gap-3.5 mb-9 flex-1">
              {plan.features.map((f) => (
                <li
                  key={f}
                  className="flex items-start gap-3 text-[15px]"
                  style={{ color: plan.featured ? "#f6f5f1" : "#16191c" }}
                >
                  <CheckmarkCircle02Icon
                    size={19}
                    className="shrink-0 mt-0.5"
                    color={plan.featured ? "#b8863a" : "#5b7a63"}
                  />
                  <span>{f}</span>
                </li>
              ))}
            </ul>

            <Button
              size="large"
              block
              className="!h-[46px] !font-semibold"
              style={
                plan.featured
                  ? {
                      backgroundColor: "#b8863a",
                      borderColor: "#b8863a",
                      color: "#16191c",
                    }
                  : {
                      backgroundColor: "transparent",
                      borderColor: "#24344a",
                      color: "#24344a",
                    }
              }
            >
              {plan.cta}
            </Button>
          </div>
        ))}
      </div>
    </section>
  );
}
