import { GearSix, House } from "@phosphor-icons/react";
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
        aria-label="Home"
        className={`segmented-btn${active === "home" ? " active" : ""}`}
        onClick={() => onChange("home")}
      >
        <House size={17} weight={active === "home" ? "fill" : "regular"} />
      </button>
      <button
        type="button"
        role="tab"
        aria-selected={active === "settings"}
        aria-label="Settings"
        className={`segmented-btn${active === "settings" ? " active" : ""}`}
        onClick={() => onChange("settings")}
      >
        <GearSix
          size={17}
          weight={active === "settings" ? "fill" : "regular"}
        />
      </button>
    </div>
  );
}
