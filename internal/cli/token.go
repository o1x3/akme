package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/o1x3/akme/internal/envx"
	"github.com/o1x3/akme/internal/token/core"
	"github.com/o1x3/akme/internal/token/tui"
	"github.com/o1x3/akme/internal/token/ui"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
)

type tokenOptions struct {
	harness      string
	rng          string
	tab          string
	interactive  bool
	jsonOut      bool
	quiet        bool
	compare      bool
	help         bool
	includeCache bool // `all` — include prompt-cache reads in totals
}

// parseTokenArgs ports tmax's positional, order-independent argument grammar.
// Bare "all" includes prompt-cache reads in headline totals (default excludes
// them). Combined-harness keywords are "combined" and "everything"; all-time
// range keywords are "alltime" and "lifetime" (all-time is also the default).
func parseTokenArgs(args []string) (tokenOptions, error) {
	o := tokenOptions{harness: core.Combined, rng: core.RangeAll, tab: ui.TabOverview}
	for _, a := range args {
		switch a {
		case "-h", "--help", "help":
			o.help = true
			return o, nil
		case "-i", "--interactive", "-t", "--tui", "tui":
			o.interactive = true

		// ---- counting mode ----
		case "all":
			// Include prompt-cache reads in totals. (Harness merge stays the
			// default; use combined/everything to name it explicitly.)
			o.includeCache = true

		// ---- harnesses ----
		case "claude", "cc", "claude-code":
			o.harness = core.Claude
		case "codex", "cx":
			o.harness = core.Codex
		case "pi", "pi.dev", "pidev":
			o.harness = core.Pi
		case "cursor", "cursor-ide", "cursor-cli", "cursor-agent":
			o.harness = core.Cursor
		case "combined", "everything":
			o.harness = core.Combined

		// ---- ranges ----
		case "30d", "month", "30":
			o.rng = core.Range30d
		case "7d", "week", "7":
			o.rng = core.Range7d
		case "alltime", "lifetime":
			o.rng = core.RangeAll

		// ---- body tabs ----
		case "overview", "-o", "--overview":
			o.tab = ui.TabOverview
		case "models", "model", "-m", "--models":
			o.tab = ui.TabModels
		case "hours", "clock", "--hours":
			o.tab = ui.TabHours
		case "punchcard", "punch", "when", "--punchcard":
			o.tab = ui.TabPunchcard
		case "trend", "spark", "series", "--trend":
			o.tab = ui.TabTrend
		case "topdays", "busiest", "--topdays":
			o.tab = ui.TabTopDays
		case "weekday", "dow", "--weekday":
			o.tab = ui.TabWeekday
		case "cost", "spend", "--cost":
			o.tab = ui.TabCost
		case "mix", "split", "--mix":
			o.tab = ui.TabMix

		// ---- standalone output modes (bypass the card) ----
		case "json", "--json", "--stats":
			o.jsonOut = true
		case "quiet", "-q", "--quiet":
			o.quiet = true
		case "compare", "vs", "--compare":
			o.compare = true

		default:
			return o, ExitError{Code: 2, Err: fmt.Errorf("unknown argument %q (try: akme help token)", a)}
		}
	}
	return o, nil
}

