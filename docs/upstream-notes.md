# Upstream notes

Things learned while industrializing this repository that belong to the workspace method, its templates or its
tools. The templates are owned elsewhere and were not edited: this file is the hand-over. Dates are 2026-11-12.

## Status (2026-11-12, workspace session)

Consumed by the workspace: the publication configuration of section 1 is now in the (private) publication
settings; the infrastructure variant, the OpenTofu and Terraform `deny` rules (now including `plan`), the
Dependabot rule for the infrastructure provider, the pinned Trivy digest, the Windows-safe format hook, the Go 1.27
tools module and a looser `Context` check are in the workspace templates, and this repository carries the same
changes. Kept as the record of what was learned.

## 1. Publication configuration, proposed (not written anywhere)

Repository description (one technical sentence, under the 350 characters GitHub allows):

> Read-only Go API that returns the URLs of Cloud Run services, with a deterministic fake provider for offline use and a tested OpenTofu deployment.

Topics: `go`, `golang`, `cloud-run`, `google-cloud`, `gcp`, `opentofu`, `infrastructure-as-code`, `rest-api`,
`distroless`, `devsecops`.

Patterns to check on the clean repository, working tree and history, before it becomes public. Regular
expressions are written so that this file does not match itself. Counters only, never the matched value.

| Category | Pattern or rule | Note |
| --- | --- | --- |
| Employer or client name | one line per name, kept in the publication configuration, never in a repository | the name is deliberately absent from this repository; the registry `ip_check` counts 0 items today |
| Personal e-mail | `[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+[.][a-z]{2,}`, allowing only the `users.noreply.github.com` form | exclude `.claude/skills/` (third-party text with `user@host` examples, covered by `skills:verify`) |
| Personal e-mail in commit metadata | author and committer fields of every commit | the registry records 1 commit with both fields exposed: the clean repository (one commit, noreply identity) is what removes it |
| WSL machine path | `/mnt/[a-z]/` | |
| Windows user path | `[A-Z]:[\\/]+Users[\\/]` | |
| School forge | `gitlab[.]esiea` | also the school forge URL form with any scheme |
| Workspace path | `[A-Z]:[\\/]+Clean_[G]ithub` | the folder name is written with a character class so that this file does not match itself |
| Private IPv4 | `\b(10\.|192\.168\.|172\.(1[6-9]\|2[0-9]\|3[01])\.)[0-9.]+\b` | registry `ip_check`: 0 today; `127.0.0.1` is loopback and allowed |
| Real GCP project IDs, service URLs, service account e-mails | review any `run.app` URL and any `.iam.gserviceaccount.com` string | the only ones in the tree are the synthetic `demo-project-1`, `*-demo-ew.a.run.app` and `caller@demo-project-1.iam...` |
| Secrets | gitleaks `dir` and `git` at 0, whole history | the original history had 0 findings (registry) |

Registry facts reused (`ip_check`, 2026-11-12): scope working tree and full history (1 commit); employer or client
items 0; real GCP project IDs 0; service accounts 0; internal URLs 0; the history had to be purged for the personal
e-mail in the commit metadata only.

## 2. Templates that should evolve

- **Infrastructure variant.** There is no template for a Go + OpenTofu repository. Reusable from this repository:
  the `infra:*` tasks in `Taskfile.yml`, the `infra` job of `ci.yml`, `dev/toolfetch` (pinned OpenTofu installed
  after a SHA-256 check, fail-closed version check), `dev/plancheck` (refuses a plan that destroys something not in
  `allowed-destroys.txt`), the `terraform` Dependabot ecosystem, the `infra/` layout with `tests/*.tftest.hcl`
  using `mock_provider`.
- **`.claude/settings.json` template has no OpenTofu or Terraform deny.** The playbook lists
  `apply|destroy|import|state|show|console`. A pattern such as `Bash(tofu apply*)` does not match
  `tofu -chdir=infra apply` nor `.tools/bin/tofu.exe ... apply`; this repository uses `Bash(*tofu* apply*)`.
  Also added: `gcloud`, `docker push`, state and plan files, `Edit` of `.terraform.lock.hcl`.
