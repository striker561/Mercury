import { copy } from "../copy";

interface Props {
  active: boolean;
}

export default function ActivityStrip({ active }: Props) {
  return (
    <div
      className={`activity-strip${active ? " activity-strip-visible" : ""}`}
      role="status"
      aria-live="polite"
      aria-hidden={!active}
    >
      <span className="activity-strip-shimmer" aria-hidden />
      {copy.activity.carrying}
    </div>
  );
}
