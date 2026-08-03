// Package tui provides the interactive Bubble Tea front-end for akme git.
package tui

import (
	"strings"
	"time"

	"github.com/o1x3/akme/internal/gitstat"
	"github.com/o1x3/akme/internal/gitstat/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

var (
	ranges = []string{gitstat.RangeYear, gitstat.Range30d, gitstat.Range7d}
	tabs   = ui.TabOrder
)

// Options carries terminal state resolved by the CLI layer.
type Options struct {
	Dark           bool
	DarkLocked     bool
	ForceTruecolor bool
}

type model struct {
	agg     *gitstat.Aggregate
	ri, tab int
	now     time.Time
	w, h    int
	hintCol lipgloss.Style
	dark    bool
	plain   bool
	opts    Options
}

// New builds the interactive model.
func New(agg *gitstat.Aggregate, rng, tab string, opts Options) model {
	m := model{
		agg:  agg,
		now:  time.Now(),
		dark: opts.Dark,
		opts: opts,
	}
	m.ri = indexOf(ranges, rng, 0)
	m.tab = indexOf(tabs, tab, 0)
	m.hintCol = lipgloss.NewStyle().Foreground(lipgloss.Color("#565668"))
	ui.Configure(m.dark, m.plain)
	return m
}

func indexOf(s []string, v string, def int) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return def
}

func (m model) Init() tea.Cmd {
	if m.opts.DarkLocked {
		return nil
	}
	return tea.RequestBackgroundColor
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case tea.BackgroundColorMsg:
		if m.opts.DarkLocked {
			break
		}
		m.dark = msg.IsDark()
		ui.Configure(m.dark, m.plain)
	case tea.ColorProfileMsg:
		m.plain = msg.Profile <= colorprofile.ASCII
		ui.Configure(m.dark, m.plain)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "tab", "m":
			m.tab = (m.tab + 1) % len(tabs)
		case "shift+tab":
			m.tab = (m.tab - 1 + len(tabs)) % len(tabs)
		case "1":
			m.ri = 0
		case "2":
			m.ri = 1
		case "3":
			m.ri = 2
		case "r":
			m.ri = (m.ri + 1) % len(ranges)
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	s := gitstat.Summarize(m.agg, ranges[m.ri], m.now)
	card := ui.RenderCard(s, tabs[m.tab])
	hint := m.hintCol.Render("tab/⇧tab views · 1/2/3 range · q quit")
	body := padBlock(lipgloss.JoinVertical(lipgloss.Left, card, "", hint))
	if m.w > 0 && m.h > 0 {
		body = lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, body)
	}
	v := tea.NewView(body)
	v.AltScreen = true
	return v
}

func padBlock(s string) string {
	lines := strings.Split(s, "\n")
	maxW := 0
	for _, l := range lines {
		if w := lipgloss.Width(l); w > maxW {
			maxW = w
		}
	}
	for i, l := range lines {
		if d := maxW - lipgloss.Width(l); d > 0 {
			lines[i] = l + strings.Repeat(" ", d)
		}
	}
	return strings.Join(lines, "\n")
}

// Run starts the interactive program. Aggregate must already be loaded.
func Run(agg *gitstat.Aggregate, rng, tab string, opts Options) error {
	var popts []tea.ProgramOption
	if opts.ForceTruecolor {
		popts = append(popts, tea.WithColorProfile(colorprofile.TrueColor))
	}
	p := tea.NewProgram(New(agg, rng, tab, opts), popts...)
	_, err := p.Run()
	return err
}
