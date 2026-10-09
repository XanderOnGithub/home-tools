package discord

import (
	"strconv"
	"time"
	"unicode/utf16"
)

// The bot's persona (decision #42): a blob with a new name and color every
// day. It's a pure function of the date and the name list, so nothing is
// stored, every restart agrees on today's persona, and tests can ask for
// any day.

// Color is a profile color (decision #22): the bot's blob uses the same
// palette as the household's avatars.
type Color string

const (
	ColorGreen  Color = "green"
	ColorBlue   Color = "blue"
	ColorOrange Color = "orange"
	ColorPurple Color = "purple"
)

// AllColors is the order the persona cycles through, one per day.
var AllColors = []Color{ColorGreen, ColorBlue, ColorOrange, ColorPurple}

// colorHex is each color's blob shade, the same as --color-avatar in
// web/packages/ui/src/tokens.css (keep them in sync), as 0xRRGGBB.
var colorHex = map[Color]uint32{
	ColorGreen:  0x4ade80,
	ColorBlue:   0x60a5fa,
	ColorOrange: 0xfb923c,
	ColorPurple: 0xc084fc,
}

// eyeHex is --color-avatar-eye: the blob's eyes, the same on every color.
const eyeHex = 0x1c1917

// Persona is who the bot is on one day.
type Persona struct {
	Name  string `json:"name"`
	Color Color  `json:"color"`
	Date  string `json:"date"` // YYYY-MM-DD, local time
}

// dayNumber counts days since 1970-01-01 for t's calendar date in t's own
// time zone, so the persona changes at local midnight, not UTC's.
func dayNumber(t time.Time) int64 {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Unix() / 86400
}

// PersonaFor returns the persona for t's date: colors simply rotate
// (never the same two days running); names take turns (see rotate).
func PersonaFor(t time.Time, names []string) Persona {
	day := dayNumber(t)
	p := Persona{Color: AllColors[mod(day, int64(len(AllColors)))], Date: t.Format(time.DateOnly)}
	if len(names) > 0 {
		p.Name = names[rotate(day, len(names), "names")]
	}
	return p
}

// rotate returns which of n items has turn k (k = 0, 1, 2, …): the items
// go in a shuffled order, a new shuffle each pass of n turns, so every
// item gets a turn before any repeats, the order isn't predictable, and
// no item has two turns in a row. Each pass is a Fisher–Yates shuffle
// seeded by salt + the pass number: O(n) time and space per call, no
// state to store (n ≤ 500, so microseconds).
func rotate(k int64, n int, salt string) int {
	switch n {
	case 1:
		return 0
	case 2:
		return int(mod(k, 2)) // alternate: a shuffle could repeat one
	}
	pass, i := floorDiv(k, int64(n)), mod(k, int64(n))
	order := shuffled(int64(n), salt, pass)
	// Passes are independent shuffles, so the last item of one pass could
	// open the next. Swapping the first two of the new pass fixes that; it
	// never moves the pass's last item (n ≥ 3), so no new clash appears.
	if prev := shuffled(int64(n), salt, pass-1); order[0] == prev[n-1] {
		order[0], order[1] = order[1], order[0]
	}
	return order[i]
}

func floorDiv(a, b int64) int64 {
	q := a / b
	if a%b != 0 && a < 0 {
		q-- // Go truncates toward zero; dates before 1970 need floor
	}
	return q
}

// shuffled returns 0..n-1 in an order fixed by salt and seed (Fisher–Yates).
func shuffled(n int64, salt string, seed int64) []int {
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	rng := newRand(salt + ":" + strconv.FormatInt(seed, 10))
	for i := n - 1; i > 0; i-- {
		j := int(rng.next() * float64(i+1))
		order[i], order[j] = order[j], order[i]
	}
	return order
}

func mod(a, b int64) int64 { return ((a % b) + b) % b }

// NextMidnight returns the first local midnight after t: when the persona
// changes next.
func NextMidnight(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d+1, 0, 0, 0, 0, t.Location())
}

// rand is the same generator the frontend's blob uses (seededRandom in
// web/packages/ui/src/profiles/blob): an FNV-1a hash of the seed, stepped
// by a linear congruential generator. Same seed → same numbers in Go and
// TypeScript, so a blob drawn here matches the one the UI would draw.
type rand struct{ h uint32 }

// newRand seeds a generator. Like the frontend, it hashes UTF-16 code
// units (JavaScript's charCodeAt), so a name like "Zoë" seeds the same.
func newRand(seed string) *rand {
	h := uint32(2166136261) // FNV-1a offset basis
	for _, c := range utf16.Encode([]rune(seed)) {
		h ^= uint32(c)
		h *= 16777619 // FNV prime
	}
	return &rand{h: h}
}

// next returns a number in [0, 1).
func (r *rand) next() float64 {
	r.h = r.h*1664525 + 1013904223 // uint32 wraps, like Math.imul + >>> 0
	return float64(r.h) / (1 << 32)
}
