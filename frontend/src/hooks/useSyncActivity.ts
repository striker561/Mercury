import { useState, useCallback, useRef, useEffect } from "react";

export function useSyncActivity(peerCount: number, paused: boolean) {
  const [active, setActive] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout>>();

  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, []);

  const pulse = useCallback(() => {
    if (peerCount > 0 && !paused) {
      setActive(true);
      if (timerRef.current) clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => setActive(false), 800);
    }
  }, [peerCount, paused]);

  return { active, pulse };
}
