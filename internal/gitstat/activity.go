package gitstat

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Activity ranges. "year" is the collected window (~52 weeks); 30d/7d filter within it.
const (
	RangeYear = "year"
	Range30d  = "30d"
	Range7d   = "7d"
)

const (
	heatmapWeeks = 52
	collectDays  = 52*7 + 6 // cover the leftmost Sunday before a full 52-week grid
)

// Aggregate is one load of first-parent commits on a repo's default branch.
type Aggregate struct {
	Path      string
	Name      string
	Base      string
	Head      string
	Fetched   bool
	FetchNote string

	// Branch diff vs base (collected once for the branch tab).
	Added   int
	Removed int
	Files   int

	ByDay       map[string]int            // YYYY-MM-DD → commit count
	ByDayHour   map[string][24]int        // YYYY-MM-DD → hour histogram
	ByAuthorDay map[string]map[string]int // author → day → commits
	Commits     int
}

// Heatmap is a GitHub-style contribution grid: 7 rows (Sun..Sat) by N weeks.
// Cell values are commit counts; -1 means future (blank).
type Heatmap struct {
	Cells    [7][]int64
	Weeks    int
	Max      int64
	FirstDay time.Time
}

// AuthorStat is one author ranked by commit count.
type AuthorStat struct {
	Name    string
	Commits int
}

// DayStat is one civil day's commit count.
type DayStat struct {
	Day     string
	Date    time.Time
	Commits int
}

// Summary is a range-filtered view derived from an Aggregate.
type Summary struct {
	Path      string
	Name      string
	Base      string
	Head      string
	Range     string
	Fetched   bool
	FetchNote string

	Commits       int
	ActiveDays    int
	Authors       int
	PeakHour      int
	CurrentStreak int
	LongestStreak int

	Heatmap    Heatmap
	AuthorList []AuthorStat
	Hours      [24]int
	Weekday    [7]int
	Punch      [7][24]int
	TopDays    []DayStat
	Daily      []int64
	DailyDays  int

	Added   int
	Removed int
	Files   int
}

// RangeDays returns the number of civil days a range covers, or 0 for year/all-window.
func RangeDays(r string) int {
	switch r {
	case Range7d:
		return 7
	case Range30d:
		return 30
	default:
		return 0
	}
}

// HasData reports whether the summary has any commits in the selected window.
func (s Summary) HasData() bool {
	return s.Commits > 0
}

// Load collects first-parent activity on the folder's default branch for ~one year.
func Load(ctx context.Context, folder string) (*Aggregate, error) {
	if strings.TrimSpace(folder) == "" {
		return nil, errors.New("folder cannot be empty")
	}

	clean := filepath.Clean(folder)
	info, err := os.Stat(clean)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", clean, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s: not a directory", clean)
	}
	if err := run(ctx, clean, "rev-parse", "--is-inside-work-tree"); err != nil {
		return nil, fmt.Errorf("%s: not a git repository", clean)
	}

	base := defaultBranch(ctx, clean)
	fetched := true
	fetchNote := ""
	if err := fetchBase(ctx, clean, base); err != nil {
		fetched = false
		fetchNote = "fetch failed; using local refs"
	}

	head := output(ctx, clean, "rev-parse", "--abbrev-ref", "HEAD")
	if head == "" {
		head = "HEAD"
	}

	added, removed, files, err := diffNumstat(ctx, clean, base)
	if err != nil {
		// Detached or missing base: keep zeros rather than failing the whole dashboard.
		added, removed, files = 0, 0, 0
	}

	agg := &Aggregate{
		Path:        clean,
		Name:        filepath.Base(clean),
		Base:        base,
		Head:        head,
		Fetched:     fetched,
		FetchNote:   fetchNote,
		Added:       added,
		Removed:     removed,
		Files:       files,
		ByDay:       map[string]int{},
		ByDayHour:   map[string][24]int{},
		ByAuthorDay: map[string]map[string]int{},
	}

	since := civil(time.Now()).AddDate(0, 0, -(collectDays - 1))
	if err := collectLog(ctx, clean, base, since, agg); err != nil {
		return nil, fmt.Errorf("%s: log against %s failed: %w", clean, base, err)
	}
	return agg, nil
}

