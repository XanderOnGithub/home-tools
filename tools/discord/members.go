package discord

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// People in the bot's Discord servers, for the People page (decision #61):
// search by name to verify someone without looking up their user ID, and
// show everyone with their picture.
//
// Discord's Search Guild Members works without the privileged Server
// Members intent (#42 keeps the bot on the guilds intent only); listing
// every member would need it. Avatars are fetched here and served from
// /api, like the Minecraft heads (#40): the browser never calls Discord,
// and the page's CSP stays 'self'.

// Member is one person, as the People page shows them.
type Member struct {
	UserID   string `json:"user_id"`
	Name     string `json:"name"`     // server nickname, else display name, else username
	Username string `json:"username"` // the unique @handle
}

const (
	maxSearchLen  = 32 // Discord's username limit; longer can't match
	searchResults = 10
)

// SearchMembers finds people whose username or nickname starts with
// query, in every server the bot is in. Bots are left out. At most
// searchResults, sorted by name. O(g) requests for g servers (a
// household's handful).
func (b *Bot) SearchMembers(query string) ([]Member, error) {
	query = strings.TrimSpace(query)
	if query == "" || len([]rune(query)) > maxSearchLen {
		return nil, fmt.Errorf("%w: search for 1–%d characters", ErrInvalid, maxSearchLen)
	}
	b.mu.RLock()
	connected := b.conn.connected
	b.mu.RUnlock()
	if !connected || b.sess == nil {
		return nil, errOffline
	}
	var found [][]*discordgo.Member
	for _, g := range b.guildIDs() {
		list, err := b.sess.GuildMembersSearch(g, query, searchResults)
		if err != nil {
			return nil, fmt.Errorf("search %s: %w", g, err)
		}
		found = append(found, list)
	}
	return mergeMembers(found, searchResults), nil
}

// mergeMembers turns each server's results into one list: people (not
// bots) once each, by name, at most limit. Someone in two servers keeps
// the first server's nickname. Pure, for tests.
func mergeMembers(lists [][]*discordgo.Member, limit int) []Member {
	seen := map[string]bool{}
	var out []Member
	for _, list := range lists {
		for _, m := range list {
			if m == nil || m.User == nil || m.User.Bot || seen[m.User.ID] {
				continue
			}
			seen[m.User.ID] = true
			out = append(out, Member{UserID: m.User.ID, Name: m.DisplayName(), Username: m.User.Username})
		}
	}
	slices.SortFunc(out, func(a, b Member) int {
		return cmp.Or(compareFold(a.Name, b.Name), strings.Compare(a.UserID, b.UserID))
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

// faces caches people's avatars: a day once fetched, an hour after a
// failure (so a missing one isn't asked for on every page load). Bounded:
// past maxFaces entries the cache starts over, which for a household is
// never.
type faces struct {
	mu   sync.Mutex
	by   map[string]face // user ID →
	urlF func(ctx context.Context, userID string) (string, error)
	get  func(ctx context.Context, url string) ([]byte, string, error)
	now  func() time.Time
}

type face struct {
	img  []byte // nil = couldn't get one
	mime string
	at   time.Time
}

const (
	maxFaces    = 500
	faceTTL     = 24 * time.Hour
	faceFailTTL = time.Hour
	faceMaxSize = 1 << 20 // a 64 px avatar is a few KB; anything bigger is wrong
)

var errNoFace = errors.New("no avatar")

// avatar returns userID's picture and its type, from the cache when it's
// fresh.
func (f *faces) avatar(ctx context.Context, userID string) ([]byte, string, error) {
	f.mu.Lock()
	c, ok := f.by[userID]
	f.mu.Unlock()
	if ok {
		ttl := faceTTL
		if c.img == nil {
			ttl = faceFailTTL
		}
		if f.now().Sub(c.at) < ttl {
			if c.img == nil {
				return nil, "", errNoFace
			}
			return c.img, c.mime, nil
		}
	}
	img, mime, err := f.fetch(ctx, userID)
	f.mu.Lock()
	if f.by == nil || len(f.by) >= maxFaces {
		f.by = make(map[string]face)
	}
	f.by[userID] = face{img: img, mime: mime, at: f.now()} // a failure is cached too (img nil)
	f.mu.Unlock()
	if err != nil {
		return nil, "", errors.Join(errNoFace, err)
	}
	return img, mime, nil
}

func (f *faces) fetch(ctx context.Context, userID string) ([]byte, string, error) {
	url, err := f.urlF(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	if !strings.HasPrefix(url, "https://cdn.discordapp.com/") {
		return nil, "", fmt.Errorf("avatar url %q isn't Discord's", url)
	}
	return f.get(ctx, url)
}

// avatarURL looks userID up (Get User: no intent needed) and returns their
// 64 px avatar, or Discord's default one for people without their own.
func (b *Bot) avatarURL(_ context.Context, userID string) (string, error) {
	b.mu.RLock()
	connected := b.conn.connected
	b.mu.RUnlock()
	if !connected || b.sess == nil {
		return "", errOffline
	}
	u, err := b.sess.User(userID)
	if err != nil {
		return "", fmt.Errorf("user %s: %w", userID, err)
	}
	return u.AvatarURL("64"), nil
}

var faceClient = &http.Client{Timeout: 10 * time.Second}

// download gets an image over HTTPS, at most faceMaxSize.
func download(ctx context.Context, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	res, err := faceClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()
	mime := res.Header.Get("Content-Type")
	if res.StatusCode != http.StatusOK || !strings.HasPrefix(mime, "image/") {
		return nil, "", fmt.Errorf("avatar: %s, %q", res.Status, mime)
	}
	img, err := io.ReadAll(io.LimitReader(res.Body, faceMaxSize+1))
	if err != nil {
		return nil, "", err
	}
	if len(img) > faceMaxSize {
		return nil, "", errors.New("avatar: too big")
	}
	return img, mime, nil
}
