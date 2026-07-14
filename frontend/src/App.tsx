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
      </div>

      <Settings />

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

      <FileTransfer />
    </div>
  );
}

export default App;
