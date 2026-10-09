package discord

import (
	"bytes"
	"image/gif"
	"math"
	"testing"
	"time"
)

func TestPersonaFor(t *testing.T) {
	names := []string{"Jim", "Larry", "Jerry", "Gary", "Steve"}
	day := time.Date(2026, 10, 9, 15, 0, 0, 0, time.Local)

	// Same day, any time: same persona (it changes at local midnight only).
	if a, b := PersonaFor(day, names), PersonaFor(time.Date(2026, 10, 9, 23, 59, 0, 0, time.Local), names); a != b {
		t.Errorf("same day differs: %v vs %v", a, b)
	}

	// Over many days: never the same name or color two days running, and
	// every name once per pass of len(names) days.
	prev := PersonaFor(day.AddDate(0, 0, -1), names)
	seen := map[string]int{}
	for i := range 1000 {
		p := PersonaFor(day.AddDate(0, 0, i), names)
		if p.Name == prev.Name || p.Color == prev.Color {
			t.Fatalf("day %d repeats yesterday: %v after %v", i, p, prev)
		}
		seen[p.Name]++
		prev = p
	}
	for _, n := range names {
		if seen[n] != 200 {
			t.Errorf("%s got %d of 1000 days, want 200", n, seen[n])
		}
	}

	for _, tt := range []struct {
		names []string
		want  int // distinct names over a week
	}{{[]string{"Jim"}, 1}, {[]string{"Jim", "Larry"}, 2}, {nil, 1}} {
		got := map[string]bool{}
		for i := range 7 {
			got[PersonaFor(day.AddDate(0, 0, i), tt.names).Name] = true
		}
		if len(got) != tt.want {
			t.Errorf("names %v: %d distinct, want %d", tt.names, len(got), tt.want)
		}
	}

	if got := PersonaFor(day, names).Date; got != "2026-10-09" {
		t.Errorf("date = %q", got)
	}
}

func TestNextMidnight(t *testing.T) {
	got := NextMidnight(time.Date(2026, 12, 31, 18, 30, 0, 0, time.UTC))
	if want := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("NextMidnight = %v, want %v", got, want)
	}
}

// The blob must match the frontend's for the same seed (decision #22):
// values from running web/packages/ui/src/profiles/blob in Node.
func TestBlobMatchesFrontend(t *testing.T) {
	pts := blobOutline("Zoë") // non-ASCII: hashed as UTF-16 like charCodeAt
	if p := pts[0]; math.Abs(p.x-90.0) > 0.05 || math.Abs(p.y-50.0) > 0.05 {
		t.Errorf("first point = %v, want (90.0, 50.0)", p)
	}
	f := faceFor("Ted")
	want := blobFace{x: 52.92660987190902, y: 44.42787266871892, gap: 11.560543533880264, rx: 5.644645321415737, ry: 9.984546254388988, tilt: -0.3344896836206317}
	if math.Abs(f.x-want.x) > 1e-9 || math.Abs(f.ry-want.ry) > 1e-9 || math.Abs(f.tilt-want.tilt) > 1e-9 {
		t.Errorf("face = %+v, want %+v", f, want)
	}
}

func TestMaskFill(t *testing.T) {
	// A 4×4 square from (1.5, 1.5): fully covered inside, half on the
	// edges, a quarter on the corners; total area exactly 16.
	m := newMask(8, 8)
	m.fill([]point{{1.5, 1.5}, {5.5, 1.5}, {5.5, 5.5}, {1.5, 5.5}})
	var sum float64
	for _, a := range m.a {
		sum += float64(a)
	}
	if math.Abs(sum-16) > 1e-4 {
		t.Errorf("area = %v, want 16", sum)
	}
	if m.at(3, 3) != 1 || math.Abs(float64(m.at(1, 3))-0.5) > 0.11 || m.at(0, 0) != 0 {
		t.Errorf("inside %v, edge %v, outside %v", m.at(3, 3), m.at(1, 3), m.at(0, 0))
	}
	// A circle's area converges to πr² (sampling error is small).
	c := newMask(64, 64)
	poly := ellipse(20, 20)
	for i := range poly {
		poly[i] = point{poly[i].x + 32, poly[i].y + 32}
	}
	c.fill(poly)
	sum = 0
	for _, a := range c.a {
		sum += float64(a)
	}
	// A 48-gon inscribed in r = 20 has area ½·48·r²·sin(2π/48).
	if want := 0.5 * 48 * 400 * math.Sin(2*math.Pi/48); math.Abs(sum-want) > 1 {
		t.Errorf("circle area = %v, want ≈ %v", sum, want)
	}
}

func TestAvatars(t *testing.T) {
	p := Persona{Name: "Jim", Color: ColorBlue}
	data, err := AvatarGIF(p)
	if err != nil {
		t.Fatal(err)
	}
	g, err := gif.DecodeAll(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, d := range g.Delay {
		total += d
		if d < 2 {
			t.Errorf("delay %d cs: browsers slow frames under 2 cs to 10 cs", d)
		}
	}
	if len(g.Image) < 10 || g.LoopCount != 0 || total < 500 || total > 1000 {
		t.Errorf("%d frames, loop %d, %d cs total", len(g.Image), g.LoopCount, total)
	}
	if len(data) > 1<<20 {
		t.Errorf("GIF is %d bytes; keep it small for upload", len(data))
	}
	if _, err := AvatarPNG(p); err != nil {
		t.Fatal(err)
	}
}
