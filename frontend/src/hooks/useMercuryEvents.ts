import { useEffect } from "react";
import { Events } from "@wailsio/runtime";

export function useMercuryEvents(
    onChange: () => void,
    onActivity?: () => void,
) {
    useEffect(() => {
        const off = Events.On("dashboard:changed", () => {
            onActivity?.();
            onChange();
        });
        return off;
    }, [onChange, onActivity]);
}
