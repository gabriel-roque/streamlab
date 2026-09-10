#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=lib/common.sh
source "$SCRIPT_DIR/lib/common.sh"

input="samples/raw/big-buck-bunny-1080p-normal.mp4"
output_dir="samples/generated/ladder"
overwrite=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --input|-i) [[ $# -ge 2 ]] || usage_error "$1 requires a value"; input="$2"; shift 2 ;;
    --output-dir|-o) [[ $# -ge 2 ]] || usage_error "$1 requires a value"; output_dir="$2"; shift 2 ;;
    --overwrite) overwrite=true; shift ;;
    -h|--help)
      cat <<'EOF'
Generate the fixed H.264/AAC ladder, skipping renditions larger than the input.

Usage: scripts/generate-ladder.sh [--input PATH] [--output-dir PATH] [--overwrite]

Profiles: 240p/400k, 360p/800k, 480p/1200k, 720p/2500k, 1080p/5000k.
EOF
      media_help
      exit 0
      ;;
    *) usage_error "unknown argument: $1" ;;
  esac
done

input="$(workspace_file "$input")"
output_dir="$(workspace_file "$output_dir")"
[[ -f "$input" ]] || die "input media not found: $input"
mkdir -p "$output_dir"

source_height="$(run_media ffprobe -v error -select_streams v:0 -show_entries stream=height -of default=noprint_wrappers=1:nokey=1 "$input")"
[[ "$source_height" =~ ^[0-9]+$ ]] || die "could not determine source video height"

printf 'profile\twidth\theight\tvideo_bitrate\n' > "$output_dir/ladder.tsv"

profiles=("240 426 400k" "360 640 800k" "480 854 1200k" "720 1280 2500k" "1080 1920 5000k")
for profile in "${profiles[@]}"; do
  read -r height width bitrate <<< "$profile"
  (( height <= source_height )) || continue

  target="$output_dir/${height}p.mp4"
  printf '%s\t%s\t%s\t%s\n' "${height}p" "$width" "$height" "$bitrate" >> "$output_dir/ladder.tsv"
  if [[ -f "$target" && "$overwrite" == false ]]; then
    printf 'exists, skipping: %s\n' "$target"
    continue
  fi

  printf 'encoding %sp (%s)\n' "$height" "$bitrate"
  run_media ffmpeg -hide_banner -loglevel warning -y -i "$input" \
    -map 0:v:0 -map 0:a:0? \
    -vf "scale=${width}:${height}:force_original_aspect_ratio=decrease,pad=${width}:${height}:(ow-iw)/2:(oh-ih)/2" \
    -c:v libx264 -preset "${X264_PRESET:-veryfast}" -profile:v main \
    -b:v "$bitrate" -maxrate "$bitrate" -bufsize "$((${bitrate%k} * 2))k" \
    -g 48 -keyint_min 48 -sc_threshold 0 -pix_fmt yuv420p \
    -c:a aac -b:a 128k -ar 48000 -movflags +faststart "$target"
done

printf 'ladder written to %s\n' "$output_dir"
