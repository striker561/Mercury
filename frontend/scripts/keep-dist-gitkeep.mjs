import { writeFileSync } from "node:fs";

// Vite empties dist/, which would remove the placeholder the Go embed needs.
writeFileSync(
  "dist/.gitkeep",
  "# DO NOT DELETE — placeholder so main.go's `//go:embed all:frontend/dist`\n" +
    "# matches in a fresh checkout. CI runs `go vet ./...` before the frontend\n" +
    "# build populates this directory; without this file the embed fails with\n" +
    '# "pattern all:frontend/dist: no matching files found".\n' +
    "# (.gitignore explicitly un-ignores this file: `!frontend/dist/.gitkeep`)\n",
);
