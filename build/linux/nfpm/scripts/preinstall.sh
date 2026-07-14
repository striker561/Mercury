#!/bin/bash
# Runs before apt/dnf installs Mercury package files.
#
# Runtime libraries (GTK, WebKit) are NOT installed here. They are declared
# in nfpm.yaml `depends:` and pulled in automatically by the package manager.
# Use this script only for pre-install checks (e.g. warn if an old binary is running).
if command -v pgrep >/dev/null 2>&1 && pgrep -x mercury >/dev/null 2>&1; then
  echo "Mercury is running. Quit from the tray before upgrading." >&2
fi
exit 0
