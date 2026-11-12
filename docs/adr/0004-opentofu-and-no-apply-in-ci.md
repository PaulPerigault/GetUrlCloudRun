# ADR 0004: OpenTofu, tested offline, never applied from CI

Status: accepted (2026-11-12)

## Context

The first version had Terraform files and a `deploy.yml` that built an image, then ran
`terraform apply -auto-approve` on every push to `main`. Three problems:

- The Terraform was invalid: interpolations had been stripped (`account_id = "-sa"`, `member = "serviceAccount:"`),
  so it could not have planned.
- A push to `main` changed production with no reviewed plan and no approval, using a long-lived federated identity
  with write rights.
- Nothing validated the configuration before that point.

## Decision

- The infrastructure is OpenTofu in `infra/`, with an exact provider version and a committed lock file for five
  platforms. The OpenTofu patch version is pinned in `infra/.opentofu-version`. `dev/toolfetch` installs it into
  `.tools/bin` after checking the SHA-256 of the archive against `infra/opentofu.sha256`, and `task infra:*` fails
  if the installed version differs. No tool is downloaded without a fingerprint.
- CI validates only: `fmt`, `tflint` (bundled ruleset, no downloaded plugin), `validate`, `test` with
  `mock_provider "google"`, and a Trivy misconfiguration scan. It holds no cloud credential. `deploy.yml` is deleted.
- Deployment is manual, announced, and described in `infra/README.md`. A plan is reviewed by hand, then
  `task infra:plan:check` fails if it destroys or replaces something not listed in `infra/allowed-destroys.txt`.
  The service has `deletion_protection = true`.
- State is outside the repository (partial `gcs` backend), versioned; client-side encryption through `TF_ENCRYPTION`.
- Dependabot watches the provider but its pull requests are never auto-merged.

## Consequences

- Nothing in this repository can change a cloud project on its own.
- The first real plan is the owner's and has not been done: the tests prove the configuration's logic against a
  mocked provider, not the behaviour of Google's API.
- Shipping a new image is a manual step, so the deployed version can lag `main`.

## Alternatives rejected

- Keep Terraform: HashiCorp's licence change is the reason for OpenTofu, and the repository's standard is OpenTofu.
- A CI job with `plan` only, through Workload Identity Federation: needs a credential in CI and a decision about
  its scope; written as a decision issue (v0.7.0) rather than assumed.
- `apply` from CI behind an environment approval: still an apply from a pipeline, which the method forbids.

## Main risk and mitigation

Risk: manual deployment drifts from the repository. Mitigation: `plancheck` and a drift-check decision in the
backlog; the runbook is short and version controlled; `deletion_protection` guards the one irreversible action.
