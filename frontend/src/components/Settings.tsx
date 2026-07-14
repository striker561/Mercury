import { useState, useEffect, useCallback, useRef } from "react";
import { Eye, EyeSlash } from "@phosphor-icons/react";
import { MercuryApp } from "../../bindings/mercury/app";
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
      <div className="section-label">Sync</div>
      <div className="group">
        <div className="group-row">
          <span className="group-row-label">Passphrase</span>
          <div className="passphrase-wrap">
            <input
              ref={inputRef}
              type={show ? "text" : "password"}
              value={pass}
              onChange={(e) => setPass(e.target.value)}
              placeholder="Shared secret"
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
          <span className="group-row-label">Sync enabled</span>
          <button
            type="button"
            className={`toggle${!paused ? " on" : ""}`}
            onClick={togglePause}
            role="switch"
            aria-checked={!paused}
            aria-label="Sync enabled"
          />
        </div>
        <div className="group-row-actions">
          <button
            type="button"
            className="btn btn-primary"
            onClick={save}
            disabled={!pass}
          >
            Save passphrase
          </button>
          <span
            className={`save-feedback${saved ? " visible" : ""}`}
            aria-live="polite"
          >
            Saved
          </span>
        </div>
      </div>

      <div className="section-label">Files</div>
      <div className="group">
        <div className="group-row">
          <span className="group-row-label">Save to</span>
          <span className="group-row-value" title={folder}>
            {folder}
          </span>
          <button type="button" className="btn btn-ghost btn-sm" onClick={pick}>
            Change
          </button>
        </div>
        <div className="group-row">
          <span className="group-row-label">Accept files</span>
          <button
            type="button"
            className={`toggle${accept ? " on" : ""}`}
            onClick={() => toggle("allow_files", accept, setAccept)}
            role="switch"
            aria-checked={accept}
            aria-label="Accept files"
          />
        </div>
        <div className="group-row" style={{ opacity: accept ? 1 : 0.45 }}>
          <span className="group-row-label">Auto-accept</span>
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
      </div>

      <div className="section-label">General</div>
      <div className="group">
        <div className="group-row">
          <span className="group-row-label">Start on login</span>
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
    </>
  );
}
