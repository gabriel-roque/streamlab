import type { Video } from "../types";
import { Icon } from "./Icon";
import { StatusBadge } from "./StatusBadge";

function formatDuration(seconds: number) {
  if (!seconds) return "--:--";
  const minutes = Math.floor(seconds / 60);
  return `${String(minutes).padStart(2, "0")}:${String(Math.floor(seconds % 60)).padStart(2, "0")}`;
}

export function VideoCard({
  video,
  onOpen,
}: {
  video: Video;
  onOpen: (video: Video) => void;
}) {
  const openable = video.status === "READY";
  return (
    <article className={`video-card ${!openable ? "video-card-muted" : ""}`}>
      <button
        className="thumbnail"
        onClick={() => openable && onOpen(video)}
        disabled={!openable}
        aria-label={
          openable ? `Open ${video.title}` : `${video.title} unavailable`
        }
      >
        {video.thumbnail ? (
          <img src={video.thumbnail} alt="" loading="lazy" />
        ) : (
          <div className="thumbnail-placeholder">
            <Icon
              name={video.status === "FAILED" ? "x" : "activity"}
              size={26}
            />
          </div>
        )}
        <span className="thumbnail-overlay" />
        {openable && (
          <span className="thumbnail-play">
            <Icon name="play" size={18} />
          </span>
        )}
        <span className="duration">{formatDuration(video.duration)}</span>
      </button>
      <div className="card-copy">
        <div className="card-title-row">
          <h3>{video.title}</h3>
          <button
            className="icon-button small"
            aria-label={`Options for ${video.title}`}
          >
            <Icon name="settings" size={16} />
          </button>
        </div>
        <p>{video.description}</p>
        <div className="card-meta">
          <StatusBadge status={video.status} />
          {video.status === "PROCESSING" && (
            <span className="progress-copy">{video.progress}% encoded</span>
          )}
          {video.status === "FAILED" && (
            <span className="failure-copy">{video.failureReason}</span>
          )}
        </div>
        {video.status === "PROCESSING" && (
          <div className="progress-track">
            <span style={{ width: `${video.progress ?? 0}%` }} />
          </div>
        )}
      </div>
    </article>
  );
}
