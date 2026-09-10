import type { PlaybackInfo, Video } from "../types";

const configuredApiUrl = (import.meta.env.VITE_API_URL as string | undefined)?.trim();
export const API_BASE = (configuredApiUrl || "/api").replace(/\/$/, "");

type JsonObject = Record<string, unknown>;

type PlaybackEvent = {
  session_id?: string;
  video_id?: string;
  type: string;
  position?: number;
  payload?: Record<string, unknown>;
};

class ApiError extends Error {
  constructor(public readonly status: number) {
    super(`API ${status}`);
    this.name = "ApiError";
  }
}

function isObject(value: unknown): value is JsonObject {
  return typeof value === "object" && value !== null;
}

function stringValue(value: unknown): string | undefined {
  return typeof value === "string" && value.length > 0 ? value : undefined;
}

function numberValue(value: unknown, fallback = 0): number {
  return typeof value === "number" && Number.isFinite(value) ? value : fallback;
}

function objectValue(value: unknown): JsonObject {
  return isObject(value) ? value : {};
}

function resolveUrl(value: unknown): string | undefined {
  const url = stringValue(value);
  if (!url || /^https?:\/\//i.test(url) || url.startsWith("blob:")) return url;
  if (url.startsWith("/")) {
    if (url === API_BASE || url.startsWith(`${API_BASE}/`)) return url;
    return `${API_BASE}${url}`;
  }
  return `${API_BASE}/${url}`;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  if (init?.body && !(init.body instanceof FormData) && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  const response = await fetch(`${API_BASE}${path}`, { ...init, headers });
  if (!response.ok) throw new ApiError(response.status);
  if (response.status === 204) return undefined as T;
  const text = await response.text();
  return (text ? JSON.parse(text) : undefined) as T;
}

function unwrapVideo(value: unknown): JsonObject {
  const object = objectValue(value);
  return isObject(object.video) ? object.video : object;
}

function normalizeVideo(value: unknown, fallback: Partial<Video> = {}): Video {
  const item = unwrapVideo(value);
  const variants = Array.isArray(item.variants) ? item.variants.filter(isObject) : [];
  const variant = variants.find((candidate) => numberValue(candidate.height) >= 1080) ?? variants[0] ?? {};
  const metadata = objectValue(item.metadata);
  const manifests = objectValue(item.manifests);
  const source = resolveUrl(item.source_url) ?? stringValue(item.source) ?? fallback.source ?? "";
  const hlsManifest = resolveUrl(manifests.hls ?? manifests.HLS);
  const status = stringValue(item.status) as Video["status"] | undefined;
  const codec = stringValue(item.codec) ?? ([stringValue(variant.video_codec), stringValue(variant.audio_codec)].filter(Boolean).join(" / ") || fallback.codec || "Detecting");
  const failureReason = stringValue(item.failure_reason ?? item.failureReason) ?? fallback.failureReason;

  return {
    id: stringValue(item.id) ?? fallback.id ?? `video-${Date.now()}`,
    title: stringValue(item.title) ?? fallback.title ?? "Untitled video",
    status: status === "UPLOADING" || status === "PROCESSING" || status === "READY" || status === "FAILED" ? status : fallback.status ?? "PROCESSING",
    duration: numberValue(item.duration_seconds ?? item.duration, fallback.duration ?? 0),
    width: numberValue(item.width ?? variant.width, fallback.width ?? 0),
    height: numberValue(item.height ?? variant.height, fallback.height ?? 0),
    codec,
    bitrate: numberValue(item.bitrate ?? variant.bitrate, fallback.bitrate ?? 0),
    createdAt: stringValue(item.created_at ?? item.createdAt) ?? fallback.createdAt ?? new Date().toISOString(),
    thumbnail: stringValue(item.thumbnail) ?? fallback.thumbnail ?? "",
    source: source || hlsManifest || "",
    protocol: stringValue(item.protocol) === "HLS" || Boolean(hlsManifest) ? "HLS" : fallback.protocol ?? "MP4",
    description: stringValue(item.description) ?? stringValue(metadata.description) ?? fallback.description ?? "",
    ...(numberValue(item.progress, -1) >= 0 ? { progress: numberValue(item.progress) } : fallback.progress !== undefined ? { progress: fallback.progress } : {}),
    ...(failureReason ? { failureReason } : {}),
  };
}

function normalizePlayback(value: unknown, videoId: string): PlaybackInfo {
  const item = objectValue(value);
  const manifests = objectValue(item.manifests);
  const protocols = objectValue(item.protocols);
  const explicitManifest = resolveUrl(item.manifest);
  const manifest = explicitManifest ?? (manifests.hls || manifests.HLS ? resolveUrl(protocols.hls ?? manifests.hls ?? manifests.HLS) : undefined);
  const source = resolveUrl(item.source_url) ?? resolveUrl(item.source);
  return {
    videoId: stringValue(item.video_id ?? item.videoId) ?? videoId,
    protocol: stringValue(item.protocol) === "HLS" || Boolean(manifest) ? "HLS" : "MP4",
    ...(manifest ? { manifest } : {}),
    ...(source ? { source } : {}),
    ...(stringValue(item.preview) ? { preview: resolveUrl(item.preview) } : {}),
    ...(isObject(item.tracks) ? { tracks: item.tracks as PlaybackInfo["tracks"] } : {}),
  };
}

export async function listVideos(): Promise<Video[]> {
  const response = await request<unknown>("/videos");
  const items = Array.isArray(response) ? response : isObject(response) && Array.isArray(response.videos) ? response.videos : [];
  return items.filter(isObject).map((item) => normalizeVideo(item));
}

export async function getPlayback(videoId: string): Promise<PlaybackInfo> {
  return normalizePlayback(await request<unknown>(`/videos/${encodeURIComponent(videoId)}/playback`), videoId);
}

export async function createPlaybackSession(videoId: string): Promise<{ id: string }> {
  const response = await request<unknown>("/playback/sessions", {
    method: "POST",
    body: JSON.stringify({ video_id: videoId }),
  });
  const item = objectValue(response);
  const session = isObject(item.session) ? item.session : item;
  const id = stringValue(session.id);
  if (!id) throw new Error("API session did not return an id");
  return { id };
}

export async function sendPlaybackEvent(payload: PlaybackEvent): Promise<void> {
  await request<void>("/playback/events", { method: "POST", body: JSON.stringify(payload) });
}

function xhrUpload(path: string, formData: FormData, onProgress: (progress: number) => void): Promise<unknown> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("POST", `${API_BASE}${path}`);
    xhr.setRequestHeader("Accept", "application/json");
    xhr.upload.onprogress = (event) => {
      if (event.lengthComputable) onProgress(Math.round((event.loaded / event.total) * 100));
    };
    xhr.onload = () => {
      if (xhr.status < 200 || xhr.status >= 300) {
        reject(new ApiError(xhr.status));
        return;
      }
      if (xhr.status === 204 || !xhr.responseText) {
        resolve(undefined);
        return;
      }
      try {
        resolve(JSON.parse(xhr.responseText));
      } catch {
        resolve(undefined);
      }
    };
    xhr.onerror = () => reject(new Error("Network unavailable"));
    xhr.send(formData);
  });
}

