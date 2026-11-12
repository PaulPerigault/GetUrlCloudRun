# ADR 0001: standard-library HTTP, no web framework

Status: accepted (2026-11-12)

## Context

The first version used Gin and `gin-contrib/cors`, with `AllowOrigins: ["*"]`. The service has three routes, all
`GET`, and no templating, binding or validation that a framework would help with. Gin and its dependencies
(`bytedance/sonic`, `go-playground/validator`, `ugorji/go`, and others) were most of the indirect lines of
`go.mod`; the pinned version was Gin v1.9.1. Since Go 1.22, `net/http` routes by method and path
(`GET /api/v1/services/{name}`), which is all this service needs.

## Decision

Use `net/http` from the standard library. Cross-cutting behaviour is a handful of small middlewares in
`internal/httpmw` (recover, access log, CORS allow-list, global token bucket, request deadline), each in its own
file with its own tests. `depguard` forbids `gin-gonic/gin` and `gin-contrib/cors` so they do not come back.

## Consequences

- The only direct dependencies are the Cloud Run client, `google.golang.org/api` and `grpc`, all in
  `internal/provider/gcp`.
- The middlewares and the error writer are ours to maintain: about 190 lines, 98 percent covered.
- CORS is deny-by-default and exact-match; there is no wildcard mode to misconfigure.

## Alternatives rejected

- Keep Gin: more code to audit and update than the service itself, for features it does not use.
- `chi` or `gorilla/mux`: a smaller dependency than Gin but still one, for a router the standard library now has.

## Main risk and mitigation

Risk: a hand-written middleware has a bug a mature one would not. Mitigation: each middleware is small, tested
with table tests (listed, unlisted and absent origins, preflight, empty bucket, refill cap, panic, query string
not logged), and `-race` runs in CI.
