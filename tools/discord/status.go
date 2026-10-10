package discord

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// The status feature (decision #45): the bot's custom status changes every
// hour, on the hour, local time. Phrases take turns like the names
// (rotate, keyed by the hour), so none repeats before all were shown.
// With live games on, an hour when someone is playing may show that
// instead, but only if the previous hour didn't: game news is a treat,
// not the new normal.

// statusState is what the status loop remembers (in memory: a restart
// just starts fresh).
type statusState struct {
	mu        sync.Mutex
	text      string // showing now ("" = none)
	resend    bool   // Discord forgot it (reconnect): send even if unchanged
	lastGame  int64  // hour number of the last game status (0 = never)
	appliedAt int64  // hour number of the last change
}

// hourNumber counts hours since 1970 on t's local clock (not UTC's), so a
// new phrase starts at each local full hour.
func hourNumber(t time.Time) int64 {
	return dayNumber(t)*24 + int64(t.Hour())
}

// pickStatus returns the status for hour h: a game status when one is
// given and the previous hour wasn't one, else the hour's phrase with
// {name} filled in. game reports whether the game status won.
func pickStatus(s StatusFeature, h int64, lastGame int64, gameText, name string) (text string, game bool) {
	if s.LiveGames && gameText != "" && lastGame != h-1 {
		return gameText, true
	}
	if len(s.Phrases) == 0 {
		return "", false
	}
	p := s.Phrases[rotate(h, len(s.Phrases), "status")]
	return strings.ReplaceAll(p, "{name}", name), false
}

// gameStatus words who's playing, or "" when nobody is: "Watching Steve
// play Minecraft", "Watching 3 blobs play Valheim". The busiest server
// wins; ties go to the first in the list.
func gameStatus(servers []GameServer) string {
	var best *GameServer
	for i := range servers {
		s := &servers[i]
		if s.State == nil || !s.State.Running || s.Players == nil || s.Players.Online == 0 {
			continue
		}
		if best == nil || s.Players.Online > best.Players.Online {
			best = s
		}
	}
	switch {
	case best == nil:
		return ""
	case best.Players.Online == 1 && len(best.Players.Names) == 1:
		return fmt.Sprintf("Watching %s play %s", best.Players.Names[0], best.Name)
	case best.Players.Online == 1:
		return "Watching a blob play " + best.Name
	}
	return fmt.Sprintf("Watching %d blobs play %s", best.Players.Online, best.Name)
}

// runStatus sets the status now, at every full hour, and when woken
// (settings saved, reconnected: Discord forgets a bot's status then).
func (b *Bot) runStatus(ctx context.Context) {
	for {
		b.applyStatus(ctx)
		now := b.now()
		next := now.Truncate(time.Hour).Add(time.Hour) // the next full hour
		timer := time.NewTimer(next.Sub(now) + time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		case <-b.statusWake:
			timer.Stop()
		}
	}
}

// applyStatus picks this hour's status and sends it if it changed.
func (b *Bot) applyStatus(ctx context.Context) {
	cfg := b.store.Config()
	s := cfg.Features.Status
	now := b.now()
	h := hourNumber(now)
	st := &b.status

	var text string
	if s.Enabled {
		st.mu.Lock()
		lastGame, applied := st.lastGame, st.appliedAt
		st.mu.Unlock()
		gameText := ""
		if s.LiveGames && applied != h { // only on the hourly change, not on every save
			if list, err := b.gameServers(ctx, time.Minute); err == nil {
				gameText = gameStatus(list)
			}
		}
		var game bool
		text, game = pickStatus(s, h, lastGame, gameText, PersonaFor(now, cfg.Names).Name)
		st.mu.Lock()
		if game {
			st.lastGame = h
		} else if s.LiveGames && applied == h && st.lastGame == h {
			text = st.text // woken mid-hour during a game hour: keep it
		}
		st.mu.Unlock()
	}

	st.mu.Lock()
	defer st.mu.Unlock()
	st.appliedAt = h
	if text == st.text && !st.resend {
		return
	}
	if err := b.sess.UpdateCustomStatus(text); err != nil { // "" clears it
		b.log.Warn("discord status", "err", err)
		return
	}
	st.text, st.resend = text, false
	b.log.Info("discord status", "text", text)
}

// Status returns the status showing now ("" = none).
func (b *Bot) Status() string {
	b.status.mu.Lock()
	defer b.status.mu.Unlock()
	return b.status.text
}
