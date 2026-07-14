import type { DashboardState } from "../types/mercury";

interface Props {
  state: DashboardState;
}

export default function StatusBar({ state }: Props) {
  const { peers, paused } = state;
  const count = peers.length;

  let title: string;
  let sub: string;

  if (paused) {
    title = "Paused";
    sub = "Clipboard sync is paused. Resume in Settings when ready.";
  } else if (count > 0) {
    title = `Connected · ${count} device${count === 1 ? "" : "s"}`;
    sub = "Copy on one machine — paste on another.";
  } else {
    title = "Waiting for peers";
    sub = "Make sure other devices run Mercury with the same passphrase.";
  }

  return (
    <div className="status-bar">
      <div className="status-bar-title">{title}</div>
      <div className="status-bar-sub">{sub}</div>
    </div>
  );
}
