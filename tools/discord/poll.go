package discord

import (
	"context"
	"errors"
	"fmt"
	mrand "math/rand/v2" // rand is the blob generator in this package
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
)

// The poll feature (decision #44): every few days, at a set local time,
// the bot posts a native Discord poll from the configured questions.
//
// The schedule is a pure function of the date: poll days are the days
// whose number is divisible by every_days. The question is a random one
// from the packs that are on that hasn't been asked this round (#47).
// state.json keeps the date of the last poll (a restart never posts the
// same slot twice) and which questions were asked (nextQuestion).

// pollSlot is one scheduled poll: its date and when it's due.
type pollSlot struct {
	date string    // YYYY-MM-DD, local
	at   time.Time // the date at post_at, local
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
		return pollSlot{date: date, at: day.Add(time.Duration(hh)*time.Hour + time.Duration(mm)*time.Minute)}
	}
	return pollSlot{} // unreachable for a valid config
}

// nextQuestion picks the next poll's question: a random one (pick(n)
// returns 0..n-1) from the packs that are on, not yet asked this round.
// When all of them have been, a new round starts: they're forgotten and
// all fair game again, except last (the question asked last), so nothing
// is asked twice in a row. It returns the question and the asked record
// to keep once it's posted, with date added and questions that no longer
// exist dropped. Packs that are off keep their record. O(n) for n
// questions. The poll must have at least one question that's on.
func nextQuestion(p PollFeature, asked map[string]string, last, date string, pick func(int) int) (PollQuestion, map[string]string) {
	next := make(map[string]string, len(asked)+1)
	for _, pack := range p.Packs {
		for _, q := range pack.Questions {
			if d, ok := asked[q.key()]; ok {
				next[q.key()] = d
			}
		}
	}
	on := p.questions()
	pool := slices.DeleteFunc(slices.Clone(on), func(q PollQuestion) bool { _, ok := next[q.key()]; return ok })
	if len(pool) == 0 { // round over: start a new one
		for _, q := range on {
			delete(next, q.key())
		}
		pool = slices.DeleteFunc(slices.Clone(on), func(q PollQuestion) bool { return q.key() == last })
		if len(pool) == 0 {
			pool = on // a single question: it has to repeat
		}
	}
	q := pool[pick(len(pool))]
	next[q.key()] = date
	return q, next
}

// NextPoll returns when the next poll goes out, or the zero time when the
// poll is off.
func (b *Bot) NextPoll() time.Time {
	p := b.store.Config().Features.Poll
	if !p.Enabled || len(p.questions()) == 0 {
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
		if p := b.store.Config().Features.Poll; p.Enabled && len(p.questions()) > 0 {
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
	case len(p.questions()) == 0:
		return errors.New("switch on a pack with questions first")
	}
	err := b.postPoll(p, nextPollSlot(b.now(), p, b.store.LastPoll()))
	wake(b.pollWake) // reschedule from the new last slot
	return err
}

// errOffline: the bot isn't connected, so nothing can be posted.
var errOffline = errors.New("the bot isn't connected to Discord")

// postPoll posts the next question and records the slot as done. One at
// a time (pollMu): the loop and "Post one now" must not pick from the
// same asked record.
func (b *Bot) postPoll(p PollFeature, slot pollSlot) error {
	b.pollMu.Lock()
	defer b.pollMu.Unlock()
	b.mu.RLock()
	connected := b.conn.connected
	b.mu.RUnlock()
	if !connected || b.sess == nil {
		return errOffline
	}
	asked, last := b.store.AskedPolls()
	q, asked := nextQuestion(p, asked, last, slot.date, mrand.IntN)
	poll := &discordgo.Poll{Question: discordgo.PollMedia{Text: q.Question}, AllowMultiselect: q.Multi, Duration: p.DurationHours}
	for _, a := range q.Answers {
		poll.Answers = append(poll.Answers, discordgo.PollAnswer{Media: &discordgo.PollMedia{Text: a}})
	}
	if _, err := b.sess.ChannelMessageSendComplex(p.ChannelID, &discordgo.MessageSend{Poll: poll, AllowedMentions: noPings}); err != nil {
		return fmt.Errorf("post poll: %w", err)
	}
	b.log.Info("discord poll posted", "date", slot.date, "question", q.Question)
	return b.store.SetPollPosted(slot.date, q.key(), asked)
}
