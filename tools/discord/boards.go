package discord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Status boards (decision #42): each configured (server, channel) pair is
// one message the bot keeps editing, so the channel never fills up. The
// message ID is remembered in state.json, so a restart of Home Tools
// keeps editing the same message.
//
// Every boardInterval the bot asks the games tool once (one request for
// all servers), renders each board, and edits only the ones whose content
// changed: a hash per board (O(1) compare) instead of an edit per tick,
// which keeps it far below Discord's rate limits. Uptime needs no edits:
// it's a Discord timestamp (<t:…:R>), which every viewer's app counts up.

const boardInterval = 30 * time.Second

// boardKey identifies a board in state.json.
func boardKey(b StatusBoard) string { return b.ServerID + ":" + b.ChannelID }

// runBoards syncs the boards every boardInterval and when woken.
func (b *Bot) runBoards(ctx context.Context) {
	sent := make(map[string]uint64)  // board key → hash of what's showing
	failing := make(map[string]bool) // boards whose last update failed, so it's logged once
	t := time.NewTicker(boardInterval)
	defer t.Stop()
	for {
		b.syncBoards(ctx, sent, failing)
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		case <-b.boardsWake:
		}
	}
}

// syncBoards brings every board up to date and removes the messages of
// boards that were taken out of the config.
func (b *Bot) syncBoards(ctx context.Context, sent map[string]uint64, failing map[string]bool) {
	cfg := b.store.Config()
	byID := map[string]GameServer{}
	list, listErr := b.gameServers(ctx, 0) // always fresh: this is the once-per-tick fetch
	for _, s := range list {
		byID[s.ID] = s
	}
	if listErr != nil {
		b.log.Debug("discord boards: games list", "err", listErr)
	}

	keep := make(map[string]bool, len(cfg.StatusBoards))
	for _, board := range cfg.StatusBoards {
		key := boardKey(board)
		keep[key] = true
		srv, found := byID[board.ServerID]
		e := boardEmbed(board.ServerID, srv, found, listErr)
		h := hashEmbed(e)
		if sent[key] == h {
			continue
		}
		e.Timestamp = b.now().Format(time.RFC3339) // "Updated …" (not hashed: it always changes)
		if err := b.putBoard(board, e); err != nil {
			if !failing[key] {
				b.log.Warn("discord status board", "server", board.ServerID, "channel", board.ChannelID, "err", err)
			}
			failing[key] = true
			continue
		}
		delete(failing, key)
		sent[key] = h
	}

	for key, msgID := range b.store.StatusMessages() {
		if keep[key] {
			continue
		}
		_, channel, _ := strings.Cut(key, ":")
		if err := b.sess.ChannelMessageDelete(channel, msgID); err != nil && !isUnknown(err) {
			b.log.Warn("discord delete status board", "key", key, "err", err)
			continue // try again next tick
		}
		if err := b.store.SetStatusMessage(key, ""); err != nil {
			b.log.Error("discord status board state", "err", err)
		}
		delete(sent, key)
		delete(failing, key)
	}
}

// putBoard edits the board's message, or posts a new one when there's
// none yet or it was deleted in Discord.
func (b *Bot) putBoard(board StatusBoard, e *discordgo.MessageEmbed) error {
	key := boardKey(board)
	if id := b.store.StatusMessage(key); id != "" {
		_, err := b.sess.ChannelMessageEditComplex(&discordgo.MessageEdit{
			ID: id, Channel: board.ChannelID, Embeds: &[]*discordgo.MessageEmbed{e},
		})
		if !isUnknown(err) {
			return err // nil: edited
		}
		// Someone deleted it: post a fresh one below.
	}
	m, err := b.sess.ChannelMessageSendComplex(board.ChannelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{e}, AllowedMentions: noPings,
	})
	if err != nil {
		return err
	}
	return b.store.SetStatusMessage(key, m.ID)
}

// isUnknown reports whether Discord said the message is gone.
func isUnknown(err error) bool {
	var rest *discordgo.RESTError
	return errors.As(err, &rest) && rest.Message != nil && rest.Message.Code == discordgo.ErrCodeUnknownMessage
}

// Board colors: the state at a glance (the text says it too).
const (
	boardGreen = 0x22c55e
	boardAmber = 0xf59e0b
	boardGray  = 0x71717a
	boardRed   = 0xef4444
)

// boardEmbed renders one server's status. found = the games tool listed
// it; listErr = the games tool couldn't be asked at all.
func boardEmbed(id string, s GameServer, found bool, listErr error) *discordgo.MessageEmbed {
	e := &discordgo.MessageEmbed{Title: s.Name, Footer: &discordgo.MessageEmbedFooter{Text: "Live status · updates by itself"}}
	switch {
	case listErr != nil:
		e.Title, e.Color = id, boardRed
		e.Description = "⚠️ **Unknown**: can't reach the game servers right now."
		return e
	case !found:
		e.Title, e.Color = id, boardRed
		e.Description = "⚠️ **Unknown server**: it's no longer set up in Home Tools."
		return e
	case s.Busy != "":
		e.Color = boardAmber
		e.Description = "🔄 **" + map[string]string{"start": "Starting", "stop": "Stopping", "restart": "Restarting"}[s.Busy] + "…**"
		return e
	case s.State == nil:
		e.Color = boardRed
		e.Description = "⚠️ **Unavailable**: " + s.Error
		return e
	case !s.State.Running:
		e.Color = boardGray
		e.Description = "⚫ **Stopped**"
		return e
	}
	e.Color = boardGreen
	e.Description = "🟢 **Running**"
	if !s.State.StartedAt.IsZero() {
		e.Description += fmt.Sprintf(" · started <t:%d:R>", s.State.StartedAt.Unix())
	}
	switch {
	case s.Players != nil:
		e.Fields = append(e.Fields, &discordgo.MessageEmbedField{Name: "Players", Value: playersText(s.Players.Online, s.Players.Max, s.Players.Names)})
	case s.PlayersError != "":
		e.Fields = append(e.Fields, &discordgo.MessageEmbedField{Name: "Players", Value: "Unknown right now"})
	}
	return e
}

// playersText: "2 of 20 online\nSteve, Alex", "Nobody online".
func playersText(online, max int, names []string) string {
	var head string
	switch {
	case online == 0:
		head = "Nobody online"
	case max > 0:
		head = fmt.Sprintf("%d of %d online", online, max)
	default:
		head = fmt.Sprintf("%d online", online)
	}
	if len(names) == 0 {
		return head
	}
	list := []rune(discordEscape(strings.Join(names, ", ")))
	if len(list) > 950 { // an embed field holds 1024 characters
		list = append(list[:950], '…')
	}
	return head + "\n" + string(list)
}

// discordEscape stops names like "__init__" from turning into formatting.
func discordEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, "*", `\*`, "_", `\_`, "~", `\~`, "`", "\\`", "|", `\|`, ">", `\>`).Replace(s)
}

// hashEmbed fingerprints an embed's content (FNV-64a of its JSON).
func hashEmbed(e *discordgo.MessageEmbed) uint64 {
	data, _ := json.Marshal(e) // an embed of strings and ints always encodes
	h := fnv.New64a()
	h.Write(data)
	return h.Sum64()
}
