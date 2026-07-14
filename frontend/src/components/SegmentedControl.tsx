import type { Tab } from "../types/mercury";

interface Props {
  active: Tab;
  onChange: (tab: Tab) => void;
}

export default function SegmentedControl({ active, onChange }: Props) {
  return (
    <div className="segmented" role="tablist" aria-label="Navigation">
      <button
        type="button"
        role="tab"
        aria-selected={active === "home"}
        className={`segmented-btn${active === "home" ? " active" : ""}`}
        onClick={() => onChange("home")}
      >
        Home
      </button>
      <button
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
