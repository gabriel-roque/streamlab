#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=lib/common.sh
source "$SCRIPT_DIR/lib/common.sh"

input=""
kind="auto"
expected_codec=""
expected_width=""
expected_height=""
duration_tolerance="${DURATION_TOLERANCE_SECONDS:-1.0}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --input|-i) [[ $# -ge 2 ]] || usage_error "$1 requires a value"; input="$2"; shift 2 ;;
    --kind) [[ $# -ge 2 ]] || usage_error "--kind requires a value"; kind="$2"; shift 2 ;;
    --codec) [[ $# -ge 2 ]] || usage_error "--codec requires a value"; expected_codec="$2"; shift 2 ;;
    --width) [[ $# -ge 2 ]] || usage_error "--width requires a value"; expected_width="$2"; shift 2 ;;
    --height) [[ $# -ge 2 ]] || usage_error "--height requires a value"; expected_height="$2"; shift 2 ;;
    --help|-h)
      cat <<'EOF'
Validate a media file, HLS directory, or DASH directory.

Usage: scripts/validate-media.sh --input PATH [--kind file|hls|dash]
                                 [--codec CODEC] [--width N] [--height N]
EOF
      media_help
      exit 0
      ;;
    *) usage_error "unknown argument: $1" ;;
  esac
done

[[ -n "$input" ]] || usage_error "--input is required"
input="$(workspace_file "$input")"
[[ -e "$input" ]] || die "media path not found: $input"

if [[ "$kind" == auto ]]; then
  if [[ -d "$input" && -f "$input/master.m3u8" ]]; then kind=hls
  elif [[ -d "$input" && -f "$input/manifest.mpd" ]]; then kind=dash
  else kind=file
  fi
fi

case "$kind" in
  file)
    [[ -f "$input" ]] || die "file validation requires a regular file"
    codec="$(run_media ffprobe -v error -select_streams v:0 -show_entries stream=codec_name -of default=nw=1:nk=1 "$input")"
    width="$(run_media ffprobe -v error -select_streams v:0 -show_entries stream=width -of default=nw=1:nk=1 "$input")"
    height="$(run_media ffprobe -v error -select_streams v:0 -show_entries stream=height -of default=nw=1:nk=1 "$input")"
    duration="$(run_media ffprobe -v error -show_entries format=duration -of default=nw=1:nk=1 "$input")"
    [[ -n "$codec" && "$width" =~ ^[0-9]+$ && "$height" =~ ^[0-9]+$ ]] || die "missing video stream metadata"
    [[ "$duration" =~ ^[0-9]+([.][0-9]+)?$ ]] || die "missing or invalid duration"
    [[ -z "$expected_codec" || "$codec" == "$expected_codec" ]] || die "codec mismatch: expected $expected_codec, got $codec"
    [[ -z "$expected_width" || "$width" == "$expected_width" ]] || die "width mismatch: expected $expected_width, got $width"
    [[ -z "$expected_height" || "$height" == "$expected_height" ]] || die "height mismatch: expected $expected_height, got $height"
    printf 'valid file: codec=%s size=%sx%s duration=%ss\n' "$codec" "$width" "$height" "$duration"
    ;;
  hls)
    master="$input/master.m3u8"
    [[ -f "$master" ]] || die "HLS master playlist not found: $master"
    grep -q '^#EXTM3U' "$master" || die "HLS master is missing #EXTM3U"
    while IFS= read -r line; do
      [[ -n "$line" ]] || continue
      [[ "$line" == \#* ]] && continue
      playlist="$input/$line"
      [[ -f "$playlist" ]] || die "HLS variant playlist missing: $playlist"
      grep -q '^#EXTM3U' "$playlist" || die "invalid HLS variant: $playlist"
      while IFS= read -r segment; do
        [[ -n "$segment" ]] || continue
        [[ "$segment" == \#* ]] && continue
        [[ -f "$(dirname "$playlist")/$segment" ]] || die "HLS segment missing: $segment"
      done < <(grep -v '^#' "$playlist" || true)
    done < <(grep -v '^#' "$master" || true)
    printf 'valid HLS package: %s\n' "$input"
    ;;
  dash)
    manifest="$input/manifest.mpd"
    [[ -f "$manifest" ]] || die "DASH manifest not found: $manifest"
    grep -q '<MPD' "$manifest" || die "DASH manifest is missing MPD root"
    grep -q '<Representation' "$manifest" || die "DASH manifest has no representations"
    grep -q 'mediaInitializationRange\|initialization=' "$manifest" || die "DASH manifest has no initialization reference"
    printf 'valid DASH manifest: %s\n' "$manifest"
    ;;
  *) die "kind must be file, hls, or dash" ;;
esac
