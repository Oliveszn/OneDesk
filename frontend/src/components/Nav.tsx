import { useState } from "react";
import { Button } from "antd";
import { Menu01Icon, Cancel01Icon } from "hugeicons-react";

const links = [
  { href: "#modules", label: "Modules" },
  { href: "#pricing", label: "Pricing" },
];

export default function Nav({ onLogin, onRegister }) {
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <header
      className="sticky top-0 z-40 border-b"
      style={{ borderColor: "#e1ded6", backgroundColor: "#f6f5f1" }}
    >
      <div
        className="max-w-6xl mx-auto px-6 h-18 flex items-center justify-between"
        style={{ height: "72px" }}
      >
        <a href="#top" className="flex items-center gap-2">
          <span
            className="w-7 h-7 flex items-center justify-center text-sm font-bold"
            style={{ backgroundColor: "#24344a", color: "#f6f5f1" }}
          >
            O
          </span>
          <span
            className="font-semibold tracking-tight text-[17px]"
            style={{ color: "#16191c" }}
          >
            One<span style={{ color: "#b8863a" }}>Desk</span>
          </span>
        </a>

        <nav className="hidden md:flex items-center gap-8">
          {links.map((l) => (
            <a
              key={l.href}
              href={l.href}
              className="text-[15px] font-medium transition-opacity hover:opacity-70"
              style={{ color: "#52585d" }}
            >
              {l.label}
            </a>
          ))}
        </nav>

        <div className="hidden md:flex items-center gap-3">
          <button
            onClick={onLogin}
            className="text-[15px] font-medium px-3 py-2 transition-opacity hover:opacity-70"
            style={{ color: "#16191c" }}
          >
            Log in
          </button>
          <Button type="primary" onClick={onRegister}>
            Get started
          </Button>
        </div>

        <button
          className="md:hidden p-2"
          onClick={() => setMobileOpen((v) => !v)}
          aria-label={mobileOpen ? "Close menu" : "Open menu"}
          aria-expanded={mobileOpen}
        >
          {mobileOpen ? <Cancel01Icon size={22} /> : <Menu01Icon size={22} />}
        </button>
      </div>

      {mobileOpen && (
        <div
          className="md:hidden border-t px-6 py-4 flex flex-col gap-4"
          style={{ borderColor: "#e1ded6" }}
        >
          {links.map((l) => (
            <a
              key={l.href}
              href={l.href}
              onClick={() => setMobileOpen(false)}
              className="text-[15px] font-medium"
              style={{ color: "#52585d" }}
            >
              {l.label}
            </a>
          ))}
          <div className="flex flex-col gap-2 pt-2">
            <Button
              onClick={() => {
                setMobileOpen(false);
                onLogin();
              }}
            >
              Log in
            </Button>
            <Button
              type="primary"
              onClick={() => {
                setMobileOpen(false);
                onRegister();
              }}
            >
              Get started
            </Button>
          </div>
        </div>
      )}
    </header>
  );
}
