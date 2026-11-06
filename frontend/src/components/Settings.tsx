import { useState, useEffect, useCallback, useRef } from "react";
import {
  ArrowsClockwise,
  Eye,
  EyeSlash,
  FolderSimple,
  Power,
} from "@phosphor-icons/react";
import { Browser } from "@wailsio/runtime";
import { MercuryApp } from "../../bindings/mercury/app";
import { copy } from "../copy";
import type { AppSettings } from "../types/mercury";

interface Props {
  focusPassphrase?: boolean;
  onPassphraseSaved?: () => void;
}

const REPO = "https://github.com/striker561/Mercury";

const LINKS = [
  { label: copy.settings.source, href: REPO },
  { label: copy.settings.issues, href: `${REPO}/issues` },
  { label: copy.settings.docs, href: "https://striker561.github.io/Mercury" },
];

function openLink(url: string) {
  Promise.resolve()
    .then(() => Browser.OpenURL(url))
    .catch(() => window.open(url, "_blank", "noopener"));
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
  const [savedPass, setSavedPass] = useState("");
  const [saveError, setSaveError] = useState("");
  const [version, setVersion] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    MercuryApp.GetVersion()
      .then((v: string) => setVersion(v))
      .catch(() => setVersion(""));
  }, []);

  useEffect(() => {
    MercuryApp.GetAllSettings().then((s: AppSettings | null) => {
      if (!s) return;
      setPass(s.passphrase ?? "");
      setSavedPass(s.passphrase ?? "");
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

  const persist = useCallback(async (): Promise<boolean> => {
    const value = pass.trim();
    if (!value) {
      setSaveError(copy.settings.passphraseRequired);
      return false;
    }
    setSaveError("");
    await MercuryApp.SetPassphrase(value);
    const state = await MercuryApp.GetDashboardState();
    if (state?.syncError) {
      setSaveError(state.syncError);
      return false;
    }
    setPass(value);
    setSavedPass(value);
    onPassphraseSaved?.();
    return true;
  }, [pass, onPassphraseSaved]);

  const toggle = (k: string, v: boolean, fn: (x: boolean) => void) => {
    fn(!v);
    MercuryApp.SetSetting(k, !v ? "true" : "false");
  };

  const togglePause = useCallback(async () => {
    if (pass !== savedPass) {
      const ok = await persist();
      if (!ok) return;
    }
    const p = await MercuryApp.TogglePause();
    setPaused(!!p);
  }, [pass, savedPass, persist]);

  const pick = () => {
    MercuryApp.PickReceivedFolder().then((p: string | null) => {
      if (p) setFolder(p);
    });
  };

  return (
    <>
      <header className="settings-head">
        <h1 className="settings-title">{copy.settings.pageTitle}</h1>
        <p className="settings-lead">{copy.settings.pageLead}</p>
      </header>

      <section className="settings-section">
        <div className="section-label">
          <ArrowsClockwise size={12} weight="bold" aria-hidden />
          {copy.settings.covenant}
        </div>
        <div className="group">
          <div className="group-row group-row-stack">
            <span className="group-row-label">{copy.settings.passphrase}</span>
            <div className="passphrase-wrap">
              <input
                ref={inputRef}
                type={show ? "text" : "password"}
                value={pass}
                onChange={(e) => {
                  setPass(e.target.value);
                  setSaveError("");
                }}
                onBlur={() => {
                  if (pass !== savedPass) void persist();
                }}
                onKeyDown={(e) => {
                  if (e.key === "Enter") e.currentTarget.blur();
                }}
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
          <div className="group-note">{copy.settings.covenantHint}</div>
          {saveError && <p className="group-note save-error">{saveError}</p>}
          <div className="group-row">
            <span className="group-row-label">{copy.settings.syncEnabled}</span>
            <button
              type="button"
              className={`toggle${!paused ? " on" : ""}`}
              onClick={togglePause}
              role="switch"
              aria-checked={!paused}
              aria-label="Sync enabled"
              disabled={!pass.trim()}
            />
          </div>
          <div className="group-note">{copy.settings.syncEnabledHint}</div>
          {!pass.trim() && (
            <div className="group-note">{copy.settings.passphraseRequired}</div>
          )}
        </div>
      </section>

      <section className="settings-section">
        <div className="section-label">
          <FolderSimple size={12} weight="bold" aria-hidden />
          {copy.settings.offerings}
        </div>
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
          <div className="group-note">{copy.settings.offeringsHint}</div>
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
          <div
            className={`group-row${accept ? "" : " group-row-disabled"}`}
            data-disabled={!accept}
          >
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
          <div
            className={`group-note${accept ? "" : " group-note-disabled"}`}
            data-disabled={!accept}
          >
            {copy.settings.autoAcceptHint}
          </div>
        </div>
      </section>

      <section className="settings-section">
        <div className="section-label">
          <Power size={12} weight="bold" aria-hidden />
          {copy.settings.presence}
        </div>
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

      <footer className="settings-about">
        <div className="settings-about-label">{copy.settings.about}</div>
        <div className="about-links">
          {LINKS.map((link) => (
            <button
              key={link.href}
              type="button"
              className="link-btn"
              onClick={() => openLink(link.href)}
            >
              {link.label}
            </button>
          ))}
        </div>
        <div className="about-meta">
          {version ? `Mercury ${version}` : "Mercury"} · {copy.settings.license}
        </div>
      </footer>
    </>
  );
}
