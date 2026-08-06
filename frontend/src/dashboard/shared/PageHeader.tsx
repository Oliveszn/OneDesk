import type { ReactNode } from "react";

interface PageHeaderProps {
  title: string;
  description?: string;
  actions?: ReactNode;
}
export default function PageHeader({
  title,
  description,
  actions,
}: PageHeaderProps) {
  return (
    <div className="flex items-start justify-between gap-6 px-8 pt-8 pb-6">
      <div>
        <h1
          className="text-[22px] font-bold tracking-tight"
          style={{ color: "#16191c" }}
        >
          {title}
        </h1>
        {description && (
          <p className="text-[14px] mt-1" style={{ color: "#52585d" }}>
            {description}
          </p>
        )}
      </div>
      {actions && (
        <div className="flex items-center gap-3 shrink-0">{actions}</div>
      )}
    </div>
  );
}
