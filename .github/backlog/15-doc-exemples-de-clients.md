---
title: "docs: client examples that call the private service with an identity token"
labels: ["type:doc", "pkg:docs", "prio:normale", "taille:trait"]
milestone: "v1.0.0 Stable API"
assignee: PaulPerigault
---

## Context

ADR 0003 makes the deployed service private: a caller needs an identity token with the invoker role. The README shows one `curl` line in `infra/README.md` and nothing for a program calling the service.

## Goal

A reader can call the service from a shell, from a Go program and from a Node program running on Google Cloud, without reading the Cloud Run IAM documentation first.

## Acceptance criteria

- [ ] `docs/clients.md` shows the three examples, each with the audience the token must carry (the service URL).
- [ ] The Go example compiles under `go vet` as a test file in `dev/` or a documented snippet checked by a test; it uses the standard library plus the official ID-token helper only if its dependency is justified in the file.
- [ ] `task check:docs` passes: every command and path in the file exists.
- [ ] No real project ID or URL appears in the examples (placeholders only).

## Package

docs

## Size

trait

## Priority

normale

## Milestone

v1.0.0 Stable API

## Out of scope

Client libraries. Setting up the caller's service account.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
