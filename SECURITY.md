# Security policy

## Reporting a vulnerability

Do not open a public issue for a vulnerability. Use GitHub's
[private vulnerability reporting](https://github.com/PaulPerigault/GetUrlCloudRun/security/advisories/new)
on this repository. Expect an answer within a few days.

## Scope and design choices

- The service is read-only: it calls the Cloud Run Admin API to list and get services, nothing else. Its
  service account has `roles/run.viewer` and no other role.
- Authentication of callers is delegated to Cloud Run IAM. The deployed service is private unless
  `allow_unauthenticated` is set explicitly in `infra/`. The code has no password, token or API key.
- No credential file is read. Application Default Credentials come from the Cloud Run service account, or from
  `gcloud auth application-default login` on a workstation.
- Provider errors are logged on the server and never returned to the client. Every input is bounded: header
  size, timeouts, request rate, service name grammar.
- CORS is refused unless the exact origin is listed in `ALLOWED_ORIGINS`. The wildcard is rejected at start-up.
- The fake provider returns recorded, synthetic data. It is a test and demo tool; the deployed configuration
  sets `PROVIDER=gcp` and a test asserts it.
- CI never applies infrastructure and holds no cloud credential.

Only the latest release receives fixes.
