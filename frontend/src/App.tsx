import { useState, useEffect } from "react";
import { MercuryApp } from "../bindings/mercury/app";

const SETTINGS_KEYS = {
  passphrase: "passphrase",
  syncEnabled: "sync_enabled",
  paused: "paused",
  allowFiles: "allow_files",
  receivedFolder: "received_folder",
  autostart: "autostart",
};

function App() {
  const [passphrase, setPassphrase] = useState("");
  const [showPassphrase, setShowPassphrase] = useState(false);
  const [syncEnabled, setSyncEnabled] = useState(false);
  const [allowFiles, setAllowFiles] = useState(true);
  const [autostart, setAutostart] = useState(false);
  const [peers, setPeers] = useState<string[]>([]);
  const [receivedFolder, setReceivedFolder] = useState("~/Mercury/");
  const [version, setVersion] = useState("");

  useEffect(() => {
    // One IPC call loads version + all settings.
    MercuryApp.GetAllSettings().then((s) => {
      if (!s) return;
      setVersion(s["version"] ?? "");
      setPassphrase(s[SETTINGS_KEYS.passphrase] ?? "");
      setSyncEnabled(s[SETTINGS_KEYS.syncEnabled] === "true");
      setAllowFiles(s[SETTINGS_KEYS.allowFiles] !== "false");
      setAutostart(s[SETTINGS_KEYS.autostart] === "true");
      setReceivedFolder(s[SETTINGS_KEYS.receivedFolder] ?? "~/Mercury/");
    });
    loadPeers();
    const interval = setInterval(loadPeers, 5000);
    return () => clearInterval(interval);
  }, []);

  const loadPeers = () => {
    MercuryApp.GetPeers().then((p) => {
      if (p)
        setPeers(p.map((peer: any) => peer.hostname || peer.id || "unknown"));
    });
  };

  const handlePassphraseSave = () => {
    MercuryApp.SetPassphrase(passphrase);
    setSyncEnabled(true);
  };

  const handleToggleSync = () => {
    MercuryApp.TogglePause().then((paused) => {
      setSyncEnabled(!paused);
    });
  };

  const handleToggleFiles = () => {
    const next = !allowFiles;
    setAllowFiles(next);
    MercuryApp.SetSetting(SETTINGS_KEYS.allowFiles, next ? "true" : "false");
  };

  const handleToggleAutostart = () => {
    const next = !autostart;
    setAutostart(next);
    MercuryApp.SetSetting(SETTINGS_KEYS.autostart, next ? "true" : "false");
  };

  return (
    <div className="settings">
      <div className="settings-header">
        <h1>Mercury</h1>
        <span className="settings-version">v{version}</span>
      </div>

      {/* Section 1 — Sync */}
      <section className="settings-section">
        <h2>Sync</h2>
        <div className="settings-field">
          <label>Passphrase</label>
          <div className="password-row">
            <input
              type={showPassphrase ? "text" : "password"}
              value={passphrase}
              onChange={(e) => setPassphrase(e.target.value)}
              placeholder="Enter shared passphrase"
            />
            <button
              className="btn-icon"
              onClick={() => setShowPassphrase(!showPassphrase)}
            >
              {showPassphrase ? "Hide" : "Show"}
            </button>
          </div>
        </div>
        <div className="settings-actions">
          <button
            className="btn-primary"
            onClick={handlePassphraseSave}
            disabled={!passphrase}
          >
            Save
          </button>
          <button className="btn-secondary" onClick={handleToggleSync}>
            {syncEnabled ? "Pause" : "Resume"}
          </button>
        </div>
      </section>

      {/* Section 2 — Peers */}
      <section className="settings-section">
        <h2>Peers</h2>
        {peers.length === 0 ? (
          <p className="settings-empty">No peers on network</p>
        ) : (
          <ul className="peer-list">
            {peers.map((peer, i) => (
              <li key={i} className="peer-item">
                <span className="peer-dot" />
                {peer}
              </li>
            ))}
          </ul>
        )}
      </section>

      {/* Section 3 — Files */}
      <section className="settings-section">
        <h2>Files</h2>
        <div className="settings-field">
          <label>Received files folder</label>
          <div className="folder-row">
            <span className="folder-path">{receivedFolder}</span>
            <button className="btn-secondary" disabled>
              Browse
            </button>
          </div>
        </div>
        <div className="toggle-row">
          <label className="toggle-label">
            <span>Accept incoming files</span>
            <input
              type="checkbox"
              className="toggle"
              checked={allowFiles}
              onChange={handleToggleFiles}
            />
          </label>
        </div>
      </section>

      {/* Section 4 — Preferences */}
      <section className="settings-section">
        <h2>Preferences</h2>
        <div className="toggle-row">
          <label className="toggle-label">
            <span>Start on login</span>
            <input
              type="checkbox"
              className="toggle"
              checked={autostart}
              onChange={handleToggleAutostart}
            />
          </label>
        </div>
      </section>
    </div>
  );
}

export default App;
