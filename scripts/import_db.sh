#!/usr/bin/env bash
set -euo pipefail

DEFAULT_DB_DIR="${MONEYWIZ_DB_DIR:-$HOME/.moneywiz-mcp}"
DEFAULT_DB_PATH="$DEFAULT_DB_DIR/ipadMoneyWiz.sqlite"
DEFAULT_BACKUP_DIR="$DEFAULT_DB_DIR/backups"

SOURCE_PATH=""
TARGET_DB_PATH="$DEFAULT_DB_PATH"
BACKUP_DIR="$DEFAULT_BACKUP_DIR"
CREATE_BACKUP=1

log() {
  printf '%s\n' "$*"
}

fail() {
  printf 'Error: %s\n' "$*" >&2
  exit 1
}

usage() {
  cat <<EOF
Usage: $(basename "$0") [options] <source>

Import a MoneyWiz export into a stable local path:
  $DEFAULT_DB_PATH

<source> can be:
  - export folder (iMoneyWiz-Data-Backup-*)
  - direct sqlite file (ipadMoneyWiz.sqlite)

Options:
  --target <path>            Target sqlite path (default: $DEFAULT_DB_PATH)
  --backup-dir <path>        Backup directory (default: $DEFAULT_BACKUP_DIR)
  --no-backup                Skip backup when replacing existing DB
  -h, --help                 Show help
EOF
}

resolve_file_path() {
  local path="$1"
  local dir base
  dir="$(cd "$(dirname "$path")" && pwd)"
  base="$(basename "$path")"
  printf '%s/%s\n' "$dir" "$base"
}

normalize_source_path() {
  local input="$1"
  local abs
  abs="$(resolve_file_path "$input")"
  [[ -e "$abs" ]] || fail "Source not found: $abs"

  if [[ -d "$abs" ]]; then
    abs="${abs%/}/ipadMoneyWiz.sqlite"
  fi
  [[ -f "$abs" ]] || fail "SQLite file not found at source: $abs"
  SOURCE_PATH="$abs"
}

prepare_target_paths() {
  case "$TARGET_DB_PATH" in
    /*) ;;
    *) TARGET_DB_PATH="$(pwd)/$TARGET_DB_PATH" ;;
  esac
  case "$BACKUP_DIR" in
    /*) ;;
    *) BACKUP_DIR="$(pwd)/$BACKUP_DIR" ;;
  esac
  mkdir -p "$(dirname "$TARGET_DB_PATH")"
  mkdir -p "$BACKUP_DIR"
}

backup_existing_if_needed() {
  if [[ ! -f "$TARGET_DB_PATH" ]]; then
    return 0
  fi
  if [[ "$CREATE_BACKUP" -ne 1 ]]; then
    return 0
  fi

  if cmp -s "$SOURCE_PATH" "$TARGET_DB_PATH"; then
    log "Source is identical to target. No backup needed."
    return 0
  fi

  local ts
  ts="$(date +"%Y%m%d-%H%M%S")"
  local backup_path="$BACKUP_DIR/ipadMoneyWiz-$ts.sqlite"
  cp "$TARGET_DB_PATH" "$backup_path"
  log "Backup created: $backup_path"
}

copy_db() {
  cp "$SOURCE_PATH" "$TARGET_DB_PATH"
  chmod 0644 "$TARGET_DB_PATH" || true
  log "Imported database to: $TARGET_DB_PATH"
}

print_next_steps() {
  cat <<EOF

Next steps:
- Install/re-register with:
  ./scripts/install.sh
- Or pin the exact path manually:
  ./scripts/install.sh --db "$TARGET_DB_PATH"
- Or set:
  export MONEYWIZ_DB_PATH="$TARGET_DB_PATH"
EOF
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --target)
        TARGET_DB_PATH="$2"
        shift 2
        ;;
      --backup-dir)
        BACKUP_DIR="$2"
        shift 2
        ;;
      --no-backup)
        CREATE_BACKUP=0
        shift
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      -*)
        fail "Unknown option: $1"
        ;;
      *)
        if [[ -n "$SOURCE_PATH" ]]; then
          fail "Only one source argument is allowed."
        fi
        SOURCE_PATH="$1"
        shift
        ;;
    esac
  done
}

main() {
  parse_args "$@"
  [[ -n "$SOURCE_PATH" ]] || fail "Missing source argument. See --help."

  normalize_source_path "$SOURCE_PATH"
  prepare_target_paths
  backup_existing_if_needed
  copy_db
  print_next_steps
}

main "$@"
