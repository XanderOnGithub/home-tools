package discord

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/bwmarrin/discordgo"
)

// Slash commands (decision #42). The logic works on plain values
// (invocation in, reply out, through a responder), so it's tested without
// Discord; discordgo only appears at the edges (definitions, adapters).

// commands are the slash commands as Discord shows them.
func commands() []*discordgo.ApplicationCommand {
	gamesChoices := make([]*discordgo.ApplicationCommandOptionChoice, len(sensGames))
	for i, g := range sensGames {
		gamesChoices[i] = &discordgo.ApplicationCommandOptionChoice{Name: g.Name, Value: g.Slug}
	}
	minSens, minDPI := 0.0001, 1.0
	everywhere := &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild, discordgo.InteractionContextBotDM}
	guildOnly := &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild}
	server := func(desc string) *discordgo.ApplicationCommandOption {
		return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionString, Name: "server", Description: desc, Required: true, Autocomplete: true}
	}
	player := &discordgo.ApplicationCommandOption{
		Type: discordgo.ApplicationCommandOptionString, Name: "player", Required: true, MaxLength: 100,
		Description: "Minecraft: their username. Valheim: their SteamID64 or Steam profile link.",
	}
	return []*discordgo.ApplicationCommand{
		{
			Name: "sens", Description: "Convert your mouse sensitivity from one game to another", Contexts: everywhere,
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionString, Name: "from_game", Description: "The game you're used to", Required: true, Choices: gamesChoices},
				{Type: discordgo.ApplicationCommandOptionNumber, Name: "value", Description: "Your sensitivity there", Required: true, MinValue: &minSens},
				{Type: discordgo.ApplicationCommandOptionString, Name: "to_game", Description: "The game to convert to", Required: true, Choices: gamesChoices},
				{Type: discordgo.ApplicationCommandOptionInteger, Name: "from_dpi", Description: "Your mouse DPI now (optional)", MinValue: &minDPI, MaxValue: 64000},
				{Type: discordgo.ApplicationCommandOptionInteger, Name: "to_dpi", Description: "Your mouse DPI in the new game, if it changes (optional)", MinValue: &minDPI, MaxValue: 64000},
			},
		},
		{
			Name: "restart", Description: "Restart a game server (verified people only)", Contexts: guildOnly,
			Options: []*discordgo.ApplicationCommandOption{server("The server to restart")},
		},
		{
			Name: "whitelist", Description: "Let players into a game server (verified people only)", Contexts: guildOnly,
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "add", Description: "Let a player in", Options: []*discordgo.ApplicationCommandOption{server("The server"), player}},
				{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "remove", Description: "Take a player off the list", Options: []*discordgo.ApplicationCommandOption{server("The server"), player}},
				{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "list", Description: "Who's on the list", Options: []*discordgo.ApplicationCommandOption{server("The server")}},
			},
		},
	}
}

// invocation is one use of a command, as plain values.
type invocation struct {
	command  string         // sens, restart, whitelist
	sub      string         // whitelist's add, remove, list
	opts     map[string]any // string, float64 (numbers) or int64 (integers)
	focused  string         // autocomplete: the option being typed
	userID   string
	userName string
}

func (in invocation) str(name string) string { s, _ := in.opts[name].(string); return s }

func (in invocation) num(name string) float64 { f, _ := in.opts[name].(float64); return f }

func (in invocation) integer(name string) int { i, _ := in.opts[name].(int64); return int(i) }

// reply is a message the bot answers with.
type reply struct {
	content string
	embed   *discordgo.MessageEmbed
	private bool // only the person who asked sees it (ephemeral)
}

// responder answers one interaction. Discord wants an answer within 3 s;
// slow work answers with later() first, then edit() within 15 min.
type responder interface {
	send(reply) error         // the answer
	later(private bool) error // "thinking…"
	edit(reply) error         // replace the answer (after send or later)
}

// handle runs a command.
func (b *Bot) handle(ctx context.Context, in invocation, r responder) error {
	switch in.command {
	case "sens":
		return r.send(b.sens(in))
	case "restart":
		return b.restart(ctx, in, r)
	case "whitelist":
		return b.whitelist(ctx, in, r)
	}
	return r.send(reply{content: "I don't know that command (anymore).", private: true})
}

