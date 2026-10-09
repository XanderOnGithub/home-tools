package discord

import (
	"errors"
	"math"
)

// Sensitivity conversion, carried over from Starport-Assistant (decision
// #42). Every game's sensitivity is expressed relative to Overwatch 2 at
// 1.0: the same physical turn (cm per 360°) at the same DPI.

// sensGame is one game the converter knows.
type sensGame struct {
	Slug  string  // option value in the command
	Name  string  // shown to people
	Ratio float64 // this game's sensitivity for Overwatch 2's 1.0
}

// sensGames is sorted by name: the command's choice list shows in this
// order. Lookups go through sensBySlug (O(1)).
var sensGames = []sensGame{
	{"apex", "Apex Legends", 0.3},
	{"arcraiders", "ARC Raiders", 4.849},
	{"battlefield6", "Battlefield 6", 2.468},
	{"cod", "Call of Duty", 1.0},
	{"cs2", "Counter-Strike 2", 0.3},
	{"deadlock", "Deadlock", 0.150},
	{"tarkov", "Escape From Tarkov", 0.053},
	{"fortnite", "Fortnite", 1.188},
	{"marvelrivals", "Marvel Rivals", 0.37714},
	{"overwatch2", "Overwatch 2", 1.0},
	{"roblox", "Roblox", 0.017},
	{"rust", "Rust", 0.059},
	{"splitgate2", "Splitgate 2", 0.591},
	{"tf2", "Team Fortress 2", 0.3},
	{"thefinals", "THE FINALS", 6.6},
	{"valheim", "Valheim", 0.132},
	{"valorant", "Valorant", 0.09434},
}

var sensBySlug = func() map[string]sensGame {
	m := make(map[string]sensGame, len(sensGames))
	for _, g := range sensGames {
		m[g.Slug] = g
	}
	return m
}()

// overwatchYaw is Overwatch's degrees per mouse count at sensitivity 1.
const overwatchYaw = 0.0066

// sensResult is a conversion: the new sensitivity, and the physical
// distance for a full turn when the DPI is known (0 otherwise).
type sensResult struct {
	From, To sensGame
	Sens     float64 // rounded to 3 decimals, as games take it
	CM360    float64 // cm of mouse travel per 360° turn; 0 = no DPI given
}

// convertSens converts value from one game to another. fromDPI/toDPI are
// optional (0): with both, the result also makes up for a DPI change.
func convertSens(from, to string, value float64, fromDPI, toDPI int) (sensResult, error) {
	f, ok1 := sensBySlug[from]
	t, ok2 := sensBySlug[to]
	switch {
	case !ok1 || !ok2:
		return sensResult{}, errors.New("unknown game")
	case !(value > 0) || math.IsInf(value, 0):
		return sensResult{}, errors.New("sensitivity must be more than 0")
	case fromDPI < 0 || toDPI < 0:
		return sensResult{}, errors.New("DPI can't be negative")
	}
	// Same turn in both games: scale out of the first ratio, into the second.
	sens := value / f.Ratio * t.Ratio
	if fromDPI > 0 && toDPI > 0 {
		sens *= float64(fromDPI) / float64(toDPI) // more counts per inch → less sensitivity
	}
	r := sensResult{From: f, To: t, Sens: math.Round(sens*1000) / 1000}
	if fromDPI > 0 {
		// Degrees per inch of travel, then inches → cm for a full turn.
		degPerInch := value / f.Ratio * overwatchYaw * float64(fromDPI)
		r.CM360 = 360 / degPerInch * 2.54
	}
	return r, nil
}
