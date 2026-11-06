import { useState, useCallback, useRef } from "react";

const STRIKE_MS = 340;

export function useSyncActivity(peerCount: number, paused: boolean) {
  const [strikes, setStrikes] = useState(0);
  const lastStrike = useRef(0);

  const pulse = useCallback(() => {
    if (peerCount === 0 || paused) return;
    const now = Date.now();
    if (now - lastStrike.current < STRIKE_MS) return;
    lastStrike.current = now;
    setStrikes((n) => n + 1);
  }, [peerCount, paused]);

  return { strikes, pulse };
}
