import { Alert01Icon } from "hugeicons-react";

// Every capped resource (products, orders, users,) shows this same banner shape once a tenant is close to or at their Free-plan limit
export default function PlanLimitBanner({ resourceLabel, used, cap }) {
  if (cap == null) return null; // unlimited plan — nothing to warn about
  const atLimit = used >= cap;
  const nearLimit = !atLimit && used / cap >= 0.8;

  if (!atLimit && !nearLimit) return null;

  return (
    <div
      className="mx-8 mb-6 flex items-center gap-3 px-4 py-3 text-[14px]"
      style={{
        backgroundColor: atLimit
          ? "rgba(156,74,60,0.08)"
          : "rgba(184,134,58,0.1)",
        border: `1px solid ${atLimit ? "#9c4a3c" : "#b8863a"}`,
        color: "#16191c",
      }}
    >
      <Alert01Icon
        size={18}
        color={atLimit ? "#9c4a3c" : "#b8863a"}
        className="shrink-0"
      />
      <span>
        {atLimit
          ? `You've reached your Free plan's ${resourceLabel} limit (${used}/${cap}).`
          : `You're close to your Free plan's ${resourceLabel} limit (${used}/${cap}).`}{" "}
        <a
          href="/dashboard/billing"
          className="font-semibold underline underline-offset-2"
        >
          Upgrade to Paid
        </a>{" "}
        to remove it.
      </span>
    </div>
  );
}
