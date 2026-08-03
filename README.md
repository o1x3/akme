# akme

<p align="center">
  <a href="https://www.npmjs.com/package/akme-cli"><img src="https://img.shields.io/npm/v/akme-cli?style=for-the-badge&logo=npm&logoColor=white&color=CB3837" alt="npm version"></a>
  <a href="https://www.npmjs.com/package/akme-cli"><img src="https://img.shields.io/npm/dm/akme-cli?style=for-the-badge&logo=npm&logoColor=white&color=CB3837" alt="npm downloads"></a>
  <a href="https://github.com/o1x3/akme/releases/latest"><img src="https://img.shields.io/github/v/release/o1x3/akme?style=for-the-badge&logo=github&logoColor=white&color=181717" alt="GitHub release"></a>
</p>

<p align="center">
  <a href="https://github.com/o1x3/akme/actions/workflows/release.yml"><img src="https://img.shields.io/github/actions/workflow/status/o1x3/akme/release.yml?branch=main&style=for-the-badge&logo=githubactions&logoColor=white&label=CI" alt="CI"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go 1.25"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge" alt="MIT License"></a>
  <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey?style=for-the-badge" alt="macOS and Linux">
</p>

<p align="center"><b>personal developer CLI</b></p>

<p align="center">git contribution graph · branch stats · coding-agent token usage · self-update</p>

## Install

npm (installs the CLI; does not run a command):

```sh
npm install -g akme-cli
# same: install or update CLI to latest GitHub release + greeting
npx akme-cli
```

Then:

```sh
akme help
akme token -i
```

Bare `npx akme-cli` / `npm install -g` download (or update) the native binary from GitHub releases. If a legacy `nx` install of this CLI is on your machine, they also migrate it to `akme` on your PATH (same cleanup as the curl installer). Pass args to run the CLI (`npx akme-cli token`). Set `AKME_NPM_QUIET=1` to skip the greeting.

curl (installs to `~/.local/bin/akme`):

```sh
curl -fsSL https://raw.githubusercontent.com/o1x3/akme/main/scripts/install.sh | sh
```

Custom install dir (must be user-writable so self-update works):

```sh
curl -fsSL https://raw.githubusercontent.com/o1x3/akme/main/scripts/install.sh | AKME_INSTALL_DIR="$HOME/bin" sh
```

## Help

```sh
akme help                 # top-level commands
akme help git             # git overview
akme help git activity    # contribution dashboard
akme help git view        # overview / authors / punchcard / …
akme help git stat        # multi-folder branch diff
akme help token           # full token reference
akme help token harness   # one token topic
akme help update          # self-update
```

Token topics: `harness`, `range`, `view`, `output`, `flags`, `env`, `exit`, `examples` (or `akme help token topics`).

Same pages via domain help: `akme git help`, `akme git help activity`, `akme git help stat`, `akme token --help`, `akme update --help`.

## Commands

### `akme git <folder>`

```sh
akme git <folder> [range] [view] [-i]
```

```sh
akme git .
akme git . -i
akme git . 30d punchcard
```

GitHub-style contribution dashboard for one local clone. Counts **first-parent** commits on the detected default branch (`origin/HEAD`, else `origin/main`). Soft-fetches that branch only; continues on local refs if fetch fails. Collects ~52 weeks with one `git log` (no GitHub API).

| Arg | Default | Values |
| --- | --- | --- |
| range | `year` | `year` (`52w`, `alltime`), `30d` (`month`, `30`), `7d` (`week`, `7`) |
| view | `overview` | `overview`, `authors`, `hours`, `punchcard`, `trend`, `topdays`, `weekday`, `branch` |

Flags: `-i` / `--tui` for interactive mode (`tab` views · `1`/`2`/`3` range · `q` quit).

### `akme git stat`

```sh
akme git stat [--jobs <n>] <folder> [folder...]
```

```sh
akme git stat gigauser gigauser-backend-prod the-exchange
```

Each folder is relative to your cwd. Fetches the remote default branch on `origin` (from `origin/HEAD`, else `origin/main`), then prints changed files / added / removed for `<base>...HEAD`. Folders run concurrently; `--jobs <n>` sets the limit.

### `akme token`

```sh
akme token [harness] [range] [view] [-i]
akme token [harness] [range] (json | quiet | compare)
```

Terminal dashboard for coding-agent token usage (alias: `akme tokens`). Args are positional and order-independent.

| Arg | Default | Values |
| --- | --- | --- |
| harness | `all` | `claude` (`cc`, `claude-code`), `codex` (`cx`), `pi` (`pi.dev`, `pidev`), `cursor` (`cursor-ide`, `cursor-cli`, `cursor-agent`), `all` (`combined`, `everything`) |
| range | all-time | `alltime` (`lifetime`), `30d` (`month`, `30`), `7d` (`week`, `7`) |
| view | `overview` | `overview`, `models`, `hours`, `punchcard`, `trend`, `topdays`, `weekday`, `cost`, `mix` |

Output modes (skip the card): `json` (NDJSON for `all`), `quiet` / `-q` (one prompt-safe line), `compare` / `vs` (side by side).

Flags: `-i` / `--tui` for interactive mode, `-h` / `--help`.

