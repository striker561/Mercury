import { copy } from "../copy";
import type { DashboardState } from "../types/mercury";

interface Props {
  state: DashboardState;
}

export default function StatusBar({ state }: Props) {
  const { peers, paused, hint, gnomeTrayTip } = state;
  const count = peers.length;

  let title: string;
  let sub: string;

  if (paused) {
    title = copy.status.pausedTitle;
    sub = copy.status.pausedSub;
  } else if (hint) {
    title = copy.status.issueTitle;
    sub = hint;
  } else if (count > 0) {
    title = copy.status.connectedTitle(count);
    sub = copy.status.connectedSub;
  } else {
    title = copy.status.waitingTitle;
    sub = copy.status.waitingSub;
  }

  return (
    <div className="status-bar">
      <div className="status-bar-title">{title}</div>
      <div className={`status-bar-sub${hint ? " warn" : ""}`}>{sub}</div>
      {gnomeTrayTip && !hint && (
        <div className="status-bar-sub" style={{ marginTop: 6 }}>
          {copy.status.gnomeTip}
        </div>
      )}
    </div>
  );
}
