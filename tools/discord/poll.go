package discord

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// The poll feature (decision #44): every few days, at a set local time,
// the bot posts a native Discord poll from the configured questions.
//
// Like the persona, the schedule is a pure function of the date: poll
// days are the days whose number is divisible by every_days, and the
// question is that poll's turn in the shuffled rotation (rotate), so
// questions don't repeat until all have been asked. The only state is the
// date of the last poll posted (state.json), so a restart never posts the
// same slot twice.

// pollSlot is one scheduled poll: its date and when it's due.
type pollSlot struct {
	date string    // YYYY-MM-DD, local
	at   time.Time // the date at post_at, local
	turn int64     // which poll this is (day / every_days): picks the question
}

// nextPollSlot returns the first poll slot after the last one posted
// (last = "" for none), looking from today. A slot earlier today that
// wasn't posted (the bot was down) is still returned: it's overdue, so it
// goes out at once. O(every_days): at most a week of days is checked.
func nextPollSlot(now time.Time, p PollFeature, last string) pollSlot {
	var hh, mm int
	fmt.Sscanf(p.PostAt, "%d:%d", &hh, &mm) // validated "HH:MM"
	y, mo, d := now.Date()
	for i := range 2*p.EveryDays + 1 { // two periods always contain an unposted slot
		day := time.Date(y, mo, d+i, 0, 0, 0, 0, now.Location())
		n := dayNumber(day)
		date := day.Format(time.DateOnly)
		if mod(n, int64(p.EveryDays)) != 0 || date <= last { // ISO dates compare as strings
			continue
		}
		return pollSlot{date: date, at: day.Add(time.Duration(hh)*time.Hour + time.Duration(mm)*time.Minute), turn: n / int64(p.EveryDays)}
	}
	return pollSlot{} // unreachable for a valid config
}

// questionFor picks slot's question.
func questionFor(p PollFeature, s pollSlot) PollQuestion {
	return p.Questions[rotate(s.turn, len(p.Questions), "polls")]
}

// NextPoll returns when the next poll goes out, or the zero time when the
// poll is off.
func (b *Bot) NextPoll() time.Time {
	p := b.store.Config().Features.Poll
	if !p.Enabled || len(p.Questions) == 0 {
		return time.Time{}
	}
	return nextPollSlot(b.now(), p, b.store.LastPoll()).at
}

// runPolls posts each poll when it's due. It sleeps until the next slot,
// and wakes early when the settings change.
func (b *Bot) runPolls(ctx context.Context) {
	retry := time.Duration(0)
	for {
		wait := time.Hour // poll off: just re-check now and then (saves wake it too)
		if p := b.store.Config().Features.Poll; p.Enabled && len(p.Questions) > 0 {
			slot := nextPollSlot(b.now(), p, b.store.LastPoll())
			wait = slot.at.Sub(b.now())
			if wait <= 0 && retry == 0 {
				if err := b.postPoll(p, slot); err != nil {
					b.log.Warn("discord poll", "date", slot.date, "err", err)
					retry = 10 * time.Minute // don't hammer Discord; try again later
				}
				continue
			}
			if retry > 0 {
				wait, retry = retry, 0
			}
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		case <-b.pollWake:
			timer.Stop()
			retry = 0
		}
	}
}

// PostPollNow posts the next poll at once (it takes the next slot, so the
// scheduled one is skipped and the question never repeats back to back).
func (b *Bot) PostPollNow() error {
	p := b.store.Config().Features.Poll
	switch {
	case !p.Enabled:
		return errors.New("turn the poll on first")
	case len(p.Questions) == 0:
		return errors.New("add a question first")
	}
	err := b.postPoll(p, nextPollSlot(b.now(), p, b.store.LastPoll()))
	wake(b.pollWake) // reschedule from the new last slot
	return err
}

// errOffline: the bot isn't connected, so nothing can be posted.
var errOffline = errors.New("the bot isn't connected to Discord")

// postPoll posts slot's question and records the slot as done.
func (b *Bot) postPoll(p PollFeature, slot pollSlot) error {
	b.mu.RLock()
	connected := b.conn.connected
	b.mu.RUnlock()
	if !connected || b.sess == nil {
		return errOffline
	}
	q := questionFor(p, slot)
	poll := &discordgo.Poll{Question: discordgo.PollMedia{Text: q.Question}, AllowMultiselect: q.Multi, Duration: p.DurationHours}
	for _, a := range q.Answers {
		poll.Answers = append(poll.Answers, discordgo.PollAnswer{Media: &discordgo.PollMedia{Text: a}})
	}
	if _, err := b.sess.ChannelMessageSendComplex(p.ChannelID, &discordgo.MessageSend{Poll: poll, AllowedMentions: noPings}); err != nil {
		return fmt.Errorf("post poll: %w", err)
	}
	b.log.Info("discord poll posted", "date", slot.date, "question", q.Question)
	return b.store.SetLastPoll(slot.date)
}
