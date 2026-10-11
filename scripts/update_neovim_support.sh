#!/usr/bin/env bash
set -euo pipefail

# update_neovim_support.sh - Build and sync all Gatwd editor support for LazyVim
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NVIM_CONFIG="${HOME}/.config/nvim"

echo "=== Updating Gatwd Support in Neovim / LazyVim ==="

# 1. Compile and install gatwd binary to ~/go/bin
echo "[1/4] Compiling gatwd binary..."
cd "${REPO_DIR}"
go install github.com/joergreinhardt/gatwd/cmd/gatwd
echo "  -> Installed gatwd to $(which gatwd 2>/dev/null || echo "${HOME}/go/bin/gatwd")"

# 2. Run gatwd-gen to regenerate editor configs from spec/spec.go
echo "[2/4] Regenerating editor configurations from spec/spec.go..."
go run ./cmd/gatwd-gen

# 3. Ensure target directories exist in ~/.config/nvim
echo "[3/4] Ensuring Neovim directories exist..."
mkdir -p "${NVIM_CONFIG}/lua/plugins" \
         "${NVIM_CONFIG}/syntax" \
         "${NVIM_CONFIG}/ftdetect" \
         "${NVIM_CONFIG}/snippets"

# 4. Establish / refresh symbolic links
echo "[4/4] Refreshing symbolic links..."
ln -sf "${REPO_DIR}/editors/neovim/lua/plugins/gatwd.lua" "${NVIM_CONFIG}/lua/plugins/gatwd.lua"
ln -sf "${REPO_DIR}/editors/neovim/syntax/gatwd.vim"      "${NVIM_CONFIG}/syntax/gatwd.vim"
ln -sf "${REPO_DIR}/editors/neovim/ftdetect/gatwd.lua"    "${NVIM_CONFIG}/ftdetect/gatwd.lua"
ln -sf "${REPO_DIR}/editors/neovim/snippets/gatwd.json"   "${NVIM_CONFIG}/snippets/gatwd.json"

echo "✅ All Gatwd LazyVim plugins, completions, linters, and snippets are up to date!"
