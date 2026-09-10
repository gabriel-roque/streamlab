#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=lib/common.sh
source "$SCRIPT_DIR/lib/common.sh"

input="samples/raw/big-buck-bunny-1080p-normal.mp4"
ladder_dir=""
output_dir="samples/generated/dash"
segment_seconds="${DASH_SEGMENT_SECONDS:-6}"
overwrite=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --input|-i) [[ $# -ge 2 ]] || usage_error "$1 requires a value"; input="$2"; shift 2 ;;
    --ladder-dir) [[ $# -ge 2 ]] || usage_error "--ladder-dir requires a value"; ladder_dir="$2"; shift 2 ;;
    --output-dir|-o) [[ $# -ge 2 ]] || usage_error "$1 requires a value"; output_dir="$2"; shift 2 ;;
    --segment-seconds) [[ $# -ge 2 ]] || usage_error "--segment-seconds requires a value"; segment_seconds="$2"; shift 2 ;;
    --overwrite) overwrite=true; shift ;;
    -h|--help)
      cat <<'EOF'
Generate a VOD MPEG-DASH MPD from the same aligned ladder used by HLS.

Usage: scripts/generate-dash.sh [--input PATH] [--ladder-dir PATH]
                                [--output-dir PATH] [--segment-seconds N]
                                [--overwrite]
EOF
      media_help
      exit 0
      ;;
    *) usage_error "unknown argument: $1" ;;
  esac
done

input="$(workspace_file "$input")"
output_dir="$(workspace_file "$output_dir")"
if [[ -n "$ladder_dir" ]]; then
  ladder_dir="$(workspace_file "$ladder_dir")"
else
  ladder_dir="$output_dir/.ladder"
  mkdir -p "$ladder_dir"
  ladder_args=(--input "$input" --output-dir "$ladder_dir")
  [[ "$overwrite" == true ]] && ladder_args+=(--overwrite)
  "$SCRIPT_DIR/generate-ladder.sh" "${ladder_args[@]}"
fi

[[ -d "$ladder_dir" ]] || die "ladder directory not found: $ladder_dir"
mkdir -p "$output_dir"
manifest="$output_dir/manifest.mpd"
if [[ -f "$manifest" && "$overwrite" == false ]]; then
  printf 'exists, skipping: %s\n' "$manifest"
  exit 0
fi

profiles=(240 360 480 720 1080)
inputs=()
maps=()
video_streams=()
audio_streams=()
index=0
has_audio=false
for height in "${profiles[@]}"; do
  rendition="$ladder_dir/${height}p.mp4"
  [[ -f "$rendition" ]] || continue
  inputs+=( -i "$rendition" )
  maps+=( -map "$index:v:0" )
  video_streams+=( "$index" )
  if [[ "$index" == 0 ]] && run_media ffprobe -v error -select_streams a:0 -show_entries stream=index -of default=nw=1:nk=1 "$rendition" >/dev/null 2>&1; then
    has_audio=true
  fi
  ((index += 1))
done

(( index > 0 )) || die "no ladder renditions found in $ladder_dir"
if [[ "$has_audio" == true ]]; then
  maps+=( -map 0:a:0 )
  audio_streams+=( "$index" )
fi

video_set="id=0,streams=$(IFS=,; printf '%s' "${video_streams[*]}")"
adaptation_sets="$video_set"
if [[ "$has_audio" == true ]]; then
  audio_set="id=1,streams=$(IFS=,; printf '%s' "${audio_streams[*]}")"
  adaptation_sets="$adaptation_sets $audio_set"
fi

run_media ffmpeg -hide_banner -loglevel warning -y "${inputs[@]}" "${maps[@]}" \
  -c copy -f dash -seg_duration "$segment_seconds" -use_template 1 -use_timeline 1 \
  -adaptation_sets "$adaptation_sets" \
  -init_seg_name 'init-$RepresentationID$.$ext$' \
  -media_seg_name 'chunk-$RepresentationID$-$Number%05d$.$ext$' "$manifest"

printf 'DASH manifest written to %s\n' "$manifest"
