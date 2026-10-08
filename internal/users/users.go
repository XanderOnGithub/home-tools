// Package users holds the household profiles shared by every tool: the
// "Who's using this?" picker. A tool keeps its own per-user data (e.g.
// fitness sessions) keyed by User.ID and checks IDs against this store.
//
// Stored as <dir>/<id>.json (data/users/ in the default layout).
package users

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/XanderOnGithub/home-tools/internal/jsonfile"
)

// ErrInvalid marks data that breaks a model rule. Callers check it with
// errors.Is (e.g. to answer HTTP 400 instead of 500).
var ErrInvalid = errors.New("invalid")

// User is one household profile.
type User struct {
	ID       string `json:"id"` // slug, e.g. "xander"; becomes a file/folder name
	Name     string `json:"name"`
	Color    Color  `json:"color"`              // avatar background and UI accent
	Units    Units  `json:"units"`              // display preference; data is always metric
	Birthday string `json:"birthday,omitempty"` // optional, "YYYY-MM-DD"
	Archived bool   `json:"archived,omitempty"`
}

// Color is a profile's color. Each value maps to a hand-tuned,
// contrast-checked accent palette in the web apps' tokens.css; there are
// no free-form colors.
type Color string

const (
	ColorGreen  Color = "green"
	ColorBlue   Color = "blue"
	ColorOrange Color = "orange"
	ColorPurple Color = "purple"
)

// AllColors lists every valid Color (for validation and UI options).
var AllColors = []Color{ColorGreen, ColorBlue, ColorOrange, ColorPurple}

// Units is how a user wants measurements displayed. Storage is always
// metric (kg, m, s); the UI converts.
type Units string

const (
	UnitsMetric   Units = "metric"
	UnitsImperial Units = "imperial"
)

// AllUnits lists every valid Units value (for validation and UI options).
var AllUnits = []Units{UnitsMetric, UnitsImperial}

// Validate checks u's own rules:
//   - ID is a valid ID (it becomes a file name, and folder names in tools)
//   - Name is not blank
//   - Color and Units are known values (both required)
//   - Birthday, if set, is a real "YYYY-MM-DD" date, not in the future
func (u User) Validate() error {
	if !jsonfile.ValidID(u.ID) {
		return fmt.Errorf("%w user: bad ID %q", ErrInvalid, u.ID)
	}
	if strings.TrimSpace(u.Name) == "" {
		return fmt.Errorf("%w user %s: missing name", ErrInvalid, u.ID)
	}
	if !slices.Contains(AllColors, u.Color) {
		return fmt.Errorf("%w user %s: unknown color %q", ErrInvalid, u.ID, u.Color)
	}
	if !slices.Contains(AllUnits, u.Units) {
		return fmt.Errorf("%w user %s: unknown units %q", ErrInvalid, u.ID, u.Units)
	}
	if u.Birthday != "" {
		// A plain date string, not time.Time: a birthday has no time of
		// day or time zone, and "2000-01-02" is what a person would write.
		day, err := time.Parse(time.DateOnly, u.Birthday)
		if err != nil {
			return fmt.Errorf("%w user %s: birthday %q is not YYYY-MM-DD", ErrInvalid, u.ID, u.Birthday)
		}
		if day.After(time.Now()) {
			return fmt.Errorf("%w user %s: birthday is in the future", ErrInvalid, u.ID)
		}
	}
	return nil
}
