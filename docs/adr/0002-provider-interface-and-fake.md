# ADR 0002: a provider interface and a deterministic fake

Status: accepted (2026-11-12)

## Context

The first version called the Cloud Run client directly from the handlers, with `*gcp.Client` as a concrete type.
The only test checked `/health` with a `nil` client. Nothing about listing or getting a service could be tested
or demonstrated without a GCP account, and `GetServiceByName` listed every service to find one.

## Decision

`internal/provider` defines the contract (`Provider` with `ListServices`, `GetService`, `Name`; `Service`;
`ErrNotFound`; `ValidName`) and depends on the standard library only. Two implementations:

- `provider/gcp`: the Cloud Run Admin API v2, Application Default Credentials. `GetService` is one API call.
- `provider/fake`: recorded, synthetic fixtures embedded in the binary (`services.json`), sorted, copied on read.

`PROVIDER=fake|gcp` selects one at start-up. The default is `gcp`, so a missing configuration stops the service
instead of silently serving invented data in production; `.env.example` sets `fake` for local work. The Google
client is tested against an in-process gRPC server that implements the same protocol on loopback, so the real
client code runs without Google. A test asserts that the OpenTofu configuration sets `PROVIDER=gcp`.

## Consequences

- `task demo`, `task dev`, the container demo and every handler test run offline and deterministically.
- The handlers never see a Google type; `depguard` limits Google imports to `provider/gcp`.
- The fake can drift from the real API. The in-process server and the shared `provider.ValidName` limit it; a
  provider contract test shared by both implementations is a backlog item (v0.3.0).

## Alternatives rejected

- Mock the Google client in the handler tests only: no offline demo, and the client code stays untested.
- A `gcloud` or HTTP stub server: heavier than an interface and tests a transport the code does not own.
- Default to `fake`: friendlier, but a misconfigured deployment would answer with fixtures. Fail closed wins.

## Main risk and mitigation

Risk: `/health` reports `provider: fake` in production by mistake. Mitigation: the default is `gcp`; the start-up
log line and `/health` both name the provider; the OpenTofu test fails if the deployed configuration is not `gcp`.
