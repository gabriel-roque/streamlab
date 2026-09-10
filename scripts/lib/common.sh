#!/usr/bin/env bash

set -euo pipefail

STREAMLAB_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FFMPEG_IMAGE="${FFMPEG_IMAGE:-jrottenberg/ffmpeg:6.1-ubuntu}"

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

usage_error() {
  printf 'error: %s\n' "$*" >&2
  exit 2
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

workspace_file() {
  local path="$1"
  if [[ "$path" == /* ]]; then
    case "$path" in
      "$STREAMLAB_ROOT"/*) printf '%s\n' "${path#"$STREAMLAB_ROOT"/}" ;;
      *) die "path must be inside the workspace: $path" ;;
    esac
  else
    printf '%s\n' "${path#./}"
  fi
}

run_media() {
  local tool="$1"
  shift
  local mode="${MEDIA_TOOL:-auto}"

  if [[ "$mode" != docker && "$mode" != local && "$mode" != auto ]]; then
    die "MEDIA_TOOL must be auto, local, or docker"
  fi

  if [[ "$mode" != docker ]] && command -v "$tool" >/dev/null 2>&1; then
    "$tool" "$@"
    return
  fi

  if [[ "$mode" == local ]]; then
    die "$tool not found; install FFmpeg or use MEDIA_TOOL=docker"
  fi
  require_cmd docker
  docker run --rm \
    --user "$(id -u):$(id -g)" \
    --volume "$STREAMLAB_ROOT:/workspace" \
    --workdir /workspace \
    --entrypoint "$tool" \
    "$FFMPEG_IMAGE" "$@"
}

media_help() {
  cat <<'EOF'
Media tools are selected with MEDIA_TOOL:

  MEDIA_TOOL=auto    use local ffmpeg/ffprobe, then Docker (default)
  MEDIA_TOOL=local   require local ffmpeg/ffprobe
  MEDIA_TOOL=docker  use Docker even when local tools exist

The Docker fallback requires Docker and pulls FFMPEG_IMAGE on first use.
EOF
}

cd "$STREAMLAB_ROOT"
