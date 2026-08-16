import { useState, useCallback } from "react";
import { X } from "@phosphor-icons/react";
import { MercuryApp } from "../bindings/mercury/app";
import { copy } from "./copy";
import { useDashboard } from "./hooks/useDashboard";
import { useMercuryEvents } from "./hooks/useMercuryEvents";
import { useSyncActivity } from "./hooks/useSyncActivity";
import Settings from "./components/Settings";
import FileTransfer from "./components/FileTransfer";
import Welcome from "./components/Welcome";
import ConnectionHero from "./components/ConnectionHero";
import ActivityStrip from "./components/ActivityStrip";
import PeerList from "./components/PeerList";
import SegmentedControl from "./components/SegmentedControl";
import type { Tab } from "./types/mercury";

function App() {
  const { state, refresh } = useDashboard();
  const [tab, setTab] = useState<Tab>("home");
  const [focusPassphrase, setFocusPassphrase] = useState(false);

  const { peers, paused, transfers } = state;
  const { active: syncActive, pulse: pulseSync } = useSyncActivity(
    peers.length,
    paused,
  );

  useMercuryEvents(refresh, pulseSync);

  const dotClass = paused
    ? "status-dot paused"
    : peers.length > 0
      ? "status-dot connected pulse"
      : "status-dot idle";

  const statusLabel = paused
    ? copy.header.paused
    : peers.length > 0
      ? copy.header.peers(peers.length)
      : copy.header.idle;

  const isTransferring = transfers.some(
    (t) => t.status === "sending" || t.status === "receiving",
  );

  const handleGetStarted = useCallback(() => {
    setTab("settings");
    setFocusPassphrase(true);
  }, []);

  const handleTabChange = useCallback((next: Tab) => {
    setTab(next);
    if (next !== "settings") setFocusPassphrase(false);
  }, []);

  const handlePassphraseSaved = useCallback(() => {
    setFocusPassphrase(false);
    refresh();
  }, [refresh]);

  const hideWindow = useCallback(() => {
    MercuryApp.HideWindow();
  }, []);

  const showWelcome = tab === "home" && !state.hasPassphrase;

  return (
    <>
      <header className="header">
        <div className="header-brand">
          <img
            src="/mercury-logo.png"
            alt=""
            className={`header-logo${isTransferring ? " header-logo-active" : ""}`}
            aria-hidden
          />
          <span className="header-title">Mercury</span>
        </div>
        <div className="header-spacer" />
        <div className="header-actions">
          <div className="header-status">
            <span className={dotClass} aria-hidden />
            <span>{statusLabel}</span>
          </div>
          <SegmentedControl active={tab} onChange={handleTabChange} />
          <button
            type="button"
            className="window-close"
            onClick={hideWindow}
            aria-label="Close window"
          >
            <X size={14} weight="bold" />
          </button>
        </div>
      </header>

      <main
        className={`content${tab === "settings" ? " content-settings" : ""}`}
      >
        <div key={tab} className="content-panel">
          {tab === "settings" ? (
            <Settings
              focusPassphrase={focusPassphrase}
              onPassphraseSaved={handlePassphraseSaved}
            />
          ) : (
            <>
              {showWelcome && <Welcome onGetStarted={handleGetStarted} />}
              <ConnectionHero state={state} />
              <ActivityStrip active={syncActive} />
              <FileTransfer
                offers={state.offers}
                transfers={state.transfers}
                onChange={refresh}
              />
              <PeerList peers={state.peers} />
            </>
          )}
        </div>
      </main>
    </>
  );
}

export default App;
