package games

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Minecraft player heads (decision #40): the 8×8 face (plus the hat
// layer) cropped from the player's skin, which comes from Mojang by the
// UUID the server reports (`list uuids`). Fetched by this server and
// cached, so browsers only ever talk to us. Offline-mode UUIDs aren't
// Mojang accounts: they get no head, and the UI shows its plain icon.

const (
	headFresh   = 24 * time.Hour  // skins rarely change
	headRetry   = 1 * time.Hour   // after a failed lookup, don't ask Mojang again for a while
	headTimeout = 5 * time.Second // per lookup (profile + skin)
	maxSkin     = 1 << 20         // skins are a few KB; refuse anything silly
)

type heads struct {
	client  *http.Client
	profile string // Mojang session server profile URL, ending in "/"; tests swap it

	mu    sync.Mutex
	uuids map[string]string // player name → UUID, from `list uuids`
	cache map[string]head   // UUID → face
}

type head struct {
	png     []byte // nil = lookup failed
	fetched time.Time
}

func newHeads() *heads {
	return &heads{
		client:  &http.Client{Timeout: headTimeout},
		profile: "https://sessionserver.mojang.com/session/minecraft/profile/",
	}
}

// remember records name → uuid, as the server just reported it.
func (h *heads) remember(name, uuid string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.uuids == nil {
		h.uuids = make(map[string]string)
	}
	h.uuids[name] = uuid
}

var errNoHead = errors.New("no head for this player")

// face returns the PNG of name's face, from the cache when it's fresh.
func (h *heads) face(ctx context.Context, name string) ([]byte, error) {
	h.mu.Lock()
	uuid, ok := h.uuids[name]
	c, cached := h.cache[uuid]
	h.mu.Unlock()
	if !ok {
		return nil, errNoHead
	}
	if cached {
		age := time.Since(c.fetched)
		switch {
		case c.png != nil && age < headFresh:
			return c.png, nil
		case c.png == nil && age < headRetry:
			return nil, errNoHead
		}
	}

	face, err := h.fetch(ctx, uuid)
	next := head{png: face, fetched: time.Now()} // a failure is cached too, so Mojang isn't asked on every look
	if err != nil && c.png != nil {
		// Keep showing the old face; ask again in headRetry.
		next = head{png: c.png, fetched: time.Now().Add(headRetry - headFresh)}
	}
	h.mu.Lock()
	if h.cache == nil {
		h.cache = make(map[string]head)
	}
	h.cache[uuid] = next
	h.mu.Unlock()
	if next.png == nil {
		return nil, fmt.Errorf("head of %s: %w", name, err)
	}
	return next.png, nil
}

// fetch asks Mojang for the profile, downloads its skin and crops the face.
func (h *heads) fetch(ctx context.Context, uuid string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, headTimeout)
	defer cancel()

	body, err := h.get(ctx, h.profile+strings.ReplaceAll(uuid, "-", ""))
	if err != nil {
		return nil, err
	}
	// {"properties":[{"name":"textures","value":"<base64 JSON>"}]}
	var profile struct {
		Properties []struct{ Name, Value string } `json:"properties"`
	}
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("profile: %w", err)
	}
	var skinURL string
	for _, p := range profile.Properties {
		if p.Name != "textures" {
			continue
		}
		raw, err := base64.StdEncoding.DecodeString(p.Value)
		if err != nil {
			return nil, fmt.Errorf("textures: %w", err)
		}
		var tex struct {
			Textures struct{ SKIN struct{ URL string } } `json:"textures"`
		}
		if err := json.Unmarshal(raw, &tex); err != nil {
			return nil, fmt.Errorf("textures: %w", err)
		}
		skinURL = tex.Textures.SKIN.URL
	}
	// Only ever download from Mojang's texture host, whatever the profile says.
	if u, err := url.Parse(skinURL); err != nil || u.Host != "textures.minecraft.net" {
		return nil, fmt.Errorf("unexpected skin URL %q", skinURL)
	}
	skin, err := h.get(ctx, strings.Replace(skinURL, "http://", "https://", 1))
	if err != nil {
		return nil, err
	}
	return cropFace(skin)
}

func (h *heads) get(ctx context.Context, u string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { // 204 = no such profile (offline mode)
		return nil, fmt.Errorf("GET %s: %s", u, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, maxSkin))
}

// cropFace cuts the face out of a skin: the 8×8 square at (8, 8), with
// the hat layer at (40, 8) drawn over it (its transparent pixels let the
// face show through). The result is 8×8; the UI scales it up pixelated.
func cropFace(skin []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(skin))
	if err != nil {
		return nil, fmt.Errorf("skin: %w", err)
	}
	if b := img.Bounds(); b.Dx() < 64 || b.Dy() < 32 {
		return nil, fmt.Errorf("skin is %v, want at least 64×32", b.Size())
	}
	o := img.Bounds().Min
	face := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	draw.Draw(face, face.Bounds(), img, o.Add(image.Pt(8, 8)), draw.Src)
	draw.Draw(face, face.Bounds(), img, o.Add(image.Pt(40, 8)), draw.Over)
	var out bytes.Buffer
	if err := png.Encode(&out, face); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
