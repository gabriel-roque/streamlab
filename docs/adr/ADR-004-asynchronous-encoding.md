# ADR-004: Asynchronous Encoding

- **Status:** accepted
- **Data:** 2026-09-10

## Context

Transcoding is CPU/GPU-intensive, can outlast the HTTP timeout, and can fail
independently of the upload request.

## Decision

The upload writes the file and enqueues a `transcode` job in an in-memory queue.
A worker processes analysis/packaging in the same process, with states
`QUEUED`, `RUNNING`, `RETRYING`, `SUCCEEDED`, and `DLQ`, with up to three retries
(four total executions including the first). The current implementation has no
broker, lease, `VideoUploaded` event, or idempotency key; these are left for the
lab's evolution.

## Alternatives

- Synchronous encoding in the API: immediate feedback, but blocks resources and
  scales poorly.
- Database polling via cron: simple, but increases contention and latency.
- External orchestrator: possible later, but excessive for the initial phase.

## Consequences

The client queries `/videos/{id}/status` and does not assume that upload implies
`READY`. The current retry uses a short linear delay in memory, and the DLQ has
no HTTP replay route; restarting the API loses the queue. Backoff, leases, and a
persistent DLQ remain architecture exercises.
