package discord

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/url"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Bot is the Discord side: one gateway connection (outbound websocket; no
// port is opened), the slash commands, the status boards and the daily
// persona. Without a token it stays offline and the settings page still
// works.
type Bot struct {
	store *Store
	games Games
	log   *slog.Logger
	token string
	now   func() time.Time // the clock; tests replace it

	sess *discordgo.Session // set once in Run, before any handler runs

	mu   sync.RWMutex
	conn connState

	servers   serverCache
	cooldowns cooldowns
	avatars   avatarCache
	applied   map[string]Persona // guild ID → persona last set there (under mu)

	boardsWake  chan struct{} // refresh the status boards now
	personaWake chan struct{} // re-check the persona in every guild
}

// connState is what the settings page shows about the connection.
type connState struct {
	connected bool
	err       string // why it isn't connected, for people
	user      string // the bot's Discord username
	appID     string // for the invite link
}

// NewBot returns a bot for token ("" = stay offline).
func NewBot(token string, store *Store, games Games, log *slog.Logger) *Bot {
	b := &Bot{
		store: store, games: games, log: log, token: token, now: time.Now,
		applied:     make(map[string]Persona),
		boardsWake:  make(chan struct{}, 1),
		personaWake: make(chan struct{}, 1),
	}
	if token == "" {
		b.conn.err = "No bot token: set DISCORD_TOKEN for Home Tools (see deploy/README.md)."
	}
	return b
}

// wake sends a non-blocking signal: one pending wake-up is enough.
func wake(c chan struct{}) {
	select {
	case c <- struct{}{}:
	default:
	}
}

// Changed tells the bot its config changed (the settings page saved), so
// boards and persona catch up now instead of on their next tick.
func (b *Bot) Changed() {
	wake(b.boardsWake)
	wake(b.personaWake)
}

// Run connects and serves until ctx ends, then disconnects. It only
// returns early when there's no token.
func (b *Bot) Run(ctx context.Context) {
	if b.token == "" {
		return
	}
	s, err := discordgo.New("Bot " + b.token)
	if err != nil {
		b.setErr("couldn't set up the Discord client")
		b.log.Error("discord client", "err", err)
		return
	}
	// Guilds only: channel lists and our own member. No privileged intents
	// (members, message content), so nothing to approve in the portal.
	s.Identify.Intents = discordgo.IntentsGuilds
	s.AddHandler(b.onReady)
	s.AddHandler(func(_ *discordgo.Session, _ *discordgo.GuildCreate) { wake(b.personaWake) })
	s.AddHandler(func(_ *discordgo.Session, _ *discordgo.Disconnect) {
		b.mu.Lock()
		b.conn.connected, b.conn.err = false, "Lost the connection to Discord; reconnecting…"
		b.mu.Unlock()
	})
	s.AddHandler(b.onInteraction)
	b.sess = s

	// The first connect can fail (network down at boot): retry with a
	// growing pause. After that, discordgo reconnects by itself.
	for pause := 5 * time.Second; ; pause = min(pause*2, 5*time.Minute) {
		err := s.Open()
		if err == nil {
			break
		}
		b.log.Warn("discord connect", "err", err, "retry_in", pause)
		b.setErr(connectError(err))
		select {
		case <-ctx.Done():
			return
		case <-time.After(pause):
		}
	}
	defer s.Close()

	var wg sync.WaitGroup
	wg.Go(func() { b.runBoards(ctx) })
	wg.Go(func() { b.runPersona(ctx) })
	<-ctx.Done()
	wg.Wait()
}

// connectError words a failed connect for the settings page.
func connectError(err error) string {
	var rest *discordgo.RESTError
	if errors.As(err, &rest) && rest.Response != nil && rest.Response.StatusCode == 401 {
		return "Discord refused the bot token. Check DISCORD_TOKEN."
	}
	return "Couldn't connect to Discord; retrying."
}

func (b *Bot) setErr(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.conn.connected, b.conn.err = false, msg
}

