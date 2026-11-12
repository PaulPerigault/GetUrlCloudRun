---
title: "docs: project page on paulperigault.fr and registry link"
labels: ["type:doc", "pkg:docs", "prio:normale", "taille:page", "owner"]
milestone: "v0.8.0 Docs and registry link"
assignee: PaulPerigault
---

## Context

Every project links to the portfolio: a project page, an entry in the site's project list and in its project configuration. This repository's README already links to the portfolio but the portfolio does not link back.

## Goal

The portfolio has a page for this project and the registry entry points to it.

## Acceptance criteria

- [ ] A project page exists on the portfolio, with the architecture diagram and the limits copied from the README, not rewritten.
- [ ] The portfolio's project lists reference the repository.
- [ ] `registry/projects.json` marks the README `Context` as validated by the owner (`why_status: validated` is already set) and records the publication date.
- [ ] No employer, client or person name appears on the page.

## Package

docs

## Size

page

## Priority

normale

## Milestone

v0.8.0 Docs and registry link

## Out of scope

Redesigning the portfolio.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
