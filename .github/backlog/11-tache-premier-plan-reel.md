---
title: "docs(infra): first real plan and a written deployment runbook"
labels: ["type:tache", "pkg:infra", "prio:haute", "taille:page", "owner"]
milestone: "v0.6.0 OpenTofu infrastructure"
assignee: PaulPerigault
---

## Context

The configuration is validated and tested with a mocked provider, never planned against a real project: the provider schema, the IAM bindings and the probe settings have only been checked by `tofu validate`. Planning needs credentials, so it is the owner's action.

## Goal

The owner has run `tofu plan` on a sandbox project, read it, and the procedure in `infra/README.md` is corrected where reality differed.

## Acceptance criteria

- [ ] `tofu plan -out=plan.bin` succeeds on a sandbox project; `task infra:plan:check PLAN=infra/plan.json` passes.
- [ ] A sanitized excerpt of the plan (counts of resources to add, no project ID) is added to `docs/journal/`.
- [ ] `infra/README.md` is updated with every correction; `task check:docs` passes.
- [ ] Nothing from the plan or the state is committed (`.gitignore` covers it).

## Package

infra

## Size

page

## Priority

haute

## Milestone

v0.6.0 OpenTofu infrastructure

## Out of scope

Applying. Workload Identity Federation. Creating the state bucket or the registry.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
