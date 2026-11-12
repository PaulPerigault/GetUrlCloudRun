---
title: "feat(api): pagination on GET /api/v1/services"
labels: ["type:fonction", "pkg:api", "prio:normale", "taille:page"]
milestone: "v0.4.0 API hardening"
assignee: PaulPerigault
---

## Context

The list route returns every service of the region in one response, built from every page of the Cloud Run Admin API. A project with hundreds of services makes the response and the request time grow without bound, and the request deadline is 5 seconds by default.

## Goal

The list is bounded: a caller asks for a page size and follows a token, and an unbounded request never reaches the Admin API.

## Acceptance criteria

- [ ] `GET /api/v1/services?page_size=2` returns at most 2 items and a `next_page_token` when there are more (fake provider, test).
- [ ] `page_size` above 200 or below 1 answers `400` with a fixed message.
- [ ] The old response shape stays the default when no parameter is given, until the contract freeze (v1.0.0 issue).
- [ ] `provider.Provider` carries the page parameters; `provider/gcp` passes them to the iterator.
- [ ] `task docs:readme` shows a paginated example and `task check:docs` passes.

## Package

api

## Size

page

## Priority

normale

## Milestone

v0.4.0 API hardening

## Out of scope

Filtering or searching services by label. Cursor stability across a changing project.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
