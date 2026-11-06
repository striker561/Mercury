import type { ReactNode } from "react";

export type BoltState = "live" | "idle" | "resting" | "warn" | "error";

const BOLT = "M13.6 2.2 5.2 13.6h5.2L9.8 21.8 18.4 10.4h-5.2z";

interface Props {
  state: BoltState;
  /** Rendered size in px. */
  height?: number;
  label?: string;
  /** Bump to replay the strike. */
  strike?: number;
}

function Frame({
  state,
  height = 12,
  label,
  strike = 0,
  children,
}: Props & { children: ReactNode }) {
  return (
    <svg
      key={strike}
      className={`bolt bolt-${state}${strike > 0 ? " bolt-strike" : ""}`}
      width={height}
      height={height}
      viewBox="0 0 24 24"
      role={label ? "img" : undefined}
      aria-label={label}
      aria-hidden={label ? undefined : true}
      focusable="false"
    >
      {children}
    </svg>
  );
}

export default function Bolt({ state, height = 12, label, strike }: Props) {
  // Keep the outline ~1.9px on screen whatever the rendered size.
  const stroke = 45.6 / height;

  if (state === "warn" || state === "error") {
    return (
      <Frame state={state} height={height} label={label} strike={strike}>
        <path className="bolt-body" d="M12 3.2 21.6 20.2H2.4z" />
        <path className="bolt-stem" d="M12 9.4v4.4" />
        <circle className="bolt-dot" cx="12" cy="17" r="1.2" />
      </Frame>
    );
  }

  if (state === "resting") {
    return (
      <Frame state={state} height={height} label={label} strike={strike}>
        <rect
          x="7.6"
          y="5.2"
          width="3.2"
          height="13.6"
          rx="1.2"
          fill="none"
          stroke="currentColor"
          strokeWidth={stroke}
        />
        <rect
          x="13.2"
          y="5.2"
          width="3.2"
          height="13.6"
          rx="1.2"
          fill="none"
          stroke="currentColor"
          strokeWidth={stroke}
        />
        <path
          className="bolt-travel"
          d="M7.6 5.2h3.2v13.6H7.6zM13.2 5.2h3.2v13.6h-3.2z"
          pathLength={1}
          fill="none"
          stroke="currentColor"
          strokeWidth={stroke}
          strokeLinejoin="miter"
          strokeLinecap="butt"
        />
      </Frame>
    );
  }

  const solid = state === "live";

  return (
    <Frame state={state} height={height} label={label} strike={strike}>
      <path
        className="bolt-path"
        d={BOLT}
        pathLength={1}
        fill={solid ? "currentColor" : "none"}
        stroke="currentColor"
        strokeWidth={stroke}
        strokeLinejoin="round"
      />
      <path
        className="bolt-travel"
        d={BOLT}
        pathLength={1}
        fill="none"
        stroke="currentColor"
        strokeWidth={stroke}
        strokeLinejoin="miter"
        strokeLinecap="butt"
      />
    </Frame>
  );
}
