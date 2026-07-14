import { useState, useEffect, useCallback } from "react";
import { MercuryApp } from "../../bindings/mercury/app";

function sz(b: number): string {
  if (b < 1024) return b + " B";
  if (b < 1048576) return (b / 1024).toFixed(0) + " KB";
  return (b / 1048576).toFixed(1) + " MB";
}
function spd(bps: number): string {
  if (bps <= 0) return "";
  return bps < 1048576 ? (bps / 1024).toFixed(0) + " KB/s" : (bps / 1048576).toFixed(1) + " MB/s";
}

export default function FileTransfer() {
  const [offers, setOffers] = useState<any[]>([]);
  const [xfers, setXfers] = useState<any[]>([]);

  useEffect(() => {
    const load = () => {
      MercuryApp.GetPendingFileOffers().then((o: any) => { if (o) setOffers(o); });
      MercuryApp.GetTransferProgress().then((t: any) => { if (t) setXfers(t.filter((p: any) => p.status !== "done")); });
    };
    load();
    const id = setInterval(load, 1000);
    return () => clearInterval(id);
  }, []);

  const accept = useCallback((id: string) => { MercuryApp.AcceptFileOffer(id); setOffers(p => p.filter((o: any) => o.id !== id)); }, []);
  const reject = useCallback((id: string) => { MercuryApp.RejectFileOffer(id); setOffers(p => p.filter((o: any) => o.id !== id)); }, []);
  const cancel = useCallback((id: string) => { MercuryApp.CancelTransfer(id); setXfers(p => p.map((t: any) => t.id === id ? { ...t, status: "cancelled" } : t)); }, []);

  if (!offers.length && !xfers.length) return null;

  return (
    <div>
      <div className="section-label">Transfers</div>
      {offers.map((o: any) => (
        <div key={o.id} className="tc">
          <div className="tfn">{o.file_name}</div>
          <div className="tm">{sz(o.file_size)}</div>
          <div className="ta">
            <button className="btn btn-outline btn-sm" onClick={() => reject(o.id)}>Decline</button>
            <button className="btn btn-primary btn-sm" onClick={() => accept(o.id)}>Accept</button>
          </div>
        </div>
      ))}
      {xfers.map((t: any) => {
        const pct = t.file_size > 0 ? Math.min(100, (t.received / t.file_size) * 100) : 0;
        const busy = t.status === "sending" || t.status === "receiving";
        return (
          <div key={t.id} className="tc">
            <div className="tfn">{t.file_name}</div>
            {busy && <div className="pt"><div className={`pf ${t.status === "failed" || t.status === "cancelled" ? "red" : ""}`} style={{ width: `${pct}%` }} /></div>}
            <div className="pf2">
              <span className="pi">{sz(t.received)} / {sz(t.file_size)}{t.speed > 0 ? ` · ${spd(t.speed)}` : ""}</span>
              {busy && <button className="btn btn-ghost btn-sm" onClick={() => cancel(t.id)}>Cancel</button>}
            </div>
          </div>
        );
      })}
    </div>
  );
}
