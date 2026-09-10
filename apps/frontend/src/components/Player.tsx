import Hls from "hls.js";
import { useEffect, useRef, useState } from "react";
import type { PlayerMetrics } from "../types";
import { Icon } from "./Icon";

type Props = { source: string; protocol: "HLS" | "MP4"; title: string; preview?: string; audioTracks?: string[]; subtitles?: Array<{ label: string; language: string; src: string }>; onTelemetry?: (event: string, detail?: Record<string, unknown>) => void; onMetrics?: (metrics: PlayerMetrics) => void };

const initialMetrics: PlayerMetrics = { protocol: "MP4", resolution: "1920x1080", bitrate: "5.8 Mbps", buffer: "0.0 s", startup: "--", switches: 0 };

export function Player({ source, protocol, title, preview, audioTracks, subtitles, onTelemetry, onMetrics }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const startedAt = useRef<number>(0);
  const hlsRef = useRef<Hls | null>(null);
  const [playing, setPlaying] = useState(false);
  const [muted, setMuted] = useState(false);
  const [progress, setProgress] = useState(0);
  const [duration, setDuration] = useState(0);
  const [quality, setQuality] = useState("auto");
  const [metrics, setMetrics] = useState<PlayerMetrics>({ ...initialMetrics, protocol });
  const [error, setError] = useState(false);
  const [previewPosition, setPreviewPosition] = useState<number | null>(null);

  useEffect(() => {
    const video = videoRef.current;
    if (!video || !source) return;
    setError(false);
    setPlaying(false);
    setProgress(0);
    const updateMetrics = () => {
      const buffered = video.buffered.length ? video.buffered.end(video.buffered.length - 1) - video.currentTime : 0;
      const next = { ...metrics, buffer: `${Math.max(0, buffered).toFixed(1)} s`, resolution: video.videoWidth ? `${video.videoWidth}x${video.videoHeight}` : metrics.resolution };
      setMetrics(next);
      onMetrics?.(next);
    };
    let hls: Hls | null = null;
    if (source.includes(".m3u8")) {
      if (Hls.isSupported()) {
        hls = new Hls({ enableWorker: true, capLevelToPlayerSize: true });
        hlsRef.current = hls;
        hls.loadSource(source);
        hls.attachMedia(video);
        hls.on(Hls.Events.LEVEL_SWITCHED, (_event, data) => {
          const level = hls?.levels[data.level];
          const next = { ...metrics, switches: metrics.switches + 1, resolution: level ? `${level.width}x${level.height}` : metrics.resolution, bitrate: level ? `${(level.bitrate / 1000000).toFixed(1)} Mbps` : metrics.bitrate };
          setMetrics(next);
          onMetrics?.(next);
          onTelemetry?.("quality_change", { level: data.level });
        });
        hls.on(Hls.Events.ERROR, (_event, data) => { if (data.fatal) setError(true); });
      } else if (video.canPlayType("application/vnd.apple.mpegurl")) video.src = source;
      else setError(true);
    } else video.src = source;
    video.addEventListener("loadedmetadata", () => { setDuration(video.duration); updateMetrics(); });
    video.addEventListener("timeupdate", () => setProgress(video.duration ? (video.currentTime / video.duration) * 100 : 0));
    video.addEventListener("progress", updateMetrics);
    video.addEventListener("waiting", () => onTelemetry?.("rebuffer_start"));
    video.addEventListener("playing", () => { if (startedAt.current) { const startup = ((performance.now() - startedAt.current) / 1000).toFixed(2); const next = { ...metrics, startup: `${startup} s` }; setMetrics(next); onMetrics?.(next); startedAt.current = 0; } onTelemetry?.("playing"); });
    video.addEventListener("error", () => setError(true));
    return () => { hls?.destroy(); hlsRef.current = null; video.removeAttribute("src"); video.load(); };
  }, [source, protocol]);

  function togglePlay() {
    const video = videoRef.current;
    if (!video) return;
    if (video.paused) {
      startedAt.current = performance.now();
      try {
        const result = video.play();
        if (result && typeof result.then === "function") result.then(() => { setPlaying(true); onTelemetry?.("play"); }).catch(() => setError(true));
        else { setPlaying(true); onTelemetry?.("play"); }
      } catch { setError(true); }
    }
    else { video.pause(); setPlaying(false); onTelemetry?.("pause"); }
  }
  function seek(value: number) { const video = videoRef.current; if (!video) return; video.currentTime = (value / 100) * video.duration; setProgress(value); onTelemetry?.("seek", { position: video.currentTime }); }
  function changeQuality(value: string) { setQuality(value); const hls = hlsRef.current; if (!hls) return; if (value === "auto") hls.currentLevel = -1; else { const level = hls.levels.findIndex((item) => item.height === Number(value)); if (level >= 0) hls.currentLevel = level; } }
  function formatTime(value: number) { return `${String(Math.floor(value / 60)).padStart(2, "0")}:${String(Math.floor(value % 60)).padStart(2, "0")}`; }

  return (
    <div className="player-shell">
      <div className="player-stage">
        <video ref={videoRef} playsInline crossOrigin="anonymous" aria-label={`Player: ${title}`} onClick={togglePlay}>
          {subtitles?.map((track) => <track key={track.language} kind="subtitles" label={track.label} srcLang={track.language} src={track.src} />)}
        </video>
        {error && <div className="player-error"><Icon name="activity" size={28} /><strong>Playback unavailable</strong><span>Use the API playback URL or try again when the source is online.</span></div>}
        {!playing && !error && <button className="center-play" onClick={togglePlay} aria-label={`Reproduzir ${title}`}><Icon name="play" size={28} /></button>}
        <div className="player-topline"><span className="player-live"><span />{protocol === "HLS" ? "HLS / ADAPTIVE" : "PROGRESSIVE MP4"}</span><span className="player-title">{title}</span></div>
        <div className="player-controls">{preview && previewPosition !== null && <div className="seek-preview" style={{ left: `${previewPosition}%` }}><img src={preview} alt="" /><span>{formatTime((previewPosition / 100) * duration)}</span></div>}<input className="seek" type="range" min="0" max="100" value={progress} onMouseMove={(event) => { const rect = event.currentTarget.getBoundingClientRect(); setPreviewPosition(Math.max(0, Math.min(100, ((event.clientX - rect.left) / rect.width) * 100))); }} onMouseLeave={() => setPreviewPosition(null)} onChange={(event) => seek(Number(event.target.value))} aria-label="Posicao no video" /><div className="controls-row"><button className="player-icon" onClick={togglePlay} aria-label={playing ? "Pausar" : "Reproduzir"}><Icon name={playing ? "pause" : "play"} size={18} /></button><span className="timecode">{formatTime((progress / 100) * duration)} / {formatTime(duration)}</span><div className="control-spacer" /><label className="select-control" aria-label="Qualidade"><span>quality</span><select value={quality} onChange={(event) => changeQuality(event.target.value)}><option value="auto">Auto</option><option value="1080">1080p</option><option value="720">720p</option><option value="480">480p</option></select><Icon name="chevron-down" size={13} /></label>{audioTracks?.length ? <label className="select-control" aria-label="Audio"><span>audio</span><select defaultValue="default" onChange={(event) => { const index = audioTracks.indexOf(event.target.value); if (hlsRef.current && index >= 0) hlsRef.current.audioTrack = index; onTelemetry?.("audio_change", { language: event.target.value }); }}><option value="default">Default</option>{audioTracks.map((track) => <option value={track} key={track}>{track}</option>)}</select><Icon name="chevron-down" size={13} /></label> : <button className={`player-icon ${muted ? "is-muted" : ""}`} onClick={() => { const video = videoRef.current; if (video) video.muted = !video.muted; setMuted(!muted); }} aria-label={muted ? "Ativar audio" : "Desativar audio"}><Icon name="volume" size={18} /></button>}</div></div>
      </div>
      <div className="player-readout"><span><b>BUFFER</b> {metrics.buffer}</span><span><b>ABR</b> {quality === "auto" ? "active" : "locked"}</span><span><b>CAPTIONS</b> {subtitles?.length ? "available" : "none"}</span><span><b>CODEC</b> H.264</span></div>
    </div>
  );
}
