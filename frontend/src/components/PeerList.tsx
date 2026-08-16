import { Devices } from "@phosphor-icons/react";
import { copy } from "../copy";
import type { Peer } from "../types/mercury";

interface Props {
  peers: Peer[];
}

function peerLabel(peer: Peer): string {
  return peer.hostname || peer.id || copy.peers.unknown;
}

function peerInitial(name: string): string {
  const ch = name.trim().charAt(0);
  return ch ? ch.toUpperCase() : "?";
}

function peerIP(addr: string): string {
  return addr?.split(":")[0] ?? "";
}

export default function PeerList({ peers }: Props) {
  return (
    <div>
      <div className="section-label">{copy.peers.section}</div>
      <div className="group">
        {peers.length === 0 ? (
          <div className="empty-state">
            <Devices size={24} className="empty-state-icon" aria-hidden />
            <p>{copy.peers.empty}</p>
            <p className="hint">{copy.peers.emptyHint}</p>
          </div>
        ) : (
          <ul className="peer-list">
            {peers.map((peer) => {
              const name = peerLabel(peer);
              return (
                <li key={peer.id} className="peer-item peer-item-enter">
                  <span className="peer-avatar" aria-hidden>
                    {peerInitial(name)}
                  </span>
                  <span className="peer-name">{name}</span>
                  <span className="peer-addr">{peerIP(peer.addr)}</span>
                </li>
              );
            })}
          </ul>
        )}
      </div>
    </div>
  );
}
