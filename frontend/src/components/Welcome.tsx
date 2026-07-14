import { useState, useEffect } from "react";
import { Shield } from "@phosphor-icons/react";
import { MercuryApp } from "../../bindings/mercury/app";

export default function Welcome() {
  const [show, setShow] = useState(false);
  useEffect(() => { MercuryApp.GetSavedPassphrase().then((p: any) => setShow(!p)); }, []);

  if (!show) return null;

  return (
    <div className="wc">
      <div className="wt">Welcome to Mercury</div>
      <p className="wp">Everything you copy or send is encrypted. No one on the network can read it without your passphrase.</p>
      <div className="ws">
        <div className="wsi"><span className="wsn">1</span><span>Open Settings (gear icon)</span></div>
        <div className="wsi"><span className="wsn">2</span><span>Set a passphrase shared with your devices</span></div>
        <div className="wsi"><span className="wsn">3</span><span>Copy anything. It appears on your other machines.</span></div>
      </div>
      <div style={{ display: "flex", gap: 6, alignItems: "center", justifyContent: "center", fontSize: 11, color: "var(--text3)", marginBottom: 10 }}>
        <Shield size={12} /> End-to-end encrypted
      </div>
      <button className="btn btn-primary" onClick={() => (document.querySelector(".gear") as HTMLElement)?.click()}>Get Started</button>
    </div>
  );
}
