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
import HeaderNotice, { type NoticeItem } from "./components/HeaderNotice";
import PeerList from "./components/PeerList";
import SegmentedControl from "./components/SegmentedControl";
import type { Tab } from "./types/mercury";

function App() {
  const { state, ready, refresh } = useDashboard();
  const [tab, setTab] = useState<Tab>("home");
  const [focusPassphrase, setFocusPassphrase] = useState(false);

  const { peers, paused } = state;
  const transferring = state.transfers.some(
    (t) => t.status === "sending" || t.status === "receiving",
  );
  const { strikes, pulse: pulseSync } = useSyncActivity(peers.length, paused);

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

  const showHome = useCallback(() => {
    setTab("home");
    setFocusPassphrase(false);
  }, []);

  useMercuryEvents(refresh, pulseSync, showHome);

  const hideWindow = useCallback(() => {
    MercuryApp.HideWindow();
  }, []);

  const showWelcome = tab === "home" && !state.hasPassphrase;
  const notices: NoticeItem[] = [];
  if (state.vpnActive) {
    notices.push({
      icon: "vpn",
      title: copy.status.vpnTitle,
      text: copy.status.vpnSub,
    });
  }
  if (state.gnomeTrayTip) {
    notices.push({ icon: "tray", text: copy.status.gnomeTip });
  }
  const idleHome =
    tab === "home" &&
    !showWelcome &&
    peers.length === 0 &&
    state.transfers.length === 0 &&
    state.offers.length === 0;

  return (
    <>
      <header className="header">
        <div className="header-brand">
          <img
            src="/mercury-logo.png"
            alt=""
            className={`header-logo${transferring ? " header-logo-active" : ""}${paused ? " header-logo-resting" : ""}`}
            aria-hidden
          />
          <span className="visually-hidden">Mercury</span>
        </div>
        <div className="header-spacer" />
        <div className="header-actions">
          <HeaderNotice items={notices} />
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
        <div
          key={tab}
          className={`content-panel${showWelcome || idleHome ? " content-panel-centered" : ""}`}
        >
          {!ready ? (
            <div className="splash">
              <img
                src="/mercury-logo.png"
                alt=""
                className="splash-logo"
                aria-hidden
              />
            </div>
          ) : tab === "settings" ? (
            <Settings
              focusPassphrase={focusPassphrase}
              onPassphraseSaved={handlePassphraseSaved}
            />
          ) : showWelcome ? (
            <Welcome onGetStarted={handleGetStarted} />
          ) : (
            <>
              <ConnectionHero
                state={state}
                onRetry={refresh}
                strike={strikes}
                centered={idleHome}
              />
              {peers.length > 0 && (
                <PeerList peers={state.peers} paused={paused} />
              )}
              <FileTransfer
                offers={state.offers}
                transfers={state.transfers}
                onChange={refresh}
              />
            </>
          )}
        </div>
      </main>
    </>
  );
}

export default App;
