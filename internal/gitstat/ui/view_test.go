package ui

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/o1x3/akme/internal/gitstat"

	"charm.land/lipgloss/v2"
)

var ansi = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func sampleSummary(t *testing.T, rng string) gitstat.Summary {
	t.Helper()
	Configure(true, true)
	now := time.Date(2026, 8, 2, 12, 0, 0, 0, time.Local)
	day := now.Format("2006-01-02")
	agg := &gitstat.Aggregate{
		Name: "repo", Base: "origin/main", Head: "main",
		Added: 10, Removed: 2, Files: 3,
		ByDay:       map[string]int{day: 5},
		ByDayHour:   map[string][24]int{day: {10: 2, 11: 3}},
		ByAuthorDay: map[string]map[string]int{"Alice": {day: 5}},
		Commits:     5,
	}
	return gitstat.Summarize(agg, rng, now)
}

func TestRenderCardOverviewContainsHeatmap(t *testing.T) {
	s := sampleSummary(t, gitstat.RangeYear)
	out := RenderCard(s, TabOverview)
	for _, want := range []string{"Contributions", "overview", "year", "repo", "origin/main", "Less", "More"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestRenderCardAllTabsSmoke(t *testing.T) {
	s := sampleSummary(t, gitstat.RangeYear)
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

// Overview banner, heatmap, and footer share one right edge (contentW).
func TestOverviewCardWidthConsistent(t *testing.T) {
	s := sampleSummary(t, gitstat.RangeYear)
	out := RenderCard(s, TabOverview)
	lines := strings.Split(out, "\n")
	var widths []int
	for i, l := range lines {
		if strings.TrimSpace(ansi.ReplaceAllString(l, "")) == "" {
			continue
		}
		w := lipgloss.Width(l)
		widths = append(widths, w)
		if w != contentW {
			t.Errorf("line %d width %d, want contentW=%d:\n%q", i, w, contentW, ansi.ReplaceAllString(l, ""))
		}
	}
	if len(widths) < 10 {
		t.Fatalf("too few non-empty overview lines: %d", len(widths))
	}
}

func TestHeatGapsSpanContentWidth(t *testing.T) {
	cols := 52
	avail := contentW - gutterW
	gaps := heatGaps(cols, avail)
	if len(gaps) != cols {
		t.Fatalf("gaps len %d, want %d", len(gaps), cols)
	}
	sum := cols
	for _, g := range gaps {
		sum += g
	}
	if sum != avail {
		t.Fatalf("cells+gaps=%d, want avail=%d", sum, avail)
	}
	if heatColX(0, gaps) != gutterW {
		t.Fatalf("col0 x=%d, want gutterW=%d", heatColX(0, gaps), gutterW)
	}
	if got := heatColX(cols-1, gaps) + 1 + gaps[cols-1]; got != contentW {
		t.Fatalf("last cell end %d, want contentW=%d", got, contentW)
	}
	// Spacers must be spread, not piled onto the first gaps.
	firstHalf, secondHalf := 0, 0
	mid := (cols - 1) / 2
	for i := 0; i < mid; i++ {
		firstHalf += gaps[i]
	}
	for i := mid; i < cols-1; i++ {
		secondHalf += gaps[i]
	}
	if firstHalf == 0 || secondHalf == 0 {
		t.Fatalf("gaps not spread: firstHalf=%d secondHalf=%d gaps=%v", firstHalf, secondHalf, gaps)
	}
}

// Legend sits on the heatmap (no blank row); at most one blank before Contributions.
func TestOverviewVerticalSpacingTight(t *testing.T) {
	s := sampleSummary(t, gitstat.RangeYear)
	out := RenderCard(s, TabOverview)
	plain := ansi.ReplaceAllString(out, "")
	if i := strings.Index(plain, "Contributions"); i > 0 {
		before := plain[:i]
		if strings.HasSuffix(before, "\n\n\n") {
			t.Fatalf("too much space before Contributions:\n%s", before[max(0, len(before)-40):])
		}
	}
	// Find a heatmap weekday gutter line followed by the legend with no blank.
	if !regexp.MustCompile(`(?m)^(?:Mon |Wed |Fri |    ).+\n *Less `).MatchString(plain) {
		t.Fatalf("expected legend immediately under heatmap rows:\n%s", plain)
	}
}
