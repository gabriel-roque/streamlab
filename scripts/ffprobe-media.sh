#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=lib/common.sh
source "$SCRIPT_DIR/lib/common.sh"

input="samples/raw/big-buck-bunny-1080p-normal.mp4"
output=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --input|-i) [[ $# -ge 2 ]] || usage_error "$1 requires a value"; input="$2"; shift 2 ;;
    --output|-o) [[ $# -ge 2 ]] || usage_error "$1 requires a value"; output="$2"; shift 2 ;;
    -h|--help)
      cat <<'EOF'
Run ffprobe with stable JSON output.

Usage: scripts/ffprobe-media.sh [--input PATH] [--output PATH]
EOF
      media_help
      exit 0
      ;;
    *) usage_error "unknown argument: $1" ;;
  esac
done

input="$(workspace_file "$input")"
[[ -f "$input" ]] || die "input media not found: $input"

if [[ -n "$output" ]]; then
  output="$(workspace_file "$output")"
  mkdir -p "$(dirname "$output")"
  run_media ffprobe -v error -show_format -show_streams -of json "$input" > "$output"
  printf 'ffprobe JSON written to %s\n' "$output"
else
  run_media ffprobe -hide_banner -show_format -show_streams "$input"
fi
