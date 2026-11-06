import { useCallback } from "react";
import { Check, File, Image as ImageIcon, X } from "@phosphor-icons/react";
import { MercuryApp } from "../../bindings/mercury/app";
import { copy } from "../copy";
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

function FileGlyph({ name }: { name: string }) {
  const isImage = /\.(png|jpe?g|gif|webp|bmp|svg|heic)$/i.test(name);
  const Icon = isImage ? ImageIcon : File;
  return <Icon size={15} aria-hidden />;
}

/** 0–100. A ring, so it reads the same on every platform. */
function Ring({ pct, failed }: { pct: number | null; failed?: boolean }) {
  const r = 15;
  const c = 2 * Math.PI * r;

  return (
    <svg
      className="ring"
      width="34"
      height="34"
      viewBox="0 0 34 34"
      aria-hidden
    >
      <circle className="ring-track" cx="17" cy="17" r={r} />
      {pct !== null && (
        <circle
          className={`ring-fill${failed ? " failed" : ""}`}
          cx="17"
          cy="17"
          r={r}
          strokeDasharray={`${(c * Math.min(100, Math.max(0, pct))) / 100} ${c}`}
        />
      )}
    </svg>
  );
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

  const dismiss = useCallback(
    async (id: string) => {
      await MercuryApp.DismissTransfer(id);
      onChange();
    },
    [onChange],
  );

  if (!offers.length && !transfers.length) return null;

  return (
    <div>
      <div className="section-label">{copy.transfers.section}</div>
      {offers.map((offer) => (
        <div key={offer.id} className="transfer-card">
          <span className="transfer-ring">
            <Ring pct={null} />
            <span className="transfer-glyph">
              <FileGlyph name={offer.file_name} />
            </span>
          </span>
          <div className="transfer-info">
            <div className="transfer-name" title={offer.file_name}>
              {offer.file_name}
            </div>
            <div className="transfer-meta">
              {formatSize(offer.file_size)} · {copy.transfers.incoming}
            </div>
          </div>
          <div className="transfer-end">
            <button
              type="button"
              className="icon-btn"
              onClick={() => reject(offer.id)}
              aria-label={copy.transfers.decline}
              title={copy.transfers.decline}
            >
              <X size={14} weight="bold" />
            </button>
            <button
              type="button"
              className="icon-btn icon-btn-primary"
              onClick={() => accept(offer.id)}
              aria-label={copy.transfers.accept}
              title={copy.transfers.accept}
            >
              <Check size={14} weight="bold" />
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
          <div
            key={t.id}
            className={`transfer-card${busy ? " transfer-card-active" : ""}`}
          >
            <span className="transfer-ring">
              <Ring pct={busy ? pct : null} failed={failed} />
              <span className="transfer-glyph">
                <FileGlyph name={t.file_name} />
              </span>
            </span>
            <div className="transfer-info">
              <div className="transfer-name" title={t.file_name}>
                {t.file_name}
              </div>
              <div className="transfer-meta">
                {formatSize(t.received)} / {formatSize(t.file_size)}
                {t.speed > 0 ? ` · ${formatSpeed(t.speed)}` : ""}
                {failed && t.error ? ` · ${t.error}` : ""}
              </div>
            </div>
            <div className="transfer-end">
              {busy && <span className="transfer-pct">{Math.round(pct)}%</span>}
              {busy ? (
                <button
                  type="button"
                  className="icon-btn"
                  onClick={() => cancel(t.id)}
                  aria-label={copy.transfers.cancel}
                  title={copy.transfers.cancel}
                >
                  <X size={14} weight="bold" />
                </button>
              ) : (
                failed && (
                  <button
                    type="button"
                    className="icon-btn"
                    onClick={() => dismiss(t.id)}
                    aria-label={copy.transfers.dismiss}
                    title={copy.transfers.dismiss}
                  >
                    <X size={14} weight="bold" />
                  </button>
                )
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}