- **Possible hardening, needs the owner's agreement:** deny `tofu plan` for agents, and `gh pr create`,
  `gh issue create`, `gh pr merge`, since `encre-processus` suggests the last one (ADR 0006).
- **`dependabot-automerge.yml` merges the infrastructure provider.** The playbook says never. Here the workflow
  skips the label `pkg:infra`, which the `terraform` ecosystem entry sets.
- **`Taskfile.go.yml` `docker:scan` uses `aquasec/trivy:latest`.** Pin tag and digest (done here). The `ci.go.yml`
  `image` job uses `trivy-action`; here CI calls `task docker:scan` instead, so the local and CI scanner are the same.
- **`tools/go.mod` copied from the reference repository does not build `lefthook` with Go 1.27**:
  `github.com/go-json-experiment/json` at the 2025 pseudo-version fails (`json.SkipFunc` undefined). `go get
  github.com/go-json-experiment/json@latest` in `tools/` fixed it (v1.27.2). Check every tool with `--version`
  after any Go bump, as the playbook already advises.
- **`checkdocs` must understand `task infra:*`.** A documented task prefix with a wildcard was reported as an unknown
  task; `knownTask` in `dev/checkdocs/main.go` accepts a prefix when some task starts with it, with a test.
- **`templates/lefthook.yml` format hook fails on Windows.** Lefthook runs `sh -c "test -z "$(gofmt ...)" ..."` with the inner double quotes
  unescaped, so the shell stops with `unexpected EOF` (exit 2) on any staged Go file. Delegating to `task fmt:check`
  (Task's embedded shell handles the quoting) works; found by running the hooks, not by reading them.
- **`templates/dev/checkdocs` assumes a cobra CLI.** For a standard-library CLI, export the list of commands
  (`app.Commands`) and test that `Run` accepts each one, as done here.
- **Coverage gate.** `dev/covercheck` (per package, prefix `internal/`, minimum 85) is small and reusable;
  `task test` writes the profile then runs it.
- **The README `Context` check requires a period `YYYY-MM to YYYY-MM`.** A project whose original month is
  unknown can only satisfy it with the Git period. Here: `2026-04 to 2026-04` plus the sentence that the work
  dates from 2024. A looser form for `date_confidence: low` entries may be worth allowing.
- **A decompression bug worth a test in any download tool:** the first version of `dev/toolfetch` capped the extracted
  file at the archive limit (100 MB) and truncated the 150 MB binary silently. It now errors on overflow and has a test.
- **A full system drive stops every Go build** (link step: not enough space). Point `GOCACHE` and `GOTMPDIR` at a
  data drive before long sessions, and note it in the delivery report.

## 3. Catalogue

- `skills.lock.json`: the `terraform-skill` entry has no `audit_level`. After the full read of 2026-11-12
  (`skills/AUDIT.md`, section 9) it can be set to `full-read`, with the verdict "accepted with reserves, manual".
  This was not done here because existing entries are not modified from this repository's session.
- The sentence in `skills/AUDIT.md` section 8 that keeps `terraform-skill` out of every repository is superseded by
  section 9; the `iac` profile is loaded in this repository as manual.
- `skills:sync` copies `_licenses/` into `.claude/skills/`; `gitleaks` and `check:docs` both skip that directory.

## 5. Template `release.go.yml` (found 2026-11-12 by the first CI run on GitHub)

- `templates/.github/workflows/release.go.yml`: `actions/setup-go` has no `cache: false`. zizmor 1.32.0 reports
  `cache-poisoning` (severity high) on a tag-triggered release workflow, which fails the `zizmor` job and so `ci`.
  Fixed here and in `homelab-infra` with `cache: false`; the template and `rsa-ct-analyzer` still need the same line.
