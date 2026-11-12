# infra: OpenTofu configuration

Deploys the service to Cloud Run: the service, its runtime service account, the read-only IAM binding it needs, and
the invoker bindings that decide who may call it. Nothing here is applied by CI or by a coding agent.

## What is checked without credentials

```bash
task infra:tool       # installs the pinned OpenTofu into .tools/bin, SHA-256 verified, fails on a version mismatch
task infra:fmt        # tofu fmt -check
task infra:lint       # tflint, bundled ruleset
task infra:validate   # tofu init -backend=false, tofu validate
task infra:test       # tofu test with the mocked Google provider: no credential, no network
task infra:scan       # Trivy misconfiguration scan, pinned scanner image (needs Docker)
```

`task verify` runs the first four. CI runs all of them in the `infra` job.

## Pins

| What | Where | Checked by |
| --- | --- | --- |
| OpenTofu `1.13.9` | `.opentofu-version`, digests in `opentofu.sha256` | `task infra:tool` (download refused on a digest mismatch, check fails closed) |
| Provider `hashicorp/google` `8.11.0` | `versions.tf`, hashes for five platforms in `.terraform.lock.hcl` | `tofu init` |
| tflint `0.66.0` | `tools/go.mod` | `go mod verify` |
| Trivy `0.76.2` | `Taskfile.yml`, image digest | Docker |

The lock file is generated, never edited by hand (`tofu providers lock -platform=...`).

## Variables

| Variable | Default | Meaning |
| --- | --- | --- |
| `project_id` | required | GCP project |
| `image` | required | `registry/path:tag` or `@sha256:` digest, never `:latest` |
| `region` | `us-central1` | Region of the service and of the listing |
| `service_name`, `service_account_id` | `geturlcloudrun`, `geturlcloudrun-sa` | Names |
| `invoker_members` | `[]` | IAM members that may call the service (`roles/run.invoker`) |
| `allow_unauthenticated` | `false` | Grants `allUsers` the invoker role: the service becomes public |
| `allowed_origins` | `[]` | CORS origins passed to `ALLOWED_ORIGINS` |
| `min_instances`, `max_instances` | `0`, `3` | Scaling bounds |
| `deletion_protection` | `true` | The service cannot be deleted until this is set to `false` in a separate change |

## State

The state lives in a GCS bucket outside this repository, with object versioning on. The backend is a partial
configuration (`backend "gcs" {}`): copy `backend.hcl.example` to `backend.hcl` (ignored by Git) and
`tofu init -backend-config=backend.hcl`. State and plan files are never committed (`*.tfstate*`, `*.tfplan`,
`plan.json` are ignored). To encrypt the state client side, set the `TF_ENCRYPTION` environment variable with
OpenTofu's `pbkdf2` key provider; the passphrase is not stored anywhere in the repository.

## Deploying by hand

This is the owner's procedure. It needs credentials, so it is outside what `task verify`, CI and the coding agents
do. Build and push the image to your registry first (the registry is not managed here).

1. `tofu init -backend-config=backend.hcl`
2. `tofu plan -var project_id=... -var image=... -out=plan.bin`, then read the whole plan.
3. `tofu show -json plan.bin > plan.json`, then `task infra:plan:check PLAN=infra/plan.json`. It fails if the plan
   destroys or replaces anything that is not listed in `allowed-destroys.txt` (empty on purpose).
4. Announce the change, then `tofu apply plan.bin`.
5. Call the service with an identity token: `curl -H "Authorization: Bearer $(gcloud auth print-identity-token)" "$(tofu output -raw service_url)/health"`.

To remove the service: first a change that sets `deletion_protection = false`, applied; then the destroy, with the
address added to `allowed-destroys.txt` in the same pull request.

## What is not done here

- The Artifact Registry repository and the image push are not managed.
- The state bucket is not created by this configuration.
- Workload Identity Federation for a deployer identity is a backlog item (`.github/backlog`).
