#!/usr/bin/env bash
set -euo pipefail

# Home is ephemeral — reinstate symlinks onto the persistent /config volume.
VSCODE_HOME=/config/.vscode_home
mkdir -p "$VSCODE_HOME/.config" "$HOME/.config"
ln -sfn "$VSCODE_HOME/.cursor" "$HOME/.cursor"
ln -sfn "$VSCODE_HOME/.codex" "$HOME/.codex"
ln -sfn "$VSCODE_HOME/.config/cursor" "$HOME/.config/cursor"
ln -sfn "$VSCODE_HOME/.zshrc" "$HOME/.zshrc"

apt-get update && pip install setuptools && pip3 install hass-deps
curl https://cursor.com/install -fsS | bash
curl -fsSL https://chatgpt.com/codex/install.sh | sh
