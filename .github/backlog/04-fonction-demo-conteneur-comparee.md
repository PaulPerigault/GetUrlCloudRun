---
title: "feat(demo): task demo:docker also checks the routes, not only the healthcheck"
labels: ["type:fonction", "pkg:image", "prio:basse", "taille:trait"]
milestone: "v0.3.0 Testability and offline demo"
assignee: PaulPerigault
---

## Context

`task demo` calls every route on the real binary; `task demo:docker` only waits for the container healthcheck. A container that is healthy but mis-wired (wrong entrypoint, read-only file system problem) would pass.

## Goal

The container demo asserts the same answers as the binary demo.

## Acceptance criteria

- [ ] `dev/demo` accepts a base URL, so the same steps run against the container.
- [ ] `task demo:docker` runs the steps from a second container on the internal network and fails on any difference.
- [ ] The container keeps `read_only`, `cap_drop: [ALL]`, `no-new-privileges` and the internal network.

## Package

image

## Size

trait

## Priority

basse

## Milestone

v0.3.0 Testability and offline demo

## Out of scope

Publishing a port to the host. Running the demo in CI (the `e2e` job already runs `task demo`).

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
