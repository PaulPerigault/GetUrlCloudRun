# ADR 0003: authentication is delegated to Cloud Run IAM

Status: accepted (2026-11-12)

## Context

The service exposes the names and URLs of the services of a project, which is internal information. The first
version had no authentication and CORS open to every origin. Cloud Run already authenticates callers: a service
that is not public requires a Google-signed identity token with the `roles/run.invoker` role, checked before the
request reaches the container.

## Decision

The code has no authentication: no key, no token parsing, no user store. The access policy is infrastructure:

- The OpenTofu configuration creates the service private. `invoker_members` grants `roles/run.invoker` to named
  identities. `allow_unauthenticated = true` is the only way to bind `allUsers`, and it is a separate, visible
  switch, off by default (tests cover both states).
- The runtime service account has `roles/run.viewer` on the project and nothing else.
- Browser access is a separate question answered by `ALLOWED_ORIGINS`, exact origins only, refused by default.
  A browser still needs a token for a private service, so a front end calls it through its own backend.

## Consequences

- No credential to leak, rotate or review in this repository.
- Local runs against a real project use Application Default Credentials; against the fake provider nothing is needed.
- The README and `infra/README.md` show how to call the deployed service with an identity token.
- An unauthenticated deployment is possible but explicit and reviewable in a plan.

## Alternatives rejected

- An API key or JWT check in the code: duplicates what the platform does, adds secrets and code to audit.
- IP allow-list: Cloud Run's front end hides the client address, and the list would be easy to get wrong.

## Main risk and mitigation

Risk: someone sets `allow_unauthenticated` to ease a demo and forgets. Mitigation: the plan shows an `allUsers`
binding, `task infra:plan:check` exists to review plans, the variable is documented as making the service public,
and the project page states the default.
