package discord

import "math"

// The blob's shape and face, ported from the frontend
// (web/packages/ui/src/profiles/blob, decision #22) so the bot looks like
// a household profile: same seed → same blob. Keep the constants in sync.

const (
	blobRadius  = 41   // in the 100×100 viewBox
	blobSpread  = 0.3  // biggest bump vs deepest dip, as a share of the radius
	blobSamples = 24   // points along the outline
	curveSteps  = 12   // line segments per Bézier when flattening for the rasterizer
	ellipseSegs = 48   // line segments per eye outline
	viewBox     = 100. // the shapes' coordinate space
)

var blobWaves = []int{3, 4, 5} // bumps per wave

// point is a position in the 100×100 viewBox (or in pixels, after scaling).
type point struct{ x, y float64 }

// blobOutline returns the blob for seed as a closed polygon: the frontend's
// Catmull-Rom curve through 24 points, flattened into 24 × curveSteps
// straight segments (fine enough that no corner is visible at 512 px).
func blobOutline(seed string) []point {
	rng := newRand(seed)
	type wave struct{ bumps, strength, angle float64 }
	waves := make([]wave, len(blobWaves))
	for i, b := range blobWaves {
		// Same call order as the frontend: strength, then angle, per wave.
		strength := 0.4 + rng.next()*0.6
		angle := rng.next() * 2 * math.Pi
		waves[i] = wave{float64(b), strength, angle}
	}

	// How far each sample sticks out: the sum of all waves at its angle.
	offsets := make([]float64, blobSamples)
	lo, hi := math.Inf(1), math.Inf(-1)
	for i := range offsets {
		t := float64(i) / blobSamples * 2 * math.Pi
		for _, w := range waves {
			offsets[i] += w.strength * math.Cos(w.bumps*t+w.angle)
		}
		lo, hi = min(lo, offsets[i]), max(hi, offsets[i])
	}
	// Every blob gets exactly blobSpread between its highest bump and
	// deepest dip, so none look plain round or broken.
	mid := (lo + hi) / 2
	pts := make([]point, blobSamples)
	for i, o := range offsets {
		t := float64(i) / blobSamples * 2 * math.Pi
		r := blobRadius * (1 + (o-mid)/(hi-lo)*blobSpread)
		pts[i] = point{50 + r*math.Cos(t), 50 + r*math.Sin(t)}
	}
	return smoothPolygon(pts)
}

// smoothPolygon flattens the closed Catmull-Rom curve through pts (as
// cubic Béziers, exactly like the frontend's smoothPath) into a polygon.
func smoothPolygon(pts []point) []point {
	n := len(pts)
	at := func(i int) point { return pts[((i%n)+n)%n] }
	out := make([]point, 0, n*curveSteps)
	for i := range n {
		p0, p1, p2, p3 := at(i-1), at(i), at(i+1), at(i+2)
		c1 := point{p1.x + (p2.x-p0.x)/6, p1.y + (p2.y-p0.y)/6}
		c2 := point{p2.x - (p3.x-p1.x)/6, p2.y - (p3.y-p1.y)/6}
		for s := range curveSteps { // s = 0 includes p1; p2 starts the next segment
			t := float64(s) / curveSteps
			u := 1 - t
			out = append(out, point{
				u*u*u*p1.x + 3*u*u*t*c1.x + 3*u*t*t*c2.x + t*t*t*p2.x,
				u*u*u*p1.y + 3*u*u*t*c1.y + 3*u*t*t*c2.y + t*t*t*p2.y,
			})
		}
	}
	return out
}

// blobFace is where the eyes sit, in the viewBox (the frontend's BlobFace;
// its timing fields are unused here: the GIF has its own script).
type blobFace struct {
	x, y   float64 // center between the eyes
	gap    float64 // center to each eye
	rx, ry float64 // eye radii
	tilt   float64 // degrees
}

// faceFor returns the face for seed, drawn from its own stream
// (seed + ":face") in the frontend's order.
func faceFor(seed string) blobFace {
	rng := newRand(seed + ":face")
	between := func(lo, hi float64) float64 { return lo + rng.next()*(hi-lo) }
	return blobFace{
		x: between(47, 53), y: between(43, 48),
		gap: between(11, 14),
		rx:  between(5.5, 6.5), ry: between(8.5, 10),
		tilt: between(-6, 6),
	}
}

// ellipse returns an ellipse's outline as a polygon around the origin.
func ellipse(rx, ry float64) []point {
	out := make([]point, ellipseSegs)
	for i := range out {
		a := float64(i) / ellipseSegs * 2 * math.Pi
		out[i] = point{rx * math.Cos(a), ry * math.Sin(a)}
	}
	return out
}
