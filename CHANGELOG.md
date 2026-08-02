# Changelog

## 0.6.6

- Bare `npx akme-cli` / `npm install -g akme-cli` auto-migrate a legacy `nx` install: copy `akme` onto the user bindir (`~/.local/bin` or `AKME_INSTALL_DIR`), remove the old `nx` binary, and refresh installer PATH markers. Unrelated `nx` tools (e.g. Nrwl) are left alone via version/binary fingerprinting.

## 0.6.5

- Install greeting status lines distinguish install vs update vs already up to date.

## 0.6.4

- Restore the pink/white/dark-red akme ANSI wordmark from the HTML art (truecolor, no bogus spaces).

## 0.6.3

- Replace garbled bg-fill ANSI dump with a clean pink `akme` block wordmark.

## 0.6.2

- Bare `npx akme-cli`: show ASCII greeting first, then install/update the native CLI (never run Go usage/help).
- Expose `bin.akme-cli` alongside `bin.akme` so `npx akme-cli` resolves the wrapper explicitly.
- Single Actions workflow (`release.yml`): test job then release job (only when `VERSION` changes on `main`).

## 0.6.1

- Single npm package **`akme-cli`** only (dropped `@o1x3/akme-*` platform optionalDependencies). Install downloads the native binary from the GitHub release.
- Bare `npx akme-cli` / `npm install -g akme-cli` install or **update** the CLI to the latest GitHub release (checksum-verified), then show the install greeting.
- Release workflow publishes only `akme-cli` via Trusted Publishing.

## 0.6.0

- npm meta package is **`akme-cli`** (unscoped; `@o1x3/akme` / bare `akme` blocked by name similarity).
- Faster install experiment via platform `optionalDependencies` (superseded in 0.6.1 by GitHub-only install).
- Bare `npx` / `npm install -g` share install path with ASCII greeting.

## 0.5.1

- Rewrote the npm package README (`npm/README.md`) with matching badges and shorter install-focused copy for the npmjs.com page.

## 0.5.0

