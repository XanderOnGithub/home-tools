package discord

import (
	"math"
	"slices"
)

// A small polygon rasterizer with anti-aliasing, so the bot can draw its
// blob without an image library (the standard library has none).
//
// Coverage is computed per scanline: each pixel row is split into
// subRows horizontal lines; on each line the polygon's edges are crossed
// (nonzero winding rule) and the covered spans are added with exact
// fractional coverage at both ends. Vertical anti-aliasing comes from the
// sub-rows, horizontal from the exact span ends.
//
// Edges sit in an edge table sorted by their top y; an active list holds
// only the edges crossing the current line. For E edges, R lines and k
// active edges per line (k is 2–6 for our shapes), that's
// O(E log E + R·k log k + pixels covered), instead of testing every edge
// against every pixel (O(W·H·E)).

const subRows = 5 // sub-scanlines per pixel row: 6 coverage levels vertically

// mask is a coverage map: 0 = empty, 1 = fully covered, per pixel.
type mask struct {
	w, h int
	a    []float32 // row-major, len w*h
}

func newMask(w, h int) *mask { return &mask{w: w, h: h, a: make([]float32, w*h)} }

// at returns the coverage of pixel (x, y), clamped to [0, 1] (overlapping
// shapes can sum past 1).
func (m *mask) at(x, y int) float32 { return min(m.a[y*m.w+x], 1) }

// edge is one polygon side, oriented top to bottom.
type edge struct {
	y0, y1 float64 // top and bottom; y0 < y1
	x0     float64 // x at y0
	dxdy   float64 // x step per unit of y
	dir    int     // +1 if the polygon went downward here, -1 upward
}

// crossing is where the current scanline meets an edge.
type crossing struct {
	x   float64
	dir int
}

// fill adds the coverage of the closed polygon poly (in pixels) to m.
func (m *mask) fill(poly []point) {
	edges := make([]edge, 0, len(poly))
	for i, a := range poly {
		b := poly[(i+1)%len(poly)]
		if a.y == b.y {
			continue // horizontal: never crosses a scanline
		}
		dir := 1
		if a.y > b.y {
			a, b, dir = b, a, -1
		}
		edges = append(edges, edge{y0: a.y, y1: b.y, x0: a.x, dxdy: (b.x - a.x) / (b.y - a.y), dir: dir})
	}
	slices.SortFunc(edges, func(p, q edge) int { return cmpFloat(p.y0, q.y0) })

	var active []edge
	var xs []crossing
	next := 0 // edges[next:] haven't started yet
	weight := float32(1) / subRows
	top := max(0, int(math.Floor(edges[0].y0)))
	for py := top; py < m.h; py++ {
		row := m.a[py*m.w : (py+1)*m.w]
		for s := range subRows {
			y := float64(py) + (float64(s)+0.5)/subRows // sample at each sub-row's middle
			for next < len(edges) && edges[next].y0 <= y {
				active = append(active, edges[next])
				next++
			}
			active = slices.DeleteFunc(active, func(e edge) bool { return e.y1 <= y })
			if len(active) == 0 {
				if next == len(edges) {
					return // below the whole shape
				}
				continue
			}
			xs = xs[:0]
			for _, e := range active {
				xs = append(xs, crossing{e.x0 + (y-e.y0)*e.dxdy, e.dir})
			}
			slices.SortFunc(xs, func(p, q crossing) int { return cmpFloat(p.x, q.x) })
			// Nonzero rule: inside wherever the winding count isn't 0.
			wind := 0
			for i, c := range xs {
				was := wind
				wind += c.dir
				if was == 0 && wind != 0 {
					start := c.x
					// Find where this inside run ends.
					for j, w := i+1, wind; j < len(xs); j++ {
						w += xs[j].dir
						if w == 0 {
							addSpan(row, start, xs[j].x, weight)
							break
						}
					}
				}
			}
		}
	}
}

// addSpan adds weight × the covered fraction of each pixel in [x0, x1).
func addSpan(row []float32, x0, x1 float64, weight float32) {
	x0, x1 = max(x0, 0), min(x1, float64(len(row)))
	if x1 <= x0 {
		return
	}
	i0, i1 := int(x0), int(x1)
	if i0 == i1 {
		row[i0] += float32(x1-x0) * weight
		return
	}
	row[i0] += float32(float64(i0+1)-x0) * weight
	for i := i0 + 1; i < i1; i++ {
		row[i] += weight
	}
	if i1 < len(row) {
		row[i1] += float32(x1-float64(i1)) * weight
	}
}

func cmpFloat(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}
