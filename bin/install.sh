#!/usr/bin/env bash
# claude-dev-team bootstrap (Unix / macOS)
# Ensures uv is installed, then runs the Python installer via `uv run`.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "claude-dev-team bootstrap"

if ! command -v uv >/dev/null 2>&1; then
    echo "uv not found. Installing via astral.sh..."
    curl -LsSf https://astral.sh/uv/install.sh | sh

    # Make uv available in the current shell session
    if [ -f "$HOME/.local/bin/env" ]; then
        # shellcheck disable=SC1091
        . "$HOME/.local/bin/env"
    elif [ -f "$HOME/.cargo/env" ]; then
        # shellcheck disable=SC1091
        . "$HOME/.cargo/env"
    else
        export PATH="$HOME/.local/bin:$PATH"
    fi
fi

if ! command -v uv >/dev/null 2>&1; then
    echo "Error: uv install did not succeed." >&2
    echo "Install it manually (https://docs.astral.sh/uv/) and re-run this script." >&2
    exit 1
fi

echo "uv: $(uv --version)"
echo "Running installer..."
echo

exec uv run "$REPO_ROOT/bin/install.py" "$@"
