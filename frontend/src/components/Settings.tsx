import { useState, useEffect } from "react";
import { MercuryApp } from "../../bindings/mercury/app";

const KEYS = {
  passphrase: "passphrase",
  syncEnabled: "sync_enabled",
  paused: "paused",
  allowFiles: "allow_files",
  receivedFolder: "received_folder",
  autostart: "autostart",
  autoAccept: "auto_accept",
};

export default function Settings() {
  const [passphrase, setPassphrase] = useState("");
  const [showPassphrase, setShowPassphrase] = useState(false);
  const [syncEnabled, setSyncEnabled] = useState(false);
  const [allowFiles, setAllowFiles] = useState(true);
  const [autostart, setAutostart] = useState(false);
  const [autoAccept, setAutoAccept] = useState(false);
  const [receivedFolder, setReceivedFolder] = useState("~/Mercury/");

  useEffect(() => {
    MercuryApp.GetAllSettings().then((s: any) => {
      if (!s) return;
      setPassphrase(s[KEYS.passphrase] ?? "");
      setSyncEnabled(s[KEYS.syncEnabled] === "true");
      setAllowFiles(s[KEYS.allowFiles] !== "false");
      setAutostart(s[KEYS.autostart] === "true");
      setAutoAccept(s[KEYS.autoAccept] === "true");
      setReceivedFolder(s[KEYS.receivedFolder] ?? "~/Mercury/");
    });
  }, []);

  const handlePassphraseSave = () => {
    MercuryApp.SetPassphrase(passphrase);
    setSyncEnabled(true);
  };

  const handleToggleSync = () => {
    MercuryApp.TogglePause().then((paused: boolean) => {
      setSyncEnabled(!paused);
    });
  };

  const handleToggleFiles = () => {
    const next = !allowFiles;
    setAllowFiles(next);
    MercuryApp.SetSetting(KEYS.allowFiles, next ? "true" : "false");
  };

  const handleToggleAutostart = () => {
    const next = !autostart;
    setAutostart(next);
    MercuryApp.SetSetting(KEYS.autostart, next ? "true" : "false");
  };

  const handleToggleAutoAccept = () => {
    const next = !autoAccept;
    setAutoAccept(next);
    MercuryApp.SetSetting(KEYS.autoAccept, next ? "true" : "false");
  };

  return (
    <>
      {/* Sync */}
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

      {/* Files Config */}
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
        <div className={`toggle-row ${!allowFiles ? "toggle-disabled" : ""}`}>
          <label className="toggle-label">
            <span>Auto-accept files</span>
            <input
              type="checkbox"
              className="toggle"
              checked={autoAccept}
              disabled={!allowFiles}
              onChange={handleToggleAutoAccept}
            />
          </label>
        </div>
      </section>

      {/* Preferences */}
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
    </>
  );
}
