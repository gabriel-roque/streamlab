import { useEffect, useMemo, useState } from "react";
import { Icon } from "./components/Icon";
import { ObservabilityPanel } from "./components/ObservabilityPanel";
import { Player } from "./components/Player";
import { StatusBadge } from "./components/StatusBadge";
import { UploadModal } from "./components/UploadModal";
import { VideoCard } from "./components/VideoCard";
import { initialVideos } from "./data";
import { createPlaybackSession, getPlayback, listVideos, sendPlaybackEvent, uploadVideo } from "./lib/api";
import type { PlaybackInfo, PlayerMetrics, Video, VideoStatus } from "./types";

type Filter = "ALL" | VideoStatus;

function formatDate(date: string) { return new Intl.DateTimeFormat("en", { month: "short", day: "numeric", year: "numeric" }).format(new Date(date)); }
function localUpload(file: File, title: string): Video { return { id: `local-${Date.now()}`, title, status: "READY", duration: 0, width: 0, height: 0, codec: "Browser source", bitrate: 0, createdAt: new Date().toISOString(), thumbnail: "", source: URL.createObjectURL(file), protocol: "MP4", description: "Local preview. API is offline, so this asset is available only in this session." }; }

export default function App() {
  const [videos, setVideos] = useState(initialVideos);
  const [selected, setSelected] = useState<Video | null>(null);
  const [playback, setPlayback] = useState<PlaybackInfo | null>(null);
  const [metrics, setMetrics] = useState<PlayerMetrics>();
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState<Filter>("ALL");
  const [sort, setSort] = useState("newest");
  const [view, setView] = useState<"grid" | "list">("grid");
  const [showUpload, setShowUpload] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [apiOnline, setApiOnline] = useState(true);
  const [sessionId, setSessionId] = useState("");

  useEffect(() => { listVideos().then((items) => { setVideos(items); setApiOnline(true); }).catch(() => setApiOnline(false)); }, []);
  useEffect(() => { if (!selected) { setPlayback(null); setSessionId(""); return; } setMetrics(undefined); getPlayback(selected.id).then(setPlayback).catch(() => setPlayback({ videoId: selected.id, protocol: selected.protocol, manifest: selected.protocol === "HLS" ? selected.source : undefined, source: selected.protocol === "MP4" ? selected.source : undefined })); createPlaybackSession(selected.id).then((session) => setSessionId(session.id)).catch(() => setSessionId("")); }, [selected]);

  const filteredVideos = useMemo(() => videos.filter((video) => (filter === "ALL" || video.status === filter) && video.title.toLowerCase().includes(query.toLowerCase())).sort((a, b) => sort === "newest" ? b.createdAt.localeCompare(a.createdAt) : a.title.localeCompare(b.title)), [videos, filter, query, sort]);
  const counts = useMemo(() => ({ all: videos.length, ready: videos.filter((video) => video.status === "READY").length, processing: videos.filter((video) => video.status === "PROCESSING").length }), [videos]);

  async function handleUpload(file: File, title: string) {
    setUploading(true); setUploadProgress(0);
    try { const uploaded = await uploadVideo(file, title, setUploadProgress); setVideos((current) => [uploaded, ...current]); setShowUpload(false); }
    catch { setApiOnline(false); setVideos((current) => [localUpload(file, title), ...current]); setShowUpload(false); }
    finally { setUploading(false); }
  }
  function track(event: string, detail?: Record<string, unknown>) {
    const position = typeof detail?.position === "number" ? detail.position : undefined;
    void sendPlaybackEvent({ ...(sessionId ? { session_id: sessionId } : {}), video_id: selected?.id, type: event, ...(position === undefined ? {} : { position }), ...(detail ? { payload: detail } : {}) }).catch(() => setApiOnline(false));
  }

  if (selected) {
    const source = playback?.manifest ?? playback?.source ?? selected.source;
    const detailMetrics = metrics ?? { protocol: playback?.protocol ?? selected.protocol, resolution: selected.width ? `${selected.width}x${selected.height}` : "Pending", bitrate: selected.bitrate ? `${(selected.bitrate / 1000000).toFixed(1)} Mbps` : "--", buffer: "0.0 s", startup: "--", switches: 0 };
    return <div className="app-shell detail-view"><header className="topbar"><button className="brand" onClick={() => setSelected(null)} aria-label="Voltar para biblioteca"><span className="brand-mark"><i /><i /><i /></span><span>stream<span>lab</span></span></button><div className="topbar-right"><span className={`api-pill ${apiOnline ? "online" : "offline"}`}><i />{apiOnline ? "API connected" : "local fallback"}</span><button className="icon-button" aria-label="Configuracoes"><Icon name="settings" /></button></div></header><main className="content"><button className="back-link" onClick={() => setSelected(null)}><Icon name="arrow-left" size={16} /> Library / {selected.title}</button><div className="detail-heading"><div><p className="eyebrow">ASSET / {selected.id}</p><h1>{selected.title}</h1><p className="detail-description">{selected.description}</p></div><StatusBadge status={selected.status} /></div>{source ? <Player source={source} protocol={playback?.protocol ?? selected.protocol} title={selected.title} preview={playback?.preview} audioTracks={playback?.tracks?.audio} subtitles={playback?.tracks?.subtitles} onTelemetry={track} onMetrics={setMetrics} /> : <div className="empty-player"><Icon name="activity" size={28} /><strong>Playback manifest is not ready</strong><span>This asset is still moving through the pipeline.</span></div>}<div className="detail-grid"><section className="panel metadata-panel"><div className="section-heading"><div><p className="eyebrow">MEDIA PROBE</p><h2>Asset metadata</h2></div><Icon name="sliders" size={20} /></div><div className="metadata-list"><div><span>Dimensions</span><strong>{selected.width ? `${selected.width} x ${selected.height}` : "Pending"}</strong></div><div><span>Video codec</span><strong>{selected.codec}</strong></div><div><span>Container</span><strong>{selected.protocol === "HLS" ? "MPEG-TS / HLS" : "MP4"}</strong></div><div><span>Ingested</span><strong>{formatDate(selected.createdAt)}</strong></div></div></section><ObservabilityPanel metrics={detailMetrics} /></div></main></div>;
  }

  return <div className="app-shell"><header className="topbar"><button className="brand" onClick={() => setSelected(null)} aria-label="StreamLab home"><span className="brand-mark"><i /><i /><i /></span><span>stream<span>lab</span></span></button><nav className="main-nav" aria-label="Navegacao principal"><button className="active"><Icon name="folder" size={16} />Library</button><button onClick={() => document.getElementById("observability")?.scrollIntoView({ behavior: "smooth" })}><Icon name="activity" size={16} />Observability</button></nav><div className="topbar-right"><span className={`api-pill ${apiOnline ? "online" : "offline"}`}><i />{apiOnline ? "API connected" : "local fallback"}</span><button className="icon-button" aria-label="Configuracoes"><Icon name="settings" /></button><button className="avatar" aria-label="Conta de SL">SL</button></div></header><main className="content"><section className="hero"><div><p className="eyebrow">STREAMING LABORATORY / 01</p><h1>Media, measured.<br /><em>Not imagined.</em></h1><p className="hero-copy">A focused workspace for shipping video through ingest, encoding and adaptive playback.</p></div><div className="hero-signal"><span className="signal-kicker"><i />pipeline status</span><strong>all systems<br /><b>nominal</b></strong><span className="sparkline" aria-hidden="true"><i /><i /><i /><i /><i /><i /><i /><i /><i /><i /></span></div></section><section className="library-toolbar"><div className="section-heading"><div><p className="eyebrow">YOUR LIBRARY / {String(counts.all).padStart(2, "0")}</p><h2>Video assets</h2></div><span className="library-summary"><b>{counts.ready}</b> ready <span /> <b>{counts.processing}</b> in pipeline</span></div><div className="toolbar-row"><label className="search-field"><Icon name="search" size={17} /><input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Search assets..." aria-label="Buscar videos" /></label><label className="filter-select"><Icon name="filter" size={16} /><select value={filter} onChange={(event) => setFilter(event.target.value as Filter)} aria-label="Filtrar por status"><option value="ALL">All status</option><option value="READY">Ready</option><option value="PROCESSING">Processing</option><option value="FAILED">Failed</option></select><Icon name="chevron-down" size={14} /></label><label className="sort-select"><span>Sort</span><select value={sort} onChange={(event) => setSort(event.target.value)} aria-label="Ordenar videos"><option value="newest">Newest</option><option value="name">Name</option></select><Icon name="chevron-down" size={14} /></label><div className="view-toggle"><button className={view === "grid" ? "active" : ""} onClick={() => setView("grid")} aria-label="Visualizacao em grade"><Icon name="grid" size={16} /></button><button className={view === "list" ? "active" : ""} onClick={() => setView("list")} aria-label="Visualizacao em lista"><Icon name="list" size={16} /></button></div><button className="button button-primary upload-button" onClick={() => setShowUpload(true)}><Icon name="plus" size={17} />Upload video</button></div></section><section className={`video-grid ${view === "list" ? "list-view" : ""}`} aria-live="polite">{filteredVideos.length ? filteredVideos.map((video) => <VideoCard key={video.id} video={video} onOpen={setSelected} />) : <div className="no-results"><Icon name="search" size={24} /><strong>No assets found</strong><span>Try changing your search or status filter.</span></div>}</section><section id="observability"><ObservabilityPanel /></section></main>{showUpload && <UploadModal onClose={() => setShowUpload(false)} onUpload={handleUpload} uploading={uploading} progress={uploadProgress} />}</div>;
}
