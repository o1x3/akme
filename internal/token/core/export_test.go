package core

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewSummaryJSON(t *testing.T) {
	now := time.Date(2026, 6, 29, 12, 0, 0, 0, time.Local)
	a := newAggregate(Claude)
	a.Sessions = 3
	a.InputTokens, a.OutputTokens, a.CacheReadTokens = 1_000_000, 500_000, 4_000_000
	a.ByDayMsgs["2026-06-29"] = 10
	a.ByDayTokens["2026-06-29"] = 5_500_000
	a.ByDayModelTok["2026-06-29"] = map[string]int64{"claude-opus-4-8": 5_500_000}
	a.ByDayModelMsg["2026-06-29"] = map[string]int{"claude-opus-4-8": 10}
	a.TokensEstimated = true

	all := NewSummaryJSON(Summarize(a, RangeAll, now, true), now)
	if all.SchemaVersion != 2 {
		t.Errorf("schema_version = %d, want 2", all.SchemaVersion)
	}
	if !all.IncludeCache {
		t.Error("include_cache must be true when Summarize asked for cache")
	}
	if !all.TokenSplitExact {
		t.Error("all-time split must be marked exact")
	}
	if all.Tokens.Total != a.TotalTokens() {
		t.Errorf("tokens.total = %d, want %d", all.Tokens.Total, a.TotalTokens())
	}
	if all.Tokens.CacheRead != 4_000_000 {
		t.Errorf("cache_read = %d, want 4,000,000", all.Tokens.CacheRead)
	}
	if all.EstUSD <= 0 || !all.EstUSDAllPriced {
		t.Errorf("expected priced spend, got usd=%.4f allpriced=%v", all.EstUSD, all.EstUSDAllPriced)
	}
	if !all.TokensEstimated {
		t.Error("tokens_estimated must carry through from the aggregate")
	}

	fresh := NewSummaryJSON(Summarize(a, RangeAll, now, false), now)
	if fresh.IncludeCache {
		t.Error("include_cache must be false by default")
	}
	wantFresh := a.CountedTokens(false)
	if fresh.Tokens.Total != wantFresh {
		t.Errorf("fresh tokens.total = %d, want %d", fresh.Tokens.Total, wantFresh)
	}
	if fresh.Tokens.CacheRead != 4_000_000 {
		t.Errorf("fresh cache_read = %d, want ledger 4,000,000", fresh.Tokens.CacheRead)
	}

	// windowed range must flag the split as inexact
	wk := NewSummaryJSON(Summarize(a, Range7d, now, false), now)
	if wk.TokenSplitExact {
		t.Error("windowed split must NOT be marked exact")
	}

	// round-trips as valid JSON with snake_case keys
	b, err := json.Marshal(all)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, key := range []string{"schema_version", "token_split_exact", "tokens_estimated", "cache_read", "est_usd", "fav_model", "include_cache"} {
		if !strings.Contains(string(b), key) {
			t.Errorf("json missing key %q", key)
		}
	}
	if all.CursorStatus != nil {
		t.Error("cursor_status should be omitted for non-cursor harnesses")
	}

	cur := NewSummaryJSON(Summarize(newAggregate(Cursor), RangeAll, now, false), now)
	if cur.CursorStatus == nil {
		t.Fatal("cursor_status missing for cursor harness")
	}
	cb, err := json.Marshal(cur)
	if err != nil {
		t.Fatalf("marshal cursor: %v", err)
	}
	if !strings.Contains(string(cb), "cursor_status") {
		t.Error("json missing cursor_status")
	}
}

func TestCostCacheSaving(t *testing.T) {
	// 10M cache-read tokens on Opus: read priced 1.5, input 15 → saving ≈
	// 10M * (15-1.5)/1e6 = $135.
	models := []ModelStat{{ID: "claude-opus-4-8", Name: "Opus 4.8", Tokens: 10_000_000}}
	cb := EstimateCost(models, 0, 0, 10_000_000, 0)
	if cb.CacheSavingUSD < 134 || cb.CacheSavingUSD > 136 {
		t.Errorf("cache saving = %.2f, want ~135", cb.CacheSavingUSD)
	}
}
