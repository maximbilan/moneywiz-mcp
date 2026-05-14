#!/usr/bin/env bash
set -euo pipefail

BIN_NAME="moneywiz-mcp"
SERVER_NAME="${MONEYWIZ_MCP_SERVER_NAME:-moneywiz}"
INSTALL_DIR="${MONEYWIZ_MCP_INSTALL_DIR:-$HOME/.local/bin}"
CLAUDE_SCOPE="${MONEYWIZ_MCP_CLAUDE_SCOPE:-user}"
CLAUDE_DESKTOP_CONFIG_DEFAULT="$HOME/Library/Application Support/Claude/claude_desktop_config.json"

DB_PATH="${MONEYWIZ_DB_PATH:-}"
SKIP_CLAUDE_DESKTOP=0
SKIP_CLAUDE_CODE=0
AUTO_YES=0

log() {
  printf '%s\n' "$*"
}

fail() {
  printf 'Error: %s\n' "$*" >&2
  exit 1
}

usage() {
  cat <<EOF
Usage: $(basename "$0") [options]

Build and install ${BIN_NAME}, then register it with Claude Desktop/Claude Code.

Options:
  --db <path>                Path to MoneyWiz export folder or ipadMoneyWiz.sqlite
  --name <server-name>       MCP server name in config (default: ${SERVER_NAME})
  --install-dir <dir>        Install directory for binary (default: ${INSTALL_DIR})
  --scope <scope>            Claude Code scope: local|user|project (default: ${CLAUDE_SCOPE})
  --skip-claude-desktop      Do not edit Claude Desktop config
  --skip-claude-code         Do not run 'claude mcp add'
  -y, --yes                  Skip confirmation prompts
  -h, --help                 Show this help

Env vars:
  MONEYWIZ_DB_PATH
  MONEYWIZ_MCP_SERVER_NAME
  MONEYWIZ_MCP_INSTALL_DIR
  MONEYWIZ_MCP_CLAUDE_SCOPE
  CLAUDE_DESKTOP_CONFIG
EOF
}

confirm_default_yes() {
  local prompt="$1"
  local input=""

  if [[ "$AUTO_YES" -eq 1 ]]; then
    return 0
  fi

  if [[ -r /dev/tty ]]; then
    printf '%s' "$prompt" > /dev/tty
    IFS= read -r input < /dev/tty || true
  elif [[ -t 0 ]]; then
    printf '%s' "$prompt"
    IFS= read -r input || true
  else
    return 0
  fi

  case "$(printf '%s' "$input" | tr '[:upper:]' '[:lower:]')" in
    ""|y|yes) return 0 ;;
    *) return 1 ;;
  esac
}

resolve_file_path() {
  local path="$1"
  local dir base
  dir="$(cd "$(dirname "$path")" && pwd)"
  base="$(basename "$path")"
  printf '%s/%s\n' "$dir" "$base"
}

normalize_db_path() {
  local value="$1"
  if [[ -z "$value" ]]; then
    return 0
  fi

  if [[ -d "$value" ]]; then
    value="${value%/}/ipadMoneyWiz.sqlite"
  fi

  value="$(resolve_file_path "$value")"
  [[ -f "$value" ]] || fail "Database file not found: $value"
  DB_PATH="$value"
}

build_binary() {
  local repo_root="$1"
  local output="$2"
  command -v go >/dev/null 2>&1 || fail "Go is required to build from source."

  log "Building ${BIN_NAME}..."
  go build -o "$output" "$repo_root/cmd/main.go"
}

install_binary() {
  local source_bin="$1"
  mkdir -p "$INSTALL_DIR"
  install -m 0755 "$source_bin" "$INSTALL_DIR/$BIN_NAME"
  log "Installed $INSTALL_DIR/$BIN_NAME"
}

configure_claude_desktop() {
  local target_bin="$1"
  local config_path="${CLAUDE_DESKTOP_CONFIG:-$CLAUDE_DESKTOP_CONFIG_DEFAULT}"

  command -v python3 >/dev/null 2>&1 || {
    log "Skipping Claude Desktop config update (python3 not found)."
    return 0
  }

  mkdir -p "$(dirname "$config_path")"
  python3 - "$config_path" "$SERVER_NAME" "$target_bin" "$DB_PATH" <<'PY'
import json
import os
import sys

config_path, server_name, command_path, db_path = sys.argv[1:5]

config = {}
if os.path.exists(config_path):
    with open(config_path, "r", encoding="utf-8") as f:
        raw = f.read().strip()
    if raw:
        config = json.loads(raw)

if not isinstance(config, dict):
    raise SystemExit(f"Config root must be an object: {config_path}")

mcp_servers = config.get("mcpServers")
if not isinstance(mcp_servers, dict):
    mcp_servers = {}

entry = {"command": command_path, "args": []}
if db_path:
    entry["args"] = ["-db", db_path]

mcp_servers[server_name] = entry
config["mcpServers"] = mcp_servers

with open(config_path, "w", encoding="utf-8") as f:
    json.dump(config, f, indent=2)
    f.write("\n")
PY

  log "Updated Claude Desktop config: $config_path"
}

configure_claude_code() {
  local target_bin="$1"
  if ! command -v claude >/dev/null 2>&1; then
    log "Skipping Claude Code registration (claude CLI not found)."
    return 0
  fi

  case "$CLAUDE_SCOPE" in
    local|user|project) ;;
    *) fail "Invalid scope '$CLAUDE_SCOPE'. Use: local|user|project" ;;
  esac

  claude mcp remove --scope "$CLAUDE_SCOPE" "$SERVER_NAME" >/dev/null 2>&1 || true

  if [[ -n "$DB_PATH" ]]; then
    claude mcp add --scope "$CLAUDE_SCOPE" "$SERVER_NAME" "$target_bin" -db "$DB_PATH"
  else
    claude mcp add --scope "$CLAUDE_SCOPE" "$SERVER_NAME" "$target_bin"
  fi
  log "Registered server in Claude Code (scope: $CLAUDE_SCOPE)."
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --db)
        DB_PATH="$2"
        shift 2
        ;;
      --name)
        SERVER_NAME="$2"
        shift 2
        ;;
      --install-dir)
        INSTALL_DIR="$2"
        shift 2
        ;;
      --scope)
        CLAUDE_SCOPE="$2"
        shift 2
        ;;
      --skip-claude-desktop)
        SKIP_CLAUDE_DESKTOP=1
        shift
        ;;
      --skip-claude-code)
        SKIP_CLAUDE_CODE=1
        shift
        ;;
      -y|--yes)
        AUTO_YES=1
        shift
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        fail "Unknown argument: $1"
        ;;
    esac
  done
}

main() {
  parse_args "$@"
  normalize_db_path "$DB_PATH"

  local repo_root
  repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
  local tmpdir
  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT INT TERM

  build_binary "$repo_root" "$tmpdir/$BIN_NAME"
  install_binary "$tmpdir/$BIN_NAME"

  if confirm_default_yes "Register with detected Claude clients? [Y/n]: "; then
    if [[ "$SKIP_CLAUDE_DESKTOP" -eq 0 ]]; then
      configure_claude_desktop "$INSTALL_DIR/$BIN_NAME"
    fi
    if [[ "$SKIP_CLAUDE_CODE" -eq 0 ]]; then
      configure_claude_code "$INSTALL_DIR/$BIN_NAME"
    fi
  else
    log "Skipping client registration."
  fi

  log "Install complete."
}

main "$@"
