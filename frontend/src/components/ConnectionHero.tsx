import { Fragment } from "react";
import { copy } from "../copy";
import type { DashboardState } from "../types/mercury";

interface Props {
  state: DashboardState;
}

function ConnectionNodes({
  count,
  paused,
  connected,
}: {
  count: number;
  paused: boolean;
  connected: boolean;
}) {
  if (paused) {
    return <span className="status-dot paused" aria-hidden />;
  }

  if (connected && count > 0) {
    const display = Math.min(count, 5);
    return (
      <div className="connection-nodes" aria-hidden>
        {Array.from({ length: display }, (_, i) => (
          <Fragment key={i}>
            {i > 0 && <span className="connection-line" />}
            <span className="connection-node connected pulse" />
          </Fragment>
        ))}
        {count > 5 && <span className="connection-more">+{count - 5}</span>}
      </div>
    );
  }

  return (
    <div className="connection-nodes" aria-hidden>
      <span className="connection-node idle" />
    </div>
  );
}

export default function ConnectionHero({ state }: Props) {
  const { peers, paused, hint, vpnActive, gnomeTrayTip } = state;
  const count = peers.length;

  let title: string;
  let sub: string;
  let warn = false;
  let heroClass = "connection-hero";

  if (paused) {
    title = copy.status.pausedTitle;
    sub = copy.status.pausedSub;
  } else if (hint) {
    title = copy.status.issueTitle;
    sub = hint;
    warn = true;
    heroClass += " connection-hero-warn";
  } else if (vpnActive) {
    title = copy.status.vpnTitle;
    sub = copy.status.vpnSub;
    warn = true;
    heroClass += " connection-hero-warn";
  } else if (count > 0) {
    title = copy.status.connectedTitle(count);
    sub = copy.status.connectedSub;
  } else {
    title = copy.status.waitingTitle;
    sub = copy.status.waitingSub;
  }

  const showNodes = !hint && !vpnActive;

  return (
    <div className={heroClass}>
      <div className="connection-hero-top">
        {showNodes && (
          <ConnectionNodes
            count={count}
            paused={paused}
            connected={count > 0 && !paused}
          />
        )}
        <div className="connection-hero-text">
          <div className="connection-hero-title">{title}</div>
          <div className={`connection-hero-sub${warn ? " warn" : ""}`}>
            {sub}
          </div>
        </div>
      </div>
      {gnomeTrayTip && !hint && !vpnActive && (
        <div className="connection-hero-foot">{copy.status.gnomeTip}</div>
      )}
    </div>
  );
}
