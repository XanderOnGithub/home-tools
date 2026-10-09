package games

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// testSkin is a 64×64 skin: face red, hat layer transparent except its
// top-left pixel (blue), which must end up drawn over the face.
func testSkin(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	for y := 8; y < 16; y++ {
		for x := 8; x < 16; x++ {
			img.Set(x, y, color.NRGBA{255, 0, 0, 255})
		}
	}
	img.Set(40, 8, color.NRGBA{0, 0, 255, 255})
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestCropFace(t *testing.T) {
	out, err := cropFace(testSkin(t))
	if err != nil {
		t.Fatal(err)
	}
	face, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	if face.Bounds().Dx() != 8 || face.Bounds().Dy() != 8 {
		t.Fatalf("face is %v, want 8×8", face.Bounds())
	}
	if r, g, b, _ := face.At(0, 0).RGBA(); r != 0 || g != 0 || b == 0 {
		t.Errorf("hat pixel not drawn over the face: %v", face.At(0, 0))
	}
	if r, _, _, _ := face.At(4, 4).RGBA(); r == 0 {
		t.Errorf("face pixel = %v, want red", face.At(4, 4))
	}
	if _, err := cropFace([]byte("not a png")); err == nil {
		t.Error("cropFace accepted garbage")
	}
}

// roundTrip sends every request to a handler, whatever its host, so the
// real Mojang URLs can be used in tests.
type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func fakeMojang(t *testing.T, skinURL string, calls *atomic.Int32) *http.Client {
	textures := base64.StdEncoding.EncodeToString([]byte(`{"textures":{"SKIN":{"url":"` + skinURL + `"}}}`))
	skin := testSkin(t)
	mux := http.NewServeMux()
	mux.HandleFunc("sessionserver.mojang.com/session/minecraft/profile/{id}", func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.PathValue("id") != "069a79f444e94726a5befca90e38aaf5" {
			w.WriteHeader(http.StatusNoContent) // unknown, like an offline-mode UUID
			return
		}
		w.Write([]byte(`{"properties":[{"name":"textures","value":"` + textures + `"}]}`))
	})
	mux.HandleFunc("textures.minecraft.net/texture/{hash}", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Scheme != "https" {
			t.Errorf("skin fetched over %s", r.URL.Scheme)
		}
		w.Write(skin)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request to %s", r.URL)
		http.NotFound(w, r)
	})
	return &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, r)
		return rec.Result(), nil
	})}
}

func TestHeads(t *testing.T) {
	var calls atomic.Int32
	h := newHeads()
	h.client = fakeMojang(t, "http://textures.minecraft.net/texture/abc", &calls)
	h.remember("Steve", "069a79f4-44e9-4726-a5be-fca90e38aaf5")
	h.remember("Offline", "00000000-0000-3000-8000-000000000000")

	for range 2 { // the second answer comes from the cache
		if face, err := h.face(t.Context(), "Steve"); err != nil || !bytes.HasPrefix(face, []byte("\x89PNG")) {
			t.Fatalf("face(Steve) = %d bytes, %v", len(face), err)
		}
	}
	if _, err := h.face(t.Context(), "Nobody"); err != errNoHead {
		t.Errorf("face(Nobody) err = %v, want errNoHead", err)
	}
	for range 2 { // a failure is cached too
		if _, err := h.face(t.Context(), "Offline"); err == nil {
			t.Error("face(Offline) found a head")
		}
	}
	if n := calls.Load(); n != 2 {
		t.Errorf("Mojang asked %d times, want 2 (once per player)", n)
	}
}

func TestHeadsRefusesOtherSkinHosts(t *testing.T) {
	var calls atomic.Int32
	h := newHeads()
	h.client = fakeMojang(t, "http://evil.example/skin.png", &calls)
	h.remember("Steve", "069a79f4-44e9-4726-a5be-fca90e38aaf5")
	if _, err := h.face(t.Context(), "Steve"); err == nil || !strings.Contains(err.Error(), "unexpected skin URL") {
		t.Errorf("err = %v, want unexpected skin URL", err)
	}
}

func TestGetHead(t *testing.T) {
	docker, _ := fakeDocker(t)
	h := newTestHandlers(t, docker)
	if rec := do(h, "GET", "/api/servers/valheim/heads/Ragnhild"); rec.Code != http.StatusNotFound {
		t.Errorf("valheim head = %d, want 404", rec.Code)
	}
	if rec := do(h, "GET", "/api/servers/minecraft/heads/Steve"); rec.Code != http.StatusNotFound {
		t.Errorf("unlisted player's head = %d, want 404", rec.Code)
	}
}
