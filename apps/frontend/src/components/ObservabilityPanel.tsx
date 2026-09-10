import type { PlayerMetrics } from "../types";
import { Icon } from "./Icon";

const baseMetrics: PlayerMetrics = { protocol: "HLS", resolution: "1920x1080", bitrate: "5.8 Mbps", buffer: "23.4 s", startup: "1.24 s", switches: 4 };

export function ObservabilityPanel({ metrics = baseMetrics }: { metrics?: PlayerMetrics }) {
  const cards = [["PROTOCOL", metrics.protocol, "transport"], ["RESOLUTION", metrics.resolution, "current level"], ["BITRATE", metrics.bitrate, "effective"], ["BUFFER", metrics.buffer, "ahead"], ["STARTUP", metrics.startup, "time to first frame"], ["SWITCHES", String(metrics.switches), "quality changes"]];
  return (
    <section className="observability panel" aria-labelledby="observability-title">
      <div className="section-heading"><div><p className="eyebrow">LIVE SIGNAL / 02</p><h2 id="observability-title">Playback observability</h2></div><span className="live-indicator"><span />live session</span></div>
      <div className="metric-grid">{cards.map(([label, value, hint]) => <div className="metric-card" key={label}><span>{label}</span><strong>{value}</strong><small>{hint}</small></div>)}</div>
      <div className="signal-row"><div className="signal-label"><Icon name="activity" size={17} /><span>Edge health</span><strong>nominal</strong></div><div className="signal-bars" aria-label="Edge health 96 percent">{[1, 1, 1, 1, 1, 1, 1, 1, 1, 0].map((active, index) => <i className={active ? "active" : ""} key={index} />)}</div><span className="signal-percent">96%</span></div>
    </section>
  );
}