- Renamed the CLI from `nx` to **`akme`** (binary, help, installer, release archives `akme_<os>_<arch>.tar.gz`, cache under `~/.cache/akme`).
- GitHub repo / Go module moved to [`o1x3/akme`](https://github.com/o1x3/akme).
- Env vars are now `AKME_*` (`AKME_NO_UPDATE`, `AKME_BACKGROUND`, `AKME_TRUECOLOR`, `AKME_TOKEN_*`, `AKME_CURSOR_SESSION_TOKEN`, `AKME_CACHE_DIR`, `AKME_INSTALL_DIR`); legacy `NX_*` names still work as fallbacks.
- Installer installs `~/.local/bin/akme` and migrates old `nx` installs (removes `~/.local/bin/nx`, `/usr/local/bin/nx`, and whatever `command -v nx` resolved to; refreshes PATH markers).
- npm package `@o1x3/akme` downloads the renamed release assets (unscoped `akme` blocked by npm name-similarity); README rewritten with npm/GitHub badges.

## 0.4.6

- Added npm package `@o1x3/akme` so you can run the CLI via `npx @o1x3/akme …`, `npm install -g @o1x3/akme`, or `bunx @o1x3/akme …` (downloads the matching GitHub release binary with checksum verification).
- Release workflow publishes `@o1x3/akme` to npm on `VERSION` bumps via Trusted Publishing.

## 0.4.5

- Fixed Cursor dashboard auth cookie id: strip WorkOS connection prefixes (`github|user_…` → `user_…`) so session JWTs authenticate on every machine.
- Discover Cursor team memberships (`/api/dashboard/teams`) and prefer the teamId that returns billed events (team plans often look empty at `teamId=0`).
- Empty dashboard event lists no longer wipe local token estimates to 0 while keeping sessions/messages — the cross-machine "No tokens recorded yet" failure mode.
- Stronger stderr / `cursor_status` hints when Cursor shows activity with 0 billed tokens.

## 0.4.4

- Fixed `nx token -i` TUI alignment: pad the card to a uniform width before centering so month labels, the tab strip, and the logo no longer drift relative to full-width rows.
- Fixed contribution-graph month labels to sit on the week that contains day 1 (with collision deferral), and pad the month row to the heatmap width.
- Vertically centered the banner logo beside the stats column; padded the tab strip to the card width.
- Cursor now probes the Windows `%APPDATA%\Cursor\User\globalStorage\state.vscdb` path (in addition to macOS/Linux).
- `nx token cursor json` includes `cursor_status` (IDE path, CLI stores, auth, dashboard); empty or estimated Cursor output prints a stderr hint. Documented that sessions are machine-local while billed tokens follow the logged-in account.

## 0.4.3

- Fixed `nx token` banner stats alignment: the key·value info column now spans to the card's right edge (same as cost amounts and the footer), so values right-align cleanly and long fav-model names keep a proper dotted leader.

## 0.4.2

- Hardened `nx token` cost/models bars so model names stay fully visible: shared column sizing prefers the label over the bar, keeps a gap before the bar (no name/bar collision on the longest row), and sizes the USD column from the widest amount.

## 0.4.1

- Fixed `nx token` cost and models bars truncating model names: the name column now grows to fit the longest visible label (within the card width), so names like `GPT-5.6-sol-high` and `Opus 4.7.thinking-high` stay intact.

## 0.4.0

- Cursor token totals now prefer the Cursor dashboard usage API (same approach as cursor-usage): read the local session JWT from `state.vscdb`, fetch billed input/output/cache events, and keep local sessions/messages for activity. Falls back to local bubble/meter/chars÷4 when offline or unauthenticated.
- Added `NX_TOKEN_CURSOR_LOCAL` to force local-only Cursor counting, plus `NX_CURSOR_SESSION_TOKEN` / `CURSOR_SESSION_TOKEN` overrides.
- Improved offline Cursor CLI estimates by counting reasoning and tool-call/result text in `store.db` blobs (still estimated).
- Invalidated on-disk token aggregate cache (schema v4) after the Cursor dashboard enricher.

## 0.3.1

- Fixed Cursor model attribution under Auto: resolve the underlying model from AgentKv (`providerOptions.cursor.modelName` by `requestId`), then non-sentinel bubble/`modelConfig` names, then dominant `composerData.usageData` keys — keep "Auto" only when nothing resolved is available locally.
- Invalidated on-disk token aggregate cache (schema v3) after the Cursor model-resolution change.

## 0.3.0

- Added `nx update` to force a self-update from the latest GitHub release.
- Background self-update now checks on every invocation (no longer once per day), running alongside the command so startup stays snappy.
- Self-update and the installer-style release lookup now use `github.com/.../releases/latest` redirects instead of `api.github.com`, avoiding unauthenticated API rate limits.

## 0.2.0

- Replaced the shared `nx` wordmark on `nx token` banners with per-harness icon ASCII (nexus mark for combined; Claude asterisk, Codex brackets, Cursor cube, pi glyph).
- Improved Claude Code counting: dual roots (`~/.claude` + `~/.config/claude`, `CLAUDE_CONFIG_DIR`), recursive JSONL including subagents, final-chunk dedup by `requestId`/`stop_reason`, nested cache-creation 5m/1h summed into cache writes.
- Improved Codex counting: `CODEX_HOME`, `archived_sessions` with sessions-wins dedupe, prefer `last_token_usage`, skip duplicate cumulatives, fold reasoning into output.
- Improved Cursor counting: prefer non-zero bubble `tokenCount`, else credit `promptTokenBreakdown.totalUsedTokens` / `contextTokensUsed` once per composer, else chars/4; document local undercount vs admin dashboard.
- Added `PI_AGENT_DIR` override for pi-agent session roots.
- Invalidated on-disk token aggregate cache (schema v2) after parser semantic changes.

## 0.1.4

- Added nested help: `nx help [command...]` drills into command detail (`nx help git`, `nx help git stat`, `nx help token`, `nx help token harness`, …). Unknown help paths exit `2` and show the nearest help page.
- Expanded `nx help git` / `nx help git stat` with subcommand lists, options, and examples; `nx git help [stat]` matches the same path.
- Token help links to topic pages (`harness`, `range`, `view`, `output`, `flags`, `env`, `exit`, `examples`) via `nx help token <topic>`.
- Documented the nested-help precedent in `AGENTS.md` so new commands keep the help tree, tests, and README Help section updated together.

## 0.1.3

- Fixed self-update noise on root-owned installs (e.g. `/usr/local/bin`): when the install directory is not writable, the daily update check now skips quietly instead of printing `permission denied`.
- Changed the curl installer default to `~/.local/bin` (no `sudo`) so new installs stay user-writable and self-update works; refuses non-writable `NX_INSTALL_DIR` targets instead of escalating with `sudo`.
- Installer now ensures the install bindir is on `PATH` via the user shell profile, and re-running it migrates an older `/usr/local/bin/nx` install by removing the stale binary.

## 0.1.2

- Improved `nx token` load performance: harnesses and session files are parsed concurrently, Cursor SQLite databases open read-only in place (no temp copy unless the live file is locked), and parsed aggregates are cached under `~/.cache/nx/token` keyed by source file mtimes. Set `NX_TOKEN_NO_CACHE=1` to bypass the cache.
- Improved JSONL and Cursor blob parsing throughput with `github.com/bytedance/sonic` on the hot decode paths.

## 0.1.1

- Migrated `nx token` rendering to the charm.land v2 stack (`lipgloss/v2`, `bubbletea/v2`), removing the duplicate lipgloss v1/termenv/isatty dependency family; the whole binary now uses one styling stack. Output is unchanged.
- Improved color handling when piped or on limited terminals: ANSI is now stripped/downsampled at write time via `colorprofile`, and `NO_COLOR`/`CLICOLOR` are respected on auto-detected terminals.
- Improved truecolor fidelity: theme colors now emit their exact declared hex values (lipgloss v1 rounded some RGB channels off by one).
- Documented the `modernc.org/sqlite` dependency decision in AGENTS.md and cleaned up ported code to modern Go idioms (range-over-int, min/max, strings.Builder).

## 0.1.0

- Added the `nx token` command family (alias: `nx tokens`): a pastel terminal dashboard for AI coding-harness token usage, ported from tmax.
- Added four harnesses: Claude Code, Codex, pi.dev, and Cursor (IDE + `cursor-agent` CLI merged; Cursor tokens are estimated from transcript size (~4 bytes/token) since Cursor stores no real token counts locally).
- Added nine card views: overview, models, hours, punchcard, trend, topdays, weekday, cost, and mix, selectable with order-independent positional arguments alongside harness and range (`alltime`, `30d`, `7d`).
- Improved model display names over tmax: legacy Claude ids render as "Sonnet 3.5"-style names (e.g. `claude-3-5-sonnet-20241022` → "Sonnet 3.5", where tmax rendered "3 5.sonnet").
- Added standalone output modes: `json` (NDJSON for `all`), `quiet` one-liner for shell prompts, and `compare` side-by-side harness view.
- Added an interactive TUI (`nx token -i`) with harness/tab/range navigation.
- Added adaptive light/dark rendering with `NX_BACKGROUND` and `NX_TRUECOLOR` overrides, plus plain-text output when piped.
- Added distinct exit codes for scripting: 0 ok, 2 bad arguments, 3 no usage for the selection (output modes).

## 0.0.4

- Fixed `nx git stat` default-branch fallback when `git remote show -n origin` reports `HEAD branch: (not queried)`.

## 0.0.3

- Improved `nx git stat` performance by fetching only the detected default branch instead of all `origin` refs.
- Improved multi-folder `nx git stat` collection with concurrent folder checks while preserving input order.
- Added `nx git stat --jobs <n>` to tune multi-folder collection concurrency.

## 0.0.2

- Added the initial extensible `nx` Go CLI foundation.
- Added `nx git stat <folder> [folder...]` with pretty terminal output.
- Added verified curl installer and runtime self-update from GitHub releases.
- Added VERSION-driven release automation with GoReleaser.
- Added local format/check scripts and pre-commit hook support.
- Added architecture notes for future command additions.

## 0.0.1

- Initial `nx` CLI foundation.
- Added `nx git stat <folder> [folder...]`.
- Added daily self-update checks from GitHub releases.
