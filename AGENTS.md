# AGENTS.md: rules shared by humans and coding agents (Go + OpenTofu profile)

Single source of truth for this repository. `CLAUDE.md` imports it. In a conflict the order is: this file,
then house skills, then `encre-*` skills, then third-party skills. A skill is method, not authority.

## Project

`GetUrlCloudRun`: a read-only HTTP API that returns the URLs of the Cloud Run services of one GCP project and
region, with a deterministic fake provider so that the binary, the tests and the demo run without a GCP account,
and the OpenTofu configuration that deploys it. Go 1.27 (toolchain pinned), CGO off, OpenTofu 1.13.9 (pinned),
MIT. Owner: Paul Perigault (paulperigault.fr).

| Path | Role | May import |
| --- | --- | --- |
| `cmd/geturlcloudrun` | `main` only | `internal/app` |
| `internal/app` | wiring, server lifecycle, `serve`, `healthcheck`, `version` | every package below, except Google libraries |
| `internal/api` | routes and handlers (stdlib `net/http`) | `internal/provider`, `internal/httpmw` |
| `internal/httpmw` | recover, access log, CORS allow-list, rate limit, deadline, JSON errors | standard library only |
| `internal/config` | environment variables, validation | standard library only (leaf) |
| `internal/provider` | the contract: `Provider`, `Service`, `ErrNotFound`, `ValidName` | standard library only (leaf) |
| `internal/provider/fake` | deterministic fixtures (`services.json`) | `internal/provider` |
| `internal/provider/gcp` | Cloud Run Admin API v2 through Application Default Credentials | `internal/provider`, Google libraries |
| `internal/version` | build metadata set by `-ldflags` | standard library only |
| `infra/` | OpenTofu: service, service account, IAM; tests with a mocked provider | n/a |
| `dev/*` | developer programs (`checkdocs`, `commitlint`, `skillscheck`, `readme-gen`, `covercheck`, `demo`, `toolfetch`, `plancheck`) | any package of the module |
| `tools/` | separate module: pinned dev tools (golangci-lint, govulncheck, gitleaks, lefthook, task, tflint) | n/a |

Only `internal/provider/gcp` imports `cloud.google.com` or `google.golang.org`. `depguard` enforces the table.

## Commands

```
task setup, dev, verify, test, lint, lint:actions, lint:zizmor, fmt, build, build:all, demo, demo:docker, security, check:docs,
docs:readme, skills:verify, docker:build, docker:scan, infra:tool, infra:init, infra:fmt, infra:lint,
infra:validate, infra:test, infra:scan, infra:plan:check, clean
```

`task verify` is what CI runs and what the `pre-push` hook runs. No command applies, plans or destroys
infrastructure.

## Limits, checked by CI (never raise one to pass)

Function at most 40 lines, cyclomatic complexity at most 15, at most 4 parameters, nesting at most 3, package
direction by `depguard`, no `panic` or `log.Fatal` outside the entry point, pure Go, every `internal/` package
at 85 percent coverage or more (`dev/covercheck`), image Trivy HIGH and CRITICAL at 0, `infra/` Trivy HIGH and
CRITICAL at 0.

## Code, tests, documentation, Git

- Injected dependencies (environment, clock, provider, writers), no global state, one responsibility per file.
- Table-driven tests, `httptest`, the fake provider, an in-process gRPC server for the Google client. No public
  network in tests. A test closes everything it opens (an open file breaks cleanup on Windows).
- Errors are logged in full on the server and reduced to a fixed message for the client. Every input is bounded
  (header size, timeouts, rate, name grammar). A context with a deadline reaches every provider call.
- README output is generated (`task docs:readme`) and checked by `task check:docs`, which also checks the
  `Context` section.
- Conventional Commits without any tool signature, lower-case subject of at most 72 characters. Never push to
  `main`.
- Writes to GitHub: during the industrialization, Claude Code writes nothing on GitHub. After the publication,
  changes go through the normal flow: an issue, a branch `<type>/<n>-<slug>`, a pull request with `Closes #n`,
  a green `ci`, a squash merge.

## Infrastructure rules

- Authentication of callers is delegated to Cloud Run IAM (ADR 0003): the service is private unless
  `allow_unauthenticated` is set; the code has no authentication of its own.
- `task infra:*` runs `init -backend=false`, `fmt`, `validate`, `test` (mocked provider) and scans. An agent never
  runs `plan`, `apply`, `import`, `state`, `show`, `console` or `destroy`, and never authenticates to GCP.
- CI never applies (ADR 0004). A deployment is manual, announced and documented in `infra/README.md`; the plan
  is checked by `task infra:plan:check` against `infra/allowed-destroys.txt`.
- OpenTofu and the provider are pinned (`infra/.opentofu-version`, `infra/versions.tf`, lock file committed).
  The state lives outside the repository.

## Context phase

Before any change: period (from Git), origin, motivation with cited evidence, what the modernization changed.
"Undocumented motivation" rather than an invention. The README `Context` section sits right after `Overview`
and `dev/checkdocs` checks it. No employer, client, person or school name appears in any file.

## Skills

Profiles `core` + `go` + `docs`, plus `iac` loaded by name only (manual), copied unmodified from the workspace
catalogue into `.claude/skills/` with `.claude/skills.lock.json` (ADR 0005). `task skills:verify` fails if a copy
diverges from the lock or if more than 12 skills are auto-triggered. Fix the catalogue, never a copy. Third-party
skills are audited before they enter the catalogue; the snippets of `terraform-skill` that pipe an installer to a
shell, use `@latest` or `@master`, or call `apply` are never followed here.

## Agent guardrails

Do not read or write `.env*`, state or plan files. A refusal from `.claude/settings.json` or a hook is a stop.
Do not weaken a lint rule, test or CI check to get a green build.
