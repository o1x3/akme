package envx

import "testing"

func TestFirstPrefersAkme(t *testing.T) {
	t.Setenv("AKME_BACKGROUND", "light")
	t.Setenv("NX_BACKGROUND", "dark")
	if got := First("AKME_BACKGROUND", "NX_BACKGROUND"); got != "light" {
		t.Fatalf("got %q", got)
	}
}

func TestFirstFallsBackToNX(t *testing.T) {
	t.Setenv("AKME_BACKGROUND", "")
	t.Setenv("NX_BACKGROUND", "dark")
	if got := First("AKME_BACKGROUND", "NX_BACKGROUND"); got != "dark" {
		t.Fatalf("got %q", got)
	}
}

func TestTruthy(t *testing.T) {
	t.Setenv("AKME_NO_UPDATE", "1")
	if !Truthy("AKME_NO_UPDATE", "NX_NO_UPDATE") {
		t.Fatal("expected truthy")
	}
}
