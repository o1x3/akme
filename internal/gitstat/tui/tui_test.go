package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/o1x3/akme/internal/gitstat"
	"github.com/o1x3/akme/internal/gitstat/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func fixedModel() model {
	day := time.Now().Format("2006-01-02")
	agg := &gitstat.Aggregate{
		Name: "repo", Base: "origin/main", Head: "main",
		ByDay:       map[string]int{day: 1},
		ByDayHour:   map[string][24]int{day: {12: 1}},
		ByAuthorDay: map[string]map[string]int{"A": {day: 1}},
		Commits:     1,
	}
	return New(agg, gitstat.RangeYear, ui.TabOverview, Options{Dark: true, DarkLocked: true})
}

func key(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func TestTUIViewRenders(t *testing.T) {
	m := fixedModel()
	m.w, m.h = 100, 40
	v := m.View()
	if !strings.Contains(v.Content, "Overview") {
		t.Error("TUI view should contain the Overview tab")
	}
	if !v.AltScreen {
		t.Error("TUI view should request the alternate screen")
	}
}

func TestModelCyclesTabsAndRanges(t *testing.T) {
	m := fixedModel()
	nm, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	mm := nm.(model)
	if tabs[mm.tab] != ui.TabAuthors {
		t.Fatalf("tab = %s, want authors", tabs[mm.tab])
	}
	nm, _ = mm.Update(key('2'))
	mm = nm.(model)
	if ranges[mm.ri] != gitstat.Range30d {
		t.Fatalf("range = %s, want 30d", ranges[mm.ri])
	}
}

func TestPadBlockUniformWidth(t *testing.T) {
	in := "short\n" + strings.Repeat("x", 20) + "\nmid"
	out := padBlock(in)
	lines := strings.Split(out, "\n")
	w0 := lipgloss.Width(lines[0])
	for i, l := range lines {
		if w := lipgloss.Width(l); w != w0 {
			t.Errorf("line %d width %d, want %d", i, w, w0)
		}
	}
}
