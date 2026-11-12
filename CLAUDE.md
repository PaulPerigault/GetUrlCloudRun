@AGENTS.md

# Claude Code notes

- Start with `task test` to know the baseline, finish with `task verify` (it includes `skills:verify` and the
  `infra:*` checks).
- The skills of `.claude/skills/` are copies of the workspace catalogue: never edit them here. `terraform-skill`
  is manual: load it by name for OpenTofu work, and ignore its `apply`, `@latest`, `@master` and piped-installer
  snippets.
- Begin every industrialization with the Context phase (period, origin, motivation, what changed), then fill
  the README `Context` section.
- Publishing is the owner's decision: Claude Code prepares, the owner runs the publication task.
- `.claude/settings.json` denies the risky commands (including OpenTofu `apply|destroy|import|state|show|console`
  and `gcloud`); a denial is a stop, not an obstacle to route around.
- No tool signature in commits or pull requests (the settings leave `attribution` empty).
- If a template of the workspace should change, note it in `docs/upstream-notes.md` instead of editing it.
