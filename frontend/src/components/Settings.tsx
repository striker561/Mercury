import { useState, useEffect } from "react";
import { Eye, EyeSlash } from "@phosphor-icons/react";
import { MercuryApp } from "../../bindings/mercury/app";

export default function Settings() {
  const [pass, setPass] = useState("");
  const [show, setShow] = useState(false);
  const [paused, setPaused] = useState(false);
  const [accept, setAccept] = useState(true);
  const [auto, setAuto] = useState(false);
  const [startup, setStartup] = useState(false);
  const [folder, setFolder] = useState("~/Downloads/Mercury/");

  useEffect(() => {
    MercuryApp.GetAllSettings().then((s: any) => {
      if (!s) return;
      setPass(s.passphrase ?? "");
      setPaused(s.paused === "true");
      setAccept(s.allow_files !== "false");
      setStartup(s.autostart === "true");
      setAuto(s.auto_accept === "true");
      setFolder(s.received_folder ?? "~/Downloads/Mercury/");
    });
  }, []);

  const save = () => MercuryApp.SetPassphrase(pass);
  const toggle = (k: string, v: boolean, fn: (x: boolean) => void) => {
    fn(!v); MercuryApp.SetSetting(k, !v ? "true" : "false");
  };
  const pick = () => MercuryApp.PickReceivedFolder().then((p: any) => { if (p) setFolder(p); });

  return (
    <>
      <div className="section-label">Sync</div>
      <div className="card">
        <div className="row">
          <span className="row-label">Passphrase</span>
          <div className="pw-wrap">
            <input type={show ? "text" : "password"} value={pass}
              onChange={e => setPass(e.target.value)} placeholder="passphrase" />
            <button onClick={() => setShow(!show)}>{show ? <EyeSlash size={14} /> : <Eye size={14} />}</button>
          </div>
        </div>
        <div className="row">
          <span className="row-label">Sync</span>
          <button className={`tog ${!paused ? "on" : ""}`} onClick={() => MercuryApp.TogglePause().then(setPaused)} aria-label="Sync toggle" />
        </div>
        <div className="row" style={{ marginTop: 8, border: "none", padding: 0 }}>
          <button className="btn btn-primary" onClick={save} disabled={!pass}>Save</button>
        </div>
      </div>

      <div className="section-label" style={{ marginTop: 4 }}>Files</div>
      <div className="card">
        <div className="row">
          <span className="row-label">Save to</span>
          <span className="row-value" style={{ flex: 1, maxWidth: 140 }}>{folder}</span>
          <button className="btn btn-ghost btn-sm" onClick={pick}>Change</button>
        </div>
        <div className="row">
          <span className="row-label">Accept</span>
          <button className={`tog ${accept ? "on" : ""}`} onClick={() => toggle("allow_files", accept, setAccept)} aria-label="Toggle accept" />
        </div>
        <div className="row" style={{ opacity: accept ? 1 : 0.4 }}>
          <span className="row-label">Auto-accept</span>
          <button className={`tog ${auto ? "on" : ""}`} onClick={() => accept && toggle("auto_accept", auto, setAuto)} aria-label="Toggle auto-accept" />
        </div>
      </div>

      <div className="section-label" style={{ marginTop: 4 }}>Preferences</div>
      <div className="card">
        <div className="row">
          <span className="row-label">Start on login</span>
          <button className={`tog ${startup ? "on" : ""}`} onClick={() => toggle("autostart", startup, setStartup)} aria-label="Toggle autostart" />
        </div>
      </div>
    </>
  );
}
