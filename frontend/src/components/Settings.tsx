import { useState, useEffect, useCallback, useRef } from "react";
import { Eye, EyeSlash } from "@phosphor-icons/react";
import { MercuryApp } from "../../bindings/mercury/app";
import { copy } from "../copy";
import type { AppSettings } from "../types/mercury";

interface Props {
  focusPassphrase?: boolean;
  onPassphraseSaved?: () => void;
}

export default function Settings({
  focusPassphrase,
  onPassphraseSaved,
}: Props) {
  const [pass, setPass] = useState("");
  const [show, setShow] = useState(false);
  const [paused, setPaused] = useState(false);
  const [accept, setAccept] = useState(true);
  const [auto, setAuto] = useState(false);
  const [startup, setStartup] = useState(false);
  const [folder, setFolder] = useState("~/Downloads/Mercury/");
  const [saved, setSaved] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    MercuryApp.GetAllSettings().then((s: AppSettings | null) => {
      if (!s) return;
      setPass(s.passphrase ?? "");
      setPaused(s.paused === "true");
      setAccept(s.allow_files !== "false");
      setStartup(s.autostart === "true");
      setAuto(s.auto_accept === "true");
      setFolder(s.received_folder ?? "~/Downloads/Mercury/");
    });
  }, []);

  useEffect(() => {
    if (focusPassphrase) {
      inputRef.current?.focus();
    }
  }, [focusPassphrase]);

  const save = async () => {
    await MercuryApp.SetPassphrase(pass);
    setSaved(true);
    onPassphraseSaved?.();
    setTimeout(() => setSaved(false), 2000);
  };

  const toggle = (k: string, v: boolean, fn: (x: boolean) => void) => {
    fn(!v);
    MercuryApp.SetSetting(k, !v ? "true" : "false");
  };

  const togglePause = useCallback(async () => {
    const p = await MercuryApp.TogglePause();
    setPaused(!!p);
  }, []);

  const pick = () => {
    MercuryApp.PickReceivedFolder().then((p: string | null) => {
      if (p) setFolder(p);
    });
  };

  return (
    <>
      <section className="settings-section">
        <div className="section-label">{copy.settings.covenant}</div>
        <p className="section-hint">{copy.settings.covenantHint}</p>
        <div className="group">
          <div className="group-row group-row-stack">
            <span className="group-row-label">{copy.settings.passphrase}</span>
            <div className="passphrase-wrap">
              <input
                ref={inputRef}
                type={show ? "text" : "password"}
                value={pass}
                onChange={(e) => setPass(e.target.value)}
                placeholder={copy.settings.passphrasePlaceholder}
                aria-label="Passphrase"
              />
              <button
                type="button"
                onClick={() => setShow(!show)}
                aria-label={show ? "Hide passphrase" : "Show passphrase"}
              >
                {show ? <EyeSlash size={14} /> : <Eye size={14} />}
              </button>
            </div>
          </div>
          <div className="group-row">
            <span className="group-row-label">{copy.settings.syncEnabled}</span>
            <button
              type="button"
              className={`toggle${!paused ? " on" : ""}`}
              onClick={togglePause}
              role="switch"
              aria-checked={!paused}
              aria-label="Sync enabled"
            />
          </div>
          <div className="group-note">{copy.settings.syncEnabledHint}</div>
          <div className="group-row-actions">
            <button
              type="button"
              className="btn btn-primary"
              onClick={save}
              disabled={!pass}
            >
              {copy.settings.save}
            </button>
            <span
              className={`save-feedback${saved ? " visible" : ""}`}
              aria-live="polite"
            >
              {copy.settings.saved}
            </span>
          </div>
        </div>
      </section>

      <section className="settings-section">
        <div className="section-label">{copy.settings.offerings}</div>
        <p className="section-hint">{copy.settings.offeringsHint}</p>
        <div className="group">
          <div className="group-row group-row-stack">
            <span className="group-row-label">{copy.settings.saveTo}</span>
            <div className="group-row-path">
              <span className="group-row-value" title={folder}>
                {folder}
              </span>
              <button
                type="button"
                className="btn btn-ghost btn-sm"
                onClick={pick}
              >
                {copy.settings.change}
              </button>
            </div>
          </div>
          <div className="group-row">
            <span className="group-row-label">{copy.settings.acceptFiles}</span>
            <button
              type="button"
              className={`toggle${accept ? " on" : ""}`}
              onClick={() => toggle("allow_files", accept, setAccept)}
              role="switch"
              aria-checked={accept}
              aria-label="Accept files"
            />
          </div>
          <div className="group-note">{copy.settings.acceptFilesHint}</div>
          <div className="group-row" style={{ opacity: accept ? 1 : 0.45 }}>
            <span className="group-row-label">{copy.settings.autoAccept}</span>
            <button
              type="button"
              className={`toggle${auto ? " on" : ""}`}
              onClick={() => accept && toggle("auto_accept", auto, setAuto)}
              role="switch"
              aria-checked={auto}
              aria-label="Auto-accept files"
              disabled={!accept}
            />
          </div>
          <div className="group-note" style={{ opacity: accept ? 1 : 0.45 }}>
            {copy.settings.autoAcceptHint}
          </div>
        </div>
      </section>

      <section className="settings-section">
        <div className="section-label">{copy.settings.presence}</div>
        <p className="section-hint">{copy.settings.presenceHint}</p>
        <div className="group">
          <div className="group-row">
            <span className="group-row-label">
              {copy.settings.startOnLogin}
            </span>
            <button
              type="button"
              className={`toggle${startup ? " on" : ""}`}
              onClick={() => toggle("autostart", startup, setStartup)}
              role="switch"
              aria-checked={startup}
              aria-label="Start on login"
            />
          </div>
        </div>
      </section>
    </>
  );
}
