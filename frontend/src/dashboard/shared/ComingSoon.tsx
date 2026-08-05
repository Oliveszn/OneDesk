export default function ComingSoon({ moduleName }) {
  return (
    <div className="flex flex-col items-center justify-center text-center py-32 px-6">
      <span
        className="text-[11px] font-semibold tracking-[0.14em] uppercase px-2 py-1 mb-4"
        style={{ backgroundColor: "#e1ded6", color: "#52585d" }}
      >
        Not built yet
      </span>
      <h1
        className="text-[22px] font-bold tracking-tight mb-2"
        style={{ color: "#16191c" }}
      >
        {moduleName}
      </h1>
      <p className="text-[14px] max-w-sm" style={{ color: "#52585d" }}>
        The backend for this module is complete this screen just hasn't been
        designed yet.
      </p>
    </div>
  );
}
