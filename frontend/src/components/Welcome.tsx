import { Shield } from "@phosphor-icons/react";

interface Props {
  onGetStarted: () => void;
}

export default function Welcome({ onGetStarted }: Props) {
  return (
    <div className="welcome">
      <img
        src="/mercury-logo.png"
        alt=""
        className="welcome-logo"
        aria-hidden
      />
      <h2 className="welcome-title">Welcome to Mercury</h2>
      <p className="welcome-lead">
        Copy on one machine. Paste on another. Encrypted on your LAN.
      </p>
      <div className="welcome-steps">
        <div className="welcome-step">
          <span className="welcome-step-num">1</span>
          <span>Set a passphrase shared with your devices</span>
        </div>
        <div className="welcome-step">
          <span className="welcome-step-num">2</span>
          <span>Run Mercury on each machine on the same network</span>
        </div>
        <div className="welcome-step">
          <span className="welcome-step-num">3</span>
          <span>Copy anything. It appears on your other machines.</span>
        </div>
      </div>
      <div className="welcome-trust">
        <Shield size={13} weight="fill" aria-hidden />
        End-to-end encrypted
      </div>
      <button
        type="button"
        className="btn btn-primary btn-block"
        onClick={onGetStarted}
      >
        Set up passphrase
      </button>
    </div>
  );
}
