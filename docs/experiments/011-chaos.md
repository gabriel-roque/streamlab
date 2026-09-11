# 011 — Chaos testing

## Objective

Demonstrate predictable behavior when a dependency fails without confusing
deliberate unavailability with an unobserved incident.

## Matriz

| Failure | Expected signal | Recovery |
|---|---|---|
| kill worker/API | in-memory queue may lose work | restart and new upload |
| stop local storage | uploads/packaging fail | HTTP error, no retry guarantee |
| stop API | new requests fail | player uses already-published cache |
| stop Redis/PostgreSQL/MinIO | no effect on current code | verify that they are future dependencies |
| increase latency | p95 and rebuffering increase | timeout and circuit breaker |
| limit bandwidth | ABR reduces variant | no switch loop |
| corrupt video | FFmpeg may fail; fixture may mask it | observe the job, do not assume DLQ |

## Procedure

Define the baseline, duration, blast radius, and rollback. Execute one failure at
a time; correlate by job ID (the current API does not generate `X-Request-ID`)
and capture status, error rate, retries, DLQ, startup, rebuffering, and cache. In
Compose, the service names are `api`, `nginx`, `postgres`, `redis`, `minio`,
`prometheus`, and `grafana`; remember that the last four are not dependencies
read by the current API. Never inject chaos into an environment with user data.

## Criteria

Every failure has a detector, impact limit, documented behavior, and verifiable
recovery. A system that merely returns 500 quickly is not considered resilient.
