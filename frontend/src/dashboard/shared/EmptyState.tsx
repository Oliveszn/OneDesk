export default function EmptyState({ Icon, title, description, action }) {
  return (
    <div className="flex flex-col items-center justify-center text-center py-20 px-6">
      <div
        className="w-12 h-12 flex items-center justify-center mb-5"
        style={{ backgroundColor: "#e1ded6", color: "#52585d" }}
      >
        <Icon size={22} strokeWidth={1.75} />
      </div>
      <h3
        className="text-[16px] font-bold tracking-tight mb-1.5"
        style={{ color: "#16191c" }}
      >
        {title}
      </h3>
      {description && (
        <p className="text-[14px] max-w-sm mb-6" style={{ color: "#52585d" }}>
          {description}
        </p>
      )}
      {action}
    </div>
  );
}
