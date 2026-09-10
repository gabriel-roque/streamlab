import type { VideoStatus } from "../types";

const labels: Record<VideoStatus, string> = { READY: "READY", PROCESSING: "PROCESSING", FAILED: "FAILED", UPLOADING: "UPLOADING" };

export function StatusBadge({ status }: { status: VideoStatus }) {
  return <span className={`status status-${status.toLowerCase()}`}><span className="status-dot" />{labels[status]}</span>;
}
