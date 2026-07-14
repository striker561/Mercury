import { useEffect } from "react";
import { Events } from "@wailsio/runtime";

export function useMercuryEvents(onChange: () => void) {
  useEffect(() => {
    const off = Events.On("dashboard:changed", () => {
      onChange();
    });
    return off;
  }, [onChange]);
}
