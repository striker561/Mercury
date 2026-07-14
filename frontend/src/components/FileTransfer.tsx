import { useState, useEffect, useCallback } from "react";
import { MercuryApp } from "../../bindings/mercury/app";

interface FileOffer {
  id: string;
  file_name: string;
  file_size: number;
  peer_addr: string;
}

interface TransferProgress {
  id: string;
  file_name: string;
  file_size: number;
  received: number;
  status: string;
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + " B";
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
  return (bytes / (1024 * 1024)).toFixed(1) + " MB";
}

export default function FileTransfer() {
  const [offers, setOffers] = useState<FileOffer[]>([]);
  const [transfers, setTransfers] = useState<TransferProgress[]>([]);

  useEffect(() => {
    load();
    const id = setInterval(load, 1000);
    return () => clearInterval(id);
  }, []);

  const load = () => {
    MercuryApp.GetPendingFileOffers().then((o: any) => {
      if (o) setOffers(o);
    });
    MercuryApp.GetTransferProgress().then((t: any) => {
      if (t) setTransfers(t);
    });
  };

  const accept = useCallback((id: string) => {
    MercuryApp.AcceptFileOffer(id);
    setOffers((prev) => prev.filter((o) => o.id !== id));
  }, []);

  const reject = useCallback((id: string) => {
    MercuryApp.RejectFileOffer(id);
    setOffers((prev) => prev.filter((o) => o.id !== id));
  }, []);

  return (
    <section className="settings-section">
      <h2>Transfers</h2>

      {offers.length === 0 && transfers.length === 0 && (
        <p className="settings-empty">No active transfers</p>
      )}

      {offers.length > 0 && (
        <div className="offers-list">
          {offers.map((o) => (
            <div key={o.id} className="offer-card">
              <div className="offer-info">
                <span className="offer-name">{o.file_name}</span>
                <span className="offer-size">{formatSize(o.file_size)}</span>
              </div>
              <div className="offer-actions">
                <button className="btn-primary btn-sm" onClick={() => accept(o.id)}>
                  Accept
                </button>
                <button className="btn-secondary btn-sm" onClick={() => reject(o.id)}>
                  Decline
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {transfers.length > 0 && (
        <div className="transfers-list">
          {transfers.map((t) => (
            <div key={t.id} className="transfer-card">
              <div className="transfer-header">
                <span className="transfer-name">{t.file_name}</span>
                <span className="transfer-status">{t.status}</span>
              </div>
              <div className="transfer-bar-track">
                <div
                  className="transfer-bar-fill"
                  style={{
                    width:
                      t.file_size > 0
                        ? `${Math.min(100, (t.received / t.file_size) * 100)}%`
                        : "0%",
                  }}
                />
              </div>
              <span className="transfer-detail">
                {formatSize(t.received)} / {formatSize(t.file_size)}
              </span>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