func (a App) runToken(ctx context.Context, args []string, stdout io.Writer) error {
	_ = ctx

	o, err := parseTokenArgs(args)
	if err != nil {
		return err
	}
	if o.help {
		fmt.Fprintln(stdout, tokenHelpText())
		return nil
	}

	forced := envx.First("AKME_TRUECOLOR", "NX_TRUECOLOR") != ""
	tty := term.IsTerminal(os.Stdout.Fd())

	now := time.Now()

	// quiet/json bypass the card and emit no styling, so dispatch them before
	// any colour work — no reason to query the terminal for those. They exit 3
	// when the selected harness has no usage, so they compose in scripts and CI.
	switch {
	case o.quiet:
		return runTokenQuiet(o, now, stdout)
	case o.jsonOut:
		return runTokenJSON(o, now, stdout)
	}

	// Light/dark detection so foreground colours stay legible on any terminal.
	// akme token paints no background; it adapts to yours. AKME_BACKGROUND overrides
	// (and skips the terminal query, which momentarily raw-modes the tty).
	// Non-TTY output defaults to the dark palette, matching v1 behaviour.
	dark := true
	darkLocked := false
	switch envx.First("AKME_BACKGROUND", "NX_BACKGROUND") {
	case "light":
		dark, darkLocked = false, true
	case "dark":
		dark, darkLocked = true, true
	default:
		if tty {
			dark = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
		}
	}

	// Colour handling: lipgloss v2 styles always emit truecolor escapes;
	// downsampling/stripping happens at write time. Force 24-bit when asked
	// (capture / under-reporting terminals), strip every escape when piped or
	// redirected, otherwise downsample to whatever the terminal supports.
	var profile colorprofile.Profile
	switch {
	case forced:
		profile = colorprofile.TrueColor
	case !tty:
		profile = colorprofile.NoTTY
	default:
		profile = colorprofile.Detect(os.Stdout, os.Environ())
	}

	// Plain mode follows the effective profile, not just tty-ness, so NO_COLOR
	// or TERM=dumb on a real terminal still gets shade-glyph density cells.
	ui.Configure(dark, profile <= colorprofile.ASCII)
	out := &colorprofile.Writer{Forward: stdout, Profile: profile}

	if o.compare {
		return runTokenCompare(o, now, out)
	}

	if o.interactive {
		return tui.Run(o.harness, o.rng, o.tab, tui.Options{
			Dark:           dark,
			DarkLocked:     darkLocked,
			ForceTruecolor: forced,
			IncludeCache:   o.includeCache,
		})
	}

	agg := core.Load(o.harness)
	s := core.Summarize(agg, o.rng, now, o.includeCache)
	fmt.Fprintln(out, ui.RenderCard(s, o.tab))
	maybeCursorHint(o.harness, s)

	// Nudge toward the interactive view when on a real terminal. Update checks
	// are akme-wide (selfupdate), not per-command, so no notice is printed here.
	if tty {
		fmt.Fprintln(out, ui.Hint("run `akme token -i` for the interactive view"))
	}
	return nil
}

// tokenTag renders the "[harness·range]" provenance tag for quiet mode,
// dropping it only for the bare default (all harnesses, all time, cache
// reads excluded). Combined harness is labeled "combined"; `all` means
// include-cache and is appended as a third component when set.
func tokenTag(o tokenOptions) string {
	if o.harness == core.Combined && o.rng == core.RangeAll && !o.includeCache {
		return ""
	}
	h := o.harness
	if h == core.Combined {
		h = "combined"
	}
	tag := h + "·" + o.rng
	if o.includeCache {
		tag += "·all"
	}
	return "[" + tag + "]"
}

// runTokenQuiet prints exactly one terse, prompt-safe line of headline numbers.
// No colour, no hint, single newline. Exit 3 when the selection has no usage.
func runTokenQuiet(o tokenOptions, now time.Time, stdout io.Writer) error {
	s := core.Summarize(core.Load(o.harness), o.rng, now, o.includeCache)
	if !s.HasData() {
		fmt.Fprintln(stdout, "akme token: no usage for this selection")
		maybeCursorHint(o.harness, s)
		return ExitError{Code: 3}
	}
	fmt.Fprintf(stdout, "akme token%s %s tok · %s msgs · %dd streak · %s\n",
		tokenTag(o), core.FormatTokens(s.TotalTokens), core.FormatInt(s.Messages),
		s.CurrentStreak, s.FavModel)
	return nil
}

// runTokenJSON emits the machine-readable summary: one indented object for a
// single harness, or NDJSON (one compact object per concrete harness) for
// combined, so `akme token json | jq -s` / `akme token all json | jq -s` work.
func runTokenJSON(o tokenOptions, now time.Time, stdout io.Writer) error {
	enc := json.NewEncoder(stdout)
	if o.harness == core.Combined {
		any := false
		var cursorSum core.Summary
		for _, h := range core.Harnesses {
			s := core.Summarize(core.Load(h), o.rng, now, o.includeCache)
			_ = enc.Encode(core.NewSummaryJSON(s, now))
			any = any || s.HasData()
			if h == core.Cursor {
				cursorSum = s
			}
		}
		if !any {
			maybeCursorHint(core.Cursor, cursorSum)
			return ExitError{Code: 3}
		}
		return nil
	}
	s := core.Summarize(core.Load(o.harness), o.rng, now, o.includeCache)
	enc.SetIndent("", "  ")
	_ = enc.Encode(core.NewSummaryJSON(s, now))
	if !s.HasData() {
		maybeCursorHint(o.harness, s)
		return ExitError{Code: 3}
	}
	maybeCursorHint(o.harness, s)
	return nil
}

