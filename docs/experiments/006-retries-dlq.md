# 006 — Retries and DLQ

## Question

Does the pipeline recover from transient failures without infinitely
reprocessing a corrupted job?

## Procedure

1. Observe the job created by the upload at `GET /videos/{id}/status`.
2. In code tests, make the handler fail and observe `RETRYING`, up to three
   retries (four total executions including the first), and then `DLQ`.
3. Record that the current queue exposes no HTTP route to inject a failure,
   replay, or query the DLQ; an arbitrary file may proceed through the fixture
   and is not a reliable transcoding-error test.

The implemented behavior uses `max_retries=3` and a short linear delay in
memory, without jitter, lease, correlation ID, or persistence. For the
architecture exercise, replace it with exponential backoff with jitter and
record the reason, first error, last attempt, and correlation ID. The desired
output key remains deterministic by `video_id/profile/version`.

## Criteria

In the current queue, verify that the job reaches `DLQ` after the limit and that
the video may remain `PROCESSING`; recovery after a crash is not guaranteed
because the queue is in memory. In the target architecture, there must not be
two competing public artifacts, a crash must not lose the job, and DLQ replay
must be explicit and auditable.
