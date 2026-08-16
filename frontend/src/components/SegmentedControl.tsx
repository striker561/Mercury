import { useRef, useEffect, useState } from "react";
import type { Tab } from "../types/mercury";

interface Props {
  active: Tab;
  onChange: (tab: Tab) => void;
}

export default function SegmentedControl({ active, onChange }: Props) {
  const homeRef = useRef<HTMLButtonElement>(null);
  const settingsRef = useRef<HTMLButtonElement>(null);
  const [indicator, setIndicator] = useState({ left: 0, width: 0 });

  useEffect(() => {
    const el = active === "home" ? homeRef.current : settingsRef.current;
    if (!el) return;
    setIndicator({ left: el.offsetLeft, width: el.offsetWidth });
  }, [active]);

  return (
    <div className="segmented" role="tablist" aria-label="Navigation">
      <span
        className="segmented-indicator"
        style={{
          transform: `translateX(${indicator.left}px)`,
          width: indicator.width,
        }}
        aria-hidden
      />
      <button
        ref={homeRef}
        type="button"
        role="tab"
        aria-selected={active === "home"}
        className={`segmented-btn${active === "home" ? " active" : ""}`}
        onClick={() => onChange("home")}
      >
        Home
      </button>
      <button
        ref={settingsRef}
        type="button"
        role="tab"
        aria-selected={active === "settings"}
        className={`segmented-btn${active === "settings" ? " active" : ""}`}
        onClick={() => onChange("settings")}
      >
        Settings
      </button>
    </div>
  );
}
