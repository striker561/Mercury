import { useEffect } from "react";
import { Events } from "@wailsio/runtime";

export function useMercuryEvents(
    onChange: () => void,
    onActivity?: () => void,
    onFileOffer?: () => void,
) {
    useEffect(() => {
        const offChange = Events.On("dashboard:changed", () => {
            onActivity?.();
            onChange();
        });
        const offOffer = onFileOffer
            ? Events.On("dashboard:file-offer", onFileOffer)
            : undefined;
        return () => {
            offChange();
            offOffer?.();
        };
    }, [onChange, onActivity, onFileOffer]);
}
