# akme

[![npm](https://img.shields.io/npm/v/akme.svg)](https://www.npmjs.com/package/akme)
[![npm downloads](https://img.shields.io/npm/dm/akme.svg)](https://www.npmjs.com/package/akme)
[![GitHub release](https://img.shields.io/github/v/release/o1x3/nx.svg)](https://github.com/o1x3/nx/releases/latest)
[![CI](https://github.com/o1x3/nx/actions/workflows/ci.yml/badge.svg)](https://github.com/o1x3/nx/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

`akme` is a personal developer CLI — git branch stats, an AI coding-harness token dashboard, and self-update. Formerly `nx`.

## Install

**npm / npx / bun**

```sh
npx akme token -i
npm install -g akme
bunx akme help
```

**curl installer** (writes `~/.local/bin/akme`, migrates an old `nx` binary if present)

```sh
curl -fsSL https://raw.githubusercontent.com/o1x3/nx/main/scripts/install.sh | sh
```

Override the install directory (must be user-writable so self-update works):

```sh
curl -fsSL https://raw.githubusercontent.com/o1x3/nx/main/scripts/install.sh | AKME_INSTALL_DIR="$HOME/bin" sh
```

Re-running the installer upgrades `akme`, removes legacy `nx` binaries (`~/.local/bin/nx`, `/usr/local/bin/nx`, and whatever `command -v nx` resolved to), and refreshes PATH markers from the old nx installer.

## Help

```sh
akme help                 # top-level commands
akme help git             # git subcommands
akme help git stat        # git stat details
akme help token           # full token reference
akme help token harness   # one token topic
akme help update          # self-update
```

Topics under `akme help token`: `harness`, `range`, `view`, `output`, `flags`, `env`, `exit`, `examples` (also `akme help token topics`). Domain-local forms work too: `akme git help`, `akme git help stat`, `akme token --help`, `akme update --help`.

## Commands

```sh
akme update
akme git stat [--jobs <n>] <folder> [folder...]
```

Example:

```sh
akme git stat gigauser gigauser-backend-prod the-exchange
```

What it does:

- Treats each folder as a path relative to your current directory.
- Fetches only the detected `origin` default branch.
- Auto-detects the remote default branch from `origin/HEAD`.
- Falls back to `origin/main` if default branch detection is unavailable.
- Checks multiple folders concurrently; set `--jobs <n>` to tune concurrency.
- Prints changed files, added lines, and removed lines for `<base>...HEAD`.

```sh
akme token [harness] [range] [view] [-i]
akme token [harness] [range] (json | quiet | compare)
```

A pastel terminal dashboard for your AI coding-harness token usage (alias: `akme tokens`). All arguments are positional and order-independent.

- Harness (default `all`): `claude` (`cc`, `claude-code`), `codex` (`cx`), `pi` (`pi.dev`, `pidev`), `cursor` (`cursor-ide`, `cursor-cli`, `cursor-agent`), `all` (`combined`, `everything`).
- Range (default all-time): `alltime` (`lifetime`), `30d` (`month`, `30`), `7d` (`week`, `7`).
- View (default `overview`): `overview`, `models`, `hours`, `punchcard`, `trend`, `topdays`, `weekday`, `cost`, `mix`.
- Output modes (bypass the card): `json` (NDJSON for `all`), `quiet` (`-q`, one prompt-safe line), `compare` (`vs`, side by side).
- Flags: `-i`/`--tui` interactive mode, `-h`/`--help`.

Examples:

```sh
akme token                    # combined card, all harnesses, all time
akme token codex 7d cost      # Codex spend, last 7 days
akme token claude punchcard   # Claude Code weekday × hour grid
akme token all json | jq -s   # NDJSON summary, one object per harness
akme token -i                 # interactive: ←/→ harness · tab/⇧tab views · 1/2/3 range · q quit
```

Data sources:

| Harness | Source | Tokens |
| --- | --- | --- |
| Claude Code | `~/.claude/projects/**/*.jsonl` and `~/.config/claude/projects/**/*.jsonl` (incl. subagents) | real (final streaming chunk) |
| Codex | `~/.codex/sessions/**` + `archived_sessions/` | real (`last_token_usage` / cumulative deltas) |
| pi.dev | `~/.pi/agent/sessions/**/*.jsonl` | real |
| Cursor (IDE + CLI) | `<config>/Cursor/User/globalStorage/state.vscdb` + `~/.cursor/chats/*/*/store.db` | dashboard usage API when logged in (input/output/cache); else bubble/meter/chars÷4 |

Overrides: `CLAUDE_CONFIG_DIR`, `CODEX_HOME`, `PI_AGENT_DIR` (comma-separated). Cursor prefers billed totals from the Cursor dashboard (session JWT in `state.vscdb`); local-only mode undercounts because cache and cumulatives are server-side. Set `AKME_TOKEN_CURSOR_LOCAL=1` to skip the dashboard. Cursor Auto resolves to the underlying local model when AgentKv / `usageData` records it; otherwise it stays "Auto".

Cursor **sessions/heatmap are machine-local** (`state.vscdb` does not sync across machines or accounts). **Billed token totals follow the logged-in Cursor account** via the dashboard API (personal + discovered team memberships). Empty dashboard replies keep local estimates instead of wiping them to 0. `<config>` is `~/Library/Application Support` on macOS, `~/.config` on Linux, and `%APPDATA%` on Windows. `akme token cursor json` includes a `cursor_status` object (`ide_found`, `cli_stores`, `auth_ok`, `dashboard_ok`, `hint`); empty, estimated, or zero-token Cursor output also prints a one-line stderr hint. Diagnose with `AKME_TOKEN_NO_CACHE=1`, `AKME_TOKEN_CURSOR_LOCAL=1`, or `AKME_CURSOR_SESSION_TOKEN`.

Environment:

- `AKME_BACKGROUND=light|dark` overrides terminal background detection.
- `AKME_TRUECOLOR=1` forces 24-bit colour (useful for capture tools). Piped output is plain text.
- `AKME_TOKEN_NO_CACHE=1` bypasses the on-disk aggregate cache for `akme token`.
- `AKME_TOKEN_CURSOR_LOCAL=1` forces local-only Cursor token counting (no dashboard API).
- `AKME_CURSOR_SESSION_TOKEN` / `CURSOR_SESSION_TOKEN` override the Cursor dashboard session JWT (or `sub::jwt`).
- `CLAUDE_CONFIG_DIR`, `CODEX_HOME`, `PI_AGENT_DIR` override harness data roots.

Legacy `NX_*` names for the same variables still work as fallbacks.

Exit codes: `0` ok · `2` bad arguments · `3` no usage for the selection (output modes only, so it composes in scripts and CI).

Interactive keys: `←`/`→` switch harness, `tab`/`⇧tab` cycle views, `1`/`2`/`3` set the range, `q` quits.

## Updates

Released builds check GitHub for a newer release on every invocation (in the background while your command runs). If one exists for your OS and CPU, `akme` downloads it and replaces the current binary in place.

```sh
akme update
```

Resolution uses `https://github.com/o1x3/nx/releases/latest` (same as the installer), not the GitHub API.

Disable background update checks:

```sh
AKME_NO_UPDATE=1 akme git stat .
```

`akme update` still runs when requested. Development builds with version `dev` do not auto-update. npm/`npx` installs disable binary self-update and should update via the package manager instead.

## Development

```sh
go test ./...
go run ./cmd/akme git stat .
```

Install local git hooks:

```sh
scripts/install-hooks.sh
```

The pre-commit hook runs `scripts/format.sh` first, then `scripts/check.sh`.

## Release Automation

Releases are driven by `VERSION` and published by GoReleaser, then the same tag is published to npm as [`akme`](https://www.npmjs.com/package/akme).

Update `VERSION`, keep `npm/package.json` version in sync, add a matching `CHANGELOG.md` section, and push to `main`. GitHub Actions creates tag `v<VERSION>`, publishes macOS/Linux `amd64`/`arm64` archives (`akme_<os>_<arch>.tar.gz`), and runs `npm publish` when the `NPM_TOKEN` repository secret is set.

```sh
printf '0.5.1\n' > VERSION
# sync npm/package.json version
# add ## 0.5.1 to CHANGELOG.md
git commit -am "release: v0.5.1"
git push origin main
```

## Design Notes

The command framework stays small. Routing is in `internal/cli`; domains own behaviour:

- `internal/gitstat` — git collection
- `internal/render` — terminal rendering
- `internal/token` — token dashboard (`core` / `ui` / `tui`)
- `internal/selfupdate` — release checks and `akme update`
- `internal/envx` — `AKME_*` env reads with `NX_*` fallbacks

No Cobra yet — reversible if the tree grows enough to justify it.
