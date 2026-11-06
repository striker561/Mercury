import { copy } from "../copy";
import Bolt from "./Bolt";
import type { Peer } from "../types/mercury";

interface Props {
  peers: Peer[];
  paused?: boolean;
}

function peerLabel(peer: Peer): string {
  return peer.hostname || peer.id || copy.peers.unknown;
}

function peerIP(addr: string): string {
  return addr?.split(":")[0] ?? "";
}

export default function PeerList({ peers, paused }: Props) {
  return (
    <div>
      <div className="section-label">{copy.peers.section}</div>
      <div className="list">
        <ul className="peer-list">
          {peers.map((peer) => {
            const name = peerLabel(peer);
            return (
              <li key={peer.id} className="peer-item">
                <span className="peer-glyph">
                  <Bolt
                    state={!paused && peer.reachable ? "live" : "idle"}
                    height={12}
                    label={
                      peer.reachable
                        ? copy.peers.reachable
                        : copy.peers.checking
                    }
                  />
                </span>
                <span className="peer-name" title={name}>
                  {name}
                </span>
                <span className="peer-addr" title={peerIP(peer.addr)}>
                  {peerIP(peer.addr)}
                </span>
              </li>
            );
          })}
        </ul>
      </div>
    </div>
  );
}
