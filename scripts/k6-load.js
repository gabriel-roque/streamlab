import http from 'k6/http';
import { check, sleep } from 'k6';

const baseURL = (__ENV.BASE_URL || 'http://localhost:3000').replace(/\/$/, '');
const manifestPath = __ENV.MANIFEST_PATH || '/video/123/master.m3u8';
const configuredSegments = (__ENV.SEGMENTS || '')
  .split(',')
  .map((value) => value.trim())
  .filter((value) => value.length > 0);

export const options = {
  scenarios: {
    players: {
      executor: 'ramping-vus',
      startVUs: Number(__ENV.START_VUS || 1),
      stages: [
        { duration: __ENV.RAMP_UP || '30s', target: Number(__ENV.TARGET_VUS || 10) },
        { duration: __ENV.HOLD || '60s', target: Number(__ENV.TARGET_VUS || 10) },
        { duration: __ENV.RAMP_DOWN || '15s', target: 0 },
      ],
      gracefulRampDown: '10s',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<1000'],
    checks: ['rate>0.99'],
  },
};

function absoluteURL(path) {
  if (path.startsWith('http://') || path.startsWith('https://')) return path;
  return `${baseURL}/${path.replace(/^\//, '')}`;
}

function playlistSegments(body) {
  return body
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line && !line.startsWith('#'))
    .slice(0, 3);
}

export default function () {
  const manifest = http.get(absoluteURL(manifestPath), { tags: { resource: 'manifest' } });
  const manifestOK = check(manifest, {
    'manifest is successful': (response) => response.status >= 200 && response.status < 300,
    'manifest has content': (response) => response.body.length > 0,
  });

  if (manifestOK) {
    const segments = configuredSegments.length > 0
      ? configuredSegments
      : playlistSegments(manifest.body);
    for (const segment of segments) {
      const response = http.get(absoluteURL(segment), { tags: { resource: 'segment' } });
      check(response, { 'segment is successful': (item) => item.status >= 200 && item.status < 300 });
    }
  }
  sleep(Number(__ENV.THINK_TIME || 1));
}
