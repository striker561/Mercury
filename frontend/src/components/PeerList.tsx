import { Devices } from "@phosphor-icons/react";
import type { Peer } from "../types/mercury";

interface Props {
  peers: Peer[];
}

function peerLabel(peer: Peer): string {
  return peer.hostname || peer.id || "Unknown device";
}

function peerIP(addr: string): string {
  return addr?.split(":")[0] ?? "";
}

export default function PeerList({ peers }: Props) {
  return (
    <div>
      <div className="section-label">Devices</div>
      <div className="group">
        {peers.length === 0 ? (
          <div className="empty-state">
            <Devices size={28} className="empty-state-icon" aria-hidden />
            <p>No devices on your network</p>
            <p className="hint">
              Peers appear automatically when Mercury is running elsewhere.
            </p>
          </div>
        ) : (
          <ul className="peer-list">
            {peers.map((peer) => (
              <li key={peer.id} className="peer-item">
                <span className="status-dot connected pulse" aria-hidden />
                <span className="peer-name">{peerLabel(peer)}</span>
                <span className="peer-addr">{peerIP(peer.addr)}</span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