func collectLog(ctx context.Context, dir, base string, since time.Time, agg *Aggregate) error {
	raw, err := outputErr(ctx, dir,
		"log", "--first-parent",
		"--since="+since.Format("2006-01-02"),
		"--pretty=format:%aI%x00%an",
		base,
	)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(raw))
	// Commits can be many; allow long lines.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		ts, author, ok := strings.Cut(line, "\x00")
		if !ok {
			continue
		}
		author = strings.TrimSpace(author)
		if author == "" {
			author = "(unknown)"
		}
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(ts))
		if err != nil {
			// Fallback: date-only
			if d, dErr := time.ParseInLocation("2006-01-02", strings.TrimSpace(ts), time.Local); dErr == nil {
				t = d
			} else {
				continue
			}
		}
		day := t.Format("2006-01-02")
		hour := t.Hour()
		agg.ByDay[day]++
		h := agg.ByDayHour[day]
		h[hour]++
		agg.ByDayHour[day] = h
		if agg.ByAuthorDay[author] == nil {
			agg.ByAuthorDay[author] = map[string]int{}
		}
		agg.ByAuthorDay[author][day]++
		agg.Commits++
	}
	return scanner.Err()
}

// Summarize derives a Summary for the given range relative to now.
func Summarize(a *Aggregate, rng string, now time.Time) Summary {
	if a == nil {
		return Summary{Range: rng, PeakHour: -1}
	}
	if rng != RangeYear && rng != Range30d && rng != Range7d {
		rng = RangeYear
	}

	today := civil(now)
	var cutoff time.Time
	if d := RangeDays(rng); d > 0 {
		cutoff = today.AddDate(0, 0, -(d - 1))
	}

	inWindow := func(dayStr string) bool {
		d, err := time.ParseInLocation("2006-01-02", dayStr, time.Local)
		if err != nil || d.After(today) {
			return false
		}
		return cutoff.IsZero() || !d.Before(cutoff)
	}

	s := Summary{
		Path:      a.Path,
		Name:      a.Name,
		Base:      a.Base,
		Head:      a.Head,
		Range:     rng,
		Fetched:   a.Fetched,
		FetchNote: a.FetchNote,
		Added:     a.Added,
		Removed:   a.Removed,
		Files:     a.Files,
		PeakHour:  -1,
		DailyDays: rangeDailyDays(rng),
	}

	for day, n := range a.ByDay {
		if !inWindow(day) || n <= 0 {
			continue
		}
		s.Commits += n
		s.ActiveDays++
	}

	authors := make([]AuthorStat, 0, len(a.ByAuthorDay))
	for name, days := range a.ByAuthorDay {
		n := 0
		for day, c := range days {
			if inWindow(day) {
				n += c
			}
		}
		if n > 0 {
			authors = append(authors, AuthorStat{Name: name, Commits: n})
		}
	}
	sort.Slice(authors, func(i, j int) bool {
		if authors[i].Commits != authors[j].Commits {
			return authors[i].Commits > authors[j].Commits
		}
		return authors[i].Name < authors[j].Name
	})
	s.AuthorList = authors
	s.Authors = len(authors)

	s.Hours = hoursIn(a, inWindow)
	s.Weekday = weekdayIn(a, inWindow)
	s.Punch = punchIn(a, inWindow)
	s.TopDays = topDaysIn(a, inWindow, 10)
	s.Daily = dailySeries(a, now, s.DailyDays)
	s.PeakHour = peakHour(s.Hours)
	s.CurrentStreak, s.LongestStreak = streaks(a, now, inWindow)
	s.Heatmap = buildHeatmap(a, rng, now)
	return s
}

func rangeDailyDays(rng string) int {
	switch rng {
	case Range7d:
		return 7
	case Range30d:
		return 30
	default:
		return heatmapWeeks * 7
	}
}

func hoursIn(a *Aggregate, keep func(string) bool) [24]int {
	var hours [24]int
	for d, h := range a.ByDayHour {
		if keep(d) {
			for i, v := range h {
				hours[i] += v
			}
		}
	}
	return hours
}

func weekdayIn(a *Aggregate, keep func(string) bool) [7]int {
	var wd [7]int
	for d, n := range a.ByDay {
		if keep(d) && n > 0 {
			if t, err := time.ParseInLocation("2006-01-02", d, time.Local); err == nil {
				wd[int(t.Weekday())] += n
			}
		}
	}
	return wd
}

func punchIn(a *Aggregate, keep func(string) bool) [7][24]int {
	var grid [7][24]int
	for d, h := range a.ByDayHour {
		if !keep(d) {
			continue
		}
		t, err := time.ParseInLocation("2006-01-02", d, time.Local)
		if err != nil {
			continue
		}
		wd := int(t.Weekday())
		for i, v := range h {
			grid[wd][i] += v
		}
	}
	return grid
}

