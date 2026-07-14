import { useCallback } from "react";
import { File, Image as ImageIcon } from "@phosphor-icons/react";
import { MercuryApp } from "../../bindings/mercury/app";
import type { FileOffer, FileProgress } from "../types/mercury";

interface Props {
  offers: FileOffer[];
  transfers: FileProgress[];
  onChange: () => void;
}

function formatSize(b: number): string {
  if (b < 1024) return `${b} B`;
  if (b < 1048576) return `${Math.round(b / 1024)} KB`;
  return `${(b / 1048576).toFixed(1)} MB`;
}

function formatSpeed(bps: number): string {
  if (bps <= 0) return "";
  return bps < 1048576
    ? `${Math.round(bps / 1024)} KB/s`
    : `${(bps / 1048576).toFixed(1)} MB/s`;
}

function fileIcon(name: string) {
  const lower = name.toLowerCase();
  if (/\.(png|jpe?g|gif|webp|bmp|svg|heic)$/.test(lower)) {
    return <ImageIcon size={18} className="transfer-icon" aria-hidden />;
  }
  return <File size={18} className="transfer-icon" aria-hidden />;
}

export default function FileTransfer({ offers, transfers, onChange }: Props) {
  const accept = useCallback(
    async (id: string) => {
      await MercuryApp.AcceptFileOffer(id);
      onChange();
    },
    [onChange],
  );

  const reject = useCallback(
    async (id: string) => {
      await MercuryApp.RejectFileOffer(id);
      onChange();
    },
    [onChange],
  );

  const cancel = useCallback(
    async (id: string) => {
      await MercuryApp.CancelTransfer(id);
      onChange();
    },
    [onChange],
  );

  if (!offers.length && !transfers.length) return null;

  return (
    <div>
      <div className="section-label">Transfers</div>
      {offers.map((offer) => (
        <div key={offer.id} className="transfer-card">
          <div className="transfer-header">
            {fileIcon(offer.file_name)}
            <div className="transfer-info">
              <div className="transfer-name">{offer.file_name}</div>
              <div className="transfer-meta">
                {formatSize(offer.file_size)} · incoming
              </div>
            </div>
          </div>
          <div className="transfer-actions">
            <button
              type="button"
              className="btn btn-outline btn-sm"
              onClick={() => reject(offer.id)}
            >
              Decline
            </button>
            <button
              type="button"
              className="btn btn-primary btn-sm"
              onClick={() => accept(offer.id)}
            >
              Accept
            </button>
          </div>
        </div>
      ))}
      {transfers.map((t) => {
        const pct =
          t.file_size > 0 ? Math.min(100, (t.received / t.file_size) * 100) : 0;
        const busy = t.status === "sending" || t.status === "receiving";
        const failed = t.status === "failed" || t.status === "cancelled";

        return (
          <div key={t.id} className="transfer-card">
            <div className="transfer-header">
              {fileIcon(t.file_name)}
              <div className="transfer-info">
                <div className="transfer-name">{t.file_name}</div>
                <div className="transfer-meta">
                  {formatSize(t.received)} / {formatSize(t.file_size)}
                  {t.speed > 0 ? ` · ${formatSpeed(t.speed)}` : ""}
                </div>
              </div>
            </div>
            {busy && (
              <div className="progress-track">
                <div
                  className={`progress-fill${failed ? " failed" : ""}`}
                  style={{ width: `${pct}%` }}
                />
              </div>
            )}
            {busy && (
              <div className="progress-footer">
                <span className="progress-info">{Math.round(pct)}%</span>
                <button
                  type="button"
                  className="btn btn-ghost btn-sm"
                  onClick={() => cancel(t.id)}
                >
                  Cancel
                </button>
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}