function uploadForm(file: File, title?: string): FormData {
  const formData = new FormData();
  formData.append("file", file, file.name);
  if (title) formData.append("title", title);
  return formData;
}

export async function uploadVideo(
  file: File,
  title: string,
  onProgress: (progress: number) => void,
): Promise<Video> {
  try {
    const uploaded = await xhrUpload("/videos", uploadForm(file, title), onProgress);
    onProgress(100);
    return normalizeVideo(uploaded, {
      title,
      status: "PROCESSING",
      description: "Upload received. Processing has started.",
    });
  } catch (error) {
    // Some deployments expose metadata creation and binary upload separately.
    if (!(error instanceof ApiError) || ![404, 405, 415].includes(error.status)) throw error;
    const created = await request<unknown>("/videos", {
      method: "POST",
      body: JSON.stringify({ title, filename: file.name }),
    });
    const createdVideo = unwrapVideo(created);
    const id = stringValue(createdVideo.id);
    if (!id) throw new Error("API video creation did not return an id");
    const uploaded = await xhrUpload(`/videos/${encodeURIComponent(id)}/upload`, uploadForm(file), onProgress);
    onProgress(100);
    return normalizeVideo({ ...createdVideo, ...unwrapVideo(uploaded), id }, {
      id,
      title,
      status: "PROCESSING",
      description: "Upload received. Processing has started.",
    });
  }
}