func topDaysIn(a *Aggregate, keep func(string) bool, n int) []DayStat {
	out := make([]DayStat, 0, len(a.ByDay))
	for d, c := range a.ByDay {
		if !keep(d) || c <= 0 {
			continue
		}
		t, err := time.ParseInLocation("2006-01-02", d, time.Local)
		if err != nil {
			continue
		}
		out = append(out, DayStat{Day: d, Date: t, Commits: c})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Commits != out[j].Commits {
			return out[i].Commits > out[j].Commits
		}
		return out[i].Day > out[j].Day
	})
	if n > 0 && len(out) > n {
		out = out[:n]
	}
	return out
}

func dailySeries(a *Aggregate, end time.Time, n int) []int64 {
	if n <= 0 {
		return nil
	}
	day := civil(end)
	out := make([]int64, n)
	for i := n - 1; i >= 0; i-- {
		out[i] = int64(a.ByDay[day.Format("2006-01-02")])
		day = day.AddDate(0, 0, -1)
	}
	return out
}

func peakHour(hours [24]int) int {
	best, idx := 0, -1
	for i, v := range hours {
		if v > best {
			best, idx = v, i
		}
	}
	return idx
}

func streaks(a *Aggregate, now time.Time, keep func(string) bool) (current, longest int) {
	day := civil(now)
	// Current streak: consecutive days ending today (or yesterday if today empty).
	start := day
	if a.ByDay[day.Format("2006-01-02")] == 0 {
		start = day.AddDate(0, 0, -1)
	}
	for d := start; ; d = d.AddDate(0, 0, -1) {
		key := d.Format("2006-01-02")
		if !keep(key) || a.ByDay[key] <= 0 {
			break
		}
		current++
		if current > 400 {
			break
		}
	}

	// Longest within collected keys that pass keep.
	days := make([]string, 0, len(a.ByDay))
	for d, n := range a.ByDay {
		if keep(d) && n > 0 {
			days = append(days, d)
		}
	}
	sort.Strings(days)
	run := 0
	var prev time.Time
	for i, d := range days {
		t, err := time.ParseInLocation("2006-01-02", d, time.Local)
		if err != nil {
			continue
		}
		if i > 0 && t.Equal(prev.AddDate(0, 0, 1)) {
			run++
		} else {
			run = 1
		}
		if run > longest {
			longest = run
		}
		prev = t
	}
	return current, longest
}

func buildHeatmap(a *Aggregate, rng string, now time.Time) Heatmap {
	today := civil(now)
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	first := weekStart.AddDate(0, 0, -7*(heatmapWeeks-1))

	var rangeStart time.Time
	if d := RangeDays(rng); d > 0 {
		rangeStart = today.AddDate(0, 0, -(d - 1))
	}

	h := Heatmap{Weeks: heatmapWeeks, FirstDay: first}
	for r := range 7 {
		h.Cells[r] = make([]int64, heatmapWeeks)
	}
	for col := range heatmapWeeks {
		for row := range 7 {
			d := first.AddDate(0, 0, col*7+row)
			switch {
			case d.After(today):
				h.Cells[row][col] = -1
			case !rangeStart.IsZero() && d.Before(rangeStart):
				h.Cells[row][col] = 0
			default:
				v := int64(a.ByDay[d.Format("2006-01-02")])
				h.Cells[row][col] = v
				if v > h.Max {
					h.Max = v
				}
			}
		}
	}
	return h
}

func civil(t time.Time) time.Time {
	y, m, d := t.In(time.Local).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

// FormatHour turns a 0-23 hour into "12 PM" style.
func FormatHour(h int) string {
	if h < 0 {
		return "—"
	}
	ampm := "AM"
	hh := h
	if h == 0 {
		hh = 12
	} else if h == 12 {
		ampm = "PM"
	} else if h > 12 {
		hh = h - 12
		ampm = "PM"
	}
	return fmt.Sprintf("%d %s", hh, ampm)
}

// FormatInt adds thousands separators.
func FormatInt(n int) string {
	s := fmt.Sprintf("%d", n)
	neg := false
	if len(s) > 0 && s[0] == '-' {
		neg, s = true, s[1:]
	}
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	if neg {
		return "-" + string(out)
	}
	return string(out)
}
