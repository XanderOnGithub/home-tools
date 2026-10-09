package discord

import (
	"math"
	"slices"
	"strings"
	"testing"
)

func TestConvertSens(t *testing.T) {
	tests := []struct {
		name            string
		from, to        string
		value           float64
		fromDPI, toDPI  int
		want, wantCM360 float64
		wantErr         bool
	}{
		{"same game", "overwatch2", "overwatch2", 5, 0, 0, 5, 0, false},
		{"ow → valorant", "overwatch2", "valorant", 5, 0, 0, 0.472, 0, false},
		{"valorant → cs2", "valorant", "cs2", 0.4, 0, 0, 1.272, 0, false},
		{"dpi halves", "cs2", "cs2", 1, 800, 1600, 0.5, 51.955, false},
		// OW 5 @ 800 DPI: 5 × 0.0066 × 800 = 26.4°/inch → 360/26.4 × 2.54 cm.
		{"cm/360", "overwatch2", "overwatch2", 5, 800, 0, 5, 34.636, false},
		{"one dpi only: no dpi change", "cs2", "cs2", 1, 800, 0, 1, 51.955, false},
		{"unknown game", "doom", "cs2", 1, 0, 0, 0, 0, true},
		{"zero", "cs2", "cs2", 0, 0, 0, 0, 0, true},
		{"NaN", "cs2", "cs2", math.NaN(), 0, 0, 0, 0, true},
		{"negative dpi", "cs2", "cs2", 1, -5, 800, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertSens(tt.from, tt.to, tt.value, tt.fromDPI, tt.toDPI)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if got.Sens != tt.want || math.Abs(got.CM360-tt.wantCM360) > 0.001 {
				t.Errorf("= %v, %.3f cm; want %v, %.3f cm", got.Sens, got.CM360, tt.want, tt.wantCM360)
			}
		})
	}
}

func TestSensGamesSorted(t *testing.T) {
	// Discord shows choices in list order; at most 25 are allowed.
	if !slices.IsSortedFunc(sensGames, func(a, b sensGame) int { return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)) }) {
		t.Error("sensGames isn't sorted by name")
	}
	if len(sensGames) > 25 || len(sensBySlug) != len(sensGames) {
		t.Errorf("%d games, %d slugs", len(sensGames), len(sensBySlug))
	}
}