// onReady runs on every (re)connect: note who we are, and register the
// slash commands. Bulk overwrite is idempotent: the set always matches the
// code, and removed commands disappear.
func (b *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	b.mu.Lock()
	b.conn = connState{connected: true, user: r.User.Username}
	if r.Application != nil {
		b.conn.appID = r.Application.ID
	}
	appID := b.conn.appID
	b.mu.Unlock()
	b.log.Info("discord connected", "user", r.User.Username, "guilds", len(r.Guilds))
	if _, err := s.ApplicationCommandBulkOverwrite(appID, "", commands()); err != nil {
		b.log.Error("discord register commands", "err", err)
	}
	wake(b.boardsWake)
	wake(b.personaWake)
}

// runPersona keeps the daily persona on every server: now, at each local
// midnight, and whenever woken (joined a server, names changed).
func (b *Bot) runPersona(ctx context.Context) {
	for {
		b.applyPersona(ctx)
		// A few seconds past midnight, so the new day's date is certain.
		timer := time.NewTimer(time.Until(NextMidnight(b.now())) + 5*time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		case <-b.personaWake:
			timer.Stop()
		}
	}
}

// Persona returns today's persona.
func (b *Bot) Persona() Persona {
	return PersonaFor(b.now(), b.store.Config().Names)
}

// applyPersona sets today's nickname and avatar in every guild that
// doesn't have them yet. Per guild (Modify Current Member), not the global
// profile: no global rename limit, nothing to undo elsewhere.
func (b *Bot) applyPersona(ctx context.Context) {
	p := b.Persona()
	for _, id := range b.guildIDs() {
		if ctx.Err() != nil {
			return
		}
		b.mu.RLock()
		done := b.applied[id] == p
		b.mu.RUnlock()
		if done {
			continue
		}
		if err := b.setPersona(id, p); err != nil {
			b.log.Warn("discord persona", "guild", id, "err", err)
			continue
		}
		b.mu.Lock()
		b.applied[id] = p
		b.mu.Unlock()
		b.log.Info("discord persona", "guild", id, "name", p.Name, "color", p.Color)
	}
}

// guildIDs lists the guilds the bot is in (copied: the cache changes
// under us as Discord sends events).
func (b *Bot) guildIDs() []string {
	st := b.sess.State
	st.RLock()
	defer st.RUnlock()
	ids := make([]string, len(st.Guilds))
	for i, g := range st.Guilds {
		ids[i] = g.ID
	}
	return ids
}

// setPersona sets nick + avatar in one guild: the animated GIF, else the
// still PNG (if animated avatars are refused), else just the nickname.
func (b *Bot) setPersona(guildID string, p Persona) error {
	gifData, pngData, err := b.avatars.get(p)
	if err != nil {
		return err
	}
	endpoint := discordgo.EndpointGuildMember(guildID, "@me")
	bucket := discordgo.EndpointGuildMember(guildID, "")
	for _, avatar := range []string{
		"data:image/gif;base64," + base64.StdEncoding.EncodeToString(gifData),
		"data:image/png;base64," + base64.StdEncoding.EncodeToString(pngData),
	} {
		body := map[string]string{"nick": p.Name, "avatar": avatar}
		if _, err = b.sess.RequestWithBucketID("PATCH", endpoint, body, bucket); err == nil {
			return nil
		}
		b.log.Debug("discord avatar refused", "guild", guildID, "err", err)
	}
	return b.sess.GuildMemberNickname(guildID, "@me", p.Name)
}

// avatarCache renders each persona's images once (a day's worth of use).
type avatarCache struct {
	mu       sync.Mutex
	persona  Persona
	gif, png []byte
}

// get returns p's animated and still avatars.
func (c *avatarCache) get(p Persona) (gifData, pngData []byte, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.persona == p && c.gif != nil {
		return c.gif, c.png, nil
	}
	if gifData, err = AvatarGIF(p); err != nil {
		return nil, nil, err
	}
	if pngData, err = AvatarPNG(p); err != nil {
		return nil, nil, err
	}
	c.persona, c.gif, c.png = p, gifData, pngData
	return gifData, pngData, nil
}

// serverCache keeps the games list briefly: autocomplete asks on every
// keystroke and must answer within Discord's 3 s.
type serverCache struct {
	mu   sync.Mutex
	list []GameServer
	at   time.Time
}

