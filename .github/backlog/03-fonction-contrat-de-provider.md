---
title: "test: one contract suite run against both providers"
labels: ["type:fonction", "pkg:provider", "prio:haute", "taille:page"]
milestone: "v0.3.0 Testability and offline demo"
assignee: PaulPerigault
---

## Context

ADR 0002 accepts that the fake can drift from the real client. Today `provider/fake` and `provider/gcp` have separate tests that assert the same behaviours by hand: sorted list, short names, `ErrNotFound` for an absent or malformed name, wrapped errors, context cancellation.

## Goal

A single exported test helper in `internal/provider` states the contract once and both implementations must pass it.

## Acceptance criteria

- [ ] `internal/provider/providertest` exposes `Run(t, factory)` with table-driven cases for list order, get, not found, invalid name and cancelled context.
- [ ] `go test ./internal/provider/...` runs it for `fake` and for `gcp` (against the in-process gRPC server).
- [ ] Breaking the sort order in either implementation makes this suite fail (shown in the pull request).
- [ ] `internal/provider` stays free of Google libraries (`task lint`).

## Package

provider

## Size

page

## Priority

haute

## Milestone

v0.3.0 Testability and offline demo

## Out of scope

Contract tests against a real Google project. New provider implementations.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
