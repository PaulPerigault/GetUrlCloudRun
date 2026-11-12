# Contributing

Thanks for looking. This is a small project with a strict method; the rules are written once in
[AGENTS.md](AGENTS.md) and checked by CI.

## Set up

```bash
git clone https://github.com/PaulPerigault/GetUrlCloudRun.git
cd GetUrlCloudRun
go install github.com/go-task/task/v3/cmd/task@latest   # or: go tool -modfile=tools/go.mod task
task setup     # modules, .env, pinned tools check, Git hooks (Lefthook)
task verify    # everything CI runs
```

Go 1.27 or newer (the toolchain is pinned in `go.mod`). Docker is needed for `task docker:scan` and
`task infra:scan`. OpenTofu is installed by `task infra:tool` into `.tools/bin`, with its SHA-256 checked.
On Windows, Git for Windows provides the POSIX tools some targets use.

## Workflow

1. One issue per change, created from a form (`.github/ISSUE_TEMPLATE/`), with labels, a milestone and
   acceptance criteria. Blank issues are disabled.
2. Branch `<type>/<issue>-<slug>`, for example `feat/42-pagination`.
3. Commits follow Conventional Commits: `type(scope): lower-case subject`, 72 characters at most,
   types `feat fix docs style refactor test chore ci perf revert`. No tool signature
   (`Co-authored-by`, "Generated with"). The `commit-msg` hook checks it.
4. Open a PR to `main` with `Closes #n`. CI must be green (`ci` is the required check). Merge is squash.
5. `task verify` passes locally before you push: it is the `pre-push` hook.

## What a good change looks like

- A test for each behaviour, table-driven, next to the code. The public network is never used in tests: the
  Google client is tested against an in-process gRPC server, everything else against the fake provider.
- Limits from AGENTS.md hold: functions of 40 lines at most, `gocyclo` 15, 4 parameters, package direction
  enforced by `depguard`, 85 percent coverage per `internal/` package. Do not raise a limit to make a build pass.
- Docs move with the code: `task check:docs` fails if a documented task, command or path is gone, and
  `task docs:readme` regenerates the sample output of the README.
- Infrastructure changes touch `infra/` and its tests (`task infra:test`). Nothing in a pull request applies
  anything: the owner deploys by hand, see [infra/README.md](infra/README.md).

## Security

Never open an issue for a vulnerability: see [SECURITY.md](SECURITY.md). Do not add keys, tokens, project IDs
or service URLs of real systems to the repository, including in tests and fixtures.
