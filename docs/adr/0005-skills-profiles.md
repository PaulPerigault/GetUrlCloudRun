# ADR 0005: skill profiles for this repository

Status: accepted (2026-11-12)

## Context

The workspace keeps a catalogue of vendorized skills with a lock file and named profiles. A repository gets `core`
plus the profiles of its stack, at most 12 skills auto-triggered. This repository is Go plus OpenTofu, with a README
and ADRs; it has no web interface.

## Decision

Copy these profiles, unmodified, with the workspace skills sync tool (checked here by `task skills:verify`, part of `task verify`):

- `core`: `conventional-git`, `encre-processus`, `encre-securite` (auto), `encre-veille` (manual).
- `go`: `golang-code-style`, `golang-concurrency`, `golang-context`, `golang-error-handling`, `golang-lint`,
  `golang-testing` (auto). `golang-project-layout` and `golang-security` stay out of the profile, as in the
  reference repository: the first writes to agent configuration files without asking, the second pre-approves
  network and write tools on every Go file.
- `docs`: `encre-voix` (auto), `copy-editing` and `humanizer` (manual).
- `iac`: `terraform-skill`, **manual only**. It was read completely on 2026-11-12 (see `skills/AUDIT.md`, section 9):
  no hidden text and no instruction aimed at the agent, but it contains an installer piped to a shell, `@latest`
  and `@master` references, unpinned actions, and `apply`, `import` and state commands that this repository denies.
  It is useful for variable contracts, test structure and naming, and is loaded by name for those.
- Not added: `ux-ui`, `astro`, `rust`, `angular`, `marketing`, `copy-fr-en` (there is no web part, no Rust, no
  French copy here).

That is 10 auto-triggered skills, under the limit of 12.

## Consequences

- The hierarchy holds: repository rules, house skills, `encre-*`, third-party. The skill's snippets are never
  followed where they contradict the settings or AGENTS.md; `AGENTS.md` says so.
- `.claude/skills` is excluded from `check:docs` and allow-listed in gitleaks, because it is third-party text and
  a lock of hashes.

## Alternatives rejected

- Exclude `terraform-skill`: loses a useful reference for a repository with a real OpenTofu part.
- Load it automatically: its apply-in-CI and state-surgery advice would sit next to the rules that forbid them.

## Main risk and mitigation

Risk: an agent follows a snippet of `terraform-skill` that contradicts the rules. Mitigation: `manual` trigger,
the explicit list of forbidden snippets in the audit, the `deny` rules in `.claude/settings.json` (a denial is a stop).
