package cli

import (
	"fmt"
	"io"
	"strings"
)

// runHelp prints nested help for a command path.
// Examples: akme help · akme help git · akme help git stat · akme help token harness
func (a App) runHelp(args []string, stdout io.Writer) error {
	text, err := helpFor(args)
	if err != nil {
		return err
	}
	fmt.Fprint(stdout, text)
	return nil
}

func helpFor(args []string) (string, error) {
	if len(args) == 0 {
		return helpText(), nil
	}

	switch args[0] {
	case "git":
		return gitHelpFor(args[1:])
	case "token", "tokens":
		return tokenHelpFor(args[1:])
	case "update":
		if len(args) > 1 {
			return "", ExitError{Code: 2, Err: fmt.Errorf("unknown help topic %q\n\n%s", args[1], strings.TrimSpace(updateHelpText()))}
		}
		return updateHelpText(), nil
	case "version", "-v", "--version":
		if len(args) > 1 {
			return "", ExitError{Code: 2, Err: fmt.Errorf("unknown help topic %q\n\n%s", args[1], strings.TrimSpace(versionHelpText()))}
		}
		return versionHelpText(), nil
	case "help", "-h", "--help":
		if len(args) > 1 {
			return "", ExitError{Code: 2, Err: fmt.Errorf("unknown help topic %q\n\n%s", args[1], strings.TrimSpace(helpCommandHelpText()))}
		}
		return helpCommandHelpText(), nil
	default:
		return "", ExitError{Code: 2, Err: fmt.Errorf("unknown help topic %q\n\n%s", args[0], strings.TrimSpace(helpText()))}
	}
}

func gitHelpFor(args []string) (string, error) {
	if len(args) == 0 {
		return gitHelpText(), nil
	}
	switch args[0] {
	case "stat", "stats":
		if len(args) > 1 {
			return "", ExitError{Code: 2, Err: fmt.Errorf("unknown help topic %q\n\n%s", args[1], strings.TrimSpace(gitStatHelpText()))}
		}
		return gitStatHelpText(), nil
	case "activity", "folder", "graph":
		if len(args) > 1 {
			return "", ExitError{Code: 2, Err: fmt.Errorf("unknown help topic %q\n\n%s", args[1], strings.TrimSpace(gitActivityHelpText()))}
		}
		return gitActivityHelpText(), nil
	case "view", "views", "tab", "tabs":
		if len(args) > 1 {
			return "", ExitError{Code: 2, Err: fmt.Errorf("unknown help topic %q\n\n%s", args[1], strings.TrimSpace(gitViewHelpText()))}
		}
		return gitViewHelpText(), nil
	case "range", "ranges":
		if len(args) > 1 {
			return "", ExitError{Code: 2, Err: fmt.Errorf("unknown help topic %q\n\n%s", args[1], strings.TrimSpace(gitRangeHelpText()))}
		}
		return gitRangeHelpText(), nil
	case "interactive", "tui", "-i":
		if len(args) > 1 {
			return "", ExitError{Code: 2, Err: fmt.Errorf("unknown help topic %q\n\n%s", args[1], strings.TrimSpace(gitInteractiveHelpText()))}
		}
		return gitInteractiveHelpText(), nil
	case "help", "-h", "--help":
		return gitHelpText(), nil
	default:
		return "", ExitError{Code: 2, Err: fmt.Errorf("unknown git help topic %q\n\n%s", args[0], strings.TrimSpace(gitHelpText()))}
	}
}

func tokenHelpFor(args []string) (string, error) {
	if len(args) == 0 {
		return tokenHelpText() + "\n", nil
	}
	topic := args[0]
	text := tokenTopicHelpText(topic)
	if text == "" {
		return "", ExitError{Code: 2, Err: fmt.Errorf("unknown token help topic %q\n\n%s", topic, strings.TrimSpace(tokenHelpTopicsText()))}
	}
	if len(args) > 1 {
		return "", ExitError{Code: 2, Err: fmt.Errorf("unknown help topic %q\n\n%s", args[1], strings.TrimSpace(text))}
	}
	return text, nil
}

func helpText() string {
	return `akme is a personal development CLI.

Usage:
  akme <command> [args]
  akme help [command...]

Commands:
  git       Git repository helpers
  token     Token stats across AI coding harnesses
  update    Update akme to the latest release
  version   Show build version
  help      Show help (nest for details)

Nest help to learn more:
  akme help git
  akme help token
  akme help update
  akme help version

`
}

func updateHelpText() string {
	return `Usage:
  akme update

Check GitHub releases for a newer akme build and replace this binary in place
when one is available. Released builds also check on every other akme invocation.

Notes:
  - Resolves the latest tag via github.com/.../releases/latest (not the API)
  - Needs a writable install directory (default ~/.local/bin)
  - Development builds (version "dev") cannot self-update
  - Set AKME_NO_UPDATE=1 to disable background checks (akme update still works)

Examples:
  akme update
  AKME_NO_UPDATE=1 akme version

`
}