// sens answers /sens.
func (b *Bot) sens(in invocation) reply {
	res, err := convertSens(in.str("from_game"), in.str("to_game"), in.num("value"), in.integer("from_dpi"), in.integer("to_dpi"))
	if err != nil {
		return reply{content: "Can't convert that: " + err.Error() + ".", private: true}
	}
	e := &discordgo.MessageEmbed{
		Title:       "Sensitivity",
		Color:       int(colorHex[b.Persona().Color]),
		Description: fmt.Sprintf("**%s** → **%s**", res.From.Name, res.To.Name),
		Fields: []*discordgo.MessageEmbedField{
			{Name: res.From.Name, Value: fmt.Sprintf("`%s`", trimFloat(in.num("value"), 4)), Inline: true},
			{Name: res.To.Name, Value: fmt.Sprintf("`%s`", trimFloat(res.Sens, 3)), Inline: true},
		},
	}
	if dpi, to := in.integer("from_dpi"), in.integer("to_dpi"); dpi > 0 && to > 0 && dpi != to {
		e.Fields = append(e.Fields, &discordgo.MessageEmbedField{Name: "DPI", Value: fmt.Sprintf("%d → %d", dpi, to), Inline: true})
	}
	if res.CM360 > 0 {
		e.Footer = &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("%.1f cm for a full turn (360°)", res.CM360)}
	}
	return reply{embed: e}
}

// trimFloat formats f with at most places decimals, no trailing zeros.
func trimFloat(f float64, places int) string {
	s := strings.TrimRight(fmt.Sprintf("%.*f", places, f), "0")
	return strings.TrimSuffix(s, ".")
}

// verified checks in's user and, when they aren't verified, answers for
// them and remembers the request for the settings page.
func (b *Bot) verified(in invocation, r responder) (bool, error) {
	if b.store.IsVerified(in.userID) {
		return true, nil
	}
	if err := b.store.AddRequest(AccessRequest{UserID: in.userID, Name: in.userName, Command: in.command, At: b.now()}); err != nil {
		b.log.Error("discord access request", "err", err)
	}
	b.log.Info("discord not verified", "user", in.userName, "command", in.command)
	return false, r.send(reply{private: true, content: "Only verified people can do that. I've passed your request on: " +
		"whoever runs the Home Tools server can verify you on the Discord page there."})
}

// pickServer finds the server the person chose (autocomplete fills in
// IDs, but people can also type anything), answering for them if it's
// not there or can't do what's asked.
func (b *Bot) pickServer(ctx context.Context, in invocation, r responder, needWhitelist bool) (GameServer, bool, error) {
	list, err := b.gameServers(ctx, 5*time.Second)
	if err != nil {
		b.log.Warn("discord games list", "err", err)
		return GameServer{}, false, r.send(reply{private: true, content: "I can't reach the game servers right now. Try again in a bit."})
	}
	want := in.str("server")
	i := slices.IndexFunc(list, func(s GameServer) bool { return s.ID == want || strings.EqualFold(s.Name, want) })
	if i < 0 {
		return GameServer{}, false, r.send(reply{private: true, content: fmt.Sprintf("There's no server called “%s”. Pick one from the list as you type.", want)})
	}
	srv := list[i]
	if needWhitelist && !srv.Whitelist {
		return GameServer{}, false, r.send(reply{private: true, content: srv.Name + " has no whitelist set up in Home Tools."})
	}
	return srv, true, nil
}

// restart answers /restart: verified only, one at a time, not twice in a
// row, and the whole channel sees it happen.
func (b *Bot) restart(ctx context.Context, in invocation, r responder) error {
	if ok, err := b.verified(in, r); !ok {
		return err
	}
	srv, ok, err := b.pickServer(ctx, in, r, false)
	if !ok {
		return err
	}
	if srv.Busy != "" {
		return r.send(reply{private: true, content: fmt.Sprintf("%s is already busy (%s). Give it a minute.", srv.Name, srv.Busy)})
	}
	if wait, ok := b.cooldowns.begin(srv.ID, b.now()); !ok {
		return r.send(reply{private: true, content: fmt.Sprintf("%s was restarted moments ago. Try again in %s if it's still having trouble.", srv.Name, roundUp(wait))})
	}
	if err := r.send(reply{content: fmt.Sprintf("🔄 Restarting **%s** for %s… (this can take a minute while it saves)", srv.Name, in.userName)}); err != nil {
		return err
	}
	b.log.Info("discord restart", "server", srv.ID, "user", in.userName, "user_id", in.userID)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	err = b.games.Restart(ctx, srv.ID)
	wake(b.boardsWake) // the status boards show the new state
	var ge *GamesError
	switch {
	case err == nil:
		return r.edit(reply{content: fmt.Sprintf("✅ **%s** restarted (asked by %s).", srv.Name, in.userName)})
	case errors.As(err, &ge) && ge.Status == http.StatusConflict:
		return r.edit(reply{content: "⏳ " + ge.Message})
	case errors.As(err, &ge):
		return r.edit(reply{content: fmt.Sprintf("❌ Couldn't restart %s: %s", srv.Name, ge.Message)})
	default:
		b.log.Error("discord restart", "server", srv.ID, "err", err)
		return r.edit(reply{content: fmt.Sprintf("❌ Couldn't restart %s: I lost touch with the game servers.", srv.Name)})
	}
}

