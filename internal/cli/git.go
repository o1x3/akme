package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/o1x3/akme/internal/envx"
	"github.com/o1x3/akme/internal/gitstat"
	"github.com/o1x3/akme/internal/gitstat/tui"
	"github.com/o1x3/akme/internal/gitstat/ui"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
)

type gitActivityOptions struct {
	folder      string
	rng         string
	tab         string
	interactive bool
	help        bool
}

func parseGitActivityArgs(args []string) (gitActivityOptions, error) {
	o := gitActivityOptions{
		rng: gitstat.RangeYear,
		tab: ui.TabOverview,
	}
	for _, a := range args {
		switch a {
		case "-h", "--help", "help":
			o.help = true
			return o, nil
		case "-i", "--interactive", "-t", "--tui", "tui":
			o.interactive = true

		case "year", "52w", "alltime", "all":
			o.rng = gitstat.RangeYear
		case "30d", "month", "30":
			o.rng = gitstat.Range30d
		case "7d", "week", "7":
			o.rng = gitstat.Range7d

		case "overview", "-o", "--overview":
			o.tab = ui.TabOverview
		case "authors", "author", "--authors":
			o.tab = ui.TabAuthors
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
		case "branch", "stat", "diff", "--branch":
			o.tab = ui.TabBranch

		default:
			if strings.HasPrefix(a, "-") {
				return o, ExitError{Code: 2, Err: fmt.Errorf("unknown option %q\n\nTry: akme help git", a)}
			}
			if o.folder != "" {
				return o, ExitError{Code: 2, Err: fmt.Errorf("unexpected argument %q\n\nTry: akme help git", a)}
			}
			o.folder = a
		}
	}
	if o.help {
		return o, nil
	}
	if o.folder == "" {
		return o, ExitError{Code: 2, Err: fmt.Errorf("usage: akme git <folder> [range] [view] [-i]\n\nTry: akme help git")}
	}
	return o, nil
}

func (a App) runGitActivity(ctx context.Context, args []string, stdout io.Writer) error {
	o, err := parseGitActivityArgs(args)
	if err != nil {
		return err
	}
	if o.help {
		return a.runHelp([]string{"git"}, stdout)
	}

	forced := envx.First("AKME_TRUECOLOR", "NX_TRUECOLOR") != ""
	tty := term.IsTerminal(os.Stdout.Fd())
	now := time.Now()

	agg, err := gitstat.Load(ctx, o.folder)
	if err != nil {
		return err
	}

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

	var profile colorprofile.Profile
	switch {
	case forced:
		profile = colorprofile.TrueColor
	case !tty:
		profile = colorprofile.NoTTY
	default:
		profile = colorprofile.Detect(os.Stdout, os.Environ())
	}

	ui.Configure(dark, profile <= colorprofile.ASCII)

	if o.interactive {
		return tui.Run(agg, o.rng, o.tab, tui.Options{
			Dark:           dark,
			DarkLocked:     darkLocked,
			ForceTruecolor: forced,
		})
	}

	out := &colorprofile.Writer{Forward: stdout, Profile: profile}
	s := gitstat.Summarize(agg, o.rng, now)
	fmt.Fprintln(out, ui.RenderCard(s, o.tab))
	if tty {
		fmt.Fprintln(out, ui.Hint(fmt.Sprintf("run `akme git %s -i` for the interactive view", o.folder)))
	}
	return nil
}
