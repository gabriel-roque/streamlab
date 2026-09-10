export type VideoStatus = "UPLOADING" | "PROCESSING" | "READY" | "FAILED";

export type Video = {
  id: string;
  title: string;
  status: VideoStatus;
  duration: number;
  width: number;
  height: number;
  codec: string;
  bitrate: number;
  createdAt: string;
  thumbnail: string;
  source: string;
  protocol: "HLS" | "MP4";
  description: string;
  progress?: number;
  failureReason?: string;
};

export type PlaybackInfo = {
  videoId: string;
  protocol: "HLS" | "MP4";
  manifest?: string;
  source?: string;
  preview?: string;
  tracks?: { audio?: string[]; subtitles?: Array<{ label: string; language: string; src: string }> };
};

export type PlayerMetrics = {
  protocol: string;
  resolution: string;
  bitrate: string;
  buffer: string;
  startup: string;
  switches: number;
};