// roundUp words a wait for people: "1 min", "45 s".
func roundUp(d time.Duration) string {
	if d > time.Minute {
		return fmt.Sprintf("%d min", int((d+time.Minute-1)/time.Minute))
	}
	return fmt.Sprintf("%d s", int((d+time.Second-1)/time.Second))
}

// steamProfile finds a SteamID64 in a profile link
// (steamcommunity.com/profiles/7656…), so people can paste what they have.
var steamProfile = regexp.MustCompile(`steamcommunity\.com/profiles/(7656119[0-9]{10})`)

var steamID64 = regexp.MustCompile(`^7656119[0-9]{10}$`)

var minecraftName = regexp.MustCompile(`^[A-Za-z0-9_]{3,16}$`)

// normalizePlayer checks a player for srv's game and returns it as the
// games API wants it, or a message saying what's wrong.
func normalizePlayer(game, player string) (string, string) {
	player = strings.TrimSpace(player)
	switch game {
	case "valheim":
		if m := steamProfile.FindStringSubmatch(player); m != nil {
			return m[1], ""
		}
		if steamID64.MatchString(player) {
			return player, ""
		}
		if strings.Contains(player, "steamcommunity.com/id/") {
			return "", "That's a custom Steam link, which doesn't show the number. Ask them for their SteamID64 " +
				"(17 digits starting 7656119; steamid.io finds it from any profile link)."
		}
		return "", "Valheim needs a SteamID64: 17 digits starting 7656119, or a steamcommunity.com/profiles/… link."
	default:
		if minecraftName.MatchString(player) {
			return player, ""
		}
		return "", "That isn't a Minecraft username (3–16 letters, digits or _)."
	}
}

// whitelist answers /whitelist add|remove|list.
func (b *Bot) whitelist(ctx context.Context, in invocation, r responder) error {
	if ok, err := b.verified(in, r); !ok {
		return err
	}
	srv, ok, err := b.pickServer(ctx, in, r, true)
	if !ok {
		return err
	}
	private := in.sub == "list" // a list is for the asker; changes are for everyone to see
	var player string
	if in.sub != "list" {
		var problem string
		if player, problem = normalizePlayer(srv.Game, in.str("player")); problem != "" {
			return r.send(reply{private: true, content: problem})
		}
	}
	// RCON can take a few seconds: answer "thinking…" first.
	if err := r.later(private); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if in.sub == "list" {
		players, err := b.games.Whitelist(ctx, srv.ID)
		if err != nil {
			return r.edit(reply{content: gamesFailure(b, srv, err)})
		}
		if len(players) == 0 {
			return r.edit(reply{content: srv.Name + "'s whitelist is empty."})
		}
		return r.edit(reply{content: fmt.Sprintf("**%s** whitelist (%d):\n%s", srv.Name, len(players), strings.Join(players, ", "))})
	}
	res, err := b.games.EditWhitelist(ctx, srv.ID, player, in.sub == "add")
	if err != nil {
		return r.edit(reply{content: gamesFailure(b, srv, err)})
	}
	b.log.Info("discord whitelist", "server", srv.ID, "player", player, "change", in.sub, "user", in.userName)
	msg := fmt.Sprintf("**%s**: %s", srv.Name, sentence(res.Output))
	if res.RestartNeeded {
		msg += "\nUse `/restart` when nobody's playing."
	}
	return r.edit(reply{content: msg})
}

// gamesFailure words a games API error for Discord.
func gamesFailure(b *Bot, srv GameServer, err error) string {
	var ge *GamesError
	if errors.As(err, &ge) {
		return fmt.Sprintf("❌ %s: %s", srv.Name, ge.Message)
	}
	b.log.Error("discord games call", "server", srv.ID, "err", err)
	return fmt.Sprintf("❌ %s: I lost touch with the game servers. Try again in a bit.", srv.Name)
}

