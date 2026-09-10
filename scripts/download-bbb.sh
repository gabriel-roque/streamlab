#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=lib/common.sh
source "$SCRIPT_DIR/lib/common.sh"

url="${BBB_URL:-https://download.blender.org/demo/movies/BBB/bbb_sunflower_1080p_30fps_normal.mp4.zip}"
output="samples/raw/big-buck-bunny-1080p-normal.mp4"
sha256="${BBB_SHA256:-}"
skip_probe=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --url) [[ $# -ge 2 ]] || usage_error "--url requires a value"; url="$2"; shift 2 ;;
    --output) [[ $# -ge 2 ]] || usage_error "--output requires a value"; output="$2"; shift 2 ;;
    --sha256) [[ $# -ge 2 ]] || usage_error "--sha256 requires a value"; sha256="$2"; shift 2 ;;
    --skip-probe) skip_probe=true; shift ;;
    -h|--help)
      cat <<'EOF'
Download and validate the official Blender Foundation Big Buck Bunny source.

Usage: scripts/download-bbb.sh [--output PATH] [--url URL] [--sha256 HASH]
                               [--skip-probe]

The default is the current official ZIP archive. The script extracts its MP4
without committing either the archive or the media. BBB_URL may point to an
official uncompressed media URL as well.
EOF
      media_help
      exit 0
      ;;
    *) usage_error "unknown argument: $1" ;;
  esac
done

case "$url" in
  https://download.blender.org/*|https://archive.blender.org/*) ;;
  *) die "BBB source must be an official Blender download URL: $url" ;;
esac

output="$(workspace_file "$output")"
mkdir -p "$(dirname "$output")"
if [[ "$url" == *.zip ]]; then
  download_path="${output}.part.zip"
else
  download_path="${output}.part"
fi

if command -v curl >/dev/null 2>&1; then
  curl --fail --location --retry 3 --retry-delay 2 --output "$download_path" "$url"
elif command -v wget >/dev/null 2>&1; then
  wget --tries=3 --output-document="$download_path" "$url"
else
  die "curl or wget is required"
fi

[[ -s "$download_path" ]] || die "downloaded file is empty: $download_path"

if [[ -n "$sha256" ]]; then
  command -v sha256sum >/dev/null 2>&1 || die "sha256sum is required for checksum validation"
  printf '%s  %s\n' "$sha256" "$download_path" | sha256sum --check --status - || die "SHA-256 mismatch: $download_path"
fi

if [[ "$url" == *.zip ]]; then
  command -v unzip >/dev/null 2>&1 || die "unzip is required for the official BBB ZIP archive"
  entry="$(unzip -Z1 "$download_path" | while IFS= read -r candidate; do
    case "$candidate" in
      *.mp4) printf '%s' "$candidate"; break ;;
    esac
  done)"
  [[ -n "$entry" ]] || die "official BBB archive contains no MP4"
  unzip -p "$download_path" "$entry" > "${output}.part"
  rm -f "$download_path"
  mv "${output}.part" "$output"
else
  mv "$download_path" "$output"
fi

if [[ "$skip_probe" == false ]]; then
  run_media ffprobe -v error -show_entries format=duration:stream=codec_type,codec_name,width,height \
    -of default=noprint_wrappers=1 "$output" || die "FFprobe rejected the downloaded media"
fi

printf 'validated Big Buck Bunny: %s\n' "$output"