// gameServers returns the server list, fetched if older than maxAge. The
// lock is held during the fetch, so concurrent callers share one request.
func (b *Bot) gameServers(ctx context.Context, maxAge time.Duration) ([]GameServer, error) {
	c := &b.servers
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.list != nil && b.now().Sub(c.at) < maxAge {
		return c.list, nil
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	list, err := b.games.Servers(ctx)
	if err != nil {
		return nil, err
	}
	list = slices.DeleteFunc(list, func(s GameServer) bool { return s.Archived })
	c.list, c.at = list, b.now()
	return list, nil
}

// cooldowns spaces out restarts of one server, whoever asks: a restart
// that just finished doesn't need another, and two people tapping at once
// shouldn't queue two.
type cooldowns struct {
	mu   sync.Mutex
	last map[string]time.Time // server ID → when its last restart began
}

const restartCooldown = 2 * time.Minute

// begin claims a restart of id at now, or returns how long to wait.
func (c *cooldowns) begin(id string, now time.Time) (wait time.Duration, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if last, seen := c.last[id]; seen && now.Sub(last) < restartCooldown {
		return restartCooldown - now.Sub(last), false
	}
	if c.last == nil {
		c.last = make(map[string]time.Time)
	}
	c.last[id] = now
	return 0, true
}

// BotView is the connection and persona, for the settings page.
type BotView struct {
	Connected bool        `json:"connected"`
	Error     string      `json:"error,omitempty"`
	User      string      `json:"user,omitempty"`
	InviteURL string      `json:"invite_url,omitempty"`
	Guilds    []GuildView `json:"guilds"`
	Persona   Persona     `json:"persona"`
}

// GuildView is one Discord server the bot is in, with the text channels a
// status board can go in.
type GuildView struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Channels []ChannelView `json:"channels"`
}

type ChannelView struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	CanPost bool   `json:"can_post"` // the bot may post embeds there
}

// invitePermissions: view channels, send messages, embed links, read
// message history, change nickname. Nothing else (no admin).
const invitePermissions = 1024 | 2048 | 16384 | 65536 | 67108864

const postPermissions = discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionEmbedLinks

// View returns the bot's state for the settings page, from discordgo's
// in-memory cache (no Discord calls).
func (b *Bot) View() BotView {
	b.mu.RLock()
	c := b.conn
	b.mu.RUnlock()
	v := BotView{Connected: c.connected, Error: c.err, User: c.user, Guilds: []GuildView{}, Persona: b.Persona()}
	if c.appID != "" {
		v.InviteURL = "https://discord.com/oauth2/authorize?" + url.Values{
			"client_id":   {c.appID},
			"scope":       {"bot applications.commands"},
			"permissions": {strconv.Itoa(invitePermissions)},
		}.Encode()
	}
	if b.sess == nil || !c.connected {
		return v
	}
	// Copy what's needed under the cache's lock, then work on the copy:
	// the permission lookups below take that lock themselves.
	st := b.sess.State
	st.RLock()
	for _, g := range st.Guilds {
		gv := GuildView{ID: g.ID, Name: g.Name, Channels: []ChannelView{}}
		chans := slices.Clone(g.Channels)
		slices.SortFunc(chans, func(a, b *discordgo.Channel) int { return a.Position - b.Position })
		for _, ch := range chans {
			if ch.Type == discordgo.ChannelTypeGuildText || ch.Type == discordgo.ChannelTypeGuildNews {
				gv.Channels = append(gv.Channels, ChannelView{ID: ch.ID, Name: ch.Name})
			}
		}
		v.Guilds = append(v.Guilds, gv)
	}
	st.RUnlock()
	for _, g := range v.Guilds {
		for i := range g.Channels {
			g.Channels[i].CanPost = b.canPost(g.Channels[i].ID)
		}
	}
	slices.SortFunc(v.Guilds, func(a, b GuildView) int { return compareFold(a.Name, b.Name) })
	return v
}

// canPost reports whether the bot may post a status board in channelID.
// Unknown (permissions not cached yet) counts as yes: the board will say
// if it fails.
func (b *Bot) canPost(channelID string) bool {
	perms, err := b.sess.State.UserChannelPermissions(b.sess.State.User.ID, channelID)
	return err != nil || perms&postPermissions == postPermissions
}
