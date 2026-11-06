import { useCallback, useState } from "react";
import { MercuryApp } from "../../bindings/mercury/app";
import { copy } from "../copy";
import Bolt, { type BoltState } from "./Bolt";
import type { DashboardState } from "../types/mercury";

interface Props {
  state: DashboardState;
  onRetry?: () => void;
  strike?: number;
  centered?: boolean;
}

export default function ConnectionHero({
  state,
  onRetry,
  strike,
  centered,
}: Props) {
  const { peers, paused, hint, hasPassphrase, syncRunning, syncError } = state;
  const count = peers.length;
  const [retrying, setRetrying] = useState(false);

  const retry = useCallback(async () => {
    setRetrying(true);
    try {
      const pass = await MercuryApp.GetSavedPassphrase();
      if (pass) {
        await MercuryApp.SetPassphrase(pass);
      }
      onRetry?.();
    } finally {
      setRetrying(false);
    }
  }, [onRetry]);

  let title: string;
  let sub: string;
  let warn = false;
  let showRetry = false;

  if (paused) {
    title = copy.status.pausedTitle;
    sub = copy.status.pausedSub;
  } else if (syncError) {
    title = copy.status.syncErrorTitle;
    sub = syncError;
    warn = true;
    showRetry = hasPassphrase;
  } else if (hasPassphrase && !syncRunning) {
    title = copy.status.syncStoppedTitle;
    sub = copy.status.syncStoppedSub;
    warn = true;
    showRetry = true;
  } else if (hint) {
    title = copy.status.issueTitle;
    sub = hint;
    warn = true;
  } else if (count > 0) {
    title = copy.status.connectedTitle;
    sub = copy.status.connectedSub;
  } else {
    title = copy.status.searchingTitle;
    sub = copy.status.searchingSub;
  }

  const glyph: BoltState = paused
    ? "resting"
    : syncError || (hasPassphrase && !syncRunning)
      ? "error"
      : hint
        ? "warn"
        : count > 0
          ? "live"
          : "idle";

  return (
    <div
      className={`connection-hero${centered ? " connection-hero-solo" : ""}`}
    >
      <div className="connection-hero-top">
        <Bolt state={glyph} height={centered ? 64 : 48} strike={strike} />
        <div className="connection-hero-text" aria-live="polite">
          <div className={`connection-hero-title${warn ? " warn" : ""}`}>
            {title}
          </div>
          <div className={`connection-hero-sub${warn ? " warn" : ""}`}>
            {sub}
          </div>
        </div>
      </div>
      {showRetry && (
        <div className="connection-hero-action">
          <button
            type="button"
            className="btn btn-primary btn-sm"
            onClick={retry}
            disabled={retrying}
          >
            {retrying ? copy.status.retrying : copy.status.retry}
          </button>
        </div>
      )}
    </div>
  );
}
