# akme Architecture Notes

`akme` is a personal developer CLI, not a framework product. Keep the core small until repeated command patterns justify more structure.

## Shape

- `cmd/akme` owns process startup, build metadata, and top-level wiring.
- `internal/cli` owns command routing and should stay thin.
- Domain packages under `internal/<domain>` own behavior.
- Rendering stays separate from collection so commands remain testable without terminal snapshots.

Current domains:

- `internal/gitstat`: default-branch activity (contribution heatmap / TUI) and repository diff stats; `ui` + `tui` for the activity dashboard.
- `internal/render`: Lip Gloss terminal presentation (git stat table).
- `internal/selfupdate`: per-invocation GitHub release checks (via `releases/latest`, not the API) and binary replacement; also backs `akme update`.
- `internal/token`: coding-agent token/cost usage stats across harnesses (claude, codex, pi, cursor), with `core` collection, `ui` rendering, and `tui` interactive views.

Dependency decision: `internal/token` reads Cursor's SQLite stores through `modernc.org/sqlite` (pure Go, ~4 MB added to the stripped binary). Rejected alternatives: `mattn/go-sqlite3` needs cgo and breaks the `CGO_ENABLED=0` cross-compiled releases; shelling out to a system `sqlite3` is a fragile runtime dependency; a hand-written SQLite/WAL reader is a correctness risk.

Commands signal specific exit codes by returning `cli.ExitError` (0 ok, 2 usage error, 3 no data, 1 anything else).

## Command Model

Commands should grow as explicit namespaces:

```sh
akme <domain> <verb> [args]
```

Example:

```sh
akme git .
akme git stat repo-a repo-b
```

Do not add root-level shortcuts unless they are clearly permanent. Folder arguments are current-working-directory relative; discovery outside the provided paths belongs in a separate command.

## Nested Help

Help is a nestable discovery tree, not a single dump. Users learn the CLI by drilling in:

```sh
akme help                 # root: list top-level commands + nest hints
akme help <domain>        # domain overview + subcommands/topics
akme help <domain> <verb> # verb / topic detail
```

Examples that must keep working: `akme help git`, `akme help git activity`, `akme help git stat`, `akme help token`, `akme help token harness`.

Precedent (keep this current when commands change):

- Help routing and text live in `internal/cli/help.go` (`helpFor` / `runHelp`). Root `help` / `-h` / `--help` forwards leftover args as the help path.
- Every user-facing command path gets a matching `akme help …` page. Domain overviews list next-level nest targets; leaf pages cover usage, args, flags, and examples.
- Large surfaces (like `token`) expose focused topic pages under `akme help <domain> <topic>` in addition to the full page. Prefer nesting over stuffing more into the root blurb.
- Domain-local help should reuse the same tree (`akme git help [stat]`, `akme token --help`), not a divergent second copy of the docs.
- Unknown help paths return `cli.ExitError` code `2` and print the nearest help page so users can recover by nesting up/down.
- When you add or rename a command, subcommand, flag, or token topic: update the nested help text, the root nest hints if needed, `internal/cli/help_test.go`, and the README Help section in the same change.

## Adding Commands

Each new command should add or extend one domain package under `internal/<domain>`, then expose only the routing surface through `internal/cli`.

The expected change shape is:

- domain behavior in `internal/<domain>`
- CLI routing in `internal/cli`
- nested help pages in `internal/cli/help.go` (and topics when the surface is large)
- rendering isolated from collection when terminal output is non-trivial
- behavior tests for the domain package (plus help-path coverage for new nest targets)
- README command docs when user-facing behavior changes

## Release Model

`VERSION` is the release trigger. GoReleaser owns GitHub releases and macOS/Linux artifacts. Runtime self-update consumes the latest GitHub release asset for the current OS and architecture. The npm package `akme-cli` is a second distribution channel: `.github/workflows/release.yml` runs `scripts/check.sh`, then (when `VERSION` changed on `main`) publishes via Trusted Publishing (OIDC, `id-token: write`) — no `NPM_TOKEN`. Install / bare `npx akme-cli` download the matching GitHub release binary (checksum-verified) into the package `vendor/` dir; bare npx also updates when a newer release exists. Configure the trusted publisher on `akme-cli` (GitHub user `o1x3`, repo `akme`, workflow **`release.yml`**). Unscoped `akme` is blocked by npm's name-similarity rules.

When a user asks for a command to be built and deployed, the expected final change includes:

- the command implementation
- tests and README updates
- a patch/minor/major bump in `VERSION`
- a matching `CHANGELOG.md` section for that version
- `npm/package.json` `version` bumped to match `VERSION` (enforced by `scripts/validate-version.sh`, which also requires the package `name` to stay `akme-cli` and no `optionalDependencies`)

Pushing a `VERSION` change to `main` creates tag `v<VERSION>` and publishes both the GitHub release and the npm package. Existing installations pick it up through self-update (native binaries) or the package manager (npm/npx/bun).

## Cursor Cloud specific instructions

`akme` is a single Go CLI (no servers/databases). Standard commands live in `README.md` ("Development") and `scripts/check.sh`; use those. Notes below are only the non-obvious caveats.

- Toolchain: `go.mod` pins `go 1.25.0`; the `go` toolchain auto-downloads it on first use, so no manual Go install is needed.
- Run in dev with `go run ./cmd/akme <cmd>` (e.g. `go run ./cmd/akme git .`). Local/`go run` builds report version `dev` and never self-update, so `AKME_NO_UPDATE=1` is unnecessary for dev.
- `akme token` reads local AI-harness data dirs (`~/.claude`, `~/.codex`, `~/.pi`, Cursor SQLite stores). A fresh VM has none, so the dashboard shows "No tokens recorded yet" and output modes (`json`/`quiet`/`compare`) exit `3`. That is expected, not a failure.
- Full local gate: `scripts/check.sh` is exactly what CI runs (`.github/workflows/release.yml` test job = Go from `go.mod` + Node 22 + `scripts/check.sh`; release job follows when `VERSION` bumps), so a green local run means green CI. It runs `gofmt -l .`, `go test ./...`, `sh -n` on the shell scripts, and version validation. It ALSO checks the npm package (`node --check` on `npm/bin`, `npm/lib`, `npm/scripts` JS + `node --test npm/lib/*.test.js`) — but only when `node` is on PATH, otherwise those checks are silently skipped. The cloud VM ships Node 22, so the full gate (Go + npm) runs here; do not rely on a machine without `node` to catch npm-package regressions.
- Version gate: `scripts/validate-version.sh` requires `VERSION` to be plain `major.minor.patch`, to have a matching `## <version>` section in `CHANGELOG.md`, and to equal `npm/package.json`'s `version` (with `name` == `akme-cli` and no `optionalDependencies`). It shells out to `node` to read `npm/package.json`, so it needs `node` present when that file exists. Bumping `VERSION` without also updating `CHANGELOG.md` and `npm/package.json` fails the gate.
- Commit authorship: the repo owner does NOT want to be added as a `Co-authored-by:` on agent commits. A Cursor-managed hook (`commit-msg.cursor.co-author` under the VM's agent hooks dir) auto-appends that trailer and may be regenerated on fresh VMs; disable it (e.g. `chmod -x` the hook) before committing, and do not add a `Co-authored-by:` trailer manually.
