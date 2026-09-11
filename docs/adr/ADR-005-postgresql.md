# ADR-005: PostgreSQL for Metadata

- **Status:** accepted
- **Data:** 2026-09-10

## Context

Videos, variants, jobs, attempts, and sessions have relationships and
transitions that require querying and consistency.

## Decision

Compose provisions PostgreSQL for the catalog stage, but the current
implementation keeps videos and jobs in the in-memory `LocalStore`/`MemoryQueue`.
The API does not open a PostgreSQL connection or persist bytes in the database.

## Alternatives

- Redis as the primary source: good for ephemeral state, weak as a catalog.
- NoSQL document: flexible, but less useful for the phase's relational invariants.
- JSON files: reproducible, but not concurrent.

## Consequences

When migrating to PostgreSQL, transitions must be atomic and indexed by
`video_id`, `status`, and `created_at`. These guarantees do not exist after a
restart today, and metrics are counters for the in-memory instance.
