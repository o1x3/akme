package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/o1x3/akme/internal/gitstat"
)

func TestRenderCardOverviewContainsHeatmap(t *testing.T) {
	Configure(true, true)
	now := time.Date(2026, 8, 2, 12, 0, 0, 0, time.Local)
	agg := &gitstat.Aggregate{
		Name: "repo",
		Base: "origin/main",
		Head: "main",
		ByDay: map[string]int{
			now.Format("2006-01-02"): 3,
		},
		ByDayHour:   map[string][24]int{now.Format("2006-01-02"): {12: 3}},
		ByAuthorDay: map[string]map[string]int{"Alice": {now.Format("2006-01-02"): 3}},
		Commits:     3,
	}
	s := gitstat.Summarize(agg, gitstat.RangeYear, now)
	out := RenderCard(s, TabOverview)
	for _, want := range []string{"Contributions", "overview", "year", "repo", "origin/main", "Less", "More"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestRenderCardAllTabsSmoke(t *testing.T) {
	Configure(true, true)
	now := time.Date(2026, 8, 2, 12, 0, 0, 0, time.Local)
	day := now.Format("2006-01-02")
	agg := &gitstat.Aggregate{
		Name: "repo", Base: "origin/main", Head: "feature",
		Added: 10, Removed: 2, Files: 3,
		ByDay:       map[string]int{day: 5},
		ByDayHour:   map[string][24]int{day: {10: 2, 11: 3}},
		ByAuthorDay: map[string]map[string]int{"Alice": {day: 5}},
		Commits:     5,
	}
	s := gitstat.Summarize(agg, gitstat.RangeYear, now)
	for _, tab := range TabOrder {
		out := RenderCard(s, tab)
		if out == "" {
			t.Fatalf("empty render for tab %s", tab)
		}
		if !strings.Contains(out, TabTitle(tab)) && tab != TabOverview {
			// header chip uses TabTitle
			if !strings.Contains(out, "[ "+TabTitle(tab)+" ]") {
				t.Fatalf("tab %s missing title in:\n%s", tab, out)
			}
		}
	}
}
