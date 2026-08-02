package ui

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/o1x3/akme/internal/gitstat"

	"charm.land/lipgloss/v2"
)

// Tabs selectable by flag and cycled in the interactive TUI.
const (
	TabOverview  = "overview"
	TabAuthors   = "authors"
	TabHours     = "hours"
	TabPunchcard = "punchcard"
	TabTrend     = "trend"
	TabTopDays   = "topdays"
	TabWeekday   = "weekday"
	TabBranch    = "branch"
)

// TabOrder is the canonical cycle order.
var TabOrder = []string{
	TabOverview, TabAuthors, TabHours, TabPunchcard,
	TabTrend, TabTopDays, TabWeekday, TabBranch,
}

var tabTitles = map[string]string{
	TabOverview:  "Overview",
	TabAuthors:   "Authors",
	TabHours:     "Hours",
	TabPunchcard: "Punchcard",
	TabTrend:     "Trend",
	TabTopDays:   "Busiest Days",
	TabWeekday:   "Weekday",
	TabBranch:    "Branch",
}

var tabShort = map[string]string{
	TabOverview:  "overview",
	TabAuthors:   "authors",
	TabHours:     "hours",
	TabPunchcard: "punch",
	TabTrend:     "trend",
	TabTopDays:   "busy",
	TabWeekday:   "weekday",
	TabBranch:    "branch",
}

// TabTitle returns the display name for a tab const.
func TabTitle(tab string) string {
	if t, ok := tabTitles[tab]; ok {
		return t
	}
	return tabTitles[TabOverview]
}

const (
	contentW = 72
	gutterW  = 4
)