```sh
akme token                    # all harnesses, all time
akme token codex 7d cost      # Codex spend, last 7 days
akme token claude punchcard   # Claude Code weekday × hour grid
akme token all json | jq -s   # NDJSON, one object per harness
akme token -i                 # ←/→ harness · tab/⇧tab views · 1/2/3 range · q quit
```

#### Data sources

| Harness | Source | Tokens |
| --- | --- | --- |
| Claude Code | `~/.claude/projects/**/*.jsonl`, `~/.config/claude/projects/**/*.jsonl` (incl. subagents) | real (final streaming chunk) |
| Codex | `~/.codex/sessions/**`, `archived_sessions/` | real (`last_token_usage` / cumulative deltas) |
| pi.dev | `~/.pi/agent/sessions/**/*.jsonl` | real |
| Cursor (IDE + CLI) | `<config>/Cursor/User/globalStorage/state.vscdb`, `~/.cursor/chats/*/*/store.db` | dashboard usage API when logged in (input/output/cache); else bubble/meter/chars÷4 |

Path overrides: `CLAUDE_CONFIG_DIR`, `CODEX_HOME`, `PI_AGENT_DIR` (comma-separated).

Cursor prefers billed totals from the Cursor dashboard (session JWT in `state.vscdb`). Local-only mode undercounts because cache and cumulatives live server-side. Set `AKME_TOKEN_CURSOR_LOCAL=1` to skip the dashboard. Cursor Auto resolves to the underlying local model when AgentKv / `usageData` has it; otherwise it stays "Auto".

Cursor sessions and heatmaps are machine-local (`state.vscdb` does not sync). Billed token totals follow the logged-in Cursor account via the dashboard API (personal + discovered team memberships). Empty dashboard replies keep local estimates instead of wiping them to 0.

`<config>` is `~/Library/Application Support` on macOS, `~/.config` on Linux, and `%APPDATA%` on Windows.

`akme token cursor json` includes a `cursor_status` object (`ide_found`, `cli_stores`, `auth_ok`, `dashboard_ok`, `hint`). Empty, estimated, or zero-token Cursor output also prints a one-line stderr hint. Diagnose with `AKME_TOKEN_NO_CACHE=1`, `AKME_TOKEN_CURSOR_LOCAL=1`, or `AKME_CURSOR_SESSION_TOKEN`.

#### Environment

| Variable | Effect |
| --- | --- |
| `AKME_BACKGROUND=light\|dark` | override terminal background detection |
| `AKME_TRUECOLOR=1` | force 24-bit colour (useful for capture tools); piped output is plain text |
| `AKME_TOKEN_NO_CACHE=1` | bypass on-disk aggregate cache |
| `AKME_TOKEN_CURSOR_LOCAL=1` | local-only Cursor counting (no dashboard API) |
| `AKME_CURSOR_SESSION_TOKEN` / `CURSOR_SESSION_TOKEN` | override Cursor dashboard session JWT (or `sub::jwt`) |
| `CLAUDE_CONFIG_DIR`, `CODEX_HOME`, `PI_AGENT_DIR` | override harness data roots |

Exit codes: `0` ok, `2` bad arguments, `3` no usage for the selection (output modes only, so scripts/CI can compose on it).

Interactive keys: `←`/`→` switch harness, `tab`/`⇧tab` cycle views, `1`/`2`/`3` set range, `q` quits.

### `akme update`

```sh
akme update
```

Released builds check GitHub for a newer release on every invocation (background, while your command runs). If one exists for your OS/CPU, `akme` downloads it and replaces the binary in place.

Resolution uses `https://github.com/o1x3/akme/releases/latest` (same as the installer), not the GitHub API.

```sh
AKME_NO_UPDATE=1 akme git stat .
```

`akme update` still runs when you ask. Dev builds (`version=dev`) skip auto-update. npm/`npx` installs disable binary self-update; update those via the package manager.

## Development

```sh
go test ./...
go run ./cmd/akme git .
go run ./cmd/akme git stat .
```

Install local git hooks:

```sh
scripts/install-hooks.sh
```

Pre-commit runs `scripts/format.sh`, then `scripts/check.sh`.

## Release

Releases are driven by `VERSION`. GoReleaser publishes GitHub assets, then the same tag goes to npm as [`akme-cli`](https://www.npmjs.com/package/akme-cli). The npm package downloads the matching release binary on install / bare `npx akme-cli`.

Bump `VERSION`, sync `npm/package.json`, add a `CHANGELOG.md` section, push to `main`. One Actions workflow (`release.yml`) runs tests, then on `VERSION` bumps tags `v<VERSION>`, ships macOS/Linux archives, and publishes `akme-cli` via Trusted Publishing (OIDC, no `NPM_TOKEN`). Trusted publisher: GitHub `o1x3` / `akme` / workflow **`release.yml`**.

```sh
printf '0.6.1\n' > VERSION
# sync npm/package.json version
# add ## 0.6.1 to CHANGELOG.md
git commit -am "release: v0.6.1"
git push origin main
```

## Layout

Routing lives in `internal/cli`. Domains own the work:

- `internal/gitstat`: activity dashboard (`ui` / `tui`) + branch diff collect
- `internal/render`: terminal rendering (git stat table)
- `internal/token`: token dashboard (`core` / `ui` / `tui`)
- `internal/selfupdate`: release checks and `akme update`
- `internal/envx`: `AKME_*` env reads

No Cobra yet. Fine until the tree needs it.
