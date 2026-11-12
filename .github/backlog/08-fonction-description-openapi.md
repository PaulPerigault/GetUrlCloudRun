---
title: "feat(api): OpenAPI 3.1 description checked against the handlers"
labels: ["type:fonction", "pkg:api", "prio:normale", "taille:page"]
milestone: "v0.4.0 API hardening"
assignee: PaulPerigault
---

## Context

The routes, statuses and error bodies are described in the README only. A client author, or a future contract test, has nothing machine readable to check against.

## Goal

A hand-written OpenAPI file lives in the repository, and a test fails when a handler answers something the file does not describe.

## Acceptance criteria

- [ ] `docs/openapi.yaml` describes the three routes, every status the handlers produce and the error body.
- [ ] A Go test replays each documented case against the router (fake provider) and checks status and content type; an undocumented status makes it fail.
- [ ] `task check:docs` verifies the link from the README.
- [ ] No new runtime dependency: the check uses the standard library to read the file or a small subset of it.

## Package

api

## Size

page

## Priority

normale

## Milestone

v0.4.0 API hardening

## Out of scope

Generating server or client code from the file. Serving the file over HTTP.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