var weekdayNames = [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

// RenderCard renders the full activity dashboard for a summary.
func RenderCard(s gitstat.Summary, tab string) string {
	th := DefaultTheme()
	blocks := []string{
		renderHeader(th, s, tab),
		renderTabStrip(th, tab),
		"",
		renderBanner(th, s),
		"",
		renderBody(th, s, tab),
		"",
		renderFooter(th, s),
	}
	return strings.Join(blocks, "\n")
}

func renderBody(th Theme, s gitstat.Summary, tab string) string {
	switch tab {
	case TabAuthors:
		return renderAuthors(th, s)
	case TabHours:
		return renderHours(th, s)
	case TabPunchcard:
		return renderPunchcard(th, s)
	case TabTrend:
		return renderTrend(th, s)
	case TabTopDays:
		return renderTopDays(th, s)
	case TabWeekday:
		return renderWeekday(th, s)
	case TabBranch:
		return renderBranch(th, s)
	default:
		return sectionTitle(th, s.Heatmap.Weeks) + "\n" + renderHeatmap(th, s.Heatmap)
	}
}

func renderHeader(th Theme, s gitstat.Summary, tab string) string {
	chip := styled(th.Accent).Bold(true).Render("[ " + TabTitle(tab) + " ]")
	return rightAlign(chip, renderRange(th, s.Range), contentW)
}

func renderTabStrip(th Theme, tab string) string {
	parts := make([]string, 0, len(TabOrder))
	for _, t := range TabOrder {
		lbl := tabShort[t]
		if t == tab {
			parts = append(parts, styled(th.Accent).Bold(true).Render(lbl))
		} else {
			parts = append(parts, styled(muted()).Render(lbl))
		}
	}
	return padRight(strings.Join(parts, styled(muted()).Render(" · ")), contentW)
}

func renderRange(th Theme, rng string) string {
	plain := ascii()
	seg := func(key, lbl string) string {
		if rng == key {
			st := styled(th.Accent).Bold(true)
			if plain {
				st = st.Underline(true)
			}
			return st.Render(lbl)
		}
		return styled(muted()).Render(lbl)
	}
	return seg(gitstat.RangeYear, "year") + "  " +
		seg(gitstat.Range30d, "30d") + "  " +
		seg(gitstat.Range7d, "7d")
}

func renderBanner(th Theme, s gitstat.Summary) string {
	title := styled(th.Accent).Bold(true).Render("akme") +
		styled(muted()).Render(" · ") +
		styled(value()).Bold(true).Render(s.Name)
	base := styled(muted()).Render(s.Base)
	line1 := rightAlign(title, base, contentW)

	status := ""
	if !s.Fetched && s.FetchNote != "" {
		status = styled(adapt("#b45309", "#fbbf24")).Render(s.FetchNote)
	}

	rows := []string{
		line1,
		styled(th.Accent).Render(strings.Repeat("─", min(lipgloss.Width(line1), contentW))),
		leaderRow("commits", gitstat.FormatInt(s.Commits), contentW),
		leaderRow("active days", gitstat.FormatInt(s.ActiveDays), contentW),
		leaderRow("authors", gitstat.FormatInt(s.Authors), contentW),
		leaderRow("streak", fmt.Sprintf("%dd / %dd", s.CurrentStreak, s.LongestStreak), contentW),
		leaderRow("peak hour", gitstat.FormatHour(s.PeakHour), contentW),
		leaderRow("HEAD", s.Head, contentW),
	}
	if status != "" {
		rows = append(rows, status)
	}
	return strings.Join(rows, "\n")
}

func renderFooter(th Theme, s gitstat.Summary) string {
	left := styled(muted()).Render(fmt.Sprintf("%s · first-parent on %s", rangeLabel(s.Range), s.Base))
	right := styled(muted()).Render("akme git")
	return rightAlign(left, right, contentW)
}

func rangeLabel(r string) string {
	switch r {
	case gitstat.Range7d:
		return "last 7d"
	case gitstat.Range30d:
		return "last 30d"
	default:
		return "last year"
	}
}

// Hint renders a dim helper line under the card on a TTY.
func Hint(s string) string {
	return lipgloss.NewStyle().Foreground(muted()).Italic(true).Render("  " + s)
}

func sectionTitle(th Theme, weeks int) string {
	return styled(th.Accent).Bold(true).Render("Contributions") +
		styled(muted()).Render(fmt.Sprintf(" · last %d weeks", weeks))
}

func renderHeatmap(th Theme, h gitstat.Heatmap) string {
	if h.Weeks <= 0 {
		return ""
	}
	cols := h.Weeks
	// Compact single-width cells: gutter + cols fits contentW (4+52=56).
	if maxCols := contentW - gutterW; cols > maxCols {
		cols = maxCols
	}
	plain := ascii()
	gut := [7]string{"    ", "Mon ", "    ", "Wed ", "    ", "Fri ", "    "}

	rows := make([]string, 0, 12)
	rows = append(rows, renderMonthRow(h, cols))
	for r := range 7 {
		var sb strings.Builder
		sb.WriteString(styled(label()).Render(gut[r]))
		for col := range cols {
			sb.WriteString(heatCell(th, h.Cells[r][col], h.Max, plain))
		}
		rows = append(rows, sb.String())
	}
	rows = append(rows, "", renderLegend(th, plain))
	return strings.Join(rows, "\n")
}

func heatCell(th Theme, v, max int64, plain bool) string {
	if v < 0 {
		return " "
	}
	if plain {
		return shadeGlyphs1[th.levelIndex(v, max)]
	}
	if v == 0 {
		return styled(th.Ramp[0]).Render("■")
	}
	return styled(th.level(v, max)).Render("■")
}

func renderLegend(th Theme, plain bool) string {
	var parts []string
	parts = append(parts, styled(muted()).Render("Less "))
	for i := 0; i < 5; i++ {
		if plain {
			parts = append(parts, shadeGlyphs1[i]+" ")
		} else {
			ch := "■"
			if i == 0 {
				ch = "■"
			}
			parts = append(parts, styled(th.Ramp[i]).Render(ch)+" ")
		}
	}
	parts = append(parts, styled(muted()).Render("More"))
	return strings.Join(parts, "")
}

func renderMonthRow(h gitstat.Heatmap, cols int) string {
	gridW := gutterW + cols
	buf := make([]rune, gridW)
	for i := range buf {
		buf[i] = ' '
	}
	lastEnd := -2
	placed := time.Month(0)
	var queue []time.Month
	enqueue := func(m time.Month) {
		if m == 0 || m == placed {
			return
		}
		for _, q := range queue {
			if q == m {
				return
			}
		}
		queue = append(queue, m)
	}
	for col := range cols {
		weekStart := h.FirstDay.AddDate(0, 0, col*7)
		if col == 0 {
			enqueue(weekStart.Month())
		}
		for i := range 7 {
			d := weekStart.AddDate(0, 0, i)
			if d.Day() == 1 {
				enqueue(d.Month())
			}
		}
		if len(queue) == 0 {
			continue
		}
		x := gutterW + col
		if x < lastEnd+1 || x+1 > gridW {
			continue
		}
		mon := queue[0]
		queue = queue[1:]
		// Single-letter month for compact year grid.
		ab := []rune(time.Date(2000, mon, 1, 0, 0, 0, 0, time.UTC).Format("Jan"))[0]
		buf[x] = ab
		placed = mon
		lastEnd = x
	}
	return styled(label()).Render(string(buf[:gridW]))
}

func renderAuthors(th Theme, s gitstat.Summary) string {
	title := styled(th.Accent).Bold(true).Render("Authors") +
		styled(muted()).Render(" · "+s.Base+" · "+s.Range)
	if len(s.AuthorList) == 0 {
		return title + "\n" + styled(muted()).Render("No commits in this window.")
	}
	maxC := int64(s.AuthorList[0].Commits)
	const barW = 28
	const nameW = 24
	rows := []string{title, ""}
	limit := min(len(s.AuthorList), 10)
	for i := 0; i < limit; i++ {
		a := s.AuthorList[i]
		name := truncate(a.Name, nameW)
		filled := barCells(int64(a.Commits), maxC, barW)
		row := padRight(styled(value()).Render(name), nameW) + " " +
			hbar(th.Accent, filled, barW) + "  " +
			styled(value()).Render(gitstat.FormatInt(a.Commits))
		rows = append(rows, row)
	}
	topPct := 0
	if s.Commits > 0 {
		topPct = int(math.Round(float64(s.AuthorList[0].Commits) / float64(s.Commits) * 100))
	}
	rows = append(rows, "", styled(muted()).Render(
		fmt.Sprintf("%d authors · top owns %d%%", s.Authors, topPct)))
	return strings.Join(rows, "\n")
}

func renderHours(th Theme, s gitstat.Summary) string {
	title := styled(th.Accent).Bold(true).Render("Hours") +
		styled(muted()).Render(" · when commits land")
	var maxV int64
	for _, v := range s.Hours {
		if int64(v) > maxV {
			maxV = int64(v)
		}
	}
	if maxV == 0 {
		return title + "\n" + styled(muted()).Render("No commits in this window.")
	}
	const barW = 40
	rows := []string{title, ""}
	for h := 0; h < 24; h++ {
		v := int64(s.Hours[h])
		if v == 0 {
			continue
		}
		filled := barCells(v, maxV, barW)
		c := th.level(v, maxV)
		if h == s.PeakHour {
			c = th.Accent
		}
		rows = append(rows, fmt.Sprintf("%02d  ", h)+hbar(c, filled, barW))
	}
	rows = append(rows, "", styled(muted()).Render(
		fmt.Sprintf("peak %s · author time", gitstat.FormatHour(s.PeakHour))))
	return strings.Join(rows, "\n")
}

func renderPunchcard(th Theme, s gitstat.Summary) string {
	title := styled(th.Accent).Bold(true).Render("Punchcard") +
		styled(muted()).Render(" · weekday × hour")
	var gridMax, peakWd, peakHr int
	for wd := range 7 {
		for h := range 24 {
			if v := s.Punch[wd][h]; v > gridMax {
				gridMax, peakWd, peakHr = v, wd, h
			}
		}
	}
	if gridMax == 0 {
		return title + "\n" + styled(muted()).Render("No activity in this window.")
	}
	plain := ascii()
	ruler := make([]rune, 24)
	for i := range ruler {
		ruler[i] = ' '
	}
	put := func(pos int, s string) {
		for i := 0; i < len(s) && pos+i < 24; i++ {
			ruler[pos+i] = rune(s[i])
		}
	}
	put(0, "0")
	put(6, "6")
	put(12, "12")
	put(18, "18")
	put(22, "23")

	gut := [7]string{"Sun ", "Mon ", "Tue ", "Wed ", "Thu ", "Fri ", "Sat "}
	rows := []string{title, "", styled(label()).Render("    " + string(ruler))}
	for wd := range 7 {
		var sb strings.Builder
		sb.WriteString(styled(label()).Render(gut[wd]))
		for h := range 24 {
			v := int64(s.Punch[wd][h])
			peak := wd == peakWd && h == peakHr
			sb.WriteString(punchCell(th, v, int64(gridMax), peak, plain))
		}
		rows = append(rows, sb.String())
	}
	rows = append(rows, "", styled(muted()).Render(
		fmt.Sprintf("busiest: %s %s", weekdayNames[peakWd], gitstat.FormatHour(peakHr))))
	return strings.Join(rows, "\n")
}

func punchLevel(v, max int64) int {
	if v <= 0 {
		return 0
	}
	if max <= 1 {
		return 4
	}
	lvl := 1 + int(math.Log1p(float64(v))/math.Log1p(float64(max))*3)
	return min(lvl, 4)
}

func punchCell(th Theme, v, max int64, peak, plain bool) string {
	if v <= 0 {
		if plain {
			return "·"
		}
		return styled(muted()).Render("·")
	}
	if peak {
		if plain {
			return "█"
		}
		return styled(th.Accent).Bold(true).Render("█")
	}
	lvl := punchLevel(v, max)
	if plain {
		return shadeGlyphs1[lvl]
	}
	return styled(th.Ramp[lvl]).Render("█")
}

func renderTrend(th Theme, s gitstat.Summary) string {
	title := styled(th.Accent).Bold(true).Render("Trend") +
		styled(muted()).Render(fmt.Sprintf(" · daily commits · last %dd", s.DailyDays))
	var total int64
	for _, v := range s.Daily {
		total += v
	}
	if total == 0 {
		return title + "\n" + styled(muted()).Render("No commits in this window.")
	}
	rows := []string{title, "", sparkline(th, s.Daily), ""}
	if s.ActiveDays > 0 {
		rows = append(rows, leaderRow("avg / active day",
			fmt.Sprintf("%.1f", float64(s.Commits)/float64(s.ActiveDays)), 40))
	}
	if len(s.TopDays) > 0 {
		d := s.TopDays[0]
		rows = append(rows, leaderRow("busiest day",
			d.Date.Format("Mon Jan 2")+" · "+gitstat.FormatInt(d.Commits), 40))
	}
	if len(s.Daily) >= 14 {
		n := len(s.Daily)
		var last7, prev7 int64
		for _, v := range s.Daily[n-7:] {
			last7 += v
		}
		for _, v := range s.Daily[n-14 : n-7] {
			prev7 += v
		}
		rows = append(rows, leaderRow("week Δ", wowDelta(last7, prev7), 40))
	}
	return strings.Join(rows, "\n")
}

func wowDelta(last, prev int64) string {
	if prev == 0 {
		if last == 0 {
			return "0%"
		}
		return "+∞%"
	}
	pct := int(math.Round(float64(last-prev) / float64(prev) * 100))
	if pct > 0 {
		return fmt.Sprintf("+%d%%", pct)
	}
	return fmt.Sprintf("%d%%", pct)
}

func renderTopDays(th Theme, s gitstat.Summary) string {
	title := styled(th.Accent).Bold(true).Render("Busiest days")
	if len(s.TopDays) == 0 {
		return title + "\n" + styled(muted()).Render("No commits in this window.")
	}
	maxC := int64(s.TopDays[0].Commits)
	const barW = 24
	rows := []string{title, ""}
	for _, d := range s.TopDays {
		filled := barCells(int64(d.Commits), maxC, barW)
		rows = append(rows,
			styled(value()).Render(d.Day)+"  "+
				hbar(th.Accent, filled, barW)+"  "+
				styled(value()).Render(gitstat.FormatInt(d.Commits)))
	}
	return strings.Join(rows, "\n")
}

func renderWeekday(th Theme, s gitstat.Summary) string {
	title := styled(th.Accent).Bold(true).Render("Weekday")
	order := []int{1, 2, 3, 4, 5, 6, 0} // Mon-first
	var maxV, total int64
	busiest, busiestN := 0, 0
	for i, v := range s.Weekday {
		total += int64(v)
		if int64(v) > maxV {
			maxV = int64(v)
		}
		if v > busiestN {
			busiestN, busiest = v, i
		}
	}
	if total == 0 {
		return title + "\n" + styled(muted()).Render("No commits in this window.")
	}
	const barW = 36
	rows := []string{title, ""}
	for _, wd := range order {
		v := int64(s.Weekday[wd])
		filled := barCells(v, maxV, barW)
		rows = append(rows,
			styled(label()).Render(weekdayNames[wd]+"  ")+
				hbar(th.level(v, maxV), filled, barW))
	}
	weekend := s.Weekday[0] + s.Weekday[6]
	wpct := int(math.Round(float64(weekend) / float64(total) * 100))
	rows = append(rows, "", styled(muted()).Render(
		fmt.Sprintf("weekend %d%% · busiest %s", wpct, weekdayNames[busiest])))
	return strings.Join(rows, "\n")
}

func renderBranch(th Theme, s gitstat.Summary) string {
	title := styled(th.Accent).Bold(true).Render("Branch") +
		styled(muted()).Render(" · HEAD vs "+s.Base)
	status := "ok"
	if !s.Fetched && s.FetchNote != "" {
		status = s.FetchNote
	}
	rows := []string{
		title,
		"",
		leaderRow("branch", s.Head, contentW),
		leaderRow("base", s.Base, contentW),
		leaderRow("files", gitstat.FormatInt(s.Files), contentW),
		leaderRow("added", "+"+gitstat.FormatInt(s.Added), contentW),
		leaderRow("removed", "-"+gitstat.FormatInt(s.Removed), contentW),
		leaderRow("status", status, contentW),
	}
	return strings.Join(rows, "\n")
}

// ---- helpers ----

var sparkBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

func sparkline(th Theme, vals []int64) string {
	if len(vals) == 0 {
		return ""
	}
	var maxV int64
	for _, v := range vals {
		if v > maxV {
			maxV = v
		}
	}
	plain := ascii()
	denom := math.Log1p(float64(maxV))
	var sb strings.Builder
	for _, v := range vals {
		if v <= 0 {
			if plain {
				sb.WriteRune(sparkBlocks[0])
			} else {
				sb.WriteString(styled(muted()).Render(string(sparkBlocks[0])))
			}
			continue
		}
		idx := 1
		if denom > 0 {
			idx = 1 + int(math.Round(math.Log1p(float64(v))/denom*float64(len(sparkBlocks)-2)))
		}
		idx = min(idx, len(sparkBlocks)-1)
		g := string(sparkBlocks[idx])
		if plain {
			sb.WriteString(g)
		} else {
			sb.WriteString(styled(th.level(v, maxV)).Render(g))
		}
	}
	return sb.String()
}

func hbar(c color.Color, filled, width int) string {
	if width <= 0 {
		return ""
	}
	filled = max(min(filled, width), 0)
	return styled(c).Render(strings.Repeat("█", filled)) + strings.Repeat(" ", width-filled)
}

func barCells(v, maxV int64, width int) int {
	if maxV <= 0 || v <= 0 {
		return 0
	}
	n := int(float64(v) / float64(maxV) * float64(width))
	if n < 1 {
		n = 1
	}
	if n > width {
		n = width
	}
	return n
}

func leaderRow(key, val string, w int) string {
	dots := max(w-lipgloss.Width(key)-lipgloss.Width(val)-2, 2)
	return styled(label()).Render(key) +
		styled(muted()).Render(" "+strings.Repeat("·", dots)+" ") +
		styled(value()).Render(val)
}

func rightAlign(left, right string, w int) string {
	gap := max(w-lipgloss.Width(left)-lipgloss.Width(right), 1)
	return left + strings.Repeat(" ", gap) + right
}

func padRight(s string, w int) string {
	if d := w - lipgloss.Width(s); d > 0 {
		return s + strings.Repeat(" ", d)
	}
	return s
}

func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	if w <= 1 {
		return "…"
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)+"…") > w {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}
