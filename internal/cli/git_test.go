package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/o1x3/akme/internal/gitstat/ui"
)

func TestParseGitActivityArgs(t *testing.T) {
	o, err := parseGitActivityArgs([]string{".", "30d", "punchcard", "-i"})
	if err != nil {
		t.Fatal(err)
	}
	if o.folder != "." || o.rng != "30d" || o.tab != ui.TabPunchcard || !o.interactive {
		t.Fatalf("got %#v", o)
	}
}

func TestRunGitActivityStatic(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.email", "t@example.com")
	runGit(t, repo, "config", "user.name", "T")
	runGit(t, repo, "checkout", "-b", "main")
	writeFile(t, filepath.Join(repo, "a.txt"), "x\n")
	runGit(t, repo, "add", "a.txt")
	cmd := exec.Command("git", "commit", "-m", "c")
	cmd.Dir = repo
	when := time.Now().Format(time.RFC3339)
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+when, "GIT_COMMITTER_DATE="+when)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("commit: %v\n%s", err, out)
	}
	runGit(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	runGit(t, repo, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")

	var stdout, stderr bytes.Buffer
	err := New(BuildInfo{Version: "test"}).Run(context.Background(), []string{"git", repo, "overview"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	got := stdout.String()
	for _, want := range []string{"Contributions", "overview", "origin/main"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
