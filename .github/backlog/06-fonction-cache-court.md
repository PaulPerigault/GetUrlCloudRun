---
title: "feat(provider): short TTL cache in front of ListServices"
labels: ["type:fonction", "pkg:provider", "prio:normale", "taille:page"]
milestone: "v0.4.0 API hardening"
assignee: PaulPerigault
---

## Context

Each list request is one or more calls to the Cloud Run Admin API, which has quotas, and service URLs change rarely. The rate limit protects the API but not the latency.

## Goal

Repeated list calls inside a short window are served from memory, with a configurable duration and a clock that tests control.

## Acceptance criteria

- [ ] A decorator in its own package (`internal/provider/cache`) implements `provider.Provider` around another provider, with an injected clock.
- [ ] `CACHE_TTL` (default `0s`, meaning disabled; at most `5m`) is validated by `internal/config` and documented in the README table.
- [ ] Tests: hit inside the window, miss after it, errors are never cached, concurrent calls under `-race` call the inner provider once.
- [ ] `depguard` keeps the decorator free of `net/http` and Google libraries.

## Package

provider

## Size

page

## Priority

normale

## Milestone

v0.4.0 API hardening

## Out of scope

A shared cache between instances. Invalidation on deploy.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
