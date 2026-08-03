package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Render configuration — same approach as token/ui: no background paint;
// colours adapt to light/dark terminals; plain mode uses shade glyphs.
var (
	cfgDark  = true
	cfgPlain = false
)

// Configure sets dark/plain render mode for subsequent paints.
func Configure(dark, plain bool) {
	cfgDark = dark
	cfgPlain = plain
}

func adapt(light, dark string) color.Color {
	return lipgloss.LightDark(cfgDark)(lipgloss.Color(light), lipgloss.Color(dark))
}

func label() color.Color { return adapt("#6a6a78", "#8b8b9c") }
func value() color.Color { return adapt("#17171f", "#f2f2f7") }
func muted() color.Color { return adapt("#9a9aa6", "#565668") }
func ascii() bool        { return cfgPlain }
func styled(fg color.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(fg)
}

// Theme is the git activity palette: teal accent + GitHub-green heatmap ramp.
type Theme struct {
	Accent color.Color
	Ramp   [5]color.Color
}

// DefaultTheme returns the GitHub-green contribution palette.
func DefaultTheme() Theme {
	return Theme{
		Accent: adapt("#0e7490", "#67e8f9"), // teal / sky
		Ramp: [5]color.Color{
			adapt("#ebedf0", "#161b22"), // empty
			adapt("#9be9a8", "#0e4429"),
			adapt("#40c463", "#006d32"),
			adapt("#30a14e", "#26a641"),
			adapt("#216e39", "#39d353"), // hottest
		},
	}
}

func (t Theme) levelIndex(v, max int64) int {
	if v <= 0 {
		return 0
	}
	if max <= 0 {
		return 2
	}
	ratio := float64(v) / float64(max)
	switch {
	case ratio >= 0.6:
		return 4
	case ratio >= 0.3:
		return 3
	case ratio >= 0.1:
		return 2
	default:
		return 1
	}
}

func (t Theme) level(v, max int64) color.Color {
	return t.Ramp[t.levelIndex(v, max)]
}

// Single-width shade glyphs for compact 52-week heatmap / punchcard.
var shadeGlyphs1 = [5]string{"·", "░", "▒", "▓", "█"}