// sentence makes the game's answer read as one: a capital and a full stop.
func sentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "Done."
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	if !strings.ContainsRune(".!?", r[len(r)-1]) {
		r = append(r, '.')
	}
	return string(r)
}

// choice is one autocomplete suggestion.
type choice struct{ name, value string }

// complete suggests servers for the "server" option: the ones this
// command can use, matching what's typed (case-insensitive, anywhere in
// the name), at most 25 (Discord's limit). O(n) over a handful of servers.
func (b *Bot) complete(ctx context.Context, in invocation) []choice {
	if in.focused != "server" {
		return nil
	}
	list, err := b.gameServers(ctx, 10*time.Second)
	if err != nil {
		return nil // an empty list; the command itself will explain
	}
	typed := strings.ToLower(in.str("server"))
	out := []choice{}
	for _, s := range list {
		if in.command == "whitelist" && !s.Whitelist {
			continue
		}
		if typed == "" || strings.Contains(strings.ToLower(s.Name), typed) || strings.Contains(s.ID, typed) {
			out = append(out, choice{s.Name, s.ID})
		}
		if len(out) == 25 {
			break
		}
	}
	return out
}

func compareFold(a, b string) int { return strings.Compare(strings.ToLower(a), strings.ToLower(b)) }

// onInteraction is discordgo's entry point. discordgo runs each event
// handler in its own goroutine, so a slow restart blocks nobody.
func (b *Bot) onInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand, discordgo.InteractionApplicationCommandAutocomplete:
	default:
		return
	}
	in := toInvocation(i)
	ctx := context.Background()
	if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
		choices := []*discordgo.ApplicationCommandOptionChoice{}
		for _, c := range b.complete(ctx, in) {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{Name: c.name, Value: c.value})
		}
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionApplicationCommandAutocompleteResult,
			Data: &discordgo.InteractionResponseData{Choices: choices},
		})
		if err != nil {
			b.log.Debug("discord autocomplete", "err", err)
		}
		return
	}
	if err := b.handle(ctx, in, &interactionResponder{s: s, i: i.Interaction}); err != nil {
		b.log.Warn("discord answer", "command", in.command, "err", err)
	}
}

// toInvocation turns discordgo's interaction into plain values.
func toInvocation(i *discordgo.InteractionCreate) invocation {
	data := i.ApplicationCommandData()
	in := invocation{command: data.Name, opts: map[string]any{}}
	opts := data.Options
	if len(opts) == 1 && opts[0].Type == discordgo.ApplicationCommandOptionSubCommand {
		in.sub, opts = opts[0].Name, opts[0].Options
	}
	for _, o := range opts {
		in.opts[o.Name] = o.Value // string, float64 (all JSON numbers), bool
		if o.Type == discordgo.ApplicationCommandOptionInteger {
			in.opts[o.Name] = o.IntValue()
		}
		if o.Focused {
			in.focused = o.Name
		}
	}
	u := i.User // in DMs
	if i.Member != nil {
		u = i.Member.User
		if i.Member.Nick != "" {
			in.userName = i.Member.Nick
		}
	}
	if u != nil {
		in.userID = u.ID
		if in.userName == "" {
			in.userName = u.GlobalName
		}
		if in.userName == "" {
			in.userName = u.Username
		}
	}
	return in
}

// interactionResponder answers through Discord's interaction API.
// Mentions never ping: replies name people, they don't summon them.
type interactionResponder struct {
	s *discordgo.Session
	i *discordgo.Interaction
}

var noPings = &discordgo.MessageAllowedMentions{}

func (r *interactionResponder) send(rep reply) error {
	return r.s.InteractionRespond(r.i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: rep.content, Embeds: embeds(rep), Flags: flags(rep.private), AllowedMentions: noPings},
	})
}

func (r *interactionResponder) later(private bool) error {
	return r.s.InteractionRespond(r.i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: flags(private)},
	})
}

func (r *interactionResponder) edit(rep reply) error {
	e := embeds(rep)
	_, err := r.s.InteractionResponseEdit(r.i, &discordgo.WebhookEdit{Content: &rep.content, Embeds: &e, AllowedMentions: noPings})
	return err
}

func embeds(rep reply) []*discordgo.MessageEmbed {
	if rep.embed == nil {
		return []*discordgo.MessageEmbed{}
	}
	return []*discordgo.MessageEmbed{rep.embed}
}

func flags(private bool) discordgo.MessageFlags {
	if private {
		return discordgo.MessageFlagsEphemeral
	}
	return 0
}
