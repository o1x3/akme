package core

import (
	"os"
	"sync"
)

// cursorDashApplied is set by applyCursorDashboard when enrichment succeeds.
// Guarded because Load may be called from tests in parallel packages, but
// within a single process the token CLI is single-threaded.
var (
	cursorDashMu      sync.Mutex
	cursorDashApplied bool
)

func setCursorDashApplied(ok bool) {
	cursorDashMu.Lock()
	cursorDashApplied = ok
	cursorDashMu.Unlock()
}

func getCursorDashApplied() bool {
	cursorDashMu.Lock()
	defer cursorDashMu.Unlock()
	return cursorDashApplied
}

// CursorStatus is a lightweight probe of why Cursor collection may be empty
// or estimated on a given machine/account. Activity (sessions/heatmap) is
// machine-local; billed tokens follow the logged-in account via the dashboard.
type CursorStatus struct {
	IDEFound    bool   `json:"ide_found"`
	IDEPath     string `json:"ide_path,omitempty"`
	CLIStores   int    `json:"cli_stores"`
	AuthOK      bool   `json:"auth_ok"`
	DashboardOK bool   `json:"dashboard_ok"`
	LocalOnly   bool   `json:"local_only"`
	Hint        string `json:"hint,omitempty"`
}

// ProbeCursorStatus inspects on-disk Cursor sources and the last dashboard
// enrich attempt. Safe to call after Load("cursor") / Load("all").
func ProbeCursorStatus() CursorStatus {
	st := CursorStatus{LocalOnly: cursorDashDisabled()}
	for _, p := range cursorIDEPaths() {
		if _, err := os.Stat(p); err == nil {
			st.IDEFound = true
			st.IDEPath = p
			break
		}
	}
	st.CLIStores = len(cursorCLIPaths())
	_, st.AuthOK = resolveCursorSession()
	st.DashboardOK = getCursorDashApplied()
	st.Hint = st.hint()
	return st
}

func (st CursorStatus) hint() string {
	switch {
	case !st.IDEFound && st.CLIStores == 0:
		return "no Cursor state.vscdb or CLI stores found (activity is machine-local and does not sync across machines/accounts)"
	case st.LocalOnly:
		return "AKME_TOKEN_CURSOR_LOCAL set; skipping Cursor dashboard (local estimates only)"
	case !st.AuthOK:
		return "not logged in to Cursor (or session expired); billed tokens need a dashboard session — sign in or set AKME_CURSOR_SESSION_TOKEN"
	case !st.DashboardOK:
		return "Cursor dashboard returned no billed usage; keeping local estimates (sessions stay machine-local — confirm the same Cursor account, or try AKME_TOKEN_CURSOR_LOCAL=1)"
	default:
		return ""
	}
}

// CursorHintFor returns a one-line diagnostic when Cursor data is missing,
// still estimated, or shows activity with 0 billed tokens. Empty string means
// no hint.
func CursorHintFor(s Summary) string {
	if s.Harness != Cursor {
		return ""
	}
	st := ProbeCursorStatus()
	// Healthy billed totals from the dashboard.
	if s.TotalTokens > 0 && !s.TokensEstimated && st.DashboardOK {
		return ""
	}
	// The confusing cross-machine case: sessions/messages present, tokens 0.
	if (s.Messages > 0 || s.Sessions > 0) && s.TotalTokens == 0 {
		if st.Hint != "" {
			return st.Hint
		}
		return "Cursor activity found but 0 billed tokens — confirm the same logged-in account on this machine, or set AKME_TOKEN_CURSOR_LOCAL=1 for local estimates"
	}
	if !s.HasData() || s.TokensEstimated {
		return st.Hint
	}
	return ""
}
