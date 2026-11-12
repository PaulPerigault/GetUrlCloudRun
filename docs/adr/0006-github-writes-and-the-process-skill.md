# ADR 0006: writes to GitHub, and the `encre-processus` skill

Status: accepted (2026-11-12)

## Context

The `core` profile includes `encre-processus`, which describes a delivery flow with `gh pr create` and
`gh pr merge`. The workspace rules say that during an industrialization Claude Code writes nothing on GitHub:
the only write path to GitHub is the owner's publication tool. `AGENTS.md` ranks above skills, and the skill is
vendorized and must not be edited.

## Decision

Leave the skill unmodified and settle the conflict in `AGENTS.md`, by the hierarchy:

- During the industrialization, and until the repository is published, Claude Code does not write to GitHub.
  `.claude/settings.json` denies the risky `gh` writes (`gh repo edit|delete`, `gh release`, `gh api` writes,
  `gh secret|variable`, `gh pr merge --admin`) and any push to `main`, force push or tag.
- After the publication, changes follow the normal flow: an issue from a form, a branch `<type>/<n>-<slug>`, a pull
  request with `Closes #n`, a green `ci`, a squash merge. Where the skill and `AGENTS.md` differ, `AGENTS.md` wins.

## Consequences

- No fork of a third-party skill and no drift against the catalogue.
- The deny list covers some of what the skill suggests, not all (`gh pr create` is not denied here). The rule is
  carried by `AGENTS.md` and by the publication guard of the workspace, not by this file alone. Hardening the deny
  list needs the owner's agreement, so it is proposed in `docs/upstream-notes.md` rather than done.

## Alternatives rejected

- Edit the skill: breaks the "never modified" rule and its lock.
- Remove `encre-processus` from the profile: it also carries rules on branches and reviews that apply here.

## Main risk and mitigation

Risk: an agent follows the skill and writes to GitHub. Mitigation: the hierarchy sentence at the top of
`AGENTS.md`, the deny rules, and the workspace guard that refuses publication commands outside the authorized tasks.
