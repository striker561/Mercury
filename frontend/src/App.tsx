import { useState, useEffect } from "react";
import { MercuryApp } from "../bindings/mercury/app";

function App() {
  const [passphrase, setPassphrase] = useState("");
  const [showPassphrase, setShowPassphrase] = useState(false);
  const [syncEnabled, setSyncEnabled] = useState(false);
  const [peers, setPeers] = useState<string[]>([]);
  const [receivedFolder] = useState("~/Mercury/");
  const [version] = useState("");

  useEffect(() => {
    MercuryApp.GetVersion().then((v) => {
      // version state update handled via setVersion if needed
    });
    loadPeers();
    const interval = setInterval(loadPeers, 5000);
    return () => clearInterval(interval);
  }, []);

  const loadPeers = () => {
    MercuryApp.GetPeers().then((p) => {
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

  return (
    <div className="settings">
      <div className="settings-header">
        <h1>Mercury</h1>
        <span className="settings-version">{version}</span>
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
            Save & Enable Sync
          </button>
          <button className="btn-secondary" onClick={handleToggleSync}>
            {syncEnabled ? "Pause Sync" : "Resume Sync"}
          </button>
        </div>
      </section>

      {/* Section 2 — Peers */}
      <section className="settings-section">
        <h2>Peers</h2>
        {peers.length === 0 ? (
          <p className="settings-empty">No peers discovered yet</p>
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
              Change
            </button>
          </div>
        </div>
      </section>
    </div>
  );
}

export default App;