func helpCommandHelpText() string {
	return `Usage:
  akme help [command...]

Show help for a command path. Keep nesting to see deeper detail.

Examples:
  akme help
  akme help git
  akme help git stat
  akme help token
  akme help token harness

`
}

func versionHelpText() string {
	return `Usage:
  akme version

Show the build version, commit, and build date.

Aliases:
  -v, --version

`
}

func gitHelpText() string {
	return `Usage:
  akme git <folder> [range] [view] [-i]
  akme git stat [--jobs <n>] <folder> [folder...]
  akme help git [topic]

Show a GitHub-style contribution graph for first-parent activity on the
repo's default branch (from the local clone), or branch diff stats.

Topics:
  activity      akme git <folder> dashboard (default)
  view          overview, authors, hours, punchcard, ...
  range         year / 30d / 7d
  interactive   -i / --tui keys
  stat          multi-folder branch diff table

Nest help:
  akme help git activity
  akme help git view
  akme help git range
  akme help git interactive
  akme help git stat

Examples:
  akme git .
  akme git . -i
  akme git . 30d punchcard
  akme git stat repo-a repo-b

`
}

func gitActivityHelpText() string {
	return `Usage:
  akme git <folder> [range] [view] [-i]

Contribution dashboard for one local clone. Counts first-parent commits
on the detected default branch (origin/HEAD → origin/main fallback).

Arguments:
  <folder>   Path relative to the current working directory

Notes:
  - Soft-fetches only the default branch; continues on local refs if fetch fails
  - Collects ~52 weeks of history (one git log; no GitHub API)
  - Static card by default; pass -i for the interactive TUI

Examples:
  akme git .
  akme git ./my-repo year authors
  akme git . -i

`
}

func gitViewHelpText() string {
	return `Usage:
  akme git <folder> [view]
  akme help git view

VIEW      (default: overview)
  overview          52-week contribution heatmap
  authors           top committers on the default branch
  hours             commits by hour of day
  punchcard         weekday × hour density grid
  trend             daily commit sparkline
  topdays           busiest civil days
  weekday           commits by weekday
  branch            HEAD vs default branch files/+/-

Aliases: punch/when, busy/topdays, dow/weekday, spark/trend, diff/stat→branch

`
}

func gitRangeHelpText() string {
	return `Usage:
  akme git <folder> [range]
  akme help git range

RANGE     (default: year)
  year     last ~52 weeks of first-parent commits on the default branch
  30d      last 30 days
  7d       last 7 days

Aliases: 52w/alltime/all → year; month/30 → 30d; week/7 → 7d

In the TUI: 1=year · 2=30d · 3=7d · r cycles

`
}

func gitInteractiveHelpText() string {
	return `Usage:
  akme git <folder> -i
  akme help git interactive

Interactive mode (Bubble Tea alt screen):
  tab / ⇧tab     cycle views
  1 / 2 / 3      year / 30d / 7d
  r              cycle range
  q / esc        quit

Aliases: --interactive, -t, --tui, tui

ENV
  AKME_BACKGROUND     light|dark — override terminal background detection
  AKME_TRUECOLOR      set to force 24-bit colour

`
}

func gitStatHelpText() string {
	return `Usage:
  akme git stat [--jobs <n>] <folder> [folder...]

Show changed files, added lines, and removed lines for each folder's
branch against the remote default branch (<base>...HEAD).

Arguments:
  <folder>   Path relative to the current working directory

Options:
  --jobs <n>   Max concurrent folder checks (default: GOMAXPROCS)

Notes:
  - Fetches only the detected origin default branch
  - Auto-detects the remote default branch from origin/HEAD
  - Falls back to origin/main if default branch detection is unavailable

Examples:
  akme git stat .
  akme git stat --jobs 4 repo-a repo-b

`
}

func tokenHelpTopicsText() string {
	return `Usage:
  akme help token [topic]

Topics:
  harness   Claude / Codex / pi / Cursor sources
  range     alltime / 30d / 7d
  count     exclude cache reads (default) / all
  view      overview, models, hours, punchcard, ...
  output    json / quiet / compare modes
  flags     -i / --help
  env       AKME_* / CLAUDE_CONFIG_DIR / CODEX_HOME / PI_AGENT_DIR / CURSOR_SESSION_TOKEN
  exit      process exit codes
  examples  common invocations

Or show everything:
  akme help token
  akme token --help

`
}

