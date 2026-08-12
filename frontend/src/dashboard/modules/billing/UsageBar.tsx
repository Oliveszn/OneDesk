import { Progress } from "antd";

export default function UsageBar({ label, used, cap }) {
  if (cap == null) {
    return (
      <div className="mb-5">
        <div className="flex items-center justify-between mb-1.5">
          <span
            className="text-[14px] font-medium"
            style={{ color: "#16191c" }}
          >
            {label}
          </span>
          <span className="text-[13px] font-mono" style={{ color: "#5b7a63" }}>
            Unlimited
          </span>
        </div>
        <Progress
          percent={100}
          showInfo={false}
          strokeColor="#5b7a63"
          trailColor="#e1ded6"
          size="small"
        />
      </div>
    );
  }

  const pct = Math.min(100, Math.round((used / cap) * 100));
  const color = pct >= 100 ? "#9c4a3c" : pct >= 80 ? "#b8863a" : "#5b7a63";

  return (
    <div className="mb-5">
      <div className="flex items-center justify-between mb-1.5">
        <span className="text-[14px] font-medium" style={{ color: "#16191c" }}>
          {label}
        </span>
        <span className="text-[13px] font-mono" style={{ color: "#52585d" }}>
          {used} / {cap}
        </span>
      </div>
      <Progress
        percent={pct}
        showInfo={false}
        strokeColor={color}
        trailColor="#e1ded6"
        size="small"
      />
    </div>
  );
}
