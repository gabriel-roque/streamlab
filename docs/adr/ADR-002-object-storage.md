# ADR-002: Object Storage for Media Bytes

- **Status:** accepted
- **Data:** 2026-09-10

## Context

Originals and segments are large, immutable, and have a different read pattern
from transactional metadata.

## Decision

Keep the local filesystem as the current adapter: `media/` stores originals and
`artifacts/{video_id}/` stores manifests/segments. Compose also starts MinIO with
an S3-compatible interface for the next stage, but the current API does not
connect to it or provide uploads through presigned URLs.

## Alternatives

- MinIO/S3 from the start: represents object storage, but is not yet used by the
  executable code.
- Store bytes in PostgreSQL: couples database scale to media throughput.
- S3 from the start: realistic, but less reproducible offline.

## Consequences

The local adapter does not provide retention, versioning, signed URLs, or atomic
prefix publication. These are concerns for the MinIO/S3 stage; for now,
restarting the API loses metadata, but the directory's bytes persist.
