import { useRef, useState } from "react";
import { Icon } from "./Icon";

export function UploadModal({ onClose, onUpload, uploading, progress }: { onClose: () => void; onUpload: (file: File, title: string) => void; uploading: boolean; progress: number }) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [file, setFile] = useState<File | null>(null);
  const [title, setTitle] = useState("");

  function chooseFile(selected: File | undefined) {
    if (!selected) return;
    setFile(selected);
    if (!title) setTitle(selected.name.replace(/\.[^/.]+$/, ""));
  }

  return (
    <div className="modal-backdrop" role="presentation" onMouseDown={(event) => event.target === event.currentTarget && !uploading && onClose()}>
      <section className="upload-modal" role="dialog" aria-modal="true" aria-labelledby="upload-title">
        <div className="modal-heading"><div><p className="eyebrow">INGEST / NEW ASSET</p><h2 id="upload-title">Send a video to the lab</h2></div><button className="icon-button" onClick={onClose} disabled={uploading} aria-label="Fechar upload"><Icon name="x" /></button></div>
        <button className={`drop-zone ${file ? "drop-zone-selected" : ""}`} onClick={() => inputRef.current?.click()} onDragOver={(event) => event.preventDefault()} onDrop={(event) => { event.preventDefault(); chooseFile(event.dataTransfer.files[0]); }}>
          <span className="upload-mark"><Icon name="cloud-upload" size={24} /></span>
          <strong>{file ? file.name : "Drop a source file here"}</strong>
          <span>{file ? `${(file.size / 1024 / 1024).toFixed(1)} MB ready for ingest` : "MP4, MOV or WebM up to 2 GB"}</span>
          <input ref={inputRef} type="file" accept="video/mp4,video/quicktime,video/webm" hidden onChange={(event) => chooseFile(event.target.files?.[0])} />
        </button>
        <label className="field-label" htmlFor="video-title">Asset name</label>
        <input id="video-title" className="text-input" value={title} onChange={(event) => setTitle(event.target.value)} placeholder="e.g. Product launch / take 01" disabled={uploading} />
        {uploading && <div className="upload-progress"><div className="progress-line"><span style={{ width: `${progress}%` }} /></div><span>{progress}% uploaded. Creating renditions next.</span></div>}
        <div className="modal-actions"><button className="button button-ghost" onClick={onClose} disabled={uploading}>Cancel</button><button className="button button-primary" disabled={!file || !title.trim() || uploading} onClick={() => file && onUpload(file, title.trim())}>{uploading ? "Uploading..." : "Start ingest"}<Icon name="arrow-left" size={16} /></button></div>
      </section>
    </div>
  );
}
