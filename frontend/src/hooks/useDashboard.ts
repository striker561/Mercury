import { useState, useEffect, useCallback } from "react";
import { MercuryApp } from "../../bindings/mercury/app";
import type { DashboardState as RawDashboardState } from "../../bindings/mercury/app/models";
import type { DashboardState } from "../types/mercury";

const empty: DashboardState = {
  peers: [],
  paused: false,
  hasPassphrase: false,
  offers: [],
  transfers: [],
};

function normalize(raw: RawDashboardState): DashboardState {
  const peers = (raw.peers ?? [])
    .filter((p): p is Record<string, string> => p != null)
    .map((p) => ({
      id: p.id ?? "",
      hostname: p.hostname ?? p.id ?? "",
      addr: p.addr ?? "",
      lastSeen: p.lastSeen,
    }));

  return {
    peers,
    paused: raw.paused,
    hasPassphrase: raw.hasPassphrase,
    offers: raw.offers ?? [],
    transfers: raw.transfers ?? [],
  };
}

export function useDashboard() {
  const [state, setState] = useState<DashboardState>(empty);

  const refresh = useCallback(async () => {
    try {
      const data = await MercuryApp.GetDashboardState();
      if (data) setState(normalize(data));
    } catch {
      // Bindings unavailable outside Wails runtime.
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return { state, refresh };
}
