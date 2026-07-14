import { useState, useEffect } from "react";
import { MercuryApp } from "../bindings/mercury/app";
import Settings from "./components/Settings";
import FileTransfer from "./components/FileTransfer";

interface Peer {
  id: string;
  addr: string;
  hostname?: string;
}

function App() {
  const [peers, setPeers] = useState<Peer[]>([]);
  const [version, setVersion] = useState("");
  const [showSettings, setShowSettings] = useState(false);

  useEffect(() => {
    MercuryApp.GetAllSettings().then((s: any) => {
      if (s) setVersion(s["version"] ?? "");
    });
    loadPeers();
    const interval = setInterval(loadPeers, 5000);
    return () => clearInterval(interval);
  }, []);

  const loadPeers = () => {
    MercuryApp.GetPeers().then((p: any) => {
      if (p) setPeers(p);
    });
  };

  return (
    <div className="settings">
      <div className="settings-header">
        <h1>Mercury</h1>
        <span className="settings-version">v{version}</span>
        <button
          className="btn-gear"
          onClick={() => setShowSettings(!showSettings)}
          title={showSettings ? "Back" : "Settings"}
        >
          {showSettings ? (
            <svg
              width="16"
              height="16"
              viewBox="0 0 16 16"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
            >
              <path d="M10 2L4 8l6 6" />
            </svg>
          ) : (
            <svg
              width="16"
              height="16"
              viewBox="0 0 16 16"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
            >
              <circle cx="8" cy="8" r="2.5" />
              <path d="M8 1.5v2M8 12.5v2M1.5 8h2M12.5 8h2M3.4 3.4l1.4 1.4M11.2 11.2l1.4 1.4M3.4 12.6l1.4-1.4M11.2 4.8l1.4-1.4" />
            </svg>
          )}
        </button>
      </div>

      {showSettings ? (
        /* Full-page settings */
        <Settings />
      ) : (
        <>
          {/* Transfers — main view */}
          <FileTransfer />

          {/* Peers */}
          <section className="settings-section">
            <h2>
              Peers <span className="peer-count">{peers.length}</span>
            </h2>
            {peers.length === 0 ? (
              <p className="settings-empty">No peers on network</p>
            ) : (
              <ul className="peer-list">
                {peers.map((peer, i) => (
                  <li key={i} className="peer-item">
                    <span className="peer-dot" />
                    <span className="peer-name">
                      {peer.hostname || peer.id || "unknown"}
                    </span>
                    <span className="peer-addr">{peer.addr?.split(":")[0]}</span>
                  </li>
                ))}
              </ul>
            )}
          </section>
        </>
      )}
    </div>
  );
}

export default App;
