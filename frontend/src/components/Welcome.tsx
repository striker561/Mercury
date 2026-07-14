import { Shield } from "@phosphor-icons/react";
import { copy } from "../copy";

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
      <h2 className="welcome-title">{copy.welcome.title}</h2>
      <p className="welcome-lead">{copy.welcome.lead}</p>
      <div className="welcome-steps">
        {copy.welcome.steps.map((step, i) => (
          <div key={step} className="welcome-step">
            <span className="welcome-step-num">{i + 1}</span>
            <span>{step}</span>
          </div>
        ))}
      </div>
      <div className="welcome-trust">
        <Shield size={13} weight="fill" aria-hidden />
        {copy.welcome.trust}
      </div>
      <button
        type="button"
        className="btn btn-primary btn-block"
        onClick={onGetStarted}
      >
        {copy.welcome.cta}
      </button>
    </div>
  );
}
