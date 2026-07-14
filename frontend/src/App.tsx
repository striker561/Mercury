import { useState, useCallback } from "react";
import { X } from "@phosphor-icons/react";
import { MercuryApp } from "../bindings/mercury/app";
import { useDashboard } from "./hooks/useDashboard";
import { useMercuryEvents } from "./hooks/useMercuryEvents";
import Settings from "./components/Settings";
import FileTransfer from "./components/FileTransfer";
import Welcome from "./components/Welcome";
import StatusBar from "./components/StatusBar";
import PeerList from "./components/PeerList";
import SegmentedControl from "./components/SegmentedControl";
import type { Tab } from "./types/mercury";

function App() {
  const { state, refresh } = useDashboard();
  const [tab, setTab] = useState<Tab>("home");
  const [focusPassphrase, setFocusPassphrase] = useState(false);

  useMercuryEvents(refresh);

  const { peers, paused } = state;
  const dotClass = paused
    ? "status-dot paused"
    : peers.length > 0
      ? "status-dot connected pulse"
      : "status-dot idle";

  const statusLabel = paused
    ? "Paused"
    : peers.length > 0
      ? `${peers.length} peer${peers.length === 1 ? "" : "s"}`
      : "Idle";

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
            className="header-logo"
            aria-hidden
          />
          <span className="header-title">Mercury</span>
        </div>
        <div className="header-spacer" />
        <div className="header-actions">
          <SegmentedControl active={tab} onChange={handleTabChange} />
          <div className="header-status">
            <span className={dotClass} aria-hidden />
            <span>{statusLabel}</span>
          </div>
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
        {tab === "settings" ? (
          <Settings
            focusPassphrase={focusPassphrase}
            onPassphraseSaved={handlePassphraseSaved}
          />
        ) : (
          <>
            {showWelcome && <Welcome onGetStarted={handleGetStarted} />}
            <StatusBar state={state} />
            <FileTransfer
              offers={state.offers}
              transfers={state.transfers}
              onChange={refresh}
            />
            <PeerList peers={state.peers} />
          </>
        )}
      </main>
    </>
  );
}

export default App;
