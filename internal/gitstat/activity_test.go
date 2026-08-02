package gitstat

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadAndSummarizeActivity(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	git(t, root, "init", repo)
	git(t, repo, "config", "user.email", "alice@example.com")
	git(t, repo, "config", "user.name", "Alice")
	git(t, repo, "checkout", "-b", "main")

	now := time.Date(2026, 8, 2, 15, 0, 0, 0, time.Local)
	writeDatedCommit(t, repo, "a.txt", "one\n", "first", now)
	writeDatedCommit(t, repo, "b.txt", "two\n", "second", now.AddDate(0, 0, -10))
	writeDatedCommit(t, repo, "c.txt", "three\n", "third", now.AddDate(0, 0, -40))

	git(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	git(t, repo, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")

	agg, err := Load(context.Background(), repo)
	if err != nil {
		t.Fatal(err)
	}
	if agg.Base != "origin/main" {
		t.Fatalf("base = %q, want origin/main", agg.Base)
	}
	if agg.Commits < 3 {
		t.Fatalf("commits = %d, want >= 3", agg.Commits)
	}
	todayKey := civil(now).Format("2006-01-02")
	if agg.ByDay[todayKey] < 1 {
		t.Fatalf("today cell empty: %#v", agg.ByDay)
	}

	year := Summarize(agg, RangeYear, now)
	if !year.HasData() {
		t.Fatal("year summary empty")
	}
	if year.Heatmap.Weeks != 52 {
		t.Fatalf("weeks = %d, want 52", year.Heatmap.Weeks)
	}
	row := int(civil(now).Weekday())
	col := year.Heatmap.Weeks - 1
	if year.Heatmap.Cells[row][col] < 1 {
		t.Fatalf("today heatmap cell = %d", year.Heatmap.Cells[row][col])
	}

	week := Summarize(agg, Range7d, now)
	if week.Commits < 1 {
		t.Fatalf("7d commits = %d, want >= 1", week.Commits)
	}
	if week.Commits > year.Commits {
		t.Fatalf("7d commits %d > year %d", week.Commits, year.Commits)
	}

	month := Summarize(agg, Range30d, now)
	if month.Commits < 2 {
		t.Fatalf("30d commits = %d, want >= 2", month.Commits)
	}
	if len(month.AuthorList) < 1 || month.AuthorList[0].Name != "Alice" {
		t.Fatalf("authors = %#v", month.AuthorList)
	}
}

func TestBuildHeatmapBlanksFutureAndOutOfRange(t *testing.T) {
	now := time.Date(2026, 8, 2, 12, 0, 0, 0, time.Local)
	old := civil(now).AddDate(0, 0, -20)
	agg := &Aggregate{ByDay: map[string]int{
		civil(now).Format("2006-01-02"):                   5,
		civil(now).AddDate(0, 0, -3).Format("2006-01-02"): 2,
		old.Format("2006-01-02"):                          9,
	}}
	h := buildHeatmap(agg, Range7d, now)
	first := h.FirstDay
	for col := range h.Weeks {
		for row := range 7 {
			d := first.AddDate(0, 0, col*7+row)
			if d.Equal(old) && h.Cells[row][col] != 0 {
				t.Fatalf("out-of-range cell = %d, want 0", h.Cells[row][col])
			}
			if d.After(civil(now)) && h.Cells[row][col] != -1 {
				t.Fatalf("future cell = %d, want -1", h.Cells[row][col])
			}
		}
	}
	if h.Max != 5 {
		t.Fatalf("max = %d, want 5 (in-window only)", h.Max)
	}
}

func writeDatedCommit(t *testing.T, repo, file, body, msg string, when time.Time) {
	t.Helper()
	write(t, filepath.Join(repo, file), body)
	git(t, repo, "add", file)
	envDate := when.Format(time.RFC3339)
	cmd := exec.Command("git", "commit", "-m", msg)
	cmd.Dir = repo
	cmd.Env = append(cmd.Environ(),
		"GIT_AUTHOR_DATE="+envDate,
		"GIT_COMMITTER_DATE="+envDate,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("commit failed: %v\n%s", err, out)
	}
}
