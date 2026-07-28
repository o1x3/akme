package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCursorIDEPathsIncludeWindows(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))

	paths := cursorIDEPaths()
	var sawWin bool
	for _, p := range paths {
		if strings.Contains(p, filepath.Join("AppData", "Roaming", "Cursor")) {
			sawWin = true
		}
	}
	if !sawWin {
		t.Fatalf("cursorIDEPaths missing Windows APPDATA candidate: %v", paths)
	}
}

func TestProbeCursorStatusNoSources(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	t.Setenv("NX_TOKEN_CURSOR_LOCAL", "1")
	os.Unsetenv("NX_CURSOR_SESSION_TOKEN")
	os.Unsetenv("CURSOR_SESSION_TOKEN")
	setCursorDashApplied(false)

	st := ProbeCursorStatus()
	if st.IDEFound || st.CLIStores != 0 || st.AuthOK || st.DashboardOK {
		t.Errorf("unexpected status: %+v", st)
	}
	if st.Hint == "" || !strings.Contains(st.Hint, "machine-local") {
		t.Errorf("hint = %q, want machine-local explanation", st.Hint)
	}
}

func TestProbeCursorStatusFindsIDE(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("APPDATA", filepath.Join(home, "empty-appdata"))
	forceCursorLocal(t)
	setCursorDashApplied(false)

	path := cursorStatePath(home)
	makeSQLiteDB(t, path,
		[]any{`CREATE TABLE cursorDiskKV (key TEXT PRIMARY KEY, value BLOB)`},
		[]any{`CREATE TABLE ItemTable (key TEXT PRIMARY KEY, value BLOB)`},
	)
	st := ProbeCursorStatus()
	if !st.IDEFound || st.IDEPath != path {
		t.Errorf("IDEFound=%v path=%q, want %q", st.IDEFound, st.IDEPath, path)
	}
	if st.AuthOK {
		t.Error("AuthOK = true, want false (no access token)")
	}
	if !strings.Contains(st.Hint, "not logged in") && !strings.Contains(st.Hint, "NX_TOKEN_CURSOR_LOCAL") {
		t.Errorf("hint = %q", st.Hint)
	}
}

func TestCursorHintFor(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	forceCursorLocal(t)
	setCursorDashApplied(false)

	empty := Summary{Harness: Cursor}
	if h := CursorHintFor(empty); h == "" {
		t.Error("empty cursor summary should produce a hint")
	}
	est := Summary{Harness: Cursor, Sessions: 1, Messages: 1, TotalTokens: 10, TokensEstimated: true}
	if h := CursorHintFor(est); h == "" {
		t.Error("estimated cursor summary should produce a hint")
	}
	other := Summary{Harness: Claude, TokensEstimated: true}
	if h := CursorHintFor(other); h != "" {
		t.Errorf("non-cursor hint = %q", h)
	}
}
