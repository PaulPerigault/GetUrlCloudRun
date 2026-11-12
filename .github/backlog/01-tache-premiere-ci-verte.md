---
title: "ci: first green run of every CI job on GitHub runners"
labels: ["type:tache", "pkg:ci", "prio:haute", "taille:page"]
milestone: "v0.2.0 Go upgrade and tooling baseline"
assignee: PaulPerigault
---

## Context

The CI was written and every job was run locally through `task`, but no workflow has run on a GitHub runner yet. Windows and macOS (`go test -race`), the `image` job (Docker on the runner), `zizmor` and `codeql` are the ones most likely to show something the author's machine did not.

## Goal

`ci` is green on `main` and on a pull request, on the three operating systems, and what the runners revealed is written down.

## Acceptance criteria

- [ ] `gh run list --workflow ci.yml --branch main --limit 1 --json conclusion` reports `success`.
- [ ] The `test` job is green on `ubuntu-latest`, `macos-latest` and `windows-latest`.
- [ ] `zizmor` reports no finding, or each finding is fixed in the workflow.
- [ ] `docs/journal/` has a dated note listing every fix the first runs needed.

## Package

ci

## Size

page

## Priority

haute

## Milestone

v0.2.0 Go upgrade and tooling baseline

## Out of scope

Adding jobs. Changing a lint limit or a coverage threshold to get a green run.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
