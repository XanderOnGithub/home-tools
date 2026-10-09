package discord

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"math"
)

// The persona's avatar: its blob, animated as a looping GIF. The script
// (decision #42): rest, blink twice, look left, look right, back, loop.
// The eyes move as a pair and the body leans a little toward where they
// look, like the profile avatars on the web (ProfileAvatar.svelte).

const (
	avatarPx  = 256 // Discord shows avatars at ≤ 128 px; 2× stays sharp
	tweenMs   = 40  // one in-between frame per 40 ms (GIF delays are in 10 ms)
	eyeLevels = 16  // anti-aliasing shades between body and eye color
	// avatarZoom shrinks the blob inside the image: Discord crops avatars
	// to a circle, and at full size the bumps (up to ~47 of 50 units from
	// the center, more while leaning) touch or cross its edge. At 0.75 the
	// blob spans at most ~72% of the circle's width, with a clear margin.
	avatarZoom = 0.75
)

// pose is one moment of the face: where the eyes look, how open they
// are, and how the body leans.
type pose struct {
	lookX, lookY float64 // eye offset, viewBox units
	open         float64 // eye height: 1 = open, 0.1 = shut
	lean         float64 // body rotation, degrees
	shiftX       float64 // body nudge, viewBox units
}

var (
	rest      = pose{open: 1}
	shut      = pose{open: 0.1}
	lookLeft  = pose{lookX: -2.6, lookY: 0.4, open: 1, lean: -2, shiftX: -0.6}
	lookRight = pose{lookX: 2.6, lookY: 0.4, open: 1, lean: 2, shiftX: 0.6}
)

// step moves to a pose over tween ms (eased), then holds it for hold ms.
type step struct {
	to          pose
	tween, hold int
}

// script is the loop. Holds dominate: it should read as calm, not busy.
var script = []step{
	{rest, 0, 1600},
	{shut, 80, 40}, {rest, 80, 180}, // blink
	{shut, 80, 40}, {rest, 80, 900}, // blink
	{lookLeft, 240, 1000},
	{lookRight, 360, 1000},
	{rest, 240, 600}, // then the loop's first hold
}

// frame is one rendered pose and how long it shows.
type frame struct {
	p      pose
	millis int
}

// frames expands the script into poses: one per tweenMs while moving
// (cubic ease in-out), one long frame per hold.
func frames() []frame {
	var out []frame
	cur := script[0].to
	for _, s := range script {
		n := s.tween / tweenMs
		for i := 1; i <= n; i++ {
			out = append(out, frame{mix(cur, s.to, easeInOut(float64(i)/float64(n+1))), tweenMs})
		}
		out = append(out, frame{s.to, max(s.hold, tweenMs)})
		cur = s.to
	}
	return out
}

