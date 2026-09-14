#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

for script in scripts/*.sh scripts/lib/*.sh; do
  bash -n "$script"
done

for script in \
  scripts/download-bbb.sh \
  scripts/ffprobe-media.sh \
  scripts/generate-ladder.sh \
  scripts/generate-hls.sh \
  scripts/generate-dash.sh \
  scripts/validate-media.sh \
  scripts/smoke-http.sh \
  scripts/quick-start.sh; do
  "$script" --help >/dev/null
done

git check-ignore --quiet samples/raw/big-buck-bunny-1080p-normal.mp4
git check-ignore --quiet samples/generated/hls/master.m3u8
printf 'script checks passed\n'