func tokenTopicHelpText(topic string) string {
	switch topic {
	case "help", "-h", "--help", "topics":
		return tokenHelpTopicsText()
	case "harness", "harnesses", "source", "sources":
		return `akme token — harnesses

HARNESS   (default: combined — every harness merged)
  claude            Claude Code        ~/.claude + ~/.config/claude
  codex             OpenAI Codex       ~/.codex (sessions + archived)
  pi                pi.dev             ~/.pi/agent/sessions
  cursor            Cursor IDE + CLI   state.vscdb + ~/.cursor
  combined          every harness merged

Aliases:
  claude: cc, claude-code
  codex:  cx
  pi:     pi.dev, pidev
  cursor: cursor-ide, cursor-cli, cursor-agent
  combined: everything

Overrides: CLAUDE_CONFIG_DIR, CODEX_HOME, PI_AGENT_DIR (comma-separated paths).

Claude counts final streaming chunks (stop_reason) and includes subagent
JSONL. Codex prefers last_token_usage deltas and skips archived duplicates.
Cursor prefers the dashboard usage API (billed input/output/cache) when a
local session JWT is available; otherwise bubble tokenCount, else the
composer context meter once per chat, else ~4 bytes/token. Cursor Auto
resolves to the underlying local model (AgentKv / usageData) when available.
Set AKME_TOKEN_CURSOR_LOCAL=1 to force local-only Cursor totals.

Cursor sessions/heatmap are machine-local (state.vscdb does not sync across
machines or accounts). Billed token totals follow the logged-in Cursor
account via the dashboard API (personal teamId=0 plus discovered team
memberships). Empty dashboard responses no longer wipe local estimates.
<config> is ~/Library/Application Support on macOS, ~/.config on Linux, and
%APPDATA% on Windows. akme token cursor json includes cursor_status;
empty/estimated/zero-token Cursor output prints a stderr hint.

More: akme help token

`
	case "range", "ranges", "time":
		return `akme token — ranges

RANGE     (default: alltime)
  alltime           lifetime
  30d               last 30 days
  7d                last 7 days

Aliases:
  alltime: lifetime
  30d:     month, 30
  7d:      week, 7

More: akme help token

`
	case "count", "counting", "cache":
		return `akme token — counting

COUNTING  (default: exclude prompt-cache reads)
  all               include cache reads in totals / day series / quiet / json

Headline totals, heatmaps, trends, model shares, quiet, compare, and
tokens.total in JSON omit cache_read unless you pass all. Cache writes
stay in the default total (billed creation). The mix and cost views still
expose the full ledger split; JSON always reports cache_read / cache_write
and sets include_cache.

Not related to AKME_TOKEN_NO_CACHE (that bypasses the on-disk parse cache).

Examples:
  akme token              # fresh totals (no cache reads)
  akme token all          # full ledger including cache reads
  akme token claude all quiet

More: akme help token

`
	case "view", "views", "tab", "tabs":
		return `akme token — views

TAB       (default: overview)
  overview          headline stats + activity heatmap
  models            token share by model
  hours             activity by local hour (clock)
  punchcard         weekday × hour density grid (when)
  trend             daily-token sparkline + averages + momentum (spark)
  topdays           your busiest days, ranked (busiest)
  weekday           tokens by day of week (dow)
  cost              estimated spend + cache savings (spend)
  mix               input / output / cache token composition (split)

More: akme help token

`
	case "output", "outputs", "mode", "modes":
		return `akme token — output modes

OUTPUT MODES   (bypass the card)
  json              machine-readable summary (--stats); NDJSON for combined
  quiet             one terse line for a shell prompt (-q)
  compare           all harnesses side by side (vs)

These modes exit 3 when the selection has no usage, so they compose in scripts.

More: akme help token

`
	case "flag", "flags", "options":
		return `akme token — flags

FLAGS
  -i, --tui         interactive mode (←/→ harness · tab · 1/2/3 range · q quit)
  -h, --help        this help

Interactive aliases: --interactive, -t, tui

Counting keyword (not a flag): all — include prompt-cache reads.
See: akme help token count

More: akme help token

`
	case "env", "environment":
		return `akme token — environment

ENV
  AKME_BACKGROUND           light|dark — override terminal background detection
  AKME_TRUECOLOR            set to force 24-bit colour
  AKME_TOKEN_NO_CACHE       set to bypass the on-disk aggregate cache
  AKME_TOKEN_CURSOR_LOCAL   set to skip Cursor dashboard enrichment (local only)
  AKME_CURSOR_SESSION_TOKEN override Cursor dashboard session (JWT or sub::jwt)
  CURSOR_SESSION_TOKEN    same as AKME_CURSOR_SESSION_TOKEN
  CLAUDE_CONFIG_DIR       comma-separated Claude config roots (…/projects)
  CODEX_HOME              comma-separated Codex homes
  PI_AGENT_DIR            comma-separated pi-agent session dirs

Cursor activity (sessions/heatmap) is machine-local; billed tokens follow the
logged-in account (including team memberships). Empty dashboard replies keep
local estimates. Check cursor_status in akme token cursor json or the stderr
hint when Cursor is empty/estimated/zero-token.

Piped output is plain text. NO_COLOR / CLICOLOR are respected.

More: akme help token

`
	case "exit", "exits", "codes", "status":
		return `akme token — exit codes

EXIT
  0 ok
  2 bad args
  3 no usage for the selection (output modes only)

More: akme help token

`
	case "example", "examples":
		return `akme token — examples

EXAMPLES
  akme token               akme token codex 7d cost      akme token claude punchcard
  akme token all json      akme token 30d compare        akme token pi trend -i

More: akme help token

`
	default:
		return ""
	}
}