// maybeCursorHint prints a one-line stderr diagnostic when Cursor is empty,
// still estimated, or shows activity with 0 billed tokens.
func maybeCursorHint(harness string, s core.Summary) {
	if harness != core.Cursor {
		return
	}
	if hint := core.CursorHintFor(s); hint != "" {
		fmt.Fprintln(os.Stderr, "akme token:", hint)
	}
}

// runTokenCompare renders the side-by-side harness card. stdout is the card
// path's colorprofile writer, so piped output is stripped of escapes and
// lesser terminals get downsampled colours.
func runTokenCompare(o tokenOptions, now time.Time, stdout io.Writer) error {
	var sums []core.Summary
	for _, h := range core.Harnesses {
		if s := core.Summarize(core.Load(h), o.rng, now, o.includeCache); s.HasData() {
			sums = append(sums, s)
		}
	}
	fmt.Fprintln(stdout, ui.RenderCompare(sums, o.rng))
	if len(sums) == 0 {
		return ExitError{Code: 3}
	}
	return nil
}

func tokenHelpText() string {
	return `akme token — token stats across your AI coding harnesses

USAGE
  akme token [harness] [range] [tab] [all] [-i]
  akme token [harness] [range] [all] (json | quiet | compare)
  akme help token [topic]

HARNESS   (default: combined — every harness merged)
  claude            Claude Code        ~/.claude + ~/.config/claude
  codex             OpenAI Codex       ~/.codex (sessions + archived)
  pi                pi.dev             ~/.pi/agent/sessions
  cursor            Cursor IDE + CLI   state.vscdb + ~/.cursor
  combined          every harness merged

  Overrides: CLAUDE_CONFIG_DIR, CODEX_HOME, PI_AGENT_DIR.
  Claude uses final streaming chunks; Cursor prefers the dashboard usage
  API when logged in (real input/output/cache), else local bubble/meter/chars÷4.
  Cursor Auto resolves underlying models locally when available.
  Cursor activity is machine-local; billed tokens follow the logged-in account
  (incl. team memberships). Empty dashboard replies keep local estimates.
  Paths: macOS ~/Library/Application Support, Linux ~/.config, Windows %APPDATA%.

COUNTING  (default: exclude prompt-cache reads)
  all               include cache reads in totals / series / quiet / json

  Cache writes stay in the default total (billed creation). Mix and cost still
  show the full ledger split either way.

RANGE     (default: alltime)
  alltime           lifetime
  30d               last 30 days
  7d                last 7 days

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

OUTPUT MODES   (bypass the card)
  json              machine-readable summary (--stats); NDJSON for combined
  quiet             one terse line for a shell prompt (-q)
  compare           all harnesses side by side (vs)

FLAGS
  -i, --tui         interactive mode (←/→ harness · tab · 1/2/3 range · q quit)
  -h, --help        this help

ENV
  AKME_BACKGROUND     light|dark — override terminal background detection
  AKME_TRUECOLOR      set to force 24-bit colour
  AKME_TOKEN_NO_CACHE set to bypass the on-disk aggregate cache
  AKME_TOKEN_CURSOR_LOCAL set to skip Cursor dashboard (local estimates only)
  AKME_CURSOR_SESSION_TOKEN / CURSOR_SESSION_TOKEN — Cursor dashboard JWT
  CLAUDE_CONFIG_DIR / CODEX_HOME / PI_AGENT_DIR — harness data roots

EXIT
  0 ok · 2 bad args · 3 no usage for the selection (output modes)

EXAMPLES
  akme token               akme token codex 7d cost      akme token claude punchcard
  akme token all json      akme token 30d compare        akme token pi trend -i

Nest help for one topic:
  akme help token harness · range · count · view · output · flags · env · exit · examples`
}