func easeInOut(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

func mix(a, b pose, t float64) pose {
	l := func(x, y float64) float64 { return x + (y-x)*t }
	return pose{l(a.lookX, b.lookX), l(a.lookY, b.lookY), l(a.open, b.open), l(a.lean, b.lean), l(a.shiftX, b.shiftX)}
}

// figure is a persona's blob, ready to draw in any pose.
type figure struct {
	body   []point
	face   blobFace
	center point // the body's bounding-box center: the lean pivots here (CSS transform-box: fill-box)
	eye    []point
}

func newFigure(seed string) figure {
	f := figure{body: blobOutline(seed), face: faceFor(seed)}
	lo, hi := point{math.Inf(1), math.Inf(1)}, point{math.Inf(-1), math.Inf(-1)}
	for _, p := range f.body {
		lo = point{min(lo.x, p.x), min(lo.y, p.y)}
		hi = point{max(hi.x, p.x), max(hi.y, p.y)}
	}
	f.center = point{(lo.x + hi.x) / 2, (lo.y + hi.y) / 2}
	f.eye = ellipse(f.face.rx, f.face.ry)
	return f
}

// draw renders the figure in pose p as body and eye coverage masks.
// Transforms match the SVG: the body leans around its center, the eye
// pair sits at the face point, tilted, then looks; each eye blinks around
// its own center.
func (f figure) draw(p pose, px int) (body, eyes *mask) {
	scale := float64(px) / viewBox * avatarZoom
	offset := float64(px) * (1 - avatarZoom) / 2 // centers the shrunken viewBox
	sin, cos := math.Sincos(p.lean * math.Pi / 180)
	toPx := func(q point) point { // body transform, then viewBox → pixels
		dx, dy := q.x-f.center.x, q.y-f.center.y
		return point{(f.center.x+dx*cos-dy*sin+p.shiftX)*scale + offset, (f.center.y+dx*sin+dy*cos)*scale + offset}
	}
	body = newMask(px, px)
	poly := make([]point, len(f.body))
	for i, q := range f.body {
		poly[i] = toPx(q)
	}
	body.fill(poly)

	eyes = newMask(px, px)
	ts, tc := math.Sincos(f.face.tilt * math.Pi / 180)
	for _, side := range []float64{-1, 1} {
		poly := make([]point, len(f.eye))
		for i, q := range f.eye {
			x, y := q.x+side*f.face.gap+p.lookX, q.y*p.open+p.lookY
			poly[i] = toPx(point{f.face.x + x*tc - y*ts, f.face.y + x*ts + y*tc})
		}
		eyes.fill(poly)
	}
	return body, eyes
}

// palette: transparent, then the body color blending into the eye color
// in eyeLevels steps (anti-aliased eye edges).
func avatarPalette(c Color) color.Palette {
	pal := color.Palette{color.RGBA{}}
	skin, eye := rgb(colorHex[c]), rgb(eyeHex)
	for i := range eyeLevels {
		t := float64(i) / (eyeLevels - 1)
		pal = append(pal, color.RGBA{lerp8(skin.R, eye.R, t), lerp8(skin.G, eye.G, t), lerp8(skin.B, eye.B, t), 255})
	}
	return pal
}

// AvatarGIF renders p's animated avatar (see BlobGIF).
func AvatarGIF(p Persona) ([]byte, error) { return BlobGIF(p.Name, p.Color, avatarPx) }

// AvatarPNG renders p's avatar at rest (see BlobPNG): the fallback where
// an animated avatar isn't allowed.
func AvatarPNG(p Persona) ([]byte, error) { return BlobPNG(p.Name, p.Color, avatarPx) }

// BlobGIF renders the blob for name, px × px, animated. GIF transparency
// is on/off, so the body's outline is cut at half coverage (Discord
// shrinks the image, which smooths it); the eyes, drawn on the body, are
// fully anti-aliased.
func BlobGIF(name string, c Color, px int) ([]byte, error) {
	f := newFigure(name)
	pal := avatarPalette(c)
	anim := &gif.GIF{LoopCount: 0} // 0 = forever
	for _, fr := range frames() {
		body, eyes := f.draw(fr.p, px)
		img := image.NewPaletted(image.Rect(0, 0, px, px), pal)
		for y := range px {
			for x := range px {
				if body.at(x, y) >= 0.5 {
					img.Pix[y*img.Stride+x] = 1 + uint8(math.Round(float64(eyes.at(x, y))*(eyeLevels-1)))
				}
			}
		}
		anim.Image = append(anim.Image, img)
		anim.Delay = append(anim.Delay, fr.millis/10)
		// Clear to transparent between frames, or a leaning body would leave
		// its old outline behind.
		anim.Disposal = append(anim.Disposal, gif.DisposalBackground)
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, anim); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// BlobPNG renders the blob for name at rest, px × px, on a transparent
// background, fully anti-aliased (PNG has real alpha).
func BlobPNG(name string, c Color, px int) ([]byte, error) {
	body, eyes := newFigure(name).draw(rest, px)
	skin, eye := rgb(colorHex[c]), rgb(eyeHex)
	img := image.NewNRGBA(image.Rect(0, 0, px, px))
	for y := range px {
		for x := range px {
			a := body.at(x, y)
			if a == 0 {
				continue
			}
			t := float64(eyes.at(x, y))
			img.SetNRGBA(x, y, color.NRGBA{lerp8(skin.R, eye.R, t), lerp8(skin.G, eye.G, t), lerp8(skin.B, eye.B, t), uint8(math.Round(float64(a) * 255))})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func rgb(hex uint32) color.RGBA {
	return color.RGBA{uint8(hex >> 16), uint8(hex >> 8), uint8(hex), 255}
}

func lerp8(a, b uint8, t float64) uint8 {
	return uint8(math.Round(float64(a) + (float64(b)-float64(a))*t))
}
