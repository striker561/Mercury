import { useState, useEffect, useCallback } from "react";
import { Gear } from "@phosphor-icons/react";
import { MercuryApp } from "../bindings/mercury/app";
import Settings from "./components/Settings";
import FileTransfer from "./components/FileTransfer";
import Welcome from "./components/Welcome";

function App() {
  const [peers, setPeers] = useState<any[]>([]);
  const [showSettings, setShowSettings] = useState(false);
  const [paused, setPaused] = useState(false);

  const load = useCallback(() => {
    MercuryApp.GetPeers().then((p: any) => { if (p) setPeers(p); });
    MercuryApp.IsPaused().then((p: any) => { if (typeof p === "boolean") setPaused(p); });
  }, []);

  useEffect(() => { load(); const id = setInterval(load, 5000); return () => clearInterval(id); }, [load]);

  const dotClass = paused ? "dot yellow" : peers.length > 0 ? "dot green pulse" : "dot gray";
  const label = paused ? "Paused" : peers.length > 0 ? `${peers.length} peer${peers.length === 1 ? "" : "s"}` : "No peers";

  return (
    <>
      <div className="header">
        <img src="/mercury-logo.png" alt="" className="header-logo" />
        <span className="header-title">Mercury</span>
        <div className="header-status">
          <span className={dotClass} />
          <span>{label}</span>
        </div>
        <button className="gear" onClick={() => setShowSettings(!showSettings)}>
          <Gear size={16} weight={showSettings ? "fill" : "regular"} />
        </button>
      </div>
      <div className="content">
        {showSettings ? <Settings /> : (
          <>
            <Welcome />
            <FileTransfer />
            <div>
              <div className="section-label">Peers</div>
              <div className="card card-slim">
                {peers.length === 0 ? (
                  <div className="empty"><p>No peers on network</p></div>
                ) : (
                  <ul className="pl">
                    {peers.map((p: any, i: number) => (
                      <li key={i}>
                        <span className="dot green pulse" />
                        <span className="pn">{p.hostname || p.id || "unknown"}</span>
                        <span className="pts">{p.addr?.split(":")[0]}</span>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            </div>
          </>
        )}
      </div>
    </>
  );
}
export default App;
