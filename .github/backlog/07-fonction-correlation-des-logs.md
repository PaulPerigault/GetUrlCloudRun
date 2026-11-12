---
title: "feat(httpmw): correlate access logs with the Cloud Trace header"
labels: ["type:fonction", "pkg:httpmw", "prio:basse", "taille:trait"]
milestone: "v0.4.0 API hardening"
assignee: PaulPerigault
---

## Context

Cloud Run adds `X-Cloud-Trace-Context` to each request, and Cloud Logging groups log lines that carry the matching trace field. Today the access log and the provider error log of one request cannot be joined.

## Goal

Every log line of a request carries the trace ID, and nothing else from the headers.

## Acceptance criteria

- [ ] A middleware reads the trace ID from the header (only the ID part, validated against its grammar) and puts it in the request context.
- [ ] The access log and `api` error logs include a `trace` attribute when present; a malformed header is ignored.
- [ ] A test shows the query string and every other header are still absent from the log.

## Package

httpmw

## Size

trait

## Priority

basse

## Milestone

v0.4.0 API hardening

## Out of scope

Distributed tracing or OpenTelemetry. Sending logs anywhere other than stdout.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
