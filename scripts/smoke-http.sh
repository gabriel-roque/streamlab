#!/usr/bin/env bash

set -euo pipefail

base_url="http://localhost:3000"
paths=("/")
manifest_kind=""
manifest_path=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --url) [[ $# -ge 2 ]] || { printf 'error: --url requires a value\n' >&2; exit 2; }; base_url="$2"; shift 2 ;;
    --path) [[ $# -ge 2 ]] || { printf 'error: --path requires a value\n' >&2; exit 2; }; paths+=("$2"); shift 2 ;;
    --manifest) [[ $# -ge 2 ]] || { printf 'error: --manifest requires a value\n' >&2; exit 2; }; manifest_path="$2"; shift 2 ;;
    --kind) [[ $# -ge 2 ]] || { printf 'error: --kind requires a value\n' >&2; exit 2; }; manifest_kind="$2"; shift 2 ;;
    -h|--help)
      cat <<'EOF'
Run a small HTTP smoke test without assuming application routes.

Usage: scripts/smoke-http.sh [--url BASE_URL] [--path PATH]...
                             [--manifest PATH --kind hls|dash]

The default check is GET /. Add --path /healthz for a health endpoint.
EOF
      exit 0
      ;;
    *) printf 'error: unknown argument: %s\n' "$1" >&2; exit 2 ;;
  esac
done

command -v curl >/dev/null 2>&1 || { printf 'error: curl is required\n' >&2; exit 1; }
base_url="${base_url%/}"

for path in "${paths[@]}"; do
  [[ "$path" == /* ]] || path="/$path"
  code="$(curl --silent --show-error --location --output /dev/null --write-out '%{http_code}' "$base_url$path")" || {
    printf 'FAIL %s%s (request error)\n' "$base_url" "$path" >&2
    exit 1
  }
  [[ "$code" =~ ^2[0-9][0-9]$|^3[0-9][0-9]$ ]] || {
    printf 'FAIL %s%s (HTTP %s)\n' "$base_url" "$path" "$code" >&2
    exit 1
  }
  printf 'PASS %s%s (HTTP %s)\n' "$base_url" "$path" "$code"
done

if [[ -n "$manifest_path" ]]; then
  [[ "$manifest_kind" == hls || "$manifest_kind" == dash ]] || {
    printf 'error: --kind must be hls or dash when --manifest is used\n' >&2
    exit 2
  }
  body="$(curl --fail --silent --show-error --location "$base_url/${manifest_path#/}")"
  if [[ "$manifest_kind" == hls ]]; then
    printf '%s\n' "$body" | grep -q '^#EXTM3U' || { printf 'FAIL manifest is not HLS\n' >&2; exit 1; }
  else
    printf '%s\n' "$body" | grep -q '<MPD' || { printf 'FAIL manifest is not DASH\n' >&2; exit 1; }
  fi
  printf 'PASS %s manifest (%s)\n' "$manifest_path" "$manifest_kind"
fi
